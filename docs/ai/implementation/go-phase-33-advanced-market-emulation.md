# Go Phase-33 Advanced Market Emulation

## Summary
This phase introduces realistic execution mechanics to the backtesting engine, specifically adding configuration for Maker/Taker commission fees and price slippage.

## Context & Motivation
A strategy that appears highly profitable in a "zero-friction" simulation often fails in reality because the spread, slippage, and exchange fees eat the profit margin. By integrating `EmulatorOptions` directly into the `Engine`, strategies are forced to overcome realistic trading costs to show positive PnL. This mirrors the `backtest.Exchange` implementations found in the `bbgo` submodule and the `pt_backtester.py` logic.

## Delivered

### Emulation Configuration
- Created `EmulatorOptions` struct allowing discrete settings for `MakerFeeRate`, `TakerFeeRate`, and `SlippageRate`.
- Created `DefaultEmulatorOptions` establishing a baseline standard of 0.1% fees (typical of Binance base spot markets) and 0% slippage.

### Execution Modifiers
- Modified `Engine.processSignals` to computationally alter the executed `Price` of simulated `exchange.Order` entries.
- **For Buys:** Increases the effective execution price by compounding the slippage and taker fee percentages. `priceVal * (1.0 + SlippageRate) * (1.0 + TakerFeeRate)`
- **For Sells:** Decreases the effective execution price. `priceVal * (1.0 - SlippageRate) * (1.0 - TakerFeeRate)`
- This logic works seamlessly with the underlying `portfolio.Tracker`, accurately inflating the `CostBasis` for buys and suppressing the `RealizedPnL` for sells.

### Testing Refinements
- Added `NewEngineWithOptions` for injecting explicit models, allowing existing unit tests to continue passing under zero-friction configurations.
- Implemented `TestEngineRunFriction` asserting the exact compounded arithmetic of 5% slippage and 1% fees against a deterministic price path.

## Architectural Significance
By pushing the emulation penalty into the `Price` field of the generated `exchange.Order`, we avoid having to rewrite the `portfolio.Tracker` or introduce complex "fiat balance drain" logic into the execution kernel. The tracker already correctly accounts for the mathematical impact of an artificially high buy price or an artificially low sell price.

## Validation
- `TestEngineRunFriction` proved that buying at 100 and selling at 200 under 5% slippage and 1% fees yields a realized PnL of exactly `$82.05` instead of the nominal `$100.00`.
- All `backtest` and `strategy` tests passed successfully via `go test ./...`.

## Recommended next steps
1.  **Optimization Subsystem:** Now that the backtester is fast, supports multi-timeframe candles, and produces *realistic* PnL numbers with fees, we are ready to implement parameter tuning pipelines (e.g., Grid search) to find the best indicator lengths.
2.  **Live Candle Streaming:** Build out the `marketdata` feed and `scheduler` to subscribe to live exchange websocket K-Line feeds and dispatch them via `runtime.CandleEvent`.
3.  **Real Exchange Adapters:** Start integrating CCXT or Binance-specific REST/Websocket connections beyond the `paper` adapter.
