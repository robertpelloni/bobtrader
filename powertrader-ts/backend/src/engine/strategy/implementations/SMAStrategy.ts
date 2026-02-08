import { IStrategy } from "../IStrategy";

export class SMAStrategy implements IStrategy {
    name = "Simple Moving Average";
    interval = "1h";

    // SMA period
    period = 20;

    async populateIndicators(dataframe: any): Promise<any> {
        // Mocking indicator calculation (e.g. using tulind or talib)
        // dataframe['sma'] = calculateSMA(dataframe['close'], this.period)
        console.log(`[SMAStrategy] Calculating SMA(${this.period})...`);
        return dataframe;
    }

    async populateBuyTrend(dataframe: any): Promise<any> {
        // Logic: Buy if close > sma
        console.log("[SMAStrategy] Checking for Golden Cross...");
        return dataframe;
    }

    async populateSellTrend(dataframe: any): Promise<any> {
        // Logic: Sell if close < sma
        console.log("[SMAStrategy] Checking for Death Cross...");
        return dataframe;
    }
}
