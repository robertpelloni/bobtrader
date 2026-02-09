import { IBacktestEngine, IBacktestConfig, IBacktestResult } from "./IBacktestEngine";
import { IStrategy } from "../strategy/IStrategy";
import { HistoricalData } from "./HistoricalData";

export class BacktestEngine implements IBacktestEngine {
    async run(config: IBacktestConfig, strategy: IStrategy): Promise<IBacktestResult> {
        // console.log(`[Backtest] Starting backtest for ${config.pair} (${strategy.name})...`);

        // 1. Fetch Data
        let candles = config.data;
        if (!candles) {
            console.log(`[Backtest] Fetching data for ${config.pair}...`);
            candles = await HistoricalData.fetch(config.pair, config.timeframe, config.startDate, config.endDate);
        }

        if (!candles || candles.length === 0) {
             throw new Error("No historical data found for range.");
        }

        // console.log(`[Backtest] Loaded ${candles.length} candles.`);

        // 2. Prepare Strategy Indicators
        const withIndicators = await strategy.populateIndicators(candles);
        const withBuy = await strategy.populateBuyTrend(withIndicators);
        const fullData = await strategy.populateSellTrend(withBuy);

        // 3. Simulate Trading
        let balance = config.initialBalance;
        let position = 0; // Amount of coin
        let equityCurve: { time: number, value: number }[] = [];
        let trades: any[] = [];
        let entryPrice = 0;
        let maxEquity = balance;
        let maxDrawdown = 0;

        // Start after warmup period (e.g. 50 candles for SMA/RSI)
        const warmup = 50;

        for (let i = warmup; i < fullData.length; i++) {
            const candle = fullData[i];
            const price = candle.close;
            const date = candle.timestamp;

            // Calculate Equity (Mark to Market)
            let equity = balance;
            if (position > 0) {
                equity += (position * price);
            }

            equityCurve.push({ time: date, value: equity });

            // Update Drawdown
            if (equity > maxEquity) maxEquity = equity;
            const dd = (maxEquity - equity) / maxEquity * 100;
            if (dd > maxDrawdown) maxDrawdown = dd;

            // Strategy Signals
            // We simplify: execute on CLOSE of signal candle

            if (position === 0) {
                if (candle.buy_signal) {
                    // Buy 99% of balance
                    const amountToSpend = balance * 0.99;
                    const amount = amountToSpend / price;
                    const fee = amountToSpend * 0.001; // 0.1% fee

                    balance -= (amountToSpend + fee);
                    position = amount;
                    entryPrice = price;

                    trades.push({
                        type: 'buy',
                        price,
                        time: date,
                        amount,
                        fee
                    });
                }
            } else {
                if (candle.sell_signal) {
                    // Sell all
                    const proceeds = position * price;
                    const fee = proceeds * 0.001;

                    balance += (proceeds - fee);

                    // Calculate PnL for this trade
                    const entryCost = position * entryPrice;
                    const pnl = proceeds - entryCost - fee; // Simple PnL

                    trades.push({
                        type: 'sell',
                        price,
                        time: date,
                        amount: position,
                        fee,
                        pnl
                    });

                    position = 0;
                    entryPrice = 0;
                }
            }
        }

        // Finalize (Close open position)
        if (position > 0) {
            const lastCandle = fullData[fullData.length-1];
            const lastPrice = lastCandle.close;
            const proceeds = position * lastPrice;
            const fee = proceeds * 0.001;
            balance += (proceeds - fee);

            const entryCost = position * entryPrice;
            const pnl = proceeds - entryCost - fee;

            trades.push({ type: 'sell', price: lastPrice, time: lastCandle.timestamp, amount: position, fee, pnl, note: "force_close" });
            position = 0;
        }

        const sellTrades = trades.filter(t => t.type === 'sell');
        const totalTrades = sellTrades.length;
        const wins = sellTrades.filter(t => t.pnl > 0).length;
        const winRate = totalTrades > 0 ? (wins / totalTrades) * 100 : 0;

        // Calculate Profit Factor (Gross Profit / Gross Loss)
        const grossProfit = sellTrades.filter(t => t.pnl > 0).reduce((acc, t) => acc + t.pnl, 0);
        const grossLoss = Math.abs(sellTrades.filter(t => t.pnl < 0).reduce((acc, t) => acc + t.pnl, 0));
        const profitFactor = grossLoss > 0 ? grossProfit / grossLoss : (grossProfit > 0 ? 999 : 0);

        // Sharpe Ratio (Simplified placeholder)
        const sharpeRatio = 0.0;

        return {
            totalTrades,
            winRate,
            profitFactor,
            maxDrawdown,
            sharpeRatio,
            trades,
            equityCurve,
            chartData: fullData
        };
    }
}
