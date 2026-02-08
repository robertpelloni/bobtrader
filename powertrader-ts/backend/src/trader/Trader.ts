import { IExchangeConnector } from "../engine/connector/IExchangeConnector";
import { ConfigManager } from "../config/ConfigManager";
import { AnalyticsManager } from "../analytics/AnalyticsManager";
import { SMAStrategy } from "../engine/strategy/implementations/SMAStrategy";
import { KuCoinConnector } from "../exchanges/KuCoinConnector";

export class Trader {
    private connector: IExchangeConnector;
    private marketDataConnector: IExchangeConnector; // For strategy data
    private config: ConfigManager;
    private analytics: AnalyticsManager;
    private activeTrades: Map<string, any> = new Map();
    private dcaLevels: number[];
    private maxDcaBuys: number;
    private strategy: SMAStrategy; // POC Strategy

    constructor(connector: IExchangeConnector) {
        this.connector = connector;
        this.marketDataConnector = new KuCoinConnector(); // Use KuCoin for reliable data
        this.config = ConfigManager.getInstance();
        this.analytics = new AnalyticsManager();
        this.strategy = new SMAStrategy();

        const cfg = this.config.get("trading");
        this.dcaLevels = cfg.dca_levels || [-2.5, -5.0, -10.0];
        this.maxDcaBuys = cfg.max_dca_buys_per_24h || 2;
    }

    public async start(): Promise<void> {
        console.log("Starting Trader Engine...");
        setInterval(() => this.tick(), 10000); // 10 seconds tick
    }

    private async tick(): Promise<void> {
        const coins = this.config.get("trading.coins") as string[];
        if (!coins) return;

        for (const coin of coins) {
            try {
                await this.processCoin(coin);
            } catch (e) {
                console.error(`Error processing ${coin}:`, e);
            }
        }
    }

    private async processCoin(coin: string): Promise<void> {
        const pair = `${coin}-USD`;
        const currentPrice = await this.connector.fetchTicker(pair);

        let position = this.activeTrades.get(coin);

        // 1. Check for ENTRY (Strategy Mode)
        if (!position) {
            // Check strategy signal
            try {
                // Fetch recent candles for strategy
                const candles = await this.marketDataConnector.fetchOHLCV(`${coin}-USD`, '1h', 50);
                if (candles.length > 0) {
                    const analysis = await this.strategy.populateBuyTrend(candles);
                    const lastCandle = analysis[analysis.length - 1];

                    if (lastCandle.buy_signal) {
                        console.log(`[Trader] Strategy Buy Signal for ${coin}!`);
                        await this.enterTrade(coin, currentPrice);
                        return;
                    }
                }
            } catch (e) {
                // Strategy failure shouldn't crash trader
                console.error(`[Trader] Strategy check failed for ${coin}:`, e);
            }
            return;
        }

        // 2. Calculate PnL
        const pnlPct = ((currentPrice - position.avgPrice) / position.avgPrice) * 100;
        console.log(`[Trader] ${coin} Price: ${currentPrice} PnL: ${pnlPct.toFixed(2)}%`);

        // 3. Check for DCA (Dollar Cost Averaging)
        if (pnlPct < 0 && position.dcaCount < this.maxDcaBuys) {
            const levelIndex = position.dcaCount;
            const nextLevel = this.dcaLevels[levelIndex] !== undefined
                ? this.dcaLevels[levelIndex]
                : this.dcaLevels[this.dcaLevels.length - 1];

            if (pnlPct <= nextLevel) {
                console.log(`[Trader] Triggering DCA for ${coin} at ${pnlPct.toFixed(2)}% (Level: ${nextLevel}%)`);
                await this.executeDCA(coin, currentPrice, position);
            }
        }

        // 4. Check for Trailing Stop Sell
        const trailingCfg = this.config.get("trading");
        const startPct = position.dcaCount === 0 ? trailingCfg.pm_start_pct_no_dca : trailingCfg.pm_start_pct_with_dca;

        if (pnlPct >= startPct) {
            if (!position.trailActive) {
                position.trailActive = true;
                position.trailPeak = currentPrice;
                position.trailLine = currentPrice * (1 - (trailingCfg.trailing_gap_pct / 100));
                console.log(`[Trader] Activated Trailing Stop for ${coin} at ${currentPrice}. Line: ${position.trailLine}`);
            } else {
                if (currentPrice > position.trailPeak) {
                    position.trailPeak = currentPrice;
                    const newLine = currentPrice * (1 - (trailingCfg.trailing_gap_pct / 100));
                    if (newLine > position.trailLine) {
                        position.trailLine = newLine;
                        console.log(`[Trader] Updated Trailing Stop for ${coin}. New Line: ${position.trailLine}`);
                    }
                }
            }
        }

        if (position.trailActive && currentPrice < position.trailLine) {
            console.log(`[Trader] Trailing stop hit for ${coin}. Selling at ${currentPrice}.`);
            await this.exitTrade(coin, currentPrice, position, "trailing_stop");
        }
    }

    private async enterTrade(coin: string, price: number): Promise<void> {
        const allocPct = this.config.get("trading.start_allocation_pct") || 0.01;
        // Assume simplified balance check for now
        const amount = 100 / price; // Fixed $100 entry for demo logic

        console.log(`[Trader] Entering Trade: ${amount} ${coin} @ ${price}`);
        await this.connector.createOrder(`${coin}-USD`, 'market', 'buy', amount);

        this.activeTrades.set(coin, {
            amount: amount,
            avgPrice: price,
            dcaCount: 0,
            trailActive: false,
            trailPeak: 0,
            trailLine: 0
        });

        this.analytics.logTrade({
            symbol: coin,
            side: 'buy',
            amount: amount,
            price: price,
            type: 'entry',
            timestamp: Date.now()
        });
    }

    private async executeDCA(coin: string, price: number, position: any): Promise<void> {
        const dcaMultiplier = this.config.get("trading.dca_multiplier");
        const amount = position.amount * dcaMultiplier;

        console.log(`[Trader] Executing DCA Buy: ${amount} ${coin} @ ${price}`);

        await this.connector.createOrder(`${coin}-USD`, 'market', 'buy', amount);

        const totalCost = (position.avgPrice * position.amount) + (price * amount);
        const totalAmount = position.amount + amount;

        position.amount = totalAmount;
        position.avgPrice = totalCost / totalAmount;
        position.dcaCount++;

        this.analytics.logTrade({
            symbol: coin,
            side: 'buy',
            amount: amount,
            price: price,
            type: 'dca',
            timestamp: Date.now()
        });
    }

    private async exitTrade(coin: string, price: number, position: any, reason: string): Promise<void> {
        console.log(`[Trader] Executing Sell: ${position.amount} ${coin} @ ${price}`);

        await this.connector.createOrder(`${coin}-USD`, 'market', 'sell', position.amount);

        this.analytics.logTrade({
            symbol: coin,
            side: 'sell',
            amount: position.amount,
            price: price,
            type: reason,
            timestamp: Date.now()
        });

        this.activeTrades.delete(coin);
    }
}
