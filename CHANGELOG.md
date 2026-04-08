# Changelog

All notable changes to PowerTrader AI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.0.41] - 2026-04-08

### Added
- **Go Ultra-Project Phase-39 Binance REST Adapter**
  - Created `internal/exchange/binance/adapter.go` implementing the `exchange.Adapter` interface.
  - **Public endpoints**: `GetTickerPrice(symbol)` for live prices, `GetKlines(symbol, interval, limit)` for historical candle data, `ListMarkets()` filtering by TRADING status.
  - **Authenticated endpoints**: `Balances()` for spot account balances, `PlaceOrder()` for market/limit orders with HMAC-SHA256 request signing.
  - **Security**: HMAC-SHA256 signature generation, X-MBX-APIKEY header, timestamp + recvWindow enforcement.
  - **Testnet support**: Configurable base URL with automatic testnet routing via `Config.Testnet`.
  - **Error handling**: Structured Binance API error parsing with code/message extraction.
  - **HTTP mocking**: Full test suite using `httptest.Server` validating ticker price, kline parsing, market filtering, signature generation, and API key enforcement.
  - Registered `binance` factory in `core/app` with testnet-safe defaults.
  - Added `APIKey`, `SecretKey`, `Testnet` fields to `AccountConfig`.
  - Documented in `docs/ai/implementation/go-phase-39-binance-rest-adapter.md`.

### Verified
- `go test ./internal/exchange/binance/...` — all tests pass with mocked HTTP server.
- `go test ./internal/exchange/...` — all exchange packages pass.
- `go build ./cmd/ultratrader` — compiles cleanly.
- Binary runs with both `paper` and `binance` adapters registered.

## [2.0.40] - 2026-04-08

### Added
- **Go Ultra-Project Phase-38 Indicator-Based Strategies**
  - **MACDCrossover**: Candle strategy generating buy/sell signals on MACD histogram zero-line crossover. Configurable fast/slow/signal EMA periods.
  - **BollingerReversion**: Candle mean-reversion strategy buying at the lower Bollinger Band (oversold) and selling at the upper band (overbought). Configurable period and multiplier.
  - **ATRSizing**: Candle strategy combining SMA crossover with ATR-based dynamic position sizing. Scales order quantity inversely with volatility — smaller positions in volatile markets, larger positions in calm markets.
  - Comprehensive test coverage for all three new strategies with realistic price sequences.
  - Documented in `docs/ai/implementation/go-phase-38-indicator-strategies.md`.

### Verified
- `go test ./internal/strategy/demo/...` — all tests pass.
- `go build ./cmd/ultratrader` — compiles cleanly.

## [2.0.39] - 2026-04-08

### Added
- **Go Ultra-Project Phase-37 Expanded Indicator Library**
  - **MACD** (Moving Average Convergence Divergence): Produces MACD line, Signal line, and Histogram. Configurable fast/slow/signal periods. Built on top of EMA.
  - **Bollinger Bands**: Calculates Upper, Middle (SMA), and Lower bands with configurable multiplier (default 2.0 std deviations). Includes Bandwidth metric for volatility measurement.
  - **ATR** (Average True Range): Volatility indicator computing the greatest of high-low range, |high-prev_close|, |low-prev_close|. Wilder smoothing after warmup period.
  - Added comprehensive test coverage for all three new indicators with mathematical validation.
  - Documented architecture in `docs/ai/implementation/go-phase-37-expanded-indicator-library.md`.

### Changed
- Added `math` import to `internal/indicator/indicators.go` for standard deviation and absolute value calculations.
- Extended `indicators_test.go` with 7 new test cases covering MACD crossover, Bollinger constant/varying/insufficient data, and ATR warmup/smoothing.

### Verified
- `go test ./internal/indicator/...` — all 10 tests pass (3 original + 7 new).
- `go build ./cmd/ultratrader` — compiles cleanly.

## [2.0.38] - 2026-04-08

### Added
- **Go Ultra-Project Phase-36 Live Candle Streaming**
  - Extended `marketdata.StreamFeed` interface with `SubscribeCandles(ctx, symbol, interval)` returning a `CandleSubscription` channel.
  - Added `CandleSubscription` interface to `marketdata/feed.go` for typed candle stream consumption.
  - Implemented `SubscribeCandles` and `nextStreamCandle` in the paper feed (`marketdata/paper/feed.go`) to emit simulated OHLCV candles on a 5-second interval.
  - Created `CandleStreamService` in `strategy/scheduler/candle_stream_service.go` — subscribes to a candle feed for multiple symbols and dispatches each candle through `RunCandle()`.
  - Added `RunCandle(ctx, Candle)` method to `scheduler.Scheduler` to route candle events through `runtime.CandleEvent` and execute resulting signals.
  - Unified reporting runner: replaced `ReportingTickRunner` with `ReportingStreamRunner` that handles both `RunTick` and `RunCandle` with automatic report persistence.
  - Added `candle-stream` scheduler mode to `core/app` — when `scheduler.mode = "candle-stream"`, the app wires up `CandleSMACross` strategy and the `CandleStreamService`.
  - Wrote `candle_stream_service_test.go` with mock feed/runner validating end-to-end candle dispatch.
  - Wrote updated `tick_runner_test.go` (renamed `stubStreamRunner`) testing both tick and candle paths through `ReportingStreamRunner`.
  - Documented architecture in `docs/ai/implementation/go-phase-36-live-candle-streaming.md`.

### Changed
- `marketdata/feed.go`: Added `CandleSubscription` interface and extended `StreamFeed` with `SubscribeCandles`.
- `marketdata/paper/feed.go`: Added `candleSubscription` type, `SubscribeCandles`, `nextStreamCandle` methods.
- `reporting/runtime/tick_runner.go`: Renamed from `ReportingTickRunner` to `ReportingStreamRunner` with dual `RunTick`/`RunCandle` support.
- `reporting/runtime/tick_runner_test.go`: Renamed `stubTickRunner` to `stubStreamRunner`, added candle test coverage.
- `strategy/scheduler/scheduler.go`: Added `RunCandle` method.
- `core/app/app.go`: Added `candle-stream` mode branching with `CandleSMACross` strategy and `CandleStreamService`.

