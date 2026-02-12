import { UniswapConnector } from '../exchanges/UniswapConnector';
import { TechnicalAnalysis } from '../utils/TechnicalAnalysis';

export class LiquidityManager {
    private connector: UniswapConnector;

    constructor(connector: UniswapConnector) {
        this.connector = connector;
    }

    /**
     * Calculate optimal Uniswap V3 range (tickLower, tickUpper) based on Bollinger Bands.
     * @param priceHistory Array of closing prices.
     * @param stdDev Standard deviations for the band (e.g. 2).
     */
    public calculateOptimalRange(priceHistory: number[], stdDev: number = 2): { lower: number, upper: number } {
        if (priceHistory.length < 20) {
            throw new Error("Insufficient history for range calculation");
        }

        const bands = TechnicalAnalysis.calculateBollingerBands(priceHistory, 20, stdDev);
        const lastBand = {
            upper: bands.upper[bands.upper.length - 1],
            lower: bands.lower[bands.lower.length - 1]
        };

        return {
            lower: this.priceToTick(lastBand.lower),
            upper: this.priceToTick(lastBand.upper)
        };
    }

    /**
     * Convert price to Uniswap Tick index (simplified).
     * tick = floor(log_1.0001(price))
     */
    private priceToTick(price: number): number {
        return Math.floor(Math.log(price) / Math.log(1.0001));
    }

    /**
     * Auto-compound loop: Collect fees and add them back to liquidity.
     */
    public async autoCompound(tokenId: number): Promise<void> {
        try {
            console.log(`[LiquidityManager] Collecting fees for Token #${tokenId}...`);
            await this.connector.collectFees(tokenId);

            // In a real implementation, we would check balances of token0/token1
            // and call increaseLiquidity with the collected amounts.
            // For MVP, we stop at collection.
            console.log(`[LiquidityManager] Fees collected. Compounding logic pending implementation.`);
        } catch (e) {
            console.error(`[LiquidityManager] Auto-compound failed:`, e);
        }
    }
}
