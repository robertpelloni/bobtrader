# Changelog

All notable changes to this project will be documented in this file.

## [3.5.0] - 2026-02-25

### Added
- **Social Sentiment:** The "Socializer" engine now scores fear/greed and validates trades.
- **Arbitrage Upgrades:** Scanner now factors in actual Order Book depth and slippage instead of just top-of-book prices.
- **Robustness:** Major performance refactor of `TechnicalAnalysis.ts`. Enhanced `BacktestEngine.ts` to handle dynamic position sizes and hard stop losses.
- **UI:** Upgraded System Status page to act as a comprehensive "Submodule Dashboard" with build dates and descriptions.

## [3.3.0] - 2026-02-25

### Added
- **Web3 Integration:** Added `WalletConnect` functionality to the frontend sidebar using `ethers.js`.
- **Risk Engine:** Implemented `CorrelationMatrix` calculation and visualization in `RiskDashboard`.
- **Settings UI:** Full configuration form for Notifications (Discord/Telegram) and Trading params.
- **Strategy Sandbox:** Dynamic parameter forms for `GridStrategy` and `MACDStrategy`.
- **Documentation:** Unified agent instructions into `UNIVERSAL_LLM_INSTRUCTIONS.md` and updated `VISION.md`.

### Changed
- Refactored `StrategyFactory` to support dynamic loading.
- Updated `SystemStatus` page to clarify submodule integration.

## [3.2.0] - 2026-02-10

### Added
- **DeFi:** Liquidity Manager for Uniswap V3 (Auto-Ranging, Auto-Compounding).
- **AI:** DeepThinker LSTM Engine with Training UI.
- **Arbitrage:** Multi-exchange scanner.

---
