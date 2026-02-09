import { IStrategy } from "../IStrategy";
import { TechnicalAnalysis } from "../../../utils/TechnicalAnalysis";

export class RSIStrategy implements IStrategy {
    name = "RSI Reversal";
    interval = "1h";

    // Config
    period = 14;
    buyThreshold = 30;
    sellThreshold = 70;

    setParameters(params: any): void {
        if (params.period) this.period = params.period;
        if (params.buyThreshold) this.buyThreshold = params.buyThreshold;
        if (params.sellThreshold) this.sellThreshold = params.sellThreshold;
    }

    async populateIndicators(dataframe: any): Promise<any> {
        const closes = dataframe.map((c: any) => c.close);
        const rsi = TechnicalAnalysis.calculateRSI(closes, this.period);

        // Pad beginning with nulls/zeros to match length
        const offset = dataframe.length - rsi.length;

        return dataframe.map((c: any, i: number) => ({
            ...c,
            rsi: i >= offset ? rsi[i - offset] : null
        }));
    }

    async populateBuyTrend(dataframe: any): Promise<any> {
        return dataframe.map((c: any) => ({
            ...c,
            buy_signal: c.rsi !== null && c.rsi < this.buyThreshold ? 1 : 0
        }));
    }

    async populateSellTrend(dataframe: any): Promise<any> {
        return dataframe.map((c: any) => ({
            ...c,
            sell_signal: c.rsi !== null && c.rsi > this.sellThreshold ? 1 : 0
        }));
    }
}
