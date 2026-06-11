# Handoff

## Current State: v2.1.0 — Full Strategy Arsenal & Sentiment Intelligence

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

### Next Steps
1. **Add more exchange adapters** — Coinbase, Kraken, KuCoin for cross-exchange arbitrage
2. **WebSocket feed debugging** — WS connects but goroutine output doesn't reach channel
3. **Position sizing optimization** — Kelly criterion or volatility-adjusted sizing
4. **Strategy parameter optimization** — Walk-forward on historical data
5. **React/Vite dashboard** — Replace server-rendered HTML with SPA
6. **More candle-based strategies** — MACD, ATR sizing in stream mode
7. **Trade journal analytics** — Query persisted signals for long-term performance
8. **API key integration** — Add CryptoPanic, YouTube, Whale Alert keys for full sentiment
