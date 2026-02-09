import { ConfigManager } from "../../config/ConfigManager";
import { IExchangeConnector } from "../../engine/connector/IExchangeConnector";

interface RebalanceSignal {
    symbol: string;
    action: 'BUY' | 'SELL';
    amount: number;
    reason: string;
}

export class PortfolioRebalancer {
    private config: ConfigManager;
    private connector: IExchangeConnector;

    constructor(connector: IExchangeConnector) {
        this.connector = connector;
        this.config = ConfigManager.getInstance();
    }

    public async analyzePortfolio(): Promise<RebalanceSignal[]> {
        const coins = this.config.get("trading.coins") as string[];
        const targetAllocation = this.config.get("trading.start_allocation_pct");

        // Fetch current prices and balances
        const prices: Record<string, number> = {};
        const balances: Record<string, number> = {};
        let totalPortfolioValue = 0;

        // Mock fetching balances for now until connector fetchBalance is robust
        // const account = await this.connector.fetchBalance();

        for (const coin of coins) {
            const price = await this.connector.fetchTicker(`${coin}-USD`);
            prices[coin] = price;

            // Mock balance
            const balance = Math.random() * 0.5; // Mock holdings
            balances[coin] = balance;

            totalPortfolioValue += balance * price;
        }

        // Add USD balance
        const usdBalance = 10000; // Mock
        totalPortfolioValue += usdBalance;

        const signals: RebalanceSignal[] = [];

        // Simple Equal Weight Rebalancing logic
        // If we want equal weight for N coins + cash reserve
        const targetValuePerCoin = totalPortfolioValue * targetAllocation; // e.g. 5% per coin?
        // Or if start_allocation_pct is entry size, maybe we want to rebalance active positions to be equal?

        // Let's assume we want to trim positions that are > 2x target allocation

        for (const coin of coins) {
            const currentValue = balances[coin] * prices[coin];
            const deviation = (currentValue - targetValuePerCoin) / targetValuePerCoin;

            if (deviation > 0.5) {
                // Sell excess
                const sellValue = currentValue - targetValuePerCoin;
                const sellAmount = sellValue / prices[coin];
                signals.push({
                    symbol: coin,
                    action: 'SELL',
                    amount: sellAmount,
                    reason: `Overweight by ${(deviation*100).toFixed(1)}%`
                });
            } else if (deviation < -0.5 && usdBalance > targetValuePerCoin) {
                // Buy to target
                const buyValue = targetValuePerCoin - currentValue;
                const buyAmount = buyValue / prices[coin];
                signals.push({
                    symbol: coin,
                    action: 'BUY',
                    amount: buyAmount,
                    reason: `Underweight by ${(deviation*100).toFixed(1)}%`
                });
            }
        }

        return signals;
    }
}