### Verified
- `go build ./cmd/ultratrader` compiles cleanly.
- All 14 internal test packages pass individually.
- Binary runs and produces valid structured JSON output with full guard pipeline active.

## [2.0.37] - 2026-04-06

### Added
- **Go Ultra-Project Phase-35 Concurrent Optimization**
  - Upgraded the `GridSearchCandles` optimizer to execute parameter permutations concurrently using an idiomatic Go worker pool pattern.
  - Added `OptimizationConfig` to allow tuning of `MaxWorkers`, defaulting to `runtime.NumCPU()`.
  - Added `TestGridSearchConcurrentStress` generating 100 isolated parameter permutations and successfully evaluating them cleanly across a fixed 8-worker thread pool.
  - Wrote architectural breakdown in `docs/ai/implementation/go-phase-35-concurrent-optimization.md`.

### Changed
- Refactored `optimizer.go` interfaces to accept the new configuration blocks.

### Verified
- Executed `go test ./internal/backtest/optimizer/...`, confirming that 100 full lifecycle strategy simulations execute and sort within ~30 milliseconds due to the new concurrent pipeline.

## [2.0.36] - 2026-04-06

### Added
- **Go Ultra-Project Phase-34 Optimization Subsystem**
  - Added the `internal/backtest/optimizer` package for systematically testing strategy parameters against historical data.
  - Implemented `GridSearchCandles`, orchestrating multiple sequential `backtest.Engine` runs against a Cartesian product of parameter combinations.
  - Added `StrategyBuilder` interface to dynamically spawn strategies with injected parameters.
  - Added `ScoringFunction` abstraction (defaulting to maximizing `RealizedPnL`).
  - Added comprehensive `grid_test.go` proving grid generation accuracy and correct sorting of optimization results using the `CandleSMACross` demo strategy.
  - Documented findings in `docs/ai/implementation/go-phase-34-optimization-subsystem.md`.

### Changed
- Updated `TODO.md` tracking to check off "Add optimization subsystem".
- Updated `go-feature-assimilation-matrix.md` to reflect full optimization capabilities natively in Go.

### Verified
- Executed `go test ./internal/backtest/optimizer/...` seamlessly, demonstrating extreme in-memory speed for repetitive backtest runs.

## [2.0.35] - 2026-04-06

### Added
- **Go Ultra-Project Phase-33 Advanced Market Emulation**
  - Added `EmulatorOptions` configuration struct to the `internal/backtest` engine to handle Maker/Taker fees and execution slippage rates.
  - Implemented mathematical modifiers in `processSignals` to adjust the simulated execution price, accurately penalizing strategy profitability by raising the effective buy entry and lowering the effective sell exit.
  - Added `DefaultEmulatorOptions` enforcing a standard 0.1% fee simulation.
  - Wrote a dedicated testing suite (`TestEngineRunFriction`) validating the exact compounding arithmetic of 5% slippage alongside 1% fees.
  - Detailed findings and architecture in `docs/ai/implementation/go-phase-33-advanced-market-emulation.md`.

### Changed
- Refactored `NewEngine` to default to the 0.1% standard fee emulator.
- Added `NewEngineWithOptions` for zero-friction test overriding or complex parameter tuning.
- Updated `TODO.md` to reflect the completion of advanced market emulation for the backtesting engine.

### Verified
- `go test ./internal/backtest/...` accurately captures the simulated value drain caused by slippage and commissions.

## [2.0.34] - 2026-04-06

### Added
- **Go Ultra-Project Phase-32 Candle-Driven Backtesting & Strategy Enhancements**
  - Upgraded the strategy `Runtime` to support `CandleStrategy` interfaces alongside the existing `TickStrategy`.
  - Added `CandleHistoryProvider` interface and `MemoryCandleHistory` to the `internal/backtest` engine.
  - Implemented `Engine.RunCandles()` to allow the backtester to iterate over multi-timeframe candle datasets, executing historical simulations based on candle Close prices.
  - Added a new `CandleSMACross` demo strategy to demonstrate moving average crossovers running specifically on interval-based candles rather than raw market ticks.
  - Added comprehensive test suites validating the correct flow of Candle-based signals through both the strategy runtime and the backtesting simulation.
  - Detailed the phase in `docs/ai/implementation/go-phase-32-candle-backtesting-and-strategies.md`.

### Changed
- Updated `TODO.md` to reflect the completion of candle/multi-timeframe strategy support.
- Re-categorized `TODO.md` items under "Backtesting and Simulation" for clarity.

### Verified
- Executed `go test ./internal/strategy/... ./internal/backtest/...` passing successfully.
- Validated crossover mathematics and deterministic backtesting states within memory arrays.

## [2.0.33] - 2026-04-06

### Added
- **Go Ultra-Project Phase-31 Backtesting Foundation**
  - Created a new `internal/backtest` package to simulate strategy execution against historical market data.
  - Implemented a `HistoryProvider` interface for injecting historical data arrays (e.g., `MemoryHistory`).
  - Added an isolated `Engine` that iteratively feeds historical ticks to a strategy and intercepts its signals.
  - Implemented simulated market execution within the backtest engine that calculates realized/unrealized PnL using the underlying `portfolio.Tracker`.
  - Added unit test suite `engine_test.go` to validate total trades and expected PnL on deterministic price sequences.
  - Added detailed documentation of this step in `docs/ai/implementation/go-phase-31-backtesting-foundation.md`.

### Changed
- Updated `TODO.md` to reflect completion of the initial backtesting subsystem.
- Updated `go-feature-assimilation-matrix.md` to show the backtesting subsystem as implemented.

### Verified
- `go test ./internal/backtest/...` passes cleanly, confirming isolated simulated execution logic.
- Entire Go workspace format and compilation checks passed.

## [2.0.32] - 2026-04-06

