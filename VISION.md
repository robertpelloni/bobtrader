# PowerTrader AI - Vision & Design

**Version:** 2.0.0
**Last Updated:** 2026-01-18

---

## 1. Ultimate Vision

PowerTrader AI aims to be the **most robust, autonomous, and accessible crypto trading system** for high-conviction assets. Unlike high-frequency scalpers that bleed fees, PowerTrader is designed for **long-term accumulation and wealth generation** using a unique blend of:

1.  **kNN-based Pattern Recognition ("The Thinker"):** Instead of black-box neural networks, we use interpretable "memory" of historical price actions to predict future bounds.
2.  **Structured DCA ("The Trader"):** A tiered Dollar Cost Averaging system that mathematically guarantees a lower average entry price during downturns, turning volatility into opportunity.
3.  **No-Loss Philosophy:** The system is built on the premise of never selling a high-conviction asset (BTC/ETH) at a loss. It waits, accumulates, and exits only when profitable.
4.  **Hybrid Architecture:** Combining the raw performance of a local backend with the accessibility of a modern web interface.

The ultimate goal is a "Set and Forget" system that runs 24/7/365, surviving bear markets through accumulation and thriving in bull markets through trailing profit taking.

---

## 2. System Architecture

The project is currently transitioning from a monolithic Python desktop app to a scalable **Web Architecture**.

### 2.1. The Backend (`powertrader-ts/backend`)
*   **Runtime:** Node.js + TypeScript.
*   **Role:** The brain and execution engine.
*   **Core Modules:**
    *   **Thinker:** Runs the kNN algorithm, analyzing OHLCV data to generate `LONG`/`SHORT` signals and price bounds.
    *   **Trader:** Executes orders. Manages state (DCA levels, Trailing Stops). Connects to exchanges.
    *   **Analytics:** Stores every trade, calculates Sharpe Ratio, Win Rate, and PnL. Uses SQLite for persistence.
    *   **Config:** Centralized YAML configuration with hot-reload.
*   **Extensions:**
    *   **HyperOpt:** Genetic algorithm for parameter tuning.
    *   **PaperTrading:** Risk-free simulation.

### 2.2. The Frontend (`powertrader-ts/frontend`)
*   **Runtime:** React + Vite.
*   **Role:** The command center.
*   **Features:**
    *   Real-time Dashboards (Volume, Risk, Account).
    *   Interactive Charts (Recharts).
    *   Configuration Management.
    *   Strategy Backtesting UI.

### 2.3. Data Flow
1.  **Market Data:** Ingested via `ExchangeConnector` (KuCoin/Binance/Robinhood).
2.  **Processing:** `Thinker` normalizes data and finds historical matches.
3.  **Signal:** `Thinker` emits a prediction (e.g., "Long confidence 85%").
4.  **Execution:** `Trader` evaluates signal against current position and DCA rules.
5.  **Action:** `Trader` routes order to `RobinhoodConnector`.
6.  **Feedback:** Order result is logged to `AnalyticsManager` and pushed to Frontend via WebSocket/API.

---

## 3. Design Principles

1.  **Robustness First:** The system must handle API failures, network outages, and crashes without losing money or state. State is persisted to disk (SQLite/JSON).
2.  **Transparency:** Every decision (buy/sell/hold) must be explainable. The UI must show *why* a trade happened (e.g., "DCA Triggered: Price dropped 5%").
3.  **Modularity:** Exchanges, Strategies, and Analytics are separate modules defined by Interfaces (`IExchangeConnector`, `IStrategy`).
4.  **Safety:** API Keys are encrypted or handled securely. Risk management checks (Correlation, Max Drawdown) act as circuit breakers.

---

## 4. Future Roadmap (The "All Features" Goal)

*   **Universal Exchange Support:** Support all major CCXT exchanges.
*   **Social Sentiment:** Integrate Twitter/Reddit sentiment analysis into `Thinker`.
*   **DeFi Integration:** Direct trading on DEXs (Uniswap) via RPC.
*   **AI Evolution:** Move from simple kNN to hybrid Transformer/LSTM models for sequence prediction.
*   **Mobile App:** React Native port of the frontend.

PowerTrader AI is not just a bot; it is a financial sovereignty tool.
