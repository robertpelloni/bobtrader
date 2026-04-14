# BobTrader / PowerTrader-TS Architecture & Memory Summary

## Overview & Vision
The project (historically known as `powertrader-ts`, now transitioning into the `BobTrader` / `ultratrader-go` "Ultra-Project") is a comprehensive, multi-exchange cryptocurrency trading bot. The ultimate vision is to consolidate all best-in-class features from various submodules and open-source crypto bots into a single, highly performant, robust, and autonomous Go-based backend with a React/Vite frontend. 

## Architectural Evolution
- **Legacy Systems:** Originally built with a Node.js/Express/TypeScript backend and Python microservices (e.g., `pt_rebalancer.py`, `pt_nlp_strategy.py`, `pt_analytics.py`, `pt_thinker.py` for AI predictions). 
- **The Go Transition:** The system is undergoing a massive port to Go (`ultratrader-go/`). The architecture in Go is heavily modular, emphasizing interfaces, strict dependency injection, and concurrency-safe state management (`sync.Mutex` and `sync.RWMutex`).

## Core Modules & Go Port Status

### 1. Market Data (`internal/marketdata`)
- **Feeds:** Abstractions for `Tick` and `Candle` streaming (`Feed`, `StreamFeed`).
- **Exchanges:** Support for Binance (REST & WebSocket) and Paper Trading.
- **Aggregation:** A multi-exchange `Aggregator` combines feeds using strategies like `AveragePrice`, `MedianPrice`, and `Failover` to ensure high availability and robust pricing.

### 2. Trading & Execution (`internal/trading`)
- **Execution Service:** Tightly couples accounts, risk pipelines, exchange registries, and portfolio trackers to execute trades.
- **Portfolio Tracking:** Concurrency-safe in-memory `Tracker` manages position state, calculates cost basis, and tracks realized/unrealized PnL using real-time market data.
- **Rebalancing:** `Rebalancer` calculates drift against configured target allocations and generates buy/sell orders automatically, including wash-sale prevention heuristics.

### 3. Risk Management (`internal/risk`)
- A modular pipeline of "guards" evaluates every order intent before execution.
- Guards include maximum concentration limits, max open positions, duplicate side suppression, and circuit breakers for API resilience.

### 4. Backtesting & Simulation (`internal/backtest`)
The Go port features an institutional-grade simulation suite:
- **Multi-Symbol Synchronization:** `MultiSymbolFeed` handles concurrent data ingestion and chronologically aligns multiple asset timelines into `SyncCandle` snapshots to prevent look-ahead bias.
- **Walk-Forward Optimization:** Mitigates overfitting by dynamically chunking historical data into sequential In-Sample (training) and Out-Of-Sample (validation) windows.
- **Parameter Optimization (Grid Search):** Exhaustive, highly concurrent evaluation of parameter grids using worker pools.
- **Monte Carlo Simulation:** Uses Fisher-Yates shuffling on historical trade sequences to stress-test equity curves, identifying median drawdowns and probabilities of ruin.

### 5. Analytics & AI (`internal/analytics` & `internal/strategy/nlp`)
- **Sentiment Engine:** Aggregates and clamps sentiment scores from external providers (News, Twitter, Reddit) into a unified signal.
- **NLP Parsing:** Employs regex heuristics to convert natural language strategy descriptions (e.g., "Buy ETH when RSI drops below 30") into machine-readable `StrategyConfig` structures containing precise entry/exit conditions and risk parameters.

## Design Patterns & Decisions
- **Concurrency:** Go routines are heavily utilized for parallel optimization (Grid Search) and concurrent market data fetching. Standard library primitives (`sync.WaitGroup`, channels) orchestrate workloads.
- **Testing:** Test-driven development is prioritized. The Go implementation relies on mocking core interfaces (e.g., `mockFeed`, `mockOptEvaluator`) to ensure logic like wash-sale prevention and parameter validation works deterministically.
- **Documentation as State:** The user strictly mandates maintaining specific Markdown files as the ultimate source of truth (`TODO.md`, `ROADMAP.md`, `CHANGELOG.md`, `VERSION.md`, `AGENTS.md`). Automated version bumping and commit hygiene are required.
- **Robustness over Speed:** While low latency is a goal, the primary directives highlight "extreme reliability, secure, stable, robust" execution, favoring graceful failovers and exhaustive risk checking over raw execution speed.