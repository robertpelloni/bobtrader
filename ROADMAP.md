# PowerTrader AI - Roadmap

**Version:** 3.3.0
**Last Updated:** 2026-02-25

## Version 3.3.0 (Released) - Web3 & Risk Maturity

### Completed Features ✅

#### Risk Management
- [x] **Correlation Matrix:** Pearson coefficient calculation for portfolio assets.
- [x] **Risk Dashboard:** Heatmap visualization of asset correlations.

#### Web3 Integration
- [x] **Wallet Connect:** MetaMask/EVM wallet connection via `ethers.BrowserProvider`.
- [x] **Context:** Global `WalletContext` available to all frontend components.

#### Advanced Strategy
- [x] **Grid Strategy:** Configurable Grid Trading logic with visual parameters.
- [x] **Arbitrage Scanner:** Multi-exchange opportunity finder.
- [x] **Strategy Sandbox:** Dynamic form generation for strategy parameters.

#### System Health
- [x] **Unified Documentation:** Consolidated agent instructions and deep analysis.
- [x] **Submodule Dashboard:** Clear view of integrated modules and versions.

---

## Version 3.2.0 (Released) - DeFi & AI

### Completed Features ✅
- [x] **Liquidity Manager:** Uniswap V3 auto-ranging and fee collection.
- [x] **DeepThinker:** LSTM AI engine with Model Versioning.
- [x] **Auto-Compounding:** Reinvesting Uniswap fees automatically.

---

## Upcoming Milestones (v3.4.0+)

### 1. Security Hardening
**Goal:** Protect user funds and data.
- [ ] **Encrypted Config:** Encrypt API keys in `config.yaml` using a master password.
- [ ] **Local Auth:** Simple login screen for the Web UI.

### 2. Social Sentiment
**Goal:** Trade based on hype and fear.
- [ ] **Twitter Scraper:** Monitor keywords ($BTC, #Crypto) for sentiment spikes.
- [ ] **Fear & Greed:** Integrate alternative.me API.

### 3. Mobile App
**Goal:** Trade on the go.
- [ ] Port React frontend to React Native (Expo).

---

**Last Updated:** 2026-02-25
**Current Version:** 3.3.0
