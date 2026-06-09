# Handoff

## Current State: v2.0.54 — Repository Synchronized & Autonomous Trading Verified

The system runs as a fully autonomous paper trader using real-time Binance.US market data. All strategy parameters are configurable via JSON — no recompilation needed.

### v2.0.53 Verification (20-min live test)
| Metric | Result |
|--------|--------|
| Executed Trades | 47 |
| Win Rate | **80%** (20W / 5L) |
| Realized PnL | ~$0 (break-even, fees on sideways market) |
| Bollinger Strategy | 87% WR, +$0.14 |
| RSI Strategy | 100% WR |
| EMA Strategy | 71% WR |
| Guard Block Rate | 39% |
| Signal Log | 102 signals persisted |

### What Changed in v2.0.53–v2.0.54

**Config-driven strategy params** — All strategy tuning is now in the JSON config file:
- `strategy.risk_pct` (default: 2%)
- `strategy.trailing_activate_pct`, `trailing_gap_pct`, `trailing_stop_loss_pct`, `trailing_max_hold_minutes`
- `strategy.bollinger_period`, `bollinger_std_dev`
- `strategy.rsi_period`, `rsi_oversold`, `rsi_overbought`
- `strategy.ema_fast`, `ema_slow`

**WebSocket market data** — `market_data.source: "websocket"` option available (experimental):
- `config/autonomous-paper.json` — REST feed, 5s polling (production)
- `config/autonomous-paper-ws.json` — WebSocket feed, 1s ticks (needs goroutine debugging)

**Signal log persistence** — Signals auto-flush to `data/signals/signals.jsonl`:
- JSONL format, one signal per line
- Auto-flush every 30 seconds, final flush on graceful shutdown

**Submodule sanitization (v2.0.54)** — Removed 6 orphaned submodule references:
- Krypto-Hashers-Community/polymarket-crypto-sports-arbitrage-trading-bot
- RobertMarcellos/polymarket-copy-trading-bot
- ericjang/cryptocurrency_arbitrage
- hello2all/gamma-ray
- fluidex/dingir-exchange
- SFCQuantX/polymarket-trading-agent

**Feature branch merge (v2.0.54)** — Assimilated `assimilate-top-crypto-bots-phase-1` branch:
- `ExecutionManager` — Modular strategy coordination
- `WolfBotBollingerStrategy` — Breakout-aware Bollinger from WolfBot
- `DoubleEMATrendStrategy` — Trend-following from freqtrade patterns
- `DynamicTrailingStop`, `ProfitBank`, `PreventLoss` — Safety strategies from pycryptobot
- `MarketMakerStrategy` — Ping-pong quoting from Krypto-trading-bot
- `LiveHistoryProvider` — High-fidelity backtesting with real market data
- CCXT-inspired unified error mapping (`internal/exchange/errors.go`)
- Expanded `Order`/`Market` structs
- Sandbox, system test, and live integration verification tests
- `LiveStrategyWrapper` for production safety checks
- API credential persistence in Account models

### Architecture
```
Binance.US → MarketDataFeed → Strategy Runtime → Risk Pipeline → Paper Execution
(REST/WS)    (5s/1s ticks)    (12+ strategies)   (8 guards)    (fee simulation)
```

### Active Strategies (per symbol: BTC, ETH, SOL)
| Strategy | Type | Signal | Source |
|----------|------|--------|--------|
| EMA Crossover | Trend | Golden/death cross | Original |
| Bollinger Reversion | Mean-reversion | Band touch | Original |
| RSI Reversion | Mean-reversion | Overbought/oversold | Original |
| Trailing Take Profit | Exit | Trail after activation | Original |
| WolfBot Bollinger | Breakout | Bollinger breakout | WolfBot |
| Double EMA Trend | Trend | Dual EMA crossover | freqtrade |
| Market Maker | Liquidity | Ping-pong quoting | Krypto-trading-bot |
| Trailing Safety | Risk | Dynamic trailing stop | pycryptobot |

### How to Run
```bash
cd ultratrader-go
# REST feed (5s polling) — PRODUCTION
go run ./cmd/ultratrader --config config/autonomous-paper.json
# WebSocket feed (1s real-time) — EXPERIMENTAL
go run ./cmd/ultratrader --config config/autonomous-paper-ws.json
```
Dashboard: http://127.0.0.1:8300/

