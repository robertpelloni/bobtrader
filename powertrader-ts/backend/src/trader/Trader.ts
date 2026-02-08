import { IExchangeConnector } from "../engine/connector/IExchangeConnector";
import { ConfigManager } from "../config/ConfigManager";
import { AnalyticsManager } from "../analytics/AnalyticsManager";

export class Trader {
    private connector: IExchangeConnector;
    private config: ConfigManager;
    private analytics: AnalyticsManager;
    private activeTrades: Map<string, any> = new Map();
    private dcaLevels: number[];
    private maxDcaBuys: number;

    constructor(connector: IExchangeConnector) {
        this.connector = connector;
        this.config = ConfigManager.getInstance();
        this.analytics = new AnalyticsManager();

        const cfg = this.config.get("trading");
        this.dcaLevels = cfg.dca_levels || [-2.5, -5.0, -10.0];
        this.maxDcaBuys = cfg.max_dca_buys_per_24h || 2;
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
        const pair = `${coin}-USD`;
        const currentPrice = await this.connector.fetchTicker(pair);
        const position = this.activeTrades.get(coin);

        // 1. Check for ENTRY
        if (!position) {
            // Logic: If Thinker signal is LONG, buy
            // For now, we assume Thinker writes a signal or we poll it
            // const signal = await this.thinker.getSignal(coin);
            // if (signal === 'LONG') this.enterTrade(coin, currentPrice);
            return;
        }

        // 2. Check for DCA (Dollar Cost Averaging)
        const pnlPct = ((currentPrice - position.avgPrice) / position.avgPrice) * 100;

        if (pnlPct < 0 && position.dcaCount < this.maxDcaBuys) {
            const nextLevel = this.dcaLevels[position.dcaCount] || this.dcaLevels[this.dcaLevels.length - 1];
            if (pnlPct <= nextLevel) {
                console.log(`[Trader] Triggering DCA for ${coin} at ${pnlPct.toFixed(2)}%`);
                await this.executeDCA(coin, currentPrice, position);
            }
        }

        // 3. Check for Trailing Stop Sell
        // Logic: if pnl > start_pct, activate trail. if price < trail_line, sell.
        const trailingCfg = this.config.get("trading");
        const startPct = position.dcaCount === 0 ? trailingCfg.pm_start_pct_no_dca : trailingCfg.pm_start_pct_with_dca;

        if (pnlPct >= startPct) {
            if (!position.trailActive) {
                position.trailActive = true;
                position.trailPeak = currentPrice;
                position.trailLine = currentPrice * (1 - (trailingCfg.trailing_gap_pct / 100));
            } else {
                if (currentPrice > position.trailPeak) {
                    position.trailPeak = currentPrice;
                    position.trailLine = currentPrice * (1 - (trailingCfg.trailing_gap_pct / 100));
                }
            }
        }

        if (position.trailActive && currentPrice < position.trailLine) {
            console.log(`[Trader] Trailing stop hit for ${coin}. Selling.`);
            await this.exitTrade(coin, currentPrice, position, "trailing_stop");
        }
    }

    private async executeDCA(coin: string, price: number, position: any): Promise<void> {
        const amount = position.amount * this.config.get("trading.dca_multiplier");
        // Execute Buy Order
        // await this.connector.createOrder(...)

        // Update position average
        const cost = (position.avgPrice * position.amount) + (price * amount);
        position.amount += amount;
        position.avgPrice = cost / position.amount;
        position.dcaCount++;

        this.analytics.logTrade({
            symbol: coin,
            side: 'buy',
            amount: amount,
            price: price,
            type: 'dca'
        });
    }

    private async exitTrade(coin: string, price: number, position: any, reason: string): Promise<void> {
        // Execute Sell Order
        // await this.connector.createOrder(...)

        this.analytics.logTrade({
            symbol: coin,
            side: 'sell',
            amount: position.amount,
            price: price,
            type: reason
        });

        this.activeTrades.delete(coin);
    }
}
