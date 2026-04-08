# Go Phase-32 Candle-Driven Backtesting & Strategy Enhancements

## Summary
This phase extends the Go ultra-project's strategy execution and backtesting subsystems to natively support aggregated interval data (Candles / K-lines).

## Context & Motivation
Submodules like `bbgo` and the legacy Python `pt_thinker.py` ML pipelines heavily rely on interval-based candles (e.g., 5m, 1h, 1d) rather than pure raw tick streams. Real-world quantitative strategies typically calculate technical indicators over fixed time boundaries. This phase introduces `CandleStrategy` as a first-class citizen alongside `TickStrategy`.

## Delivered

### Strategy Architecture Updates
- Added `CandleStrategy` to `internal/strategy/runtime.go`, requiring implementations to define `OnMarketCandle(ctx, candle)`.
- Added `CandleEvent` to `strategy.Runtime` to properly route incoming candles to compatible strategies.

### Backtester Enhancements
- Refactored `backtest.Engine` creation to accept a generic `strategy.Strategy` interface.
- Introduced `CandleHistoryProvider` interface to feed historical candle data.
- Added `RunCandles(ctx, history)` method to `Engine`, mirroring `RunTicks`, to iterate over historical interval data and simulate executions at the candle's `Close` price.
- Added `MemoryCandleHistory` for injecting test fixtures.

### Demo Strategy
- Implemented `CandleSMACross` (`internal/strategy/demo/candle_sma_crossover.go`), showcasing how two Moving Averages (fast and slow) interact purely via `OnMarketCandle` to trigger golden and death crosses.

## Architectural Significance
By decoupling strategy execution interfaces into `TickStrategy` and `CandleStrategy`, the system can accommodate high-frequency tick-based arbitrage just as easily as it handles slow-moving trend-following systems. The backtester now fully supports evaluating these slower timeframe models deterministically.

## Validation
- Verified passing unit tests for the `Runtime` correctly filtering and routing `CandleEvent`s versus `TickEvent`s.
- Verified the simulated crossover logic in `CandleSMACross` fires exactly as the arithmetic dictates across sequential candle intervals.
- `go test ./internal/strategy/... ./internal/backtest/...` runs cleanly.

## Recommended next steps
1.  **Advanced Market Emulation:** Now that strategies can be backtested on candles, add simulated slippage, commission fees, and maker/taker spread mechanics to `Engine.processSignals` to make PnL reporting realistic.
2.  **Live Candle Streaming:** Build out the `marketdata` feed and `scheduler` to subscribe to live exchange websocket K-Line feeds and dispatch them via `runtime.CandleEvent`.
3.  **Optimization Engine:** Scaffold parameter tuning pipelines for the `CandleSMACross` strategy.
