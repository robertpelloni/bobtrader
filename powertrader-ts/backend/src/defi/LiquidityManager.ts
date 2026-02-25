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
     * Note: This is a simplified version. A robust version would check exact collected amounts from event logs.
     */
    public async autoCompound(tokenId: number): Promise<{ success: boolean, txHash?: string }> {
        try {
            console.log(`[LiquidityManager] Collecting fees for Token #${tokenId}...`);
            // 1. Collect Fees (Harvest to Wallet)
            await this.connector.collectFees(tokenId);

            // 2. Fetch Position to see owed fees (which should be 0 now) or just use wallet balance.
            // For this MVP, we will assume we want to reinvest a fixed ratio or whatever is in the wallet.
            // However, since we don't track the exact collected amount easily without parsing logs,
            // we will fetch the 'tokensOwed' BEFORE collection to know how much we got?
            // Or simpler: The user manually triggers "Compound" which takes available wallet balance
            // (or the fees we just collected if we could parse them).

            // Let's implement a "Check and Add" strategy.
            // We'll fetch the position *before* collect to see pending fees.
            const positions = await this.connector.fetchPositions();
            const pos = positions.find(p => p.tokenId == tokenId.toString());

            if (!pos) throw new Error("Position not found");

            const fees0 = parseFloat(ethers.formatEther(pos.fees0)); // WETH
            const fees1 = parseFloat(ethers.formatUnits(pos.fees1, 6)); // USDC

            if (fees0 === 0 && fees1 === 0) {
                console.log("[LiquidityManager] No fees to compound.");
                return { success: false };
            }

            console.log(`[LiquidityManager] Compounding ${fees0} WETH and ${fees1} USDC...`);

            // 3. Increase Liquidity with the collected amounts
            // Note: In reality, we need to handle the ratio. Uniswap requires adding in proportion to price range.
            // If we just collected fees, they might not be in the perfect ratio for the range.
            // The contract will refund the unused portion.
            const txHash = await this.connector.increaseLiquidity(tokenId, fees0, fees1);

            console.log(`[LiquidityManager] Compound Successful! TX: ${txHash}`);
            return { success: true, txHash };

        } catch (e) {
            console.error(`[LiquidityManager] Auto-compound failed:`, e);
            return { success: false };
        }
    }
}

import { ethers } from 'ethers';
