# PowerTrader AI / BobTrader — Roadmap

## Current State: v2.0.54 — Autonomous Paper Trading Platform

The Go ultra-project (`ultratrader-go/`) is now the primary development track. The legacy Python system is preserved for reference but no longer actively developed.

### Go Ultra-Project — Current Capabilities

| Category | Status | Details |
|----------|--------|---------|
| **Runtime** | ✅ | JSON config, graceful shutdown, signal handling |
| **Market Data** | ✅ | Binance.US REST feed (5s), WebSocket feed (experimental) |
| **Strategies** | ✅ | 8 strategies: EMA, Bollinger, RSI, TrailingTP, WolfBot Bollinger, DoubleEMA, MarketMaker, Safety |
| **Risk Pipeline** | ✅ | 8 guards: whitelist, max-notional, max-notional/symbol, cooldown, duplicate-symbol, duplicate-side, max-positions, max-concentration |
| **Execution** | ✅ | Paper trading with 0.1% taker fee, market-aware fills, sell-qty capping |
| **Portfolio** | ✅ | Position tracking, avg entry price, unrealized/realized PnL |
| **Diagnostics** | ✅ | Health/readiness, portfolio, orders, execution, exposure, guards, runtime reports, trends |
| **Signal Log** | ✅ | JSONL persistence, auto-flush, strategy stats API |
| **Build & Deploy** | ✅ | Docker, docker-compose, Windows/Linux binaries |
| **Backtesting** | ✅ | Multi-symbol, walk-forward, grid search, Monte Carlo |
| **Security** | ✅ | AES-GCM secrets, RBAC, rate limiting, audit logging, input validation |
| **Notifications** | ✅ | Discord, Telegram, Email channels |

### Verified Performance (v2.0.53, 20-min live test)
| Metric | Result |
|--------|--------|
| Executed Trades | 47 |
| Win Rate | 80% (20W / 5L) |
| Bollinger Strategy | 87% WR, +$0.14 PnL |
| RSI Strategy | 100% WR |
| Guard Block Rate | 39% |

---

## Near-Term Roadmap (v2.1.x)

### 1. WebSocket Feed Hardening
- [x] Debug goroutine-to-channel delivery in WS feed
- [x] Add auto-reconnection with exponential backoff
- [x] Add WebSocket health monitoring endpoint
- [x] Switch default from REST to WS once stable

### 2. Strategy Enhancement
- [ ] Kelly criterion / volatility-adjusted position sizing
- [ ] Walk-forward parameter optimization on historical data
- [ ] MACD strategy in stream mode
- [ ] ATR-based dynamic sizing
- [ ] Strategy backtesting with real market history

### 3. Real Exchange Integration
- [x] Wire execution to real Binance spot API
- [x] Order reconciliation service
- [x] Trade history sync from exchange
- [x] Circuit breaker for API resilience

### 4. Dashboard Modernization
- [ ] React/Vite SPA dashboard
- [ ] Real-time WebSocket streaming to frontend
- [ ] Interactive charts (TradingView lightweight)
- [ ] Strategy parameter tuning UI

---

## Mid-Term Roadmap (v2.2.x – v2.5.x)

### 5. Advanced Risk Management
- [ ] Drawdown monitoring with auto-shutdown
- [ ] Volatility-based position limits
- [ ] Liquidity checks before large trades
- [ ] Portfolio rebalancer with wash-sale prevention

### 6. Analytics & Reporting
- [ ] Persistent metrics/valuation history
- [ ] Trade journal analytics (long-term performance)
- [ ] Correlation analysis engine
- [ ] Market regime detection integration

### 7. Multi-Exchange Support
- [ ] KuCoin adapter
- [ ] Coinbase adapter
- [ ] Price aggregation across exchanges
- [ ] Cross-exchange arbitrage detection

### 8. AI/Research Layer
- [ ] Sentiment analysis integration (MCP servers)
- [ ] NLP strategy parsing
- [ ] Q-Learning RL optimizer
- [ ] Feature engineering pipeline

---

## Long-Term Vision

### Production-Grade Trading Platform
- **Correctness first** — Every trade must be accounted for, every fill reconciled
- **Full observability** — Every state transition inspectable via API
- **Daemon-ready** — Run 24/7 on Linux VPS with minimal ops
- **AI-assisted** — Optional AI/research layers that assist without destabilizing

### Architectural Thesis
- **OpenAlice-style platform architecture**
- **BBGO-style Go trading kernel**
- **CCXT-style exchange abstraction**
- **WolfBot-style advanced execution patterns**

---

## Legacy Python System (Preserved, Not Active)

The Python system (`pt_*.py` files) is feature-rich but frozen. Key modules:
- `pt_hub.py` (5,835 lines) — Tkinter GUI
- `pt_thinker.py` (1,381 lines) — kNN prediction AI
- `pt_trader.py` (2,421 lines) — Robinhood trading
- `pt_analytics.py` (770 lines) — SQLite trade journal
- `pt_notifications.py` (876 lines) — Multi-platform notifications
- `pt_exchanges.py` (663 lines) — Multi-exchange price aggregation

See `MODULE_INDEX.md` for Python→Go module mapping.

---

## Reference Submodules (44 active)

Research corpus of open-source trading bots for architecture/feature assimilation:

| Page | Notable Projects | Purpose |
|------|-----------------|---------|
| page-02 | OpenAlice, Krypto-trading-bot, xcrypto | Architecture, market-making |
| page-03 | WolfBot, CryptoBot, golang-crypto-trading-bot | Strategy patterns, Go kernel |
| page-04 | bbgo, ccxt, CryptoTradingFramework | Go kernel, exchange abstraction |
| page-05 | FinRL_Crypto, intelligent-trading-bot | ML strategies, AI research |
| page-06 | pycryptobot, binance-trader, LLMAgentCrypto | Risk management, LLM agents |

Full list in `SUBMODULES.md` and `.gitmodules`.

---

**Current Version:** 2.0.54
**Last Updated:** 2026-06-08
**License:** Apache 2.0
