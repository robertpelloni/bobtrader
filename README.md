# BobTrader / UltraTrader

**Autonomous cryptocurrency trading platform — Go runtime with real market data, paper trading, and 8 built-in strategies.**

> ⚠️ **This software can place real trades automatically.** You are responsible for all financial outcomes. Keep API keys private. This is not financial advice. See the [Apache 2.0 license](LICENSE) for full disclaimers.

---

## What Is This?

BobTrader (UltraTrader) is a modular, daemon-grade cryptocurrency trading platform written in Go. It connects to live Binance market data, runs configurable trading strategies through a multi-layered risk pipeline, and executes trades via a paper-trading engine with realistic fee simulation. The system is designed to be observable, safe, and extensible — a foundation you can build your own trading bot on top of.

The project also preserves a legacy Python trading system (`pt_*.py` files, ~25K lines) for reference, and maintains a research corpus of 42 open-source trading bot submodules for architecture and feature inspiration.

### Key Facts

| | |
|---|---|
| **Language** | Go 1.24.3 (primary), Python (legacy, frozen) |
| **Source Files** | 229 Go files, ~25,700 lines of Go code |
| **Market Data** | Binance.US REST (production), WebSocket (experimental) |
| **Execution** | Paper trading with 0.1% taker fee simulation |
| **Strategies** | 8 built-in (Bollinger, RSI, EMA, Trailing TP, WolfBot, DoubleEMA, MarketMaker, Safety) |
| **Risk Guards** | 8 guards (whitelist, notional, concentration, cooldown, duplicate, max-positions) |
| **API** | 20+ REST endpoints for portfolio, orders, diagnostics, config |
| **Config** | JSON-driven — no recompilation needed to tune strategies |
| **License** | Apache 2.0 |

---

## Architecture

```
                    ┌─────────────────────────────────────────────────┐
                    │              UltraTrader Go Runtime             │
                    │                                                 │
  Binance.US ──────▶  Market Data Feed (REST 5s / WS 1s)            │
  Real Prices       │         │                                       │
                    │         ▼                                       │
                    │  Strategy Runtime                              │
                    │  ┌─────────────────────────────────┐          │
                    │  │ Bollinger │ RSI │ EMA │ Trailing │          │
                    │  │ WolfBot   │ DEMA│ MktMaker│Safety│          │
                    │  └─────────────┬───────────────────┘          │
                    │                │ Signals                        │
                    │                ▼                                │
                    │  Risk Pipeline (8 Guards)                      │
                    │  ┌──────────────────────────────────┐         │
                    │  │ Whitelist │ MaxNotional │ Cooldown│         │
                    │  │ MaxConc   │ Duplicate   │ MaxPos  │         │
                    │  └─────────────┬────────────────────┘         │
                    │                │ Approved Orders               │
                    │                ▼                                │
                    │  Paper Execution Engine                       │
                    │  (Market-Aware, 0.1% Taker Fee)               │
                    │                │                                │
                    │                ▼                                │
                    │  Portfolio Tracker │ Signal Log │ Order Journal│
                    └─────────────────────────────────────────────────┘
                                     │
                                     ▼
                            HTTP API + Dashboard
                       (Portfolio, Guards, Metrics, Config)
```

---

## Quick Start

### Prerequisites

