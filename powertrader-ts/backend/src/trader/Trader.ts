import { IExchangeConnector } from "../engine/connector/IExchangeConnector";
import { ConfigManager } from "../config/ConfigManager";
import { AnalyticsManager } from "../analytics/AnalyticsManager";
import { StrategyFactory } from "../engine/strategy/StrategyFactory";
import { IStrategy } from "../engine/strategy/IStrategy";
import { KuCoinConnector } from "../exchanges/KuCoinConnector";
import { NotificationManager } from "../notifications/NotificationManager";

export class Trader {
    private connector: IExchangeConnector;
    private marketDataConnector: IExchangeConnector;
    private config: ConfigManager;
    private analytics: AnalyticsManager;
    private notifications: NotificationManager;
    private activeTrades: Map<string, any> = new Map();
    private dcaLevels: number[];
    private maxDcaBuys: number;
    private strategy: IStrategy;

    constructor(connector: IExchangeConnector) {
        this.connector = connector;
        this.marketDataConnector = new KuCoinConnector();
        this.config = ConfigManager.getInstance();
        this.analytics = new AnalyticsManager();
        this.notifications = NotificationManager.getInstance();

        const cfg = this.config.get("trading");
        this.dcaLevels = cfg.dca_levels || [-2.5, -5.0, -10.0];
        this.maxDcaBuys = cfg.max_dca_buys_per_24h || 2;

        // Load Strategy from Config
        const strategyName = cfg.active_strategy || "SMAStrategy";
        this.strategy = StrategyFactory.get(strategyName) || StrategyFactory.get("SMAStrategy")!;
        console.log(`[Trader] Loaded Strategy: ${this.strategy.name}`);
    }

    public async start(): Promise<void> {
        console.log("Starting Trader Engine...");
        setInterval(() => this.tick(), 10000);
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
        try {
            const pair = `${coin}-USD`;
            const currentPrice = await this.connector.fetchTicker(pair);

            let position = this.activeTrades.get(coin);

        // 1. Check for ENTRY
        if (!position) {
            try {
                // Fetch recent candles for strategy
                const candles = await this.marketDataConnector.fetchOHLCV(`${coin}-USD`, '1h', 50);
                if (candles.length > 0) {
                    const withIndicators = await this.strategy.populateIndicators(candles);
                    const withSignals = await this.strategy.populateBuyTrend(withIndicators);
                    const lastCandle = withSignals[withSignals.length - 1];

                    if (lastCandle.buy_signal) {
                        console.log(`[Trader] ${this.strategy.name} Buy Signal for ${coin}!`);
                        await this.enterTrade(coin, currentPrice);
                        return;
                    }
                }
            } catch (e: any) {
                console.error(`[Trader] Strategy check failed for ${coin}:`, e);
                await this.notifications.error(`Strategy check failed for ${coin}: ${e.message}`);
            }
            return;
        }

        // 2. Calculate PnL
        const pnlPct = ((currentPrice - position.avgPrice) / position.avgPrice) * 100;

        // Check Strategy Exit
        try {
             const candles = await this.marketDataConnector.fetchOHLCV(`${coin}-USD`, '1h', 50);
             if (candles.length > 0) {
                 const withIndicators = await this.strategy.populateIndicators(candles);
                 const withSignals = await this.strategy.populateSellTrend(withIndicators);
                 const lastCandle = withSignals[withSignals.length - 1];

                 if (lastCandle.sell_signal) {
                     console.log(`[Trader] ${this.strategy.name} Sell Signal for ${coin}!`);
                     await this.exitTrade(coin, currentPrice, position, "strategy_signal");
                     return;
                 }
             }
        } catch (e) {}

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
        } catch (e: any) {
            console.error(`[Trader] Error processing ${coin}:`, e);
            await this.notifications.error(`Error processing ${coin}: ${e.message}`);
        }
    }

    private async enterTrade(coin: string, price: number): Promise<void> {
        const amount = 100 / price;
        const message = `Entering Trade: ${amount.toFixed(4)} ${coin} @ $${price.toFixed(2)}`;
        console.log(`[Trader] ${message}`);
        await this.notifications.info(message);

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

        const message = `Executing DCA for ${coin}: ${amount.toFixed(4)} @ $${price.toFixed(2)}`;
        console.log(`[Trader] ${message}`);
        await this.notifications.info(message);

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
        const message = `Exiting Trade (${reason}): ${position.amount.toFixed(4)} ${coin} @ $${price.toFixed(2)}`;
        console.log(`[Trader] ${message}`);
        await this.notifications.info(message);

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
