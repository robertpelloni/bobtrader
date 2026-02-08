/**
 * Technical Analysis Library (Dependency-Free)
 * Implements core indicators for PowerTrader AI strategies.
 */

export class TechnicalAnalysis {

    /**
     * Simple Moving Average (SMA)
     * @param data Array of numbers (prices)
     * @param period Period length
     */
    static calculateSMA(data: number[], period: number): number[] {
        const sma: number[] = [];
        if (data.length < period) return sma;

        for (let i = period - 1; i < data.length; i++) {
            const slice = data.slice(i - period + 1, i + 1);
            const sum = slice.reduce((a, b) => a + b, 0);
            sma.push(sum / period);
        }
        return sma;
    }

    /**
     * Exponential Moving Average (EMA)
     * @param data Array of numbers
     * @param period Period length
     */
    static calculateEMA(data: number[], period: number): number[] {
        const ema: number[] = [];
        if (data.length < period) return ema;

        const k = 2 / (period + 1);

        // Start with SMA for first EMA value
        const initialSlice = data.slice(0, period);
        let prevEma = initialSlice.reduce((a, b) => a + b, 0) / period;
        ema.push(prevEma);

        // Calculate rest
        for (let i = period; i < data.length; i++) {
            const val = (data[i] * k) + (prevEma * (1 - k));
            ema.push(val);
            prevEma = val;
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
     */
    static calculateBollingerBands(data: number[], period: number = 20, stdDev: number = 2): { upper: number[], middle: number[], lower: number[] } {
        const middle = this.calculateSMA(data, period);
        const upper: number[] = [];
        const lower: number[] = [];

        // Align indices: SMA result starts at index 'period-1' of input data
        // We calculate bands for each SMA point
        for (let i = 0; i < middle.length; i++) {
            const dataIndex = i + period - 1;

            // Calculate StdDev for the slice ending at dataIndex
            const slice = data.slice(dataIndex - period + 1, dataIndex + 1);
            const avg = middle[i];

            const squareDiffs = slice.map(value => Math.pow(value - avg, 2));
            const avgSquareDiff = squareDiffs.reduce((a, b) => a + b, 0) / period;
            const sd = Math.sqrt(avgSquareDiff);

            upper.push(avg + (sd * stdDev));
            lower.push(avg - (sd * stdDev));
        }

        return { upper, middle, lower };
    }

    /**
     * Moving Average Convergence Divergence (MACD)
     * @param data Array of prices
     * @param fastPeriod (12)
     * @param slowPeriod (26)
     * @param signalPeriod (9)
     */
    static calculateMACD(data: number[], fastPeriod: number = 12, slowPeriod: number = 26, signalPeriod: number = 9): { macdLine: number[], signalLine: number[], histogram: number[] } {
        // We need EMA(12) and EMA(26)
        // Since EMA arrays have different starting points (offset by period), we must align them.

        const fastEMA = this.calculateEMA(data, fastPeriod);
        const slowEMA = this.calculateEMA(data, slowPeriod);

        const macdLine: number[] = [];

        // Slow EMA is shorter. It starts at index `slowPeriod - 1` of original data.
        // Fast EMA starts at index `fastPeriod - 1`.
        // We need to match them.

        // Lag of slow vs fast
        const lag = slowPeriod - fastPeriod;

        for (let i = 0; i < slowEMA.length; i++) {
            // slowEMA[i] corresponds to data index (slowPeriod-1 + i)
            // fastEMA index?
            // fastEMA starts earlier. so fastEMA[i + lag] corresponds to the same data point.

            const macdVal = fastEMA[i + lag] - slowEMA[i];
            macdLine.push(macdVal);
        }

        // Signal Line = EMA(9) of MACD Line
        const signalLine = this.calculateEMA(macdLine, signalPeriod);
        const histogram: number[] = [];

        // Align MACD line and Signal line
        const sigLag = signalPeriod - 1; // EMA starts after period items

        for (let i = 0; i < signalLine.length; i++) {
            const macdVal = macdLine[i + sigLag];
            const hist = macdVal - signalLine[i];
            histogram.push(hist);
        }

        return { macdLine, signalLine, histogram };
    }
}