- [Go 1.24+](https://go.dev/dl/)
- Internet access (for Binance.US market data)

### Build & Run

```bash
# Clone
git clone https://github.com/robertpelloni/bobtrader.git
cd bobtrader/ultratrader-go

# Build
go build -o ultratrader.exe ./cmd/ultratrader

# Run autonomous paper trading (REST feed, $10K starting balance)
./ultratrader.exe --config config/autonomous-paper.json
```

Open the dashboard: **http://127.0.0.1:8300/**

The system immediately begins pulling live BTC/ETH/SOL prices from Binance.US, generating strategy signals, and executing simulated trades.

### Run with Docker

```bash
cd ultratrader-go
docker build -t ultratrader-go .
docker run --rm -p 8300:8300 ultratrader-go
```

---

## Config Profiles

| Config File | Purpose | Market Data | Balance |
|---|---|---|---|
| `autonomous-paper.json` | **Production paper trading** | Binance.US REST 5s | $10,000 |
| `autonomous-paper-ws.json` | WebSocket paper trading (experimental) | Binance.US WS 1s | $10,000 |
| `development-timer.json` | Development with timer scheduler | Paper mock | $10,000 |
| `development-stream.json` | Development with stream scheduler | Paper mock | $10,000 |
| `paper-service.json` | Headless paper service | Paper mock | $10,000 |
| `live-trading-binance.json` | Live trading template (requires API keys) | Binance REST | — |

All strategy parameters are configurable via JSON — **no recompilation needed**:

```json
{
  "strategy": {
    "risk_pct": 2.0,
    "trailing_activate_pct": 1.0,
    "trailing_gap_pct": 0.3,
    "trailing_stop_loss_pct": 3.0,
    "bollinger_period": 20,
    "bollinger_std_dev": 2.0,
    "rsi_oversold": 35,
    "rsi_overbought": 65,
    "ema_fast": 9,
    "ema_slow": 21
  }
}
```

---

## Strategies

| Strategy | Type | Signal | Inspiration |
|---|---|---|---|
| **Bollinger Tick Reversion** | Mean-reversion | Buy at lower band, sell at upper band | Original |
| **RSI Reversion** | Mean-reversion | Buy oversold (<35), sell overbought (>65) | Original |
| **EMA Tick Crossover** | Trend-following | Golden/death cross on 9/21 EMA | Original |
| **Trailing Take Profit** | Exit | Trailing stop after 1% activation | Original |
| **WolfBot Bollinger** | Breakout | Bollinger breakout detection | Ekliptor/WolfBot |
| **Double EMA Trend** | Trend-following | Dual EMA with long-period filter | freqtrade |
| **Market Maker** | Liquidity | Ping-pong quoting | Krypto-trading-bot |
| **Dynamic Trailing Stop** | Risk | High-price tracking stop | pycryptobot |
| **ProfitBank / PreventLoss** | Safety | Multi-layered exit triggers | pycryptobot |

Each strategy runs per-symbol. The default config runs 4 strategies × 3 symbols = 12 concurrent strategy instances.

---

## Risk Pipeline

Every signal passes through 8 risk guards before execution:

| Guard | Purpose |
|---|---|
| **Symbol Whitelist** | Only trade approved symbols |
| **Max Notional** | Cap total order value |
| **Max Notional Per Symbol** | Cap per-symbol exposure |
| **Max Concentration** | Prevent portfolio over-concentration (uses live market values) |
| **Cooldown** | Side-aware rate limiting (10s buy, 5s sell) |
| **Duplicate Symbol** | Prevent rapid re-entry on same symbol |
| **Duplicate Side** | Prevent rapid repeated buy or sell |
| **Max Open Positions** | Cap total concurrent positions |

**Sell orders bypass most guards** — exits reduce exposure and should never be blocked by notional/concentration limits.

---

## API Endpoints

| Endpoint | Description |
|---|---|
| `GET /healthz` | Health check |
| `GET /readyz` | Readiness check |
| `GET /api/status` | Runtime status |
| `GET /api/portfolio` | Full portfolio with positions & PnL |
| `GET /api/portfolio-summary` | Compact portfolio summary |
| `GET /api/orders` | Order history |
| `GET /api/execution-summary` | Execution statistics |
| `GET /api/execution-diagnostics` | Detailed execution diagnostics |
| `GET /api/exposure-diagnostics` | Exposure/concentration details |
| `GET /api/metrics` | Runtime metrics |
| `GET /api/guards` | Guard names |
| `GET /api/guard-diagnostics` | Guard block reasons & counts |
| `GET /api/config` | Full config including strategy params |
| `GET /api/signals` | Recent signal log |
| `GET /api/strategy-stats` | Per-strategy win rate, PnL, signal count |
| `GET /api/runtime-reports/latest` | Latest runtime report |
| `GET /api/runtime-reports/history` | Historical reports |
| `GET /api/runtime-reports/trends` | Trend analysis (variances) |
| `GET /dashboard` | Web dashboard UI |

---

## Verified Performance

20-minute autonomous paper trading test with real Binance.US market data (BTC/ETH/SOL, $10K USDT):

| Metric | Result |
|---|---|
| Executed Trades | 47 |
| Win Rate | **80%** (20W / 5L) |
| Bollinger Strategy | 87% WR, +$0.14 PnL |
| RSI Strategy | 100% WR |
| EMA Strategy | 71% WR |
| Guard Block Rate | 39% (prevents over-trading) |

---

## Project Structure

```
bobtrader/
├── ultratrader-go/                  # Go ultra-project (primary)
│   ├── cmd/ultratrader/             # Application entrypoint
│   ├── config/                      # JSON config profiles
│   ├── internal/
│   │   ├── core/                    # App composition, config, logging, event log
│   │   ├── marketdata/              # Market data feeds (Binance REST/WS, paper)
│   │   ├── strategy/                # Trading strategies + scheduler + signal log
│   │   │   ├── demo/                # 16 strategy implementations
│   │   │   ├── scheduler/           # Enhanced scheduler + smart dispatcher
│   │   │   ├── sizing/              # Position sizing (PortfolioSizer)
│   │   │   ├── marketmaking/        # Market maker strategy
│   │   │   ├── composite/           # Signal composition/voting
│   │   │   ├── regime/              # Market regime detection
│   │   │   └── nlp/                 # NLP strategy parsing
│   │   ├── risk/                    # 8 risk guards + circuit breaker
│   │   ├── trading/                 # Execution, portfolio, reconciliation
│   │   │   ├── execution/           # ExecutionManager, WolfBot, Safety, MarketMaker
│   │   │   ├── portfolio/           # Position tracker, exposure view
│   │   │   ├── orders/              # Order journal
│   │   │   ├── account/             # Account service
│   │   │   ├── reconciliation/      # Order reconciliation
│   │   │   └── rebalancer/          # Portfolio rebalancer
│   │   ├── exchange/                # Exchange registry + adapters
│   │   │   ├── binance/             # Binance REST adapter
│   │   │   ├── paper/               # Market-aware paper exchange
│   │   │   ├── aggregator/          # Multi-exchange price aggregation
│   │   │   └── ratelimit/           # Token bucket rate limiter
│   │   ├── analytics/               # Journal, correlation, ML, RL, sentiment
│   │   ├── backtest/                # Engine + optimizer (walk-forward, grid, Monte Carlo)
│   │   ├── connectors/httpapi/      # HTTP API + dashboard
│   │   ├── notification/            # Discord, Telegram, Email
│   │   ├── enterprise/              # RBAC, audit logging
│   │   ├── persistence/             # Orders, snapshots, reports (JSONL)
│   │   ├── indicator/               # Technical indicators (MACD, Bollinger, ATR)
│   │   └── metrics/                 # Runtime metrics tracker
│   ├── Dockerfile
│   └── docker-compose.yml
│
├── pt_*.py                          # Legacy Python system (frozen, ~25K lines)
├── submodules/                      # 42 reference trading bot repos
├── docs/                            # Architecture analysis, assimilation docs
├── HANDOFF.md                       # Session-to-session handoff log
├── ROADMAP.md                       # Development roadmap
├── TODO.md                          # Task tracking
├── DEPLOY.md                        # Deployment instructions
├── VERSION.md                       # Current version (2.0.54)
└── CHANGELOG.md                     # Version history
```

---

## Legacy Python System

The original PowerTrader AI is preserved in the root `pt_*.py` files. It is a Tkinter-based desktop app with:

- **kNN price prediction AI** — instance-based predictor with online reliability weighting
- **Structured DCA system** — tiered dollar-cost-averaging with neural level triggers
- **Robinhood Crypto execution** — spot trading via Robinhood API
- **Multi-exchange price aggregation** — KuCoin + Binance + Coinbase with median/VWAP
- **Notifications** — Email, Discord, Telegram
- **25,300 lines** across 36 Python modules

The Python system is **frozen** (no longer actively developed). All new development targets the Go runtime.

---

## Research Corpus

The `submodules/` directory contains 42 open-source trading bot repositories for architecture and feature reference:

| Category | Notable Projects |
|---|---|
| Architecture | TraderAlice/OpenAlice |
| Go Kernel | c9s/bbgo |
| Exchange Abstraction | ccxt/ccxt |
| Advanced Features | Ekliptor/WolfBot |
| Market Making | ctubio/Krypto-trading-bot |
| ML Strategies | AI4Finance-Foundation/FinRL_Crypto |
| Risk Management | whittlem/pycryptobot |

The Go project uses **clean-room reimplementation** — studying architecture and behavior without direct source reuse, to avoid licensing conflicts.

---

## Testing

```bash
cd ultratrader-go

# Run core test suite
go test ./internal/core/... ./internal/strategy/... ./internal/risk/... \
       ./internal/trading/... ./internal/marketdata/... ./internal/exchange/paper/...

# Run all tests
go test ./...

# Build binary
go build -o ultratrader.exe ./cmd/ultratrader
```

---

## Configuration Reference

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
  "market_data": { "source": "rest", "initial_balance": 10000 },
  "accounts": [{
    "id": "paper-main",
    "name": "Paper Trading (Live Binance Data)",
    "enabled": true,
    "exchange": "paper-market-aware"
  }]
}
```

---

## Documentation

| File | Description |
|---|---|
| [MANUAL.md](MANUAL.md) | Legacy Python user manual |
| [DEPLOY.md](DEPLOY.md) | Deployment instructions (Go + Python) |
| [ROADMAP.md](ROADMAP.md) | Development roadmap |
| [TODO.md](TODO.md) | Task tracking |
| [HANDOFF.md](HANDOFF.md) | Session-to-session development log |
| [CHANGELOG.md](CHANGELOG.md) | Version history |
| [MODULE_INDEX.md](MODULE_INDEX.md) | Python→Go module mapping |
| [MCP_SERVERS_RESEARCH.md](MCP_SERVERS_RESEARCH.md) | 25+ MCP server research |

---

## Disclaimer

This software is for **educational and experimental purposes**. It can place real trades automatically. You are responsible for everything it does to your money and your account. Keep your API keys private. This is not financial advice. The authors are not responsible for any losses incurred or security breaches.

---

## License

Apache License 2.0 — see [LICENSE](LICENSE) for details.
