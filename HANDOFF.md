# Handoff

## System Evolution: bobtrader → fully_automated_gay_luxuxy_communism

> **CRITICAL PARADIGM SHIFT:** This repository is now absorbed into the
> `fully_automated_gay_luxuxy_communism` autonomous system. Both the Python
> legacy stack AND the Go ultra-project run simultaneously with live Binance
> capital, competing head-to-head on profitability.
>
> See [`AUTONOMOUS_DUAL_BOT_STRATEGY.md`](AUTONOMOUS_DUAL_BOT_STRATEGY.md) for the
> complete specification.

## Current State: v3.3.0-Alpha — Liquidity Execution & HFT Core

### ⚠️ Autonomous Dual-Bot Operation Active

| Component | Allocation | Status |
|-----------|-----------|--------|
| **Python Bot** | 25% of portfolio | Legacy PowerTrader AI |
| **Go Bot** | 25% of portfolio | UltraTrader Go (14 strategies) |
| **Reserve** | 50% of portfolio | Held as USDT |

**Supervisor:** An AI agent monitors both bots, modifies source, rebuilds,
restarts, and rebalances capital based on comparative performance. See
`AUTONOMOUS_DUAL_BOT_STRATEGY.md` for full protocol.

The system runs as a fully autonomous paper trader using real-time Binance.US market data with **14 strategies** across **9 trading pairs**. All strategy parameters are configurable via JSON — no recompilation needed.

### v2.1.0 — Major Feature Release

**14 Active Strategies per Symbol:**
| # | Strategy | Type | Source |
|---|----------|------|--------|
| 1 | EMA Crossover (5/13) | Technical | Original |
| 2 | Bollinger Tick (15, 1.5σ) | Technical | Original |
| 3 | RSI Reversion (10) | Technical | Original |
| 4 | Trailing Take Profit | Exit | Original |
| 5 | Tick Momentum Burst | Technical | Original |
| 6 | Tick Mean Reversion | Technical | Original |
| 7 | Double EMA Trend | Technical | freqtrade |
| 8 | Tick Price Threshold | Technical | Original |
| 9 | Sentiment-Aware | Sentiment | NEW |
| 10 | USDT Stablecoin Scalp | Stablecoin | NEW |
| 11 | USDC Stablecoin Scalp | Stablecoin | NEW |
| 12 | Weekly Cycle | Time-based | NEW |
| 13 | China Session | Time-based | NEW |
| 14 | Whale Alert | On-chain | NEW |

**9 Trading Pairs:**
BTC, ETH, SOL, XLM, ADA, DOGE, XRP, USDT, USDC

**Sentiment Intelligence Layer:**
- Fear & Greed Index (live, free API)
- Market Events (BTC halving, FOMC, ETF decisions, tax season)
- CryptoPanic News (needs free API key)
- YouTube Sentiment (monitors 8 channels: Arcane Bear, Benjamin Cowen, Coin Bureau, etc.)
- Stock Market Correlation (SPY as risk indicator)
- Whale Alert (tracks large exchange inflows/outflows)

**Time-Based Strategies:**
- Weekly Cycle: Buy Monday dip, sell Sunday peak
- China Session: Buy pre-Asia 00:00-01:00 UTC, sell Asia spike 01:30-03:00 UTC

**Stablecoin Strategies:**
- USDT Scalp: Buy 0.9992, sell 0.9999, stop 0.98
- USDC Scalp: Buy 0.9985, sell 0.9998, stop 0.97

### Config Files
| Config | Purpose |
|--------|---------|
| `config/paper-live-data.json` | Conservative: real prices, paper execution |
| `config/paper-aggressive.json` | Aggressive: 5% risk, 5s cooldown, tight Bollinger |
| `config/paper-all-strategies.json` | Full arsenal: 14 strategies, 9 symbols |
| `config/autonomous-paper.json` | Original autonomous paper trading |