### Added
- **Go Ultra-Project Phase-30 Core Indicator Library and Technical Analysis**
  - Created a new `internal/indicator` package to house technical analysis tools, laying the groundwork for complex strategy creation.
  - Implemented the `SMA` (Simple Moving Average) indicator.
  - Implemented the `EMA` (Exponential Moving Average) indicator.
  - Implemented the `RSI` (Relative Strength Index) indicator.
  - Added `demo-ema-crossover` strategy to `internal/strategy/demo`, demonstrating the integration of technical indicators within the trading runtime.
  - Added documentation reflecting the completion of the core indicator library and strategy expansion at `docs/ai/implementation/go-phase-30-core-indicator-library.md`.

### Changed
- Refactored numeric string parsing into a centralized `utils.ParseFloat` function (`internal/core/utils/conv.go`) to unify calculation logic across portfolio, strategy, and execution domains.
- Updated `TODO.md` and the feature assimilation matrix to reflect the addition of a richer strategy library and foundational technical analysis capabilities.
- Updated the `app.go` runtime composition to integrate the `EMACrossover` strategy into the active schedule when running in timer mode.

### Verified
- `go test ./...` passes inside `ultratrader-go/` ensuring accuracy of mathematical indicator implementations.
- `go run ./cmd/ultratrader` initializes successfully after Phase-30 additions, properly scheduling indicator-driven strategies.

## [2.0.31] - 2026-04-06

### Added
- **Go Ultra-Project Phase-29 Stream Strategy Library Expansion**
  - Added `TickMeanReversion`, a third stream-aware demo strategy
  - Expanded stream-mode runtime composition to support threshold, momentum, and mean-reversion tick strategies together
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-29-stream-strategy-library-expansion.md`

### Changed
- Strengthened the stream-mode runtime path by broadening the variety of event-driven strategy behaviors represented in the demo strategy library

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-29 additions and retains the stream-capable runtime flow

## [2.0.30] - 2026-04-06

### Added
- **Go Ultra-Project Phase-28 Trend Widget Expansion**
  - Added dashboard visualizations for concentration drift and blocked-count trends
  - Added derived trend metrics for dominant block count and top concentration percentage in the runtime analysis layer
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-28-trend-widget-expansion.md`

### Changed
- Enhanced the dashboard to visualize both categorical diagnostics and time-varying risk/blocked behavior more clearly
- Enhanced runtime trend analysis with stronger derived metrics over concentration and block-reason history

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-28 visualization and trend-analysis additions

## [2.0.29] - 2026-04-06

### Added
- **Go Ultra-Project Phase-27 Dashboard Diagnostics Visualization Expansion**
  - Added dashboard bar charts for exposure concentration and guard block reasons
  - Extended the dashboard's historical/operator surface beyond line charts into categorical diagnostics visualizations
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-27-dashboard-diagnostics-visualization-expansion.md`

### Changed
- Enhanced the dashboard to visualize concentration percentages and block-reason counts using existing diagnostics APIs and report history surfaces

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-27 dashboard visualization updates

## [2.0.28] - 2026-04-06

### Added
- **Go Ultra-Project Phase-26 Dashboard Visualization Layer**
  - Added inline chart rendering for portfolio value history and execution success-rate history
  - Added richer dashboard layout sections for chart-focused operator inspection
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-26-dashboard-visualization-layer.md`

### Changed
- Enhanced the Go dashboard from a text-first console into a more visual operator surface with chart panels and richer history presentation

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-26 dashboard visualization improvements

## [2.0.27] - 2026-04-06

### Added
- **Go Ultra-Project Phase-25 Dashboard Enrichment**
  - Enhanced the built-in dashboard with summary cards, report history tables, and auto-refresh controls
  - Added client-side consumption of metrics and valuation history endpoints for richer operator visibility
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-25-dashboard-enrichment.md`

### Changed
- Updated `TODO.md` to reflect completion of deployment packaging and environment profile support in the operator-surface checklist
- Expanded the dashboard from a raw JSON panel set into a more structured operator-facing surface

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-25 dashboard improvements and serves the enriched dashboard page

## [2.0.26] - 2026-04-06

### Added
- **Go Ultra-Project Phase-24 Deployment Packaging and Environment Profiles**
  - Added initial Go runtime Dockerfile
  - Added `.dockerignore`
  - Added `docker-compose.yml`
  - Added example config profiles for timer, stream, and paper-service modes under `ultratrader-go/config/`

### Changed
- Expanded `DEPLOY.md` with containerized run guidance and the current Go runtime endpoint surface
- Updated `ultratrader-go/README.md` with config-profile and container usage examples
- Updated `TODO.md` to reflect completion of deployment packaging and environment profile support

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after the packaging/documentation additions

## [2.0.25] - 2026-04-06

### Added
- **Go Ultra-Project Phase-23 Advanced Risk Guards**
  - Added `duplicate-side` guard to suppress repeated same-side execution within a configured time window
  - Added `max-notional-per-symbol` guard to cap projected notional per symbol
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-23-advanced-risk-guards.md`

### Changed
- Expanded Go runtime risk configuration with `max_notional_per_symbol` and `duplicate_side_window_ms`
- Wired the new guards into the active runtime guard pipeline
- Updated `TODO.md` to reflect completion of the additional-guards milestone

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-23 additions and retains the current diagnostics/reporting behavior

## [2.0.24] - 2026-04-06

### Added
- **Go Ultra-Project Phase-22 Stream-Aware Strategy Library Growth**
  - Added `TickMomentumBurst`, a second event-driven demo strategy for stream mode
  - Extended stream-mode runtime composition so multiple tick-aware strategies can execute together
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-22-stream-aware-strategy-library-growth.md`

### Changed
- Enhanced the paper-stream runtime path so it now supports both threshold-driven and momentum-driven tick strategies
- Strengthened the credibility of the event-driven architecture by expanding beyond a single demo stream strategy

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-22 additions and retains the stream-capable runtime path

## [2.0.23] - 2026-04-06

### Added
- **Go Ultra-Project Phase-21 Operator Dashboard Bootstrap**
  - Added a built-in HTML dashboard page for the Go runtime
  - Dashboard consumes existing status, portfolio, execution, metrics, guards, and report APIs from the browser
  - Added HTTP tests verifying dashboard serving from the runtime handler
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-21-operator-dashboard-bootstrap.md`

