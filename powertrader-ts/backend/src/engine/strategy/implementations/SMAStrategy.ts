import { IStrategy } from "../IStrategy";
import { TechnicalAnalysis } from "../../../utils/TechnicalAnalysis";

export class SMAStrategy implements IStrategy {
    name = "Simple Moving Average";
    interval = "1h";

    // SMA period
    period = 20;

    setParameters(params: any): void {
        if (params.period) this.period = params.period;
    }

    async populateIndicators(dataframe: any): Promise<any> {
        const closes = dataframe.map((c: any) => c.close);
        // Ensure data is sorted by time ascending (it should be)
        const sma = TechnicalAnalysis.calculateSMA(closes, this.period);

        // Pad with nulls
        const offset = dataframe.length - sma.length;

        return dataframe.map((c: any, i: number) => ({
            ...c,
            sma: i >= offset ? sma[i - offset] : null
        }));
    }

    async populateBuyTrend(dataframe: any): Promise<any> {
        // Buy if price > SMA
        return dataframe.map((c: any) => ({
            ...c,
            buy_signal: (c.sma !== null && c.close > c.sma) ? 1 : 0
        }));
    }

    async populateSellTrend(dataframe: any): Promise<any> {
        // Sell if Price < SMA
        return dataframe.map((c: any) => ({
            ...c,
            sell_signal: (c.sma !== null && c.close < c.sma) ? 1 : 0
        }));
    }
}
