# Changelog

All notable changes to PowerTrader AI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
