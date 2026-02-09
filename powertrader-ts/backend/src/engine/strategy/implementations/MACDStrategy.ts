import { IStrategy } from "../IStrategy";
import { TechnicalAnalysis } from "../../../utils/TechnicalAnalysis";

export class MACDStrategy implements IStrategy {
    name = "MACD Crossover";
    interval = "1h";

    // Config
    fastPeriod = 12;
    slowPeriod = 26;
    signalPeriod = 9;

    async populateIndicators(dataframe: any): Promise<any> {
        const closes = dataframe.map((c: any) => c.close);
        const { macdLine, signalLine, histogram } = TechnicalAnalysis.calculateMACD(closes, this.fastPeriod, this.slowPeriod, this.signalPeriod);

        // MACD results are shorter than input due to EMA lag
        // Offset = slowPeriod - 1 + signalPeriod - 1
        // Actually TechnicalAnalysis.calculateMACD handles internal alignment,
        // but the final array length is `data.length - slowPeriod - signalPeriod + 2` approximately.
        // We align from the end.

        const len = macdLine.length; // macdLine, signalLine, histogram are same length
        const offset = dataframe.length - len;

        return dataframe.map((c: any, i: number) => {
            if (i < offset) return { ...c, macd: null, signal: null, hist: null };

            const idx = i - offset;
            return {
                ...c,
                macd: macdLine[idx],
                signal: signalLine[idx],
                hist: histogram[idx]
            };
        });
    }

    async populateBuyTrend(dataframe: any): Promise<any> {
        // Buy when MACD crosses above Signal (Histogram goes negative -> positive)
        return dataframe.map((c: any, i: number) => {
            if (i === 0) return { ...c, buy_signal: 0 };
            const prev = dataframe[i-1];

            // Bullish Crossover
            const buy = (prev.hist !== null && c.hist !== null) && (prev.hist < 0 && c.hist > 0) ? 1 : 0;
            return { ...c, buy_signal: buy };
        });
    }

    async populateSellTrend(dataframe: any): Promise<any> {
        // Sell when MACD crosses below Signal
        return dataframe.map((c: any, i: number) => {
            if (i === 0) return { ...c, sell_signal: 0 };
            const prev = dataframe[i-1];

            // Bearish Crossover
            const sell = (prev.hist !== null && c.hist !== null) && (prev.hist > 0 && c.hist < 0) ? 1 : 0;
            return { ...c, sell_signal: sell };
        });
    }
}