### Changed
- Updated `TODO.md` to reflect completion of the initial Go runtime UI/dashboard layer
- Enhanced the HTTP root/dashboard routes to serve a lightweight operator interface over the existing diagnostics APIs

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-21 additions and serves the dashboard page through the Go HTTP layer

## [2.0.22] - 2026-04-06

### Added
- **Go Ultra-Project Phase-20 Exposure and Trend Diagnostics Expansion**
  - Added `/api/exposure-diagnostics` for operator-visible concentration and exposure summaries
  - Added concentration-oriented and block-reason-oriented trend metadata to runtime report trend analysis
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-20-exposure-and-trend-diagnostics.md`

### Changed
- Enhanced the Go diagnostics API surface with richer operator views over exposure and report trends
- Updated `TODO.md` to reflect completion of exposure/concentration diagnostics endpoints

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-20 additions and retains the existing runtime/reporting flow

## [2.0.21] - 2026-04-06

### Added
- **Go Ultra-Project Phase-19 Operator Diagnostics Surface Expansion**
  - Added `/api/portfolio-summary` for aggregate portfolio/operator views separate from raw positions
  - Added `/api/execution-diagnostics` combining execution summary and runtime metrics
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-19-operator-diagnostics-surface-expansion.md`

### Changed
- Enhanced the Go app diagnostics layer to expose richer portfolio and execution operator views
- Updated `TODO.md` to reflect completion of richer operator-facing diagnostics APIs and a distinct portfolio summary endpoint

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-19 additions and retains the existing runtime/reporting behavior

## [2.0.20] - 2026-04-06

### Added
- **Go Ultra-Project Phase-18 Runtime Report Trend Analysis**
  - Added report-trend analysis over persistent runtime reports
  - Added `/api/runtime-reports/trends` endpoint exposing derived runtime trends
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-18-runtime-report-trends.md`

### Changed
- Enhanced the Go HTTP diagnostics layer to expose interpreted report trends instead of only raw latest/history report data
- Enhanced app wiring to derive trend analytics from persistent metrics, valuation, and execution-summary reports

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-18 additions and retains the existing runtime reporting flow

## [2.0.19] - 2026-04-06

### Added
- **Go Ultra-Project Phase-17 Continuous Cycle Reporting**
  - Added reporting wrappers for both timer-driven and tick-driven scheduler execution paths
  - Added per-cycle durable report generation for metrics snapshots, portfolio valuation, and execution summaries
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-17-continuous-cycle-reporting.md`

### Changed
- Enhanced the Go app to route scheduler execution through reporting-aware wrappers so runtime history grows beyond startup-only writes
- Updated `TODO.md` to reflect completion of execution summary history over time

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-17 additions and continues to emit startup/runtime report data through the persistent reporting layer

## [2.0.18] - 2026-04-06

### Added
- **Go Ultra-Project Phase-16 Tick-Aware Strategy Runtime and Richer Paper Stream Simulation**
  - Added tick-aware strategy runtime support via `TickStrategy` and runtime `TickEvent()` handling
  - Added tick-driven demo strategy for stream-triggered threshold execution
  - Added stream-mode scheduler execution path that passes market ticks directly into the runtime
  - Enhanced the paper market-data feed with simulated varying tick sequences instead of a single repeated static price
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-16-tick-aware-runtime-and-stream-simulation.md`

### Changed
- Enhanced the Go app to select timer-mode or stream-mode strategy composition based on scheduler configuration
- Updated TODO tracking to reflect completion of richer paper stream simulation patterns

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-16 additions and retains the current paper execution loop while enabling stream-aware strategy infrastructure

## [2.0.17] - 2026-04-06

### Added
- **Go Ultra-Project Phase-15 Report History and Analytics Surface**
  - Added report history retrieval by type and limit in the runtime report store
  - Added `/api/runtime-reports/history` endpoint for operator-visible historical report access
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-15-report-history-and-analytics-surface.md`

### Changed
- Enhanced app wiring to expose report history through the HTTP diagnostics layer
- Updated `TODO.md` to reflect completion of the first runtime analytics/reporting layer milestone on top of report storage

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-15 additions and report history is available to the runtime API surface

## [2.0.16] - 2026-04-06

### Added
- **Go Ultra-Project Phase-14 Stream-Driven Strategy Consumption**
  - Added scheduler stream service that triggers runtime evaluation from subscribed market-data ticks
  - Added scheduler mode configuration to support `timer` and `stream` execution styles
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-14-stream-driven-strategy-consumption.md`

### Changed
- Enhanced the Go app to choose between timer-driven and stream-driven scheduler services based on config
- Updated TODO tracking to reflect completion of stream-driven strategy consumption and integration of stream-fed scheduling into the runtime lifecycle

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-14 additions and supports the scheduler mode configuration path

## [2.0.15] - 2026-04-06

### Added
- **Go Ultra-Project Phase-13 Execution Rates, Concentration Summaries, and Guard-Reason Metrics**
  - Added success-rate and blocked-rate calculations to runtime metrics
  - Added richer execution summary fields including unique symbol count and top-symbol activity
  - Added portfolio concentration summaries derived from live-valued positions
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-13-rates-concentration-and-reporting-summaries.md`

### Changed
- Enhanced operator diagnostics APIs so portfolio responses now include concentration data and metrics responses include rate information
- Enhanced execution/repository summaries to expose richer order-distribution context
- Updated `TODO.md` to reflect completion of richer execution diagnostics including success/block rates and concentration summaries

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-13 additions and now emits richer metric-rate information at startup

## [2.0.14] - 2026-04-06

