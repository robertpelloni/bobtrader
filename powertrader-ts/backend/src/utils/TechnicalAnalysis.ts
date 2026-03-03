/**
 * Technical Analysis Library (Dependency-Free)
 * Implements core indicators for PowerTrader AI strategies.
 */

export class TechnicalAnalysis {

    /**
     * Simple Moving Average (SMA)
     * @param data Array of numbers (prices)
     * @param period Period length
     * @returns Array of length `data.length` padded with `NaN`
     */
    static calculateSMA(data: number[], period: number): number[] {
        const sma: number[] = new Array(data.length).fill(NaN);
        if (data.length < period) return sma;

        let sum = 0;
        // Initial window
        for (let i = 0; i < period; i++) {
            sum += data[i];
        }
        sma[period - 1] = sum / period;

        // Sliding window
        for (let i = period; i < data.length; i++) {
            sum = sum - data[i - period] + data[i];
            sma[i] = sum / period;
        }
        return sma;
    }

    /**
     * Exponential Moving Average (EMA)
     * @param data Array of numbers
     * @param period Period length
     * @returns Array of length `data.length` padded with `NaN`
     */
    static calculateEMA(data: number[], period: number): number[] {
        const ema: number[] = new Array(data.length).fill(NaN);
        if (data.length < period) return ema;

        const k = 2 / (period + 1);

        // Start with SMA for first EMA value
        let sum = 0;
        for (let i = 0; i < period; i++) sum += data[i];

        let prevEma = sum / period;
        ema[period - 1] = prevEma;

        // Calculate rest
        for (let i = period; i < data.length; i++) {
            // Handle cases where data[i] might be NaN (e.g. nested indicators)
            if (isNaN(data[i])) {
                ema[i] = NaN;
            } else {
                const val = (data[i] * k) + (prevEma * (1 - k));
                ema[i] = val;
                prevEma = val;
            }
        }
        return ema;
    }

    /**
     * Relative Strength Index (RSI)
     * @param data Array of closing prices
     * @param period Period length (default 14)
     */
    static calculateRSI(data: number[], period: number = 14): number[] {
        const rsi: number[] = [];
        if (data.length <= period) return rsi;

        const gains: number[] = [];
        const losses: number[] = [];

        // Calculate price changes
        for (let i = 1; i < data.length; i++) {
            const change = data[i] - data[i - 1];
            gains.push(change > 0 ? change : 0);
            losses.push(change < 0 ? Math.abs(change) : 0);
        }

        // Initial Avg Gain/Loss
        let avgGain = gains.slice(0, period).reduce((a, b) => a + b, 0) / period;
        let avgLoss = losses.slice(0, period).reduce((a, b) => a + b, 0) / period;

        // First RSI
        let rs = avgLoss === 0 ? 100 : avgGain / avgLoss;
        rsi.push(100 - (100 / (1 + rs)));

        // Smoothed RSI
        for (let i = period; i < gains.length; i++) {
            avgGain = ((avgGain * (period - 1)) + gains[i]) / period;
            avgLoss = ((avgLoss * (period - 1)) + losses[i]) / period;

            rs = avgLoss === 0 ? 100 : avgGain / avgLoss;
            rsi.push(100 - (100 / (1 + rs)));
        }

        return rsi;
    }

    /**
     * Bollinger Bands
     * @param data Array of closing prices
     * @param period Period length (default 20)
     * @param stdDev Multiplier (default 2)
     * @returns Arrays of length `data.length` padded with `NaN`
     */
    static calculateBollingerBands(data: number[], period: number = 20, stdDev: number = 2): { upper: number[], middle: number[], lower: number[] } {
        const middle = this.calculateSMA(data, period);
        const upper: number[] = new Array(data.length).fill(NaN);
        const lower: number[] = new Array(data.length).fill(NaN);

        for (let i = period - 1; i < data.length; i++) {
            if (isNaN(middle[i])) continue;

            const avg = middle[i];
            let varianceSum = 0;

            // Calculate Variance
            for (let j = i - period + 1; j <= i; j++) {
                varianceSum += Math.pow(data[j] - avg, 2);
            }

            const sd = Math.sqrt(varianceSum / period);

            upper[i] = avg + (sd * stdDev);
            lower[i] = avg - (sd * stdDev);
        }

        return { upper, middle, lower };
    }

    /**
     * Moving Average Convergence Divergence (MACD)
     * @param data Array of prices
     * @param fastPeriod (12)
     * @param slowPeriod (26)
     * @param signalPeriod (9)
     * @returns Arrays of length `data.length` padded with `NaN`
     */
    static calculateMACD(data: number[], fastPeriod: number = 12, slowPeriod: number = 26, signalPeriod: number = 9): { macdLine: number[], signalLine: number[], histogram: number[] } {
        const fastEMA = this.calculateEMA(data, fastPeriod);
        const slowEMA = this.calculateEMA(data, slowPeriod);

        const macdLine: number[] = new Array(data.length).fill(NaN);

        for (let i = 0; i < data.length; i++) {
            if (!isNaN(fastEMA[i]) && !isNaN(slowEMA[i])) {
                macdLine[i] = fastEMA[i] - slowEMA[i];
            }
        }

        const signalLine = this.calculateEMA(macdLine, signalPeriod);
        const histogram: number[] = new Array(data.length).fill(NaN);

        for (let i = 0; i < data.length; i++) {
            if (!isNaN(macdLine[i]) && !isNaN(signalLine[i])) {
                histogram[i] = macdLine[i] - signalLine[i];
            }
        }

        return { macdLine, signalLine, histogram };
    }
}
