import { IStrategy } from "../../engine/strategy/IStrategy";
import { TechnicalAnalysis } from "../../utils/TechnicalAnalysis";

export class CointradeAdapter implements IStrategy {
    name = "Cointrade (External)";
    interval = "1h";

    constructor() {
        console.log("[Cointrade] Adapter initialized. Running TS implementation of Cointrade logic.");
    }

    async populateIndicators(dataframe: any): Promise<any> {
        // Cointrade Logic Port: RSI, Bollinger Bands, and MACD
        const closes = dataframe.map((c: any) => c.close);

        const rsi = TechnicalAnalysis.calculateRSI(closes, 14);
        const { upper, lower } = TechnicalAnalysis.calculateBollingerBands(closes, 20, 2);
        const { macdLine, signalLine } = TechnicalAnalysis.calculateMACD(closes, 12, 26, 9);

        // Align arrays (indicators are shorter than data due to warm-up periods)
        const len = dataframe.length;

        return dataframe.map((c: any, i: number) => {
            // Calculate offsets
            const rsiIdx = i - (len - rsi.length);
            const bbIdx = i - (len - upper.length);
            const macdIdx = i - (len - macdLine.length);

            return {
                ...c,
                rsi: rsiIdx >= 0 ? rsi[rsiIdx] : null,
                bb_upper: bbIdx >= 0 ? upper[bbIdx] : null,
                bb_lower: bbIdx >= 0 ? lower[bbIdx] : null,
                macd: macdIdx >= 0 ? macdLine[macdIdx] : null,
                macd_signal: macdIdx >= 0 ? signalLine[macdIdx] : null
            };
        });
    }

    async populateBuyTrend(dataframe: any): Promise<any> {
        // Cointrade Signal: Buy when RSI < 30 AND Price < Lower BB
        return dataframe.map((c: any) => {
            const buy = (c.rsi !== null && c.bb_lower !== null) &&
                        (c.rsi < 30 && c.close < c.bb_lower) ? 1 : 0;
            return { ...c, buy_signal: buy };
        });
    }

    async populateSellTrend(dataframe: any): Promise<any> {
        // Cointrade Signal: Sell when RSI > 70 AND Price > Upper BB
        return dataframe.map((c: any) => {
            const sell = (c.rsi !== null && c.bb_upper !== null) &&
                         (c.rsi > 70 && c.close > c.bb_upper) ? 1 : 0;
            return { ...c, sell_signal: sell };
        });
    }
}