### Added
- **Go Ultra-Project Phase-12 Runtime History Readback and Exposure View**
  - Added report-store readback helpers for latest reports and latest-by-type snapshots
  - Added `/api/runtime-reports/latest` endpoint for operator-visible report history access
  - Added live-valued `ExposureView` to support more realistic portfolio concentration evaluation
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-12-history-and-shutdown-integration.md`

### Changed
- Enhanced app startup to persist multiple durable report types (`startup-summary`, `metrics-snapshot`, `portfolio-valuation`)
- Enhanced app integration tests to cover startup with active HTTP runtime and coordinated shutdown behavior
- Enhanced `TODO.md` and the Go feature assimilation matrix to reflect durable report history and exposure-view progress

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-12 additions and emits persistent runtime report data

## [2.0.13] - 2026-04-06

### Added
- **Go Ultra-Project Phase-12 History and Shutdown Integration**
  - Added latest-report reading support for persistent runtime reports
  - Added `/api/runtime-reports/latest` endpoint for operator-visible report snapshots
  - Added live-valued exposure view to support more realistic concentration controls
  - Added app-level startup + shutdown integration testing with active HTTP runtime
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-12-history-and-shutdown-integration.md`

### Changed
- Enhanced app startup to persist multiple report types (`startup-summary`, `metrics-snapshot`, `portfolio-valuation`)
- Enhanced Go runtime configuration to include persistent report storage paths and concentration settings in the active runtime flow
- Enhanced the feature assimilation matrix and TODO tracking to reflect persistent history and shutdown progress

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-12 additions and continues to emit startup summary state to logs and persistent reports

## [2.0.12] - 2026-04-06

### Added
- **Go Ultra-Project Phase-11 Block Reasons and Diagnostics Depth**
  - Added structured `GuardError` propagation from the risk pipeline
  - Added block-reason tracking in runtime metrics
  - Added `/api/guard-diagnostics` endpoint exposing active guards together with block-reason-aware metrics
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-11-block-reasons-and-diagnostics-depth.md`

### Changed
- Enhanced execution service to classify blocked executions by guard name and record those classifications in metrics
- Enhanced the metrics tracker to retain per-guard block counts in addition to aggregate attempt/success/block totals
- Enhanced diagnostics documentation to reflect the deeper supervisory/guard-visibility model

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-11 additions and emits the existing startup/runtime summary logs

## [2.0.11] - 2026-04-05

### Added
- **Go Ultra-Project Phase-11 Governance, Streaming, and Reporting Baseline**
  - Added persistent runtime report storage and startup-summary persistence for the Go runtime
  - Added market-data streaming abstractions and deterministic paper tick subscription support
  - Added concentration-control groundwork and portfolio value helpers for future exposure enforcement
  - Added project-governance documents:
    - `VISION.md`
    - `MEMORY.md`
    - `DEPLOY.md`
    - `TODO.md`

### Changed
- Updated `ROADMAP.md` to explicitly track the Go ultra-project as a parallel workstream
- Updated `UNIVERSAL_LLM_INSTRUCTIONS.md`, `AGENTS.md`, `CLAUDE.md`, `GEMINI.md`, `GPT.md`, and `copilot-instructions.md` to better reflect the current dual-track Python + Go project direction and universal instruction hierarchy
- Updated `docs/ai/implementation/go-feature-assimilation-matrix.md` to reflect the current runtime/reporting/streaming state

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully with report persistence and paper-stream support

## [2.0.10] - 2026-04-05

### Added
- **Go Ultra-Project Phase-9/10 Exposure Controls, Market-Data Streams, and Persistent Runtime Reports**
  - Added persistent runtime report storage under `ultratrader-go/internal/persistence/reports`
  - Added market-data streaming abstractions and paper tick subscription support
  - Added `max-concentration` guard scaffold and wired `max-open-positions` into the runtime risk pipeline
  - Added `/api/guards` endpoint and richer operator-visible guard diagnostics
  - Added detailed implementation notes:
    - `docs/ai/implementation/go-phase-9-exposure-controls-and-marketdata-streams.md`
    - `docs/ai/implementation/go-phase-10-persistent-reports-and-exposure-controls.md`
- **Project Direction Documentation**
  - Added `VISION.md`
  - Added `MEMORY.md`
  - Added `DEPLOY.md`
  - Added `TODO.md`
  - Updated `ROADMAP.md` with a dedicated Go ultra-project parallel-track section

### Changed
- Expanded Go runtime configuration with report persistence and concentration risk settings
- Enhanced app startup to persist a durable runtime summary report containing metrics, PnL, portfolio value, and active guards
- Enhanced market-data layer with subscription-oriented interfaces to prepare for event-driven strategy evolution
- Enhanced portfolio tracker with value-query helpers for future exposure/concentration enforcement

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-9/10 additions and logs the active guard stack plus runtime summary data

## [2.0.9] - 2026-04-05

### Added
- **Go Ultra-Project Phase-8 Guard Diagnostics and Runtime Lifecycle**
  - Added `/api/guards` endpoint for operator-visible guard configuration diagnostics
  - Added `max-open-positions` guard for portfolio-aware admission control
  - Added explicit HTTP runtime lifecycle controls including `Address()` and `Shutdown()` with integration tests
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-8-guard-diagnostics-and-runtime-lifecycle.md`

### Changed
- Changed default Go runtime bind address to `127.0.0.1:0` to avoid local port collisions during development and validation
- Enhanced app diagnostics logging to include active guard names and resolved HTTP runtime address
- Enhanced portfolio tracker with open-position counting and state queries for risk enforcement
- Enhanced feature assimilation documentation to reflect runtime lifecycle control and guard diagnostics

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-8 additions and now binds to an ephemeral local port by default
- HTTP runtime lifecycle tests, guard tests, and app integration tests pass

## [2.0.8] - 2026-04-05

