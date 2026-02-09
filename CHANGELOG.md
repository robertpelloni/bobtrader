# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [2.5.0] - 2026-01-18

### Added
- **Advanced Strategy Engine**: Implemented `StrategyFactory` allowing dynamic switching between strategies.
- **New Strategies**: Added `RSIStrategy` and `MACDStrategy` using native TypeScript technical analysis.
- **Portfolio Rebalancer**: Added `PortfolioRebalancer` engine to calculate rebalancing signals based on target allocations.
- **Strategy Management UI**: New Frontend page (`StrategyManager.tsx`) to activate and configure strategies.
- **Deep Cointrade Integration**: Ported Cointrade logic (RSI/BB/MACD) directly into `CointradeAdapter` using `TechnicalAnalysis.ts`, removing external dependencies.

## [2.4.0] - 2026-01-18

### Added
- **Core Math Library**: Implemented `TechnicalAnalysis.ts` (SMA, EMA, RSI, MACD, BB) without external dependencies.
- **Real Market Data**: `KuCoinConnector` and `BinanceConnector` now fetch real OHLCV data from public APIs.
- **Strategy Integration**: `Trader` now uses `SMAStrategy` for entry signals based on real market data.
- **Strategy Sandbox Live**: The Sandbox UI now runs strategies against real market data instead of mocks.

## [2.3.0] - 2026-01-18

### Added
- **Strategy Sandbox**: New frontend page for backtesting strategies (Cointrade, SMA) with visual charts.
- **Backend Strategy API**: New endpoint `POST /api/strategy/backtest` to run simulations.
- **Multi-Exchange Skeleton**: Added `KuCoinConnector` and `BinanceConnector` classes.
- **Persistence**: `AnalyticsManager` now connects to the real `hub_data/trades.db`.

## [2.2.0] - 2026-01-18

### Added
- **Real-Time Frontend**: Upgraded `Dashboard.tsx` to use WebSockets (`useWebSocket` hook) for live trade and account updates.
- **Advanced Cointrade Simulation**: Enhanced `CointradeAdapter` to simulate complex technical indicators (MACD, RSI, Bollinger Bands) and signal generation.
- **Improved Cointrade Structure**: Prepared `backend/src/modules/cointrade` for submodule injection.
- **Dashboard Hooks**: Added `useWebSocket` hook for reusable real-time data connection.

### Changed
- **Config Management**: Standardized configuration loading across the entire stack.
- **Documentation**: Major overhaul of `VISION.md`, `MANUAL.md`, and `PROJECT_STRUCTURE.md` to reflect the TypeScript evolution.

## [2.1.0] - 2026-01-18

### Added
- **Dockerization**: Added `Dockerfile` for Backend and Frontend, plus `docker-compose.yml` for full stack deployment.
- **Real-Time Data**: Implemented WebSocket server in backend and client integration for live updates.
- **Trainer Port**: Ported AI training logic to `Trainer.ts`, removing Python dependency for pattern generation.
- **Strategy Engine**: Added `SMAStrategy` as a proof-of-concept for the new Strategy interface.

## [2.0.0] - 2026-01-18

### Added
- **TypeScript Web Architecture**: Full port of the application to a Node.js Backend and React Frontend.
    - `powertrader-ts/backend`: Express.js server with TypeScript logic for Trader and Thinker.
    - `powertrader-ts/frontend`: React + Vite dashboard for real-time monitoring and configuration.
- **Risk Management Dashboard**: New UI for visualizing Correlation Matrix and Position Sizing recommendations.
- **Volume Analysis Dashboard**: New UI for analyzing Volume Profiles and anomalies.
- **Comprehensive Documentation**: Added `MANUAL.md`, `VISION.md`, `PROJECT_STRUCTURE.md`, and `UNIVERSAL_LLM_INSTRUCTIONS.md`.
- **ConfigManager**: Centralized YAML-based configuration management for both Python and TypeScript codebases.
- **Real-Time Dashboards**: Python `pt_hub.py` updated with tabs for Volume and Risk analysis.
- **Robinhood Authentication**: Implemented real Ed25519 signing for Robinhood API in TypeScript connector.

### Changed
- **Architecture**: Shifted from monolithic Python desktop app to a modular hybrid system (Python Legacy + TypeScript Web).
- **Configuration**: Deprecated scattered JSON config files in favor of `config.yaml`.
- **Refactoring**: Updated `pt_trader.py`, `pt_thinker.py`, and `pt_hub.py` to use the new `ConfigManager`.

### Fixed
- **Settings Persistence**: Fixed an issue where changing settings in the Python GUI required a restart to take effect.
- **Repository Hygiene**: Cleaned up `__pycache__` and `node_modules` from source control.

## [1.0.0] - 2025-01-18

### Added
- Initial release of PowerTrader AI (Python Desktop App).
- kNN-based Price Prediction AI ("The Thinker").
- Structured DCA Trading Engine ("The Trader").
- Multi-Exchange Price Aggregation (KuCoin, Binance, Coinbase).
- Basic Tkinter GUI (`pt_hub.py`).
