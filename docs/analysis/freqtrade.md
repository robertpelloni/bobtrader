# Architectural Analysis: freqtrade/freqtrade (Strategies)

## Overview
Freqtrade is a highly popular, open-source algorithmic trading software written in Python. It is designed for backtesting, optimization, and live trading on several crypto exchanges. The repository `paulcpk/freqtrade-strategies-that-work` provides a collection of strategies that emphasize **multi-timeframe analysis and trend filtering**.

## Key Architectural Patterns

### 1. Unified Strategy Interface (`IStrategy`)
- **Lifecycle Methods:** Strategies implement `populate_indicators`, `populate_buy_trend`, and `populate_sell_trend`. This separates the calculation of technical indicators from the generation of entry/exit signals.
- **Data-Frame Based:** Heavily relies on Pandas DataFrames for vectorized indicator calculation (using TALib).
- **Assimilation Strategy for Go:** In `ultratrader-go`, we use a similar interface-based approach but should ensure our `OnMarketCandle` event handler provides a way to buffer historical candles for vectorized-style indicator updates.

### 2. Trend Filtering
- **EMA Trend Filter:** Strategies like `DoubleEMACrossoverWithTrend` use a very slow EMA (e.g., EMA 200) to determine the major market trend.
- **Admission Control:** Buys are only allowed if the price (or fast EMA) is above the trend filter, ensuring the bot only trades in the direction of the macro trend.
- **Assimilation Strategy for Go:** Implement a `TrendFilter` component that strategies can use to gate their signal generation logic.

### 3. Backtesting and Data Management
- **Offline Data:** Freqtrade emphasizes downloading historical data (`freqtrade download-data`) and running local simulations.
- **Minimal ROI and Stoploss:** Provides declarative ways to define profit targets and stop-loss levels outside of the core logic.
- **Assimilation Strategy for Go:** Strengthen our `internal/backtest` engine to support downloading data from exchanges and caching it locally in JSONL or SQLite formats.

## Implementation Takeaways for UltraTrader Go
1. **DoubleEMA with Trend:** Port the `ema9/21/200` logic to Go. This requires buffering at least 200 candles to calculate the trend filter accurately.
2. **Historical Data Provider:** Create a provider that can bridge the Binance REST API (`GetKlines`) with our `backtest.Engine`.
3. **Signal Vectorization:** While Go isn't Pandas-centric, we can implement a `Series` type or use slices with technical indicator libraries to achieve similar results.
