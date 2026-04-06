# Changelog

All notable changes to PowerTrader AI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
