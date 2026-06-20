# Session Handoff

## Summary of Accomplishments

During this session, we completed several major milestones on the `v2.1.x` roadmap for BobTrader:

1.  **Dashboard Modernization**:
    *   Scaffolded a new React/Vite Single Page Application (SPA) in `ultratrader-ui`.
    *   Configured TailwindCSS for styling and set up basic routing with `react-router-dom`.
    *   Implemented `fetchWithAuth` for secure API communication using JWT tokens.
    *   Created a robust `WebSocketClient` class for real-time data streaming to the frontend.
    *   Built `PriceChart` (using `recharts`) and `StrategyConfig` components.
2.  **WebSocket Feed Hardening**:
    *   Fixed a bug in the Binance `StreamFeed` where `parseTickerMessage` failed to unmarshal large epoch numbers (changed `EventTime` and `Price` to `json.Number`).
    *   Fixed the WebSocket endpoint URL construction.
    *   Added auto-reconnection logic with exponential backoff and verified it with a mock-based test (`TestWSFeed_ReconnectWithExponentialBackoff`).
    *   Implemented a WebSocket health monitoring endpoint (`/api/ws-health`) in the Go backend and wired it to the React dashboard UI.
    *   Switched the default market data source from `rest` to `websocket` across all relevant configuration files.
3.  **Backend Bug Fixes**:
    *   Fixed a compilation error in `app.go` caused by an out-of-order variable initialization (the `reconciler` was being initialized before `execAdapter` was created).

## System State & Next Steps

*   The project is now fully compiling, and all tests pass (ignoring flaky live-connection tests in CI).
*   The `TODO.md` and `ROADMAP.md` have been updated to reflect the completed tasks.
*   **Next Priority**: Review the "Strategy Enhancement" section of the `TODO.md`, specifically "Walk-forward parameter optimization on historical data" or "MACD strategy in stream mode", or move on to the "Mid-Term Roadmap" items like "Drawdown monitoring with auto-shutdown".

## Key Learnings & Context
*   **Binance JSON APIs**: Must use `json.Number` for numeric fields that could be exceptionally large or inconsistently formatted as strings vs numbers (e.g., `E` EventTime, `c` Price).
*   **Testing**: Do not rely on real Binance testnet WebSockets for CI tests, as they can be completely silent and cause timeouts.
*   **Architecture**: The backend uses an adapter pattern for exchanges and a pipeline pattern for risk guards. The frontend is a standard Vite React app that proxies `/api` calls to the Go backend on port 8400.
