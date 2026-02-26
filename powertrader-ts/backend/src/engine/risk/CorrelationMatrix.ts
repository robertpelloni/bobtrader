import { IExchangeConnector } from "../../engine/connector/IExchangeConnector";
import { ConfigManager } from "../../config/ConfigManager";

export class CorrelationMatrix {
    private connector: IExchangeConnector;
    private config: ConfigManager;

    constructor(connector: IExchangeConnector) {
        this.connector = connector;
        this.config = ConfigManager.getInstance();
    }

    // Pearson Correlation Coefficient
    private calculatePearson(x: number[], y: number[]): number {
        if (x.length !== y.length) return 0;
        const n = x.length;
        if (n === 0) return 0;

        const sumX = x.reduce((a, b) => a + b, 0);
        const sumY = y.reduce((a, b) => a + b, 0);

        const sumXY = x.reduce((sum, xi, i) => sum + xi * y[i], 0);

        const sumX2 = x.reduce((sum, xi) => sum + xi * xi, 0);
        const sumY2 = y.reduce((sum, yi) => sum + yi * yi, 0);

        const numerator = (n * sumXY) - (sumX * sumY);
        const denominator = Math.sqrt(((n * sumX2) - (sumX * sumX)) * ((n * sumY2) - (sumY * sumY)));

        return denominator === 0 ? 0 : numerator / denominator;
    }

    public async generateMatrix(coins: string[], timeframe: string = '1h', limit: number = 50): Promise<any> {
        const history: Record<string, number[]> = {};

        // Fetch history for all coins
        for (const coin of coins) {
            try {
                // Using USD pair
                const candles = await this.connector.fetchOHLCV(`${coin}-USD`, timeframe, limit);
                // Extract closes
                if (candles && candles.length > 0) {
                    history[coin] = candles.map((c: any) => c.close);
                }
            } catch (e) {
                console.warn(`[Risk] Failed to fetch history for ${coin}`);
            }
        }

        const availableCoins = Object.keys(history);
        const matrix: any = {};

        for (const coinA of availableCoins) {
            matrix[coinA] = {};
            for (const coinB of availableCoins) {
                if (coinA === coinB) {
                    matrix[coinA][coinB] = 1.0;
                } else {
                    // Ensure lengths match (truncate to shortest)
                    const len = Math.min(history[coinA].length, history[coinB].length);
                    const sliceA = history[coinA].slice(-len);
                    const sliceB = history[coinB].slice(-len);

                    matrix[coinA][coinB] = this.calculatePearson(sliceA, sliceB);
                }
            }
        }

        return {
            coins: availableCoins,
            matrix
        };
    }
}
