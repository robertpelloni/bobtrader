# AI Handoff Protocol

**Date:** 2026-02-25
**From:** Google Jules (v3.2.4 Implementer)
**To:** Next Model (Gemini 3 / Claude Opus / GPT-5)

## 1. Current State Summary

The project `PowerTrader AI` has successfully migrated from a legacy Python desktop app to a robust TypeScript Web Architecture (`powertrader-ts`).

*   **Version:** v3.2.4
*   **Core Stack:** Node.js, Express, TypeScript, React, Vite, TensorFlow.js, Ethers.js.
*   **Key Modules Active:**
    *   `Trader`: DCA & Trailing Stop engine.
    *   `Thinker`: kNN & LSTM (`DeepThinker`) prediction engines.
    *   `DeFi`: `LiquidityManager` for Uniswap V3 (Auto-Compound active).
    *   `Strategies`: SMA, RSI, MACD, Grid.
    *   `Tools`: Backtest Engine, HyperOpt, Arbitrage Scanner.

## 2. Recent Accomplishments

1.  **Arbitrage Scanner:** Implemented multi-exchange price comparison and a new Dashboard page.
2.  **Grid Strategy:** Implemented a new strategy class for grid trading logic.
3.  **AI Lab:** Enhanced with Training Loss charts, Test Inference, and Model Versioning.
4.  **Refactoring:** Centralized `StrategyFactory` and `TechnicalAnalysis` utils.

## 3. Immediate Next Steps (The "Task List")

The current user request requires **"Extreme Depth"** and **"Total Completion"**.

1.  **Risk Analysis:**
    *   **Task:** Implement `CorrelationMatrix.ts` in the backend.
    *   **Task:** Build the Heatmap visualization in `RiskDashboard.tsx`.
    *   **Goal:** Warn users if they are holding highly correlated assets (e.g., BTC & ETH often > 0.9).

2.  **Settings & Configuration:**
    *   **Task:** The `Settings` page is currently basic. It needs a full form for Notifications (Discord/Telegram) and Exchange API Keys.
    *   **Goal:** User should never have to touch `config.yaml` manually.

3.  **Security:**
    *   **Task:** Review where Private Keys are stored. Currently in `config.yaml` or Env.
    *   **Idea:** Implement an encrypted vault or simple password protection for the UI.

## 4. Known Quirks

*   **Repo Hygiene:** There might be some `*.log` or `*.db` files in the history. `gitignore` was recently updated, but check for stragglers.
*   **Submodules:** The user mentions `cointrade` as a submodule. We implemented its features via `CointradeAdapter` because we couldn't clone it. Continue this pattern if "submodules" are requested but inaccessible.

## 5. Deployment

*   **Docker:** The `docker-compose.yml` is the source of truth for running the stack.
*   **Local:** `cd powertrader-ts/backend && npm run build` + `cd powertrader-ts/frontend && npm run dev`.

---

*Proceed with the "High Priority" items in `TODO.md`.*