### How to Run
```bash
cd ultratrader-go

# Full strategy arsenal (recommended)
go run -buildvcs=false ./cmd/ultratrader --config config/paper-all-strategies.json

# Conservative paper trading
go run -buildvcs=false ./cmd/ultratrader --config config/paper-live-data.json

# Aggressive mode
go run -buildvcs=false ./cmd/ultratrader --config config/paper-aggressive.json
```

Dashboard: http://127.0.0.1:8300/

### Key Files Modified (v2.1.0)
- `internal/strategy/demo/sentiment_aware.go` — Sentiment-aware strategy
- `internal/strategy/demo/usdt_stablecoin_scalp.go` — USDT/USDC stablecoin scalping
- `internal/strategy/demo/weekly_cycle.go` — Weekly cycle (Sunday peak pattern)
- `internal/strategy/demo/china_session.go` — China session (1AM UTC volatility)
- `internal/strategy/demo/whale_alert_strategy.go` — Whale alert trading
- `internal/strategy/demo/cross_exchange_arbitrage.go` — Cross-exchange arbitrage
- `internal/strategy/demo/tick_momentum_burst.go` — Momentum burst signals
- `internal/strategy/demo/tick_mean_reversion.go` — Mean reversion signals
- `internal/analytics/sentiment/providers.go` — Fear/Greed, Market Events, Stock Correlation
- `internal/analytics/sentiment/youtube.go` — YouTube channel sentiment analysis
- `internal/analytics/sentiment/whale_alert.go` — Whale Alert API integration
- `internal/core/app/app.go` — All strategies wired into runtime
- `config/paper-all-strategies.json` — 9 symbols, 14 strategies

### Data Files
| File | Description |
|------|-------------|
| `data/signals/signals.jsonl` | All strategy signals with outcomes, PnL |
| `data/orders/orders.jsonl` | All executed orders |
| `data/reports/runtime.jsonl` | Periodic metrics/valuation snapshots |
| `data/eventlog/events.jsonl` | Application lifecycle events |
| `data/logs/app.jsonl` | Structured application log |

### API Keys (Optional — unlocks full functionality)
| Service | Key | Free Tier |
|---------|-----|-----------|
| CryptoPanic | https://cryptopanic.com/developers/api/ | Yes |
| YouTube Data | https://console.cloud.google.com | Yes |
| Alpha Vantage | https://www.alphavantage.co/support/#api-key | Yes |
| Whale Alert | https://whale-alert.io/ | 10 req/min |

### v3.3.0-Alpha — Liquidity Execution & HFT Core
This release focuses on institutional-grade execution and cross-exchange arbitrage.

**New Execution Core:**
- **VWAP Execution:** Implemented Volume Weighted Average Price order slicing in `internal/trading/execution/vwap.go`.
- **Atomic Arbitrage Leg:** Concurrent execution of multi-venue trades with `ArbitrageExecutorV2`.
- **Order Book Depth Visualization:** Interactive depth charts in the React dashboard.

**Integration Success:**
- **Go Backend serves React SPA:** The backend now serving the production React build from `web/dist`.
- **System Stability:** Verified with 108s of system simulation and 100% success rate on 26 signals during stress tests.

### Config Files
| Config | Purpose |
|--------|---------|
| `config/paper-live-data.json` | Conservative: real prices, paper execution |
| `config/paper-aggressive.json` | Aggressive: 5% risk, 5s cooldown, tight Bollinger |
| `config/paper-all-strategies.json` | Full arsenal: 14 strategies, 9 symbols |
| `config/autonomous-paper.json` | Original autonomous paper trading |

### Next Steps
1. **Multi-Hop Chain Arbitrage** — Expand atomic executor to support 3-leg triangular routes across different venues.
2. **Low-Latency Orderflow Scalper** — Implement L2-based scalping using order book imbalance.
3. **Live HFT Benchmarking** — Test VWAP slippage against benchmarks on live accounts.
4. **Real-time Depth Streaming** — Transition `/api/marketdata/depth` from mock to live Binance depth feed.
5. **Strategy Portfolio Rebalancing** — Automatically shift capital to strategies with highest rolling Sharpe ratio.
