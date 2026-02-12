# PowerTrader AI - Deep System Analysis

**Generated:** 2026-02-12
**Version:** 3.2.0

This document provides a comprehensive deep dive into the architecture, design philosophy, and operational mechanics of the PowerTrader AI ecosystem. It is intended for advanced developers and AI models tasked with maintaining or extending the system.

---

## 1. Architectural Philosophy

PowerTrader AI has evolved from a monolithic Python desktop application into a modern, distributed TypeScript microservices-style architecture.

### The "Why" of the Migration
The original Python implementation (`pt_*.py`) relied on `tkinter` for UI and blocking loops for logic. While functional, it suffered from:
1.  **UI Limitations:** Desktop GUIs are hard to access remotely.
2.  **Concurrency:** Python's GIL made true parallelism difficult for high-frequency checks across multiple exchanges.
3.  **Type Safety:** Dynamic typing led to runtime errors in complex financial logic.

The new **TypeScript Architecture** (`powertrader-ts`) solves this by:
1.  **React Frontend:** A responsive, web-based dashboard accessible from anywhere.
2.  **Node.js Backend:** Non-blocking I/O ideal for handling multiple WebSocket streams from exchanges.
3.  **Strict Typing:** TypeScript interfaces ensure data integrity across the trading pipeline.

---

## 2. Core Modules Breakdown

### A. The Trader (`src/trader/Trader.ts`)
The `Trader` is the execution heart of the system. It does *not* make decisions; it executes instructions based on signals and configuration.

*   **DCA Engine (Dollar Cost Averaging):**
    *   Logic: When a position goes against us, the Trader calculates "DCA Levels" based on percentage drops (e.g., -2.5%, -5%).
    *   Multiplier: Each subsequent buy is larger (e.g., 2.0x) to aggressively lower the average cost basis.
    *   State: Managed in-memory and persisted to SQLite via `AnalyticsManager`.

*   **Trailing Profit System:**
    *   Once a position is profitable (e.g., >1.5%), a "Trailing Stop" is activated.
    *   The stop price follows the market price up but never moves down.
    *   Execution: Market sell when price crosses below the trail.

### B. The Thinker (`src/thinker/Thinker.ts` & `DeepThinker.ts`)
The "Brain" is split into two evolutionary stages:

1.  **Standard Thinker (kNN):**
    *   **Algorithm:** k-Nearest Neighbors.
    *   **Logic:** Finds the 10 most similar historical price patterns to the current market state.
    *   **Output:** "Neural Levels" (0-7). Level 7 means 7/10 similar past instances resulted in a price increase.
    *   **Pros:** Fast, explainable, robust for mean reversion.

2.  **DeepThinker (LSTM):**
    *   **Tech Stack:** `@tensorflow/tfjs-node`.
    *   **Architecture:** Long Short-Term Memory (LSTM) Recurrent Neural Network.
    *   **Input:** Sequences of OHLCV data normalized to [0,1].
    *   **Output:** Probability of next candle closing higher.
    *   **Integration:** Currently runs in parallel; can be configured to "vet" kNN signals.

### C. Liquidity Manager (`src/defi/LiquidityManager.ts`)
The v3.2.0 addition for decentralized market making on Uniswap V3.

*   **Math:** Uses **Bollinger Bands** (20 SMA, 2 STD) to determine the "likely" trading range of an asset.
*   **Ticks:** Converts prices to Uniswap `int24` ticks (`log_1.0001(price)`).
*   **Operation:**
    *   `mint()`: Creates an NFT position within the calculated range.
    *   `collect()`: Harvests accrued trading fees.
    *   `burn()`: Removes liquidity.

---

## 3. Data Flow Architecture

The system operates on an event-driven loop triggered by market data ticks.

1.  **Ingestion:**
    *   `ExchangeConnector` (e.g., `BinanceConnector`) subscribes to WebSocket streams for Price and Account Balance.
    *   *Normalization:* All incoming data is converted to a standard `Ticker` object.

2.  **Analysis (The "Think" Phase):**
    *   `Trader` receives the tick.
    *   It queries `Thinker` for the current signal strength (0-7).
    *   It checks `LiquidityManager` for active ranges (if DeFi is enabled).

3.  **Decision (The "Act" Phase):**
    *   **Entry:** If `Signal >= StartLevel` AND `No Position`, `createOrder(BUY)`.
    *   **DCA:** If `Price <= LastBuy * (1 - Drop%)`, `createOrder(BUY)`.
    *   **Exit:** If `Price <= TrailingStop`, `createOrder(SELL)`.

4.  **Feedback (The UI Loop):**
    *   `Trader` emits `TRADE_UPDATE` events via `api/websocket.ts`.
    *   **Frontend (React Query):** Receives the event and invalidates the `usePositions` query, causing an immediate UI re-render.

---

## 4. Extension Points for Future Models

### Adding a New Exchange
1.  Implement `IExchangeConnector` interface (`src/engine/connector/IExchangeConnector.ts`).
2.  Required methods: `fetchTicker`, `fetchBalance`, `createOrder`.
3.  Register in `Trader` constructor based on `config.yaml`.

### Adding a New Strategy
1.  Extend the `Strategy` base class (`src/engine/strategy/BaseStrategy.ts`).
2.  Implement `analyze(candles: Candle[]): Signal`.
3.  Register in `StrategyFactory`.

### Enhancing AI
1.  Modify `DeepThinker.ts`.
2.  The model topology is defined in `buildModel()`.
3.  You can add layers (Dropout, Dense) or change the optimizer (Adam) here.
4.  Retrain via the "AI Lab" frontend page.

---

## 5. Legacy vs. Modern Map

| Feature | Legacy Python (`pt_*.py`) | Modern TS (`powertrader-ts`) |
| :--- | :--- | :--- |
| **GUI** | Tkinter (Desktop) | React + Tailwind (Web) |
| **Backend** | Python Main Loop | Node.js + Express |
| **Database** | JSON Files | SQLite (`hub_data/trades.db`) |
| **AI** | `sklearn` (kNN) | TensorFlow.js (LSTM) |
| **DeFi** | None | `ethers.js` + Uniswap V3 |

---

## 6. Known Limitations & Roadmap

1.  **Mobile Support:** The frontend is responsive but not a native app. React Native port is planned.
2.  **Gas Optimization:** `LiquidityManager` does not yet account for L1 gas fees (less critical on Polygon).
3.  **Backtesting:** The `BacktestEngine` is robust but slow for multi-year simulations with LSTM.

---

*This document serves as the primary technical reference for the PowerTrader AI internal architecture.*
