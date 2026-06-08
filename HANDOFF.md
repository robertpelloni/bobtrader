# Handoff Documentation

## Current State: v2.0.52 — Autonomous Paper Trading with Live Data

The system now runs as a **fully autonomous trader** using real-time Binance.US market data with simulated paper execution. Verified in 20-minute live test: 77 trades, 92.9% win rate, +$0.997 net PnL.

### Architecture Overview

```
Binance.US (REST) → MarketDataFeed → Strategy Runtime → Risk Pipeline → Paper Execution
     ↑ real prices      ↑ 5s ticks     ↑ 12 strategies    ↑ 8 guards      ↑ fee simulation
```

### Signal-to-Execution Flow (Complete)

1. **Market Data**: Binance.US REST API polls 1m kline close prices every 5s
2. **Strategy Runtime**: 12 strategies (3 entry + 1 exit per symbol) evaluate each tick
3. **Smart Dispatcher**: Position-aware signal filtering and quantity sizing
4. **Risk Pipeline**: 8 guards with sell-aware bypass logic
5. **Execution Service**: Creates order, runs risk pipeline, places via paper adapter
6. **Paper Adapter**: Fills at real market price, simulates 0.1% taker fee
7. **Portfolio Tracker**: Updates positions, tracks PnL, provides exposure data
8. **Signal Log**: Records every signal with outcome, PnL, and entry price

### Active Strategies (per symbol: BTC, ETH, SOL)

| Strategy | Type | Signal | Config |
|----------|------|--------|--------|
| EMA Crossover | Trend | Buy on golden cross, sell on death cross | EMA(9)/EMA(21) |
| Bollinger Reversion | Mean-reversion | Buy at lower band, sell at upper band | BB(20, 2σ) |
| RSI Reversion | Mean-reversion | Buy oversold, sell overbought | RSI(14, 35/65) |
| Trailing Take Profit | Exit | Trail stop after 1% profit | 1% act, 0.3% trail, 3% SL, 5min max hold |

### Risk Guards (sell-aware)

| Guard | Buy Behavior | Sell Behavior |
|-------|-------------|---------------|
| Symbol Whitelist | Block non-whitelisted | Block non-whitelisted |
| Max Notional | Block > $500 | **Bypass** (sells reduce exposure) |
| Max Per-Symbol | Block > $250 | **Bypass** |
| Cooldown | 10s per symbol | **Bypass if IsExit** |
| Duplicate Symbol | Block within 15s | **Bypass** |
| Duplicate Side | Block within 10s | **Bypass if IsExit** |
| Max Open Positions | Block > 5 | Allow |
| Max Concentration | Block > 90% | **Bypass** |

### Key Technical Decisions

- **IsExit flag on OrderIntent**: Exit signals bypass cooldown/duplicate guards — entries are throttled but exits always go through
- **Fee-corrected fills**: Paper adapter reports net quantity after 0.1% taker fee on buys; caps sell qty at held amount for fee-dust rounding
- **Position re-check**: ExecuteSignals re-checks portfolio after each execution, preventing duplicate sells
- **Neutral zone reset**: Bollinger and RSI strategies reset `lastSignal` when price returns to band middle, enabling re-entry after round-trips
- **5-minute max hold**: TrailingTakeProfit forces exit if position isn't profitable after 5 minutes
- **ExposureView with USDT balance**: Concentration calculations include cash, giving accurate portfolio-level view

### Verified Performance (20-min test)

| Metric | Value |
|--------|-------|
| Total Executed Trades | 77 |
| Win Rate | 92.9% (39W / 3L) |
| Net Realized PnL | +$0.997 |
| Bollinger WR | 93.9% |
| RSI WR | 100% |
| Guard Block Rate | 39% (preventative) |

### Files Modified This Phase

- `internal/risk/guard.go` — OrderSide, IsExit on OrderIntent
- `internal/risk/cooldown.go` — Side-aware cooldown with IsExit bypass
- `internal/risk/duplicate_side.go` — Both-side support, IsExit bypass
- `internal/risk/duplicate_symbol.go` — Sell bypass, IsExit bypass
- `internal/risk/max_notional.go` — Sell bypass
- `internal/risk/max_notional_per_symbol.go` — Sell bypass
- `internal/risk/max_concentration.go` — Sell bypass
- `internal/strategy/demo/trailing_take_profit.go` — Max hold, stop-loss, portfolio entry price
- `internal/strategy/demo/bollinger_tick_reversion.go` — Neutral zone reset
- `internal/strategy/demo/rsi_reversion.go` — Neutral zone reset, wider thresholds
- `internal/strategy/scheduler/smart_dispatcher.go` — IsExit, PnL tracking, dust check, error classification
- `internal/strategy/signal_log.go` — PnL and EntryPrice fields, win/loss tracking
- `internal/trading/portfolio/tracker.go` — AverageEntryPrice method
- `internal/trading/portfolio/exposure_view.go` — USDTBalanceReader, NewExposureViewWithBalance
- `internal/exchange/paper/market_aware.go` — Net quantity, fee dust cap
- `internal/core/app/app.go` — Strategy params, DuplicateSideGuard fix

### How to Run

```bash
cd ultratrader-go
go run ./cmd/ultratrader --config config/autonomous-paper.json
```

Dashboard: http://127.0.0.1:8300/
API: http://127.0.0.1:8300/api/portfolio

### Next Steps

1. **Binance WebSocket adapter** — Replace REST polling with real-time WebSocket streams for lower latency
2. **More sophisticated entry strategies** — Add candle-based strategies (MACD, ATR sizing) alongside tick strategies
3. **Position sizing optimization** — Kelly criterion or volatility-adjusted sizing
4. **Trade journal persistence** — Save signal log to disk for long-term analytics
5. **Strategy parameter optimization** — Walk-forward optimization on historical data
6. **Real exchange adapter** — Wire execution to real Binance spot API
7. **UI dashboard** — React/Vite dashboard consuming the API endpoints

## Completed Tasks in Go Port (Version 3.0.0)

### 1. Backtesting & Analytics
- **Multi-Symbol Synchronization:** `internal/backtest/multisymbol.go`
- **Walk-Forward Optimization:** `internal/backtest/optimizer/walkforward.go`
- **Grid Search & Monte Carlo:** `internal/backtest/optimizer/gridsearch.go`, `montecarlo.go`
- **Machine Learning Ensembles:** `internal/analytics/ml/ensemble.go`
- **Q-Learning RL Agent:** `internal/analytics/rl/qlearning.go`
- **Pattern Recognition:** `internal/analytics/patterns.go`
- **Arbitrage & Order Flow:** `internal/analytics/arbitrage.go`, `orderflow.go`

### 2. Security & Enterprise (Completed this phase)
- **Secrets Management (AES-GCM):** `internal/core/config/secrets.go`
- **Strict Input Validation:** `internal/reporting/api/validation.go`
- **API Rate Limiter (Token Bucket):** `internal/reporting/api/middleware.go`
- **SQL Injection Prevention:** `internal/persistence/db.go`
- **Client-Side Exchange Rate Limiter:** `internal/exchange/ratelimit.go`
- **Multi-Account RBAC:** `internal/enterprise/rbac.go`
- **Cryptographic Audit Logging:** `internal/enterprise/audit.go`
