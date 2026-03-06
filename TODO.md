# PowerTrader AI - Todo & Backlog

**Version:** 3.2.4
**Last Updated:** 2026-02-25

## 🔴 High Priority (Immediate)

- [x] **Risk Engine - Correlation Matrix:** Implemented `CorrelationMatrix.ts` & UI.
- [x] **Settings UI - Notifications:** Fully implemented in `Settings.tsx`.
- [x] **Wallet Connect:** Implemented globally via `WalletContext`.

## 🟡 Medium Priority (Polishing)

- [x] **Paper Trading Toggle:** Implemented with `/api/system/mode`.
- [x] **Strategy Params UI:** Dynamic form added to `StrategySandbox.tsx`.
- [x] **Mobile Responsive Tweaks:** Ensure `LiquidityDashboard` and `ArbitrageDashboard` tables stack correctly on mobile screens.
- [x] **Core Engine Robustness:** Refactor `TechnicalAnalysis.ts` for speed and ensure `BacktestEngine.ts` supports dynamic fees and better edge cases.

## 🟢 Low Priority (Future Features)

- [x] **Social Sentiment Module:** Implemented "The Socializer".
- [x] **Order Book Arbitrage:** Upgraded Arbitrage Scanner to use Order Book depth.
- [x] **Grid Bot Visualization:** Show the grid lines on the main price chart.
- [x] **On-Chain Whale Watcher:** Monitor large ERC20 stablecoin transfers to gauge institutional money flow.
- [ ] **Exchange "Fill" Websockets:** Currently we rely on polling/simulated fills. Listen to real "Order Update" websockets from Binance/KuCoin.
- [ ] **Secure UI Login:** Create `Login.tsx` and implement the UI layer for the backend auth middleware.

## 🔵 Documentation & Maintenance

- [ ] **Swagger/OpenAPI:** Generate API documentation for the backend endpoints.
- [ ] **Unit Tests:** Increase coverage for `LiquidityManager` and `ArbitrageScanner`.
