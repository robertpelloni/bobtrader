# PowerTrader AI - Todo & Backlog

**Version:** 3.2.4
**Last Updated:** 2026-02-25

## 🔴 High Priority (Immediate)

- [ ] **Risk Engine - Correlation Matrix:** Implement `CorrelationMatrix.ts` to calculate Pearson coefficients between monitored assets. Visualize in `RiskDashboard.tsx`.
- [ ] **Settings UI - Notifications:** Add a form to `Settings.tsx` to configure Discord/Telegram webhooks/keys without editing YAML.
- [ ] **Wallet Connect:** Allow users to connect their Web3 wallet (MetaMask/Rabby) in the frontend for "View Only" portfolio tracking alongside the bot's internal wallet.

## 🟡 Medium Priority (Polishing)

- [ ] **Paper Trading Toggle:** Ensure the Dashboard has a clear visual indicator of "LIVE" vs "PAPER" mode, with a quick toggle (requiring restart).
- [ ] **Strategy Params UI:** `StrategySandbox` allows selecting strategies, but passing custom parameters (period, thresholds) via UI is currently limited. Add a dynamic form builder based on strategy `setParameters`.
- [ ] **Mobile Responsive Tweaks:** Ensure `LiquidityDashboard` and `ArbitrageDashboard` tables stack correctly on mobile screens.

## 🟢 Low Priority (Future Features)

- [ ] **Social Sentiment Module:** Implement the "Socializer" (Twitter/Reddit scraper).
- [ ] **Grid Bot Visualization:** Show the grid lines on the main price chart.
- [ ] **Exchange "Fill" Websockets:** Currently we rely on polling/simulated fills. Listen to real "Order Update" websockets from Binance/KuCoin.

## 🔵 Documentation & Maintenance

- [ ] **Swagger/OpenAPI:** Generate API documentation for the backend endpoints.
- [ ] **Unit Tests:** Increase coverage for `LiquidityManager` and `ArbitrageScanner`.