### Data Files
| File | Description |
|------|-------------|
| `data/signals/signals.jsonl` | All strategy signals with outcomes, PnL |
| `data/orders/orders.jsonl` | All executed orders |
| `data/reports/runtime.jsonl` | Periodic metrics/valuation snapshots |
| `data/eventlog/events.jsonl` | Application lifecycle events |
| `data/logs/app.jsonl` | Structured application log |

### Key Files Modified (v2.0.52–v2.0.54)
- `internal/core/config/config.go` — StrategyConfig, MarketDataConfig structs
- `internal/core/app/app.go` — Config-driven strategy construction, WS feed, signal persistence
- `internal/strategy/demo/trailing_take_profit.go` — Functional option pattern
- `internal/strategy/signal_log.go` — JSONL persistence, auto-flush
- `internal/connectors/httpapi/server.go` — StrategyInfo, MarketDataInfo types
- `internal/connectors/httpapi/dashboard.go` — Strategy/MarketData config cards
- `internal/risk/guard.go` — OrderSide, IsExit on OrderIntent
- `internal/risk/cooldown.go` — Side-aware cooldown with IsExit bypass
- `internal/risk/duplicate_side.go` — Both-side support, IsExit bypass
- `internal/risk/duplicate_symbol.go` — Sell bypass, IsExit bypass
- `internal/strategy/scheduler/smart_dispatcher.go` — Position re-check, dust threshold
- `internal/exchange/paper/market_aware.go` — Net qty, fee dust cap
- `internal/marketdata/binance/ws_feed.go` — TLS dial, Host header fix, frame reader
- `internal/trading/execution/manager.go` — Modular strategy coordination (from feature branch)
- `internal/trading/execution/wolfbot_bollinger.go` — WolfBot Bollinger strategy (from feature branch)
- `internal/trading/execution/trailing_stop.go` — DynamicTrailingStop (from feature branch)
- `internal/trading/execution/safety.go` — ProfitBank, PreventLoss (from feature branch)
- `internal/trading/execution/live_strategy.go` — LiveStrategyWrapper (from feature branch)
- `internal/trading/execution/market.go` — MarketMaker strategy (from feature branch)

### Next Steps
1. **WebSocket feed debugging** — WS connects but goroutine output doesn't reach channel
2. **Position sizing optimization** — Kelly criterion or volatility-adjusted sizing
3. **Strategy parameter optimization** — Walk-forward on historical data
4. **Real exchange adapter** — Wire execution to real Binance spot API
5. **React/Vite dashboard** — Replace server-rendered HTML with SPA
6. **More candle-based strategies** — MACD, ATR sizing in stream mode
7. **Trade journal analytics** — Query persisted signals for long-term performance

## Completed Tasks in Go Port (Version 3.0.0)

### 1. Backtesting & Analytics
- **Multi-Symbol Synchronization:** `internal/backtest/multisymbol.go`
- **Walk-Forward Optimization:** `internal/backtest/optimizer/walkforward.go`
- **Grid Search & Monte Carlo:** `internal/backtest/optimizer/gridsearch.go`, `montecarlo.go`
- **Machine Learning Ensembles:** `internal/analytics/ml/ensemble.go`
- **Q-Learning RL Agent:** `internal/analytics/rl/qlearning.go`
- **Pattern Recognition:** `internal/analytics/patterns.go`
- **Arbitrage & Order Flow:** `internal/analytics/arbitrage.go`, `orderflow.go`

### 2. Security & Enterprise
- **Secrets Management (AES-GCM):** `internal/core/config/secrets.go`
- **Strict Input Validation:** `internal/reporting/api/validation.go`
- **API Rate Limiter (Token Bucket):** `internal/reporting/api/middleware.go`
- **SQL Injection Prevention:** `internal/persistence/db.go`
- **Client-Side Exchange Rate Limiter:** `internal/exchange/ratelimit.go`
- **Multi-Account RBAC:** `internal/enterprise/rbac.go`
- **Cryptographic Audit Logging:** `internal/enterprise/audit.go`

### 3. Assimilation Program (Feature Branch)
- **WolfBot Bollinger Strategy** — Breakout detection patterns
- **DoubleEMA Trend Strategy** — freqtrade-inspired trend following
- **Market Maker Strategy** — Krypto-trading-bot ping-pong quoting
- **Dynamic Trailing Stop** — pycryptobot safety patterns
- **ProfitBank / PreventLoss** — Multi-layered exit safety
- **CCXT Error Mapping** — Unified exchange error handling
- **Live History Provider** — Real-market backtesting data