### Added
- **Go Ultra-Project Phase-7 Metrics, Diagnostics, and Guard Configuration**
  - Added in-memory execution metrics tracking for attempts, successes, and blocked executions
  - Added `/api/metrics` endpoint for runtime metrics exposure
  - Added richer execution summary and portfolio API responses including PnL values
  - Added configuration support for cooldown and duplicate-execution guard windows
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-7-metrics-diagnostics-and-guard-config.md`

### Changed
- Enhanced execution service to record metrics alongside existing journaling and correlation-aware logging
- Enhanced app wiring to expose metrics and richer diagnostics through HTTP handlers
- Enhanced feature assimilation documentation to reflect runtime metrics and richer diagnostics surfaces

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-7 additions
- Metrics tests, API handler tests, scheduler tests, and app integration tests pass

## [2.0.7] - 2026-04-05

### Added
- **Go Ultra-Project Phase-6 PnL, Guard Strengthening, Metrics, and Scheduler Lifecycle**
  - Added cooldown and duplicate-symbol guards for time-aware execution protection
  - Added richer in-memory execution repository summaries and recent-symbol detection
  - Added portfolio analytics for average entry, cost basis, realized PnL, unrealized PnL, and total market value
  - Added execution summary and richer portfolio data exposure through HTTP API surfaces
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-6-pnl-guards-metrics-and-scheduler.md`

### Changed
- Enhanced the paper exchange adapter to assign deterministic fill prices for market orders
- Enhanced the app wiring to include new guards and richer portfolio/execution diagnostics
- Enhanced the scheduler service to run through a generic runner abstraction with repeated-run test coverage
- Updated the feature assimilation matrix with the new runtime-state and PnL capabilities

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-6 additions
- Risk tests, scheduler lifecycle tests, repository summary tests, and portfolio PnL tests pass

## [2.0.6] - 2026-04-05

### Added
- **Go Ultra-Project Phase-5 Observability, Valuation, and API Surfaces**
  - Added structured JSON logging package with context-propagated correlation IDs
  - Added portfolio valuation using the market-data feed
  - Added runtime HTTP read-model endpoints:
    - `/api/status`
    - `/api/portfolio`
    - `/api/orders`
  - Added detailed implementation notes at `docs/ai/implementation/go-phase-5-observability-valuation-and-api.md`

### Changed
- Expanded `ultratrader-go` configuration with logging settings
- Upgraded execution service to emit correlation-aware structured logs and persist correlation IDs in order/event artifacts
- Upgraded app bootstrap to create a real logger and expose dynamic status/portfolio/order state through the HTTP handler
- Added logger cleanup support for tests and future shutdown handling

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes successfully after Phase-5 additions
- API handler tests, logging tests, valuation tests, and app integration tests pass

## [2.0.5] - 2026-04-05

### Added
- **Go Ultra-Project Phase-4 Risk, Portfolio, and Loop Foundations**
  - Added concrete risk guards for symbol whitelist and max notional enforcement
  - Added in-memory execution repository and portfolio tracker
  - Added market-data-aware demo strategy based on paper feed price thresholds
  - Added recurring scheduler service abstraction for future daemonized trading loops
  - Added detailed implementation docs:
    - `docs/ai/implementation/go-phase-4-risk-portfolio-and-loop.md`
    - updated `docs/ai/implementation/go-feature-assimilation-matrix.md`

### Changed
- Expanded `ultratrader-go` configuration with scheduler and risk sections
- Upgraded app bootstrap to wire concrete guards, execution memory, portfolio state, and the market-data-aware demo strategy
- Upgraded execution service to save orders to an in-memory repository and apply fills to the portfolio tracker

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` still initializes successfully after Phase-4 additions
- Scheduler tests, risk tests, portfolio tests, and app integration tests all pass

## [2.0.4] - 2026-04-05

### Added
- **Go Ultra-Project Phase-3 Market Data, Journaling, and Scheduling**
  - Added append-only order journaling under `ultratrader-go/internal/persistence/orders`
  - Added market-data abstractions and a deterministic paper market-data feed
  - Added demo strategy package with a bootstrap `buy-once` strategy
  - Added strategy scheduler that converts signals into execution requests
  - Added HTTP runtime wrapper for controllable health/readiness serving
  - Added comprehensive implementation docs:
    - `docs/ai/implementation/go-phase-3-marketdata-and-scheduling.md`
    - updated `docs/ai/implementation/go-feature-assimilation-matrix.md`

### Changed
- Extended `ultratrader-go` configuration with order-journal settings
- Expanded the execution service to persist order records in addition to event-log entries
- Expanded app bootstrap to wire market data, order journaling, demo strategy execution, and scheduler-driven startup behavior

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` still initializes successfully after Phase-3 additions
- App integration tests now validate event, snapshot, and order persistence together

## [2.0.3] - 2026-04-05

### Added
- **Go Ultra-Project Phase-2 Kernel Services**
  - Added exchange registry and the first clean-room `paper` exchange adapter
  - Added execution service to connect accounts, risk guards, exchange adapters, and event logging
  - Added append-only account snapshot persistence
  - Added health/readiness HTTP handler package
  - Added strategy runtime skeleton for future signal scheduling and aggregation
  - Added detailed implementation docs:
    - `docs/ai/implementation/go-phase-2-kernel-services.md`
    - `docs/ai/implementation/go-feature-assimilation-matrix.md`

