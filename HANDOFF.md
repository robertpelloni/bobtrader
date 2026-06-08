# Handoff Documentation

## Current State: v2.0.53 — Config-Driven Autonomous Trading

The system runs as a fully autonomous paper trader using real-time Binance.US market data. All strategy parameters are now configurable via JSON — no recompilation needed.

### v2.0.53 Verification (20-min live test)

| Metric | Result |
|--------|--------|
| Executed Trades | 46 |
| Win Rate | **96.0%** (24W / 1L) |
| Realized PnL | **+$1.28** |
| Bollinger Strategy | 100% WR, +$0.94 |
| RSI Strategy | 100% WR, +$0.31 |
| EMA Strategy | 67% WR, +$0.03 |
| Guard Block Rate | 27% |

### What Changed in v2.0.53

**Config-driven strategy params** — All strategy tuning is now in the JSON config file:
- `strategy.risk_pct` (default: 2%)
- `strategy.trailing_activate_pct` (default: 1%)
- `strategy.trailing_gap_pct` (default: 0.3%)
- `strategy.trailing_stop_loss_pct` (default: 3%)
- `strategy.trailing_max_hold_minutes` (default: 5)
- `strategy.bollinger_period`, `bollinger_std_dev`
- `strategy.rsi_period`, `rsi_oversold`, `rsi_overbought`
- `strategy.ema_fast`, `ema_slow`

**WebSocket market data** — `market_data.source: "websocket"` uses Binance WS streams:
- `config/autonomous-paper.json` — REST feed, 5s polling
- `config/autonomous-paper-ws.json` — WebSocket feed, 1s ticks

**Signal log persistence** — Signals auto-flush to `data/signals/signals.jsonl`:
- JSONL format, one signal per line
- Auto-flush every 30 seconds
- Final flush on graceful shutdown

**TrailingTakeProfit refactored** — Functional option pattern:
- `WithStopLossPct(pct)`, `WithMaxHoldMinutes(min)`
- `WithPortfolioEntry(reader)`, `WithFeed(feed)`
- All params from `cfg.Strategy`

**API & Dashboard** — `/api/config` now includes `strategy` and `market_data` sections.
Dashboard config page shows Strategy and Market Data cards.

### Architecture

```
Binance.US → MarketDataFeed → Strategy Runtime → Risk Pipeline → Paper Execution
  (REST/WS)     (5s/1s ticks)   (12 strategies)    (8 guards)    (fee simulation)
```

### Active Strategies (per symbol: BTC, ETH, SOL)

| Strategy | Type | Signal | Config Source |
|----------|------|--------|---------------|
| EMA Crossover | Trend | Golden/death cross | `strategy.ema_fast`, `ema_slow` |
| Bollinger Reversion | Mean-reversion | Band touch | `strategy.bollinger_period`, `bollinger_std_dev` |
| RSI Reversion | Mean-reversion | Overbought/oversold | `strategy.rsi_period`, `rsi_oversold`, `rsi_overbought` |
| Trailing Take Profit | Exit | Trail after activation | `strategy.trailing_*` |

### Config File Structure

```json
{
  "environment": "autonomous-paper-trading",
  "server": { "enabled": true, "address": "0.0.0.0:8300" },
  "scheduler": { "enabled": true, "mode": "stream", "interval_ms": 5000 },
  "risk": {
    "max_notional": 500,
    "max_notional_per_symbol": 250,
    "allowed_symbols": ["BTCUSDT", "ETHUSDT", "SOLUSDT"],
    "cooldown_ms": 10000,
    "duplicate_window_ms": 15000,
    "duplicate_side_window_ms": 10000,
    "max_open_positions": 5,
    "max_concentration_pct": 90
  },
  "strategy": {
    "risk_pct": 2.0,
    "max_notional": 500,
    "trailing_activate_pct": 1.0,
    "trailing_gap_pct": 0.3,
    "trailing_stop_loss_pct": 3.0,
    "trailing_max_hold_minutes": 5,
    "bollinger_period": 20,
    "bollinger_std_dev": 2.0,
    "rsi_period": 14,
    "rsi_oversold": 35,
    "rsi_overbought": 65,
    "ema_fast": 9,
    "ema_slow": 21
  },
  "market_data": {
    "source": "rest",
    "initial_balance": 10000
  },
  "accounts": [{
    "id": "paper-main",
    "name": "Paper Trading (Live Binance Data)",
    "enabled": true,
    "exchange": "paper-market-aware",
    "capabilities": ["spot", "paper", "candles", "balances", "orders"]
  }]
}
```

### How to Run

```bash
cd ultratrader-go

# REST feed (5s polling)
go run ./cmd/ultratrader --config config/autonomous-paper.json

# WebSocket feed (1s real-time)
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

### Key Files Modified (v2.0.52–v2.0.53)

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

### Next Steps

1. **WebSocket feed debugging** — WS connects but goroutine output doesn't reach channel; needs deeper goroutine/concurrency investigation
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
