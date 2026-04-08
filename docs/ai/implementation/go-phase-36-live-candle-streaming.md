# Go Phase-36 Live Candle Streaming

## Summary
This phase bridges the candle-based strategy pipeline built in Phase 32 (backtester) into the **live App runtime**. Strategies that consume `CandleEvent` data can now operate in a streaming production mode alongside the existing tick-stream and timer modes.

## Context & Motivation
Phases 30–34 built a sophisticated simulation pipeline: indicators → candle strategies → backtesting engine → friction modeling → concurrent optimization. However, the live `App` runtime only supported two scheduler modes:
- `timer` — periodic polling via `strategy.Tick()`
- `stream` — tick-by-tick via `SubscribeTicks` and `runtime.TickEvent`

Multi-timeframe candle strategies (`CandleSMACross`, etc.) were isolated to the backtester. Phase 36 opens a third live mode — `candle-stream` — that subscribes to real-time OHLCV candle feeds and dispatches them through the strategy runtime.

## Delivered

### 1. Candle Subscription Interface
- Added `CandleSubscription` interface to `marketdata/feed.go` with a typed `Chan() <-chan Candle` method.
- Extended `StreamFeed` to include `SubscribeCandles(ctx, symbol, interval string) CandleSubscription`.
- This mirrors the existing `TickSubscription`/`SubscribeTicks` pattern exactly.

### 2. Paper Candle Feed
- Implemented `SubscribeCandles` on `paper.Feed` using a goroutine + `time.Ticker` (5-second interval).
- Added `nextStreamCandle` that cycles through the deterministic price sequence to produce mock OHLCV candles.
- Added `candleSubscription` adapter struct wrapping the channel.

### 3. Candle Stream Service
- Created `strategy/scheduler/candle_stream_service.go` with `CandleStreamService`.
- Subscribes to candle feeds for each configured symbol and dispatches each candle through `candleRunner.RunCandle()`.
- Follows the same goroutine-per-symbol pattern as `StreamService`.

### 4. Scheduler Candle Routing
- Added `RunCandle(ctx, Candle)` method to `scheduler.Scheduler` that calls `runtime.CandleEvent` and executes resulting signals.
- This parallels the existing `RunTick` method.

### 5. Unified Reporting Stream Runner
- Replaced `ReportingTickRunner` with `ReportingStreamRunner` in `reporting/runtime/tick_runner.go`.
- The new runner implements both `tickRunner` and `candleRunner` interfaces.
- Both `RunTick` and `RunCandle` delegate to inner runners and persist reports.
- Eliminates code duplication between tick and candle report paths.

### 6. App Integration
- Added `candle-stream` mode to `core/app/app.go`.
- When `scheduler.mode = "candle-stream"`, the app:
  - Loads both tick and candle strategies (including `CandleSMACross`).
  - Wires `CandleStreamService` with `ReportingStreamRunner` for automatic report persistence.
  - Uses `"1m"` as the default candle interval.

### 7. Testing
- `candle_stream_service_test.go` — mock feed/runner validates end-to-end candle dispatch.
- Updated `tick_runner_test.go` — renamed `stubStreamRunner`, tests both tick and candle paths.
- All 14 internal test packages pass.

## Architecture Diagram

```
┌─────────────────┐     ┌──────────────────────┐     ┌─────────────────┐
│  paper.Feed      │────▶│ CandleStreamService   │────▶│ ReportingStream │
│ SubscribeCandles │     │ (goroutine/symbol)    │     │ Runner          │
└─────────────────┘     └──────────────────────┘     └────────┬────────┘
                                                              │
                                                    ┌─────────▼────────┐
                                                    │ Scheduler        │
                                                    │ RunCandle()      │
                                                    └─────────┬────────┘
                                                              │
                                                    ┌─────────▼────────┐
                                                    │ Runtime          │
                                                    │ CandleEvent()    │
                                                    └─────────┬────────┘
                                                              │
                                                    ┌─────────▼────────┐
                                                    │ ExecutionService │
                                                    │ + Risk Pipeline  │
                                                    └──────────────────┘
```

## Configuration

Add to your JSON config:
```json
{
  "scheduler": {
    "enabled": true,
    "mode": "candle-stream",
    "interval_ms": 5000
  }
}
```

## Next Steps
1. **Additional Indicators** — Expand the library with MACD, Bollinger Bands, ATR to enrich candle strategy signals.
2. **Real Exchange Adapters** — Implement Binance websocket K-Line feeds so candle streaming uses real market data.
3. **Walk-Forward Optimization** — Combine the concurrent optimizer (Phase 35) with live candle streaming for continuous parameter re-tuning.
