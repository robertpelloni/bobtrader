# BobTrader Project Architecture & Implementation Summary

## 1. Project Evolution & Vision
BobTrader has evolved from a feature-rich but monolithic Python-based trading bot (**PowerTrader AI**) into a high-performance, modular Go-native trading platform (**Project Ultra** or `ultratrader-go`). 

*   **Core Goal:** To build the "Best Crypto Trading Suite" by clean-room assimilating the strongest architectural patterns from leading open-source projects (OpenAlice, bbgo, ccxt, WolfBot, pycryptobot, freqtrade).
*   **Key Philosophy:** Correctness and safety first, followed by observability, architecture quality, and finally feature breadth.
*   **Current State (v2.1.5):** A sophisticated hierarchical trading system supporting high-frequency micro-scalping, macro trend following, predictive ML ensembles, and cross-exchange arbitrage.

## 2. Core Go Architecture (`ultratrader-go`)
The system follows a modular "Kernel" design centered around dependency injection and interface-driven composition.

### Primary Components:
*   **App Container (`internal/core/app/app.go`):** The composition root. It manages the lifecycle of all services (HTTP server, Strategy Runtime, Scheduler, Execution Service).
*   **Exchange Abstraction (`internal/exchange/`):** Inspired by **CCXT**, providing a unified interface for multiple exchanges. It includes a robust **Paper Trading Adapter** that simulates market-aware fills, fees (0.1% taker), and slippage.
*   **Market Data Feed (`internal/marketdata/`):** Supports both REST polling and high-speed WebSocket streaming. The `marketdata.Tick` struct is volume-aware (includes `Quantity`).
*   **Risk Pipeline (`internal/risk/`):** A "Policy-Before-Execution" system. Every order must pass through a sequential chain of "Guards" (Whitelist, MaxNotional, Cooldown, Duplicate-Side, MaxConcentration, Drawdown).

## 3. Advanced Strategy Patterns
The system implements a **Hierarchical Strategy Architecture** where macro-level market regimes govern micro-level execution.

### Hierarchical Coordination:
*   **Regime Filter:** A composite pattern where a `MacroRegimeStrategy` (using EMA and ADX) classifies the market (Trending Bullish/Bearish vs. Ranging).
*   **Signal Suppression:** High-frequency strategies like `MicroScalper` are wrapped in a `RegimeFilter`, ensuring they only trade in the direction of the macro trend.

### Profit Siphoning Mechanism:
*   **`SiphoningManager`:** A strategic component that monitors realized PnL. It automatically "siphons" a configurable percentage (e.g., 10%) of scalp profits into long-term macro trend positions (e.g., BTC/ETH), effectively converting short-term volatility into durable wealth.

### Predictive Alpha & ML:
*   **Ensemble Predictor:** Aggregates signals from multiple ML models.
*   **KNN Model:** A k-Nearest Neighbors implementation that finds historical market patterns similar to the current feature vector to predict high/low price movements with associated confidence scores.

## 4. Operational & Diagnostic Features
*   **Operator Dashboard:** A professional UI served via HTTP API, providing real-time visibility into portfolio valuation, guard status, execution metrics, and strategy statistics.
*   **Trade Journal:** Persists all signals and execution summaries to JSONL for post-trade analysis and backtesting.
*   **Resilience:** Includes an API **Circuit Breaker** to prevent cascading failures during exchange outages.
*   **Dynamic Runtime:** Supports runtime injection of strategies and schedulers, enabling seamless transitions between paper and live trading modes.

## 5. Strategic Development Decisions
*   **Clean-Room Implementation:** Rather than direct source merges, the project reimplements best-in-class features in Go to ensure license compliance and code quality.
*   **Environment Profiles:** Managed via JSON configurations (e.g., `autonomous-paper.json`, `live-trading-binance.json`) to separate testing from production.
*   **Versioning Discipline:** Strict adherence to version tracking in `VERSION.md` with explicit version bumps in commit messages.
*   **Autonomous Autopilot:** Designed for continuous execution with frequent git syncs and automated repo sanitization (submodule updates).

## 6. Key Interface Definitions
*   **`strategy.Strategy`:** The base interface for all signal-generating logic.
*   **`exchange.Adapter`:** Standardizes order placement, balance fetching, and market data across CEXs.
*   **`risk.Guard`:** Defines the contract for trade-blocking safety logic.
*   **`ml.Model`:** Enables the plug-and-play addition of new machine learning algorithms.