### Changed
- Expanded `ultratrader-go` configuration to include snapshot and server settings
- Extended app bootstrap to register the paper exchange, initialize snapshot persistence, and emit bootstrap account snapshots
- Improved account service with deterministic listing and account lookup

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` still initializes successfully after the Phase-2 service additions

## [2.0.2] - 2026-04-05

### Added
- **Go Ultra-Project Phase-1 Scaffold**
  - Added new top-level Go module at `ultratrader-go/`
  - Added CLI entrypoint at `ultratrader-go/cmd/ultratrader`
  - Added foundational packages for config loading, event logging, unified account modeling, exchange capability contracts, and guard pipeline interfaces
  - Added unit tests for config loading, JSONL event logging, and guard pipeline error handling
  - Added implementation notes at `docs/ai/implementation/go-phase-1-scaffold.md`

### Verified
- `go test ./...` passes inside `ultratrader-go/`
- `go run ./cmd/ultratrader` initializes the scaffold successfully and writes the first event-log entry

## [2.0.1] - 2026-04-05

### Added
- **Submodule Research Corpus**
  - Added 50 crypto-trading repositories as organized git submodules under `submodules/page-02` through `submodules/page-06`
  - Added `SUBMODULES.md` manifest with ranking/order, paths, commit SHAs, and source URLs
  - Normalized `.gitmodules` entries to match the page-based layout

- **AI DevKit Phase Documentation**
  - Added `docs/ai/requirements/go-ultra-project-requirements.md`
  - Added `docs/ai/design/go-ultra-project-architecture.md`
  - Added `docs/ai/planning/go-ultra-project-program-plan.md`
  - Added `docs/ai/implementation/submodule-architecture-audit.md`
  - Added `docs/ai/implementation/submodule_inventory.json`

### Changed
- Established Stage-1 recommendation to use `c9s/bbgo` as the Go kernel reference and `TraderAlice/OpenAlice` as the architectural reference for the future unified Go ultra-project
- Documented licensing constraints across imported repos and formalized a clean-room reimplementation strategy instead of direct multi-project code transplantation

## [2.0.0] - 2026-01-18

### Added
- **Version Management System**
  - VERSION.md file for single source of truth version number
  - CHANGELOG.md for detailed change tracking
  - ROADMAP.md for feature planning
  - MODULE_INDEX.md for complete module inventory
  - pt_hub.py updated to display version number in window title (v2.0.0)

- **Comprehensive Documentation**
  - UNIVERSAL_LLM_INSTRUCTIONS.md - Universal AI agent instructions
  - CLAUDE.md - Anthropic Claude model-specific instructions
  - GEMINI.md - Google Gemini model-specific instructions
  - GPT.md - OpenAI GPT model-specific instructions
  - copilot-instructions.md - GitHub Copilot model-specific instructions
  - AGENTS.md - Comprehensive agent instruction documentation

- **Analytics Integration System** (pt_analytics.py - 771 lines)
  - SQLite-based persistent trade journal
  - TradeJournal class for logging entries, DCAs, and exits
  - PerformanceTracker class for metrics calculation
  - get_dashboard_metrics() function for real-time data
  - Trade group ID tracking for linking related trades
  - Graceful fallback if analytics module unavailable
  - Automatic logging integrated into pt_trader.py

- **Analytics Dashboard** (pt_analytics_dashboard.py - 262 lines)
  - KPICard widget for displaying key metrics
  - PerformanceTable widget for period comparisons
  - AnalyticsWidget main class integrating components
  - Real-time KPIs: Total trades, win rate, today's P&L, max drawdown
  - Period comparison tables (all-time, 7/30 days, 30 days)
  - Mtime-cached refresh (5 second default interval)
  - Integrated ANALYTICS tab in pt_hub.py GUI

- **Multi-Exchange Price Aggregation** (pt_exchanges.py - 1006 lines)
  - ExchangeManager unified interface for KuCoin, Binance, Coinbase
  - pt_thinker_exchanges.py wrapper module (96 lines)
  - get_aggregated_current_price() - Median/VWAP across exchanges
  - get_candle_from_exchanges() - OHLCV candles with fallback
  - detect_arbitrage_opportunities() - Cross-exchange spread monitoring
  - KuCoin primary source, Binance/Coinbase fallbacks
  - Arbitrage detection integrated into pt_thinker.py prediction loop
  - Robinhood current price unchanged (still execution source)

- **Notification System** (pt_notifications.py - 406 lines)
  - Unified notification interface via NotificationManager
  - EmailNotifier - Gmail integration via yagmail
  - DiscordNotifier - Webhook-based Discord notifications
  - TelegramNotifier - Bot token-based via python-telegram-bot
  - NotificationConfig dataclass for JSON-based configuration
  - Platform-specific rate limiting (Email: 2/hr, Discord: 30/min, Telegram: 20/min)
  - Notification levels: INFO, WARNING, ERROR, CRITICAL
  - NotificationDatabase for SQLite logging of sent notifications
  - CLI interface for configuration and testing
  - Integration points ready for pt_analytics.py and event logging

- **Volume Analysis System** (pt_volume.py - 237 lines)
  - VolumeMetrics dataclass (SMA_10, SMA_50, EMA_12, VWAP)
  - VolumeAnalyzer class with calculation methods
  - detect_anomaly() - Z-score based anomaly detection
  - calculate_trend() - Increasing/decreasing/stable detection
  - VolumeCLI for backtesting volume strategies
  - Integration points ready for pt_thinker.py and pt_analytics.py

- **Version Management System**
  - VERSION.md file for single source of truth version number
  - CHANGELOG.md for detailed change tracking
  - ROADMAP.md for feature planning
  - Version display in GUI
  - Automated version bumping with commits

 - **Comprehensive Documentation**
   - NOTIFICATIONS_README.md - Complete notification system documentation
   - NOTIFICATION_INTEGRATION.md - Integration guide for notifications
   - MODULE_INDEX.md - Submodule inventory with versions and locations
   - UNIVERSAL_LLM_INSTRUCTIONS.md - Universal AI agent instructions
   - CLAUDE.md - Claude model-specific instructions
   - GEMINI.md - Gemini model-specific instructions
   - GPT.md - GPT model-specific instructions
   - AGENTS.md - Comprehensive agent instruction documentation
   - MCP_SERVERS_RESEARCH.md - Research on 25+ MCP servers and financial libraries

 - **Multi-Asset Correlation Analysis** (pt_correlation.py - 447 lines)
   - CorrelationCalculator class for computing correlations
   - Portfolio correlation based on position sizes (weighted)
   - Historical correlation tracking with 7/30/90-day periods
   - Diversification alerts for high correlations (>0.8 threshold)
   - Correlation matrix calculation for multiple assets
   - Integration points ready for pt_thinker.py and pt_analytics.py

 - **Volatility-Adjusted Position Sizing** (pt_position_sizing.py - 414 lines)
   - VolatilityMetrics and PositionSizingResult dataclasses
   - PositionSizer class with ATR (Average True Range) calculation
   - True Range calculation for accurate volatility measurement
   - Risk-adjusted position sizing with configurable min/max (1% to 10%)
   - Volatility factor adjustment based on ATR %
     - Low volatility (<1%): 1.5x position size
     - Medium volatility (1-2%): 1.25x position size
     - High volatility (>5%): 0.75x position size
     - Very high volatility (>8%): 0.5x position size
   - Market volatility data retrieval from analytics database
   - Complete sizing recommendation system with volatility level classification
    - Main testing function with sample data generation

 - **Configuration Management System** (pt_config.py - 628 lines)
   - TradingConfig dataclass for all trading settings (entry, DCA, profit margin)
   - NotificationConfig dataclass for notification platforms and rate limiting
   - ExchangeConfig dataclass for API keys (KuCoin, Binance, Coinbase)
   - AnalyticsConfig dataclass for analytics database and retention settings
   - PositionSizingConfig dataclass for risk management settings
   - CorrelationConfig dataclass for correlation analysis settings
   - SystemConfig dataclass for logging level and debug mode
   - PowerTraderConfig unified configuration dataclass
   - ConfigValidator class for schema validation and constraint checking
   - ConfigManager singleton with hot-reload support
   - YAML-based configuration (more readable than JSON)
   - Environment variable overrides with POWERTRADER_ prefix
   - Migration path from existing gui_settings.json
   - File watcher for automatic config reloading
   - Callback system for configuration change notifications
   - Export methods (dict, JSON) for GUI integration
    - Default configuration file generation
    - Comprehensive main testing function with examples

 - **Structured Logging System** (pt_logging.py - 538 lines)
   - LogEntry dataclass for structured log entries
   - LogConfig dataclass for logging settings (level, file, rotation)
   - StructuredFormatter for JSON log output
   - ConsoleFormatter for human-readable console logs
   - CriticalLogHandler for critical log notifications
   - StructuredLogger class with rotation and retention policies
   - LogViewer class for dashboard integration
   - setup_logging() function for application-wide logging
   - get_logger() function for module-specific loggers
   - Log rotation by file size (configurable max size)
   - Backup log retention policy (configurable count)
   - Critical notification integration with pt_notifications.py
   - Log search functionality (query by level/module)
   - Recent logs retrieval for dashboard
   - Log summary generation (by level/module)
   - Specialized logging methods (trade, prediction, api_call)
   - Console output support with color-coded levels
   - JSON file logging for structured data
   - Main testing function with comprehensive examples

### Changed
- **pt_trader.py** - Integrated analytics logging into _record_trade() method (~50 lines)
  - Added TradeJournal import with graceful fallback
  - Added trade group ID tracking dictionary
  - Added analytics logging calls in buy/DCA/sell branches
  - Error handling prevents trading disruption

- **pt_thinker.py** - Enhanced price fetching and arbitrage monitoring (~30 lines)
  - Added pt_thinker_exchanges import
  - Added get_aggregated_current_price() and detect_arbitrage_opportunities() calls
  - Integrated arbitrage monitoring in step_coin() prediction loop

- **pt_hub.py** - Added ANALYTICS tab and version display (~40 lines)
  - Added AnalyticsWidget integration
  - Added dashboard refresh in main _tick() loop
  - Added version number display in GUI header
  - Added VERSION.md integration for dynamic version display
  - Added TradeJournal import with graceful fallback
  - Added trade group ID tracking dictionary
  - Added analytics logging calls in buy/DCA/sell branches
  - Error handling prevents trading disruption

- **pt_thinker.py** - Enhanced price fetching and arbitrage monitoring (~30 lines)
  - Added pt_thinker_exchanges import
  - Added get_aggregated_current_price() and detect_arbitrage_opportunities() calls
  - Integrated arbitrage monitoring in step_coin() prediction loop

- **pt_hub.py** - Added ANALYTICS tab and version display (~40 lines)
  - Added AnalyticsWidget integration
  - Added dashboard refresh in main _tick() loop
  - Added version number display in GUI header

- **requirements.txt** - Updated dependencies
  - Added yagmail for Gmail notifications
  - Added discord-webhook for Discord notifications
  - Added python-telegram-bot for Telegram notifications
  - Added requests for webhook API calls

### Fixed
- Graceful error handling for analytics module unavailability
- Robust fallback mechanisms for exchange price fetching
- Rate limiting to prevent API bans for notification platforms

### Technical Notes
- Single-point integration pattern for analytics logging
- Trade group IDs link entries/DCAs/exits for proper tracking
- Multi-exchange aggregation uses KuCoin as primary to maintain consistency
- Notifications are non-blocking via asyncio
- All new modules follow existing codebase patterns

## [1.0.0] - 2025-01-18

### Initial Release
- Core trading system with 4 main modules:
  - pt_hub.py - GUI and orchestration hub
  - pt_thinker.py - kNN-based price prediction AI
  - pt_trader.py - Trade execution with structured DCA
  - pt_trainer.py - AI training system
  - pt_backtester.py - Historical strategy testing

- Trading Strategy Features:
  - Instance-based (kNN/kernel-style) price predictor
  - Online per-instance reliability weighting
  - Multi-timeframe trading signals (1hr to 1wk)
  - Neural Levels for signal strength (LONG/SHORT)
  - Trade entry: LONG >= 3 and SHORT == 0
  - Structured DCA with 2 max DCAs per 24hr window
  - Trailing profit margin (5% no DCA, 2.5% with DCA)
  - Trailing margin gap: 0.5%

- Robinhood Integration:
  - Spot trading only
  - No stop loss (by design)
  - No liquidation risk
  - Real API key generation wizard in settings

- GUI Features:
  - Dark theme interface
  - Real-time price charts with neural level overlays
  - Trade status monitoring
  - Training progress tracking
  - Settings management
  - Multiple coin support with folder-based organization

- Documentation:
  - README.md with setup instructions
  - Apache 2.0 license

---

## Version Format
- **MAJOR**: Incompatible API changes
- **MINOR**: Backwards-compatible functionality additions
- **PATCH**: Backwards-compatible bug fixes

## Links
- [Repository](https://github.com/your-username/PowerTrader_AI)
- [Issues](https://github.com/your-username/PowerTrader_AI/issues)
- [Documentation](README.md)

---

**DO NOT TRUST THE POWERTRADER FORK FROM Drizztdowhateva!!!**

This is my personal trading bot that I decided to make open source. This system is meant to be a foundation/framework for you to build your dream bot!
