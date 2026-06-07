# Bobtrader / PowerTrader AI — Complete Repository Analysis

> Generated: 2026-06-06 | Version: 2.0.48
> **25,301 lines Python** · **20,565 lines Go** · **91 Go test files** · **47+ external submodules** · **16 API endpoints**

---

## 1. Repository Overview

This is a **dual-track trading systems workspace** with two simultaneous goals:

| Track | Language | Lines | Role |
|-------|----------|-------|------|
| **PowerTrader AI** | Python | 25,301 | Production crypto trading bot with kNN AI |
| **Ultra-Project** | Go | 20,565 | Next-gen modular trading platform (clean-room port) |
| **Submodules** | Various | N/A | 47+ external repos for architecture research |

### Core Philosophy
- **Spot-only, long-term DCA trading** — no stop-losses, no futures, no leverage
- **"No loss selling"** — HODL through downturns, buy more on dips
- **kNN-based price prediction** — instance-based learning with per-pattern reliability weighting
- **Clean-room Go port** — assimilating best architecture ideas without copying source

---

## 2. Python System: PowerTrader AI (v2.0.0)

### 2.1 Architecture

```
┌──────────────────────────────────────────────────────────────────────┐
│                     pt_hub.py (5,835 lines)                          │
│               Main GUI / Orchestrator (Tkinter)                      │
│                                                                      │
│  ┌──────────┬───────────┬────────────┬──────────┬──────────┐        │
│  │Dashboard │ Charts    │ Analytics  │ Volume   │ Risk     │        │
│  │Tab       │ Tab       │ Tab        │ Tab      │ Tab      │        │
│  └────┬─────┴─────┬─────┴─────┬──────┴────┬─────┴────┬─────┘        │
│       │           │           │           │          │               │
│  ┌────┴────┐ ┌────┴────┐ ┌───┴────┐ ┌────┴────┐ ┌───┴──────┐       │
│  │Thinker  │ │Trader   │ │Analytics│ │Volume   │ │Risk Mgmt │       │
│  │1,381 L  │ │2,309 L  │ │770 L   │ │1,026 L  │ │146 L     │       │
│  └────┬────┘ └────┬────┘ └───┬────┘ └────┬────┘ └───┬──────┘       │
│       │           │          │           │          │               │
│  ┌────┴────┐      │    ┌─────┴─────┐     │    ┌─────┴──────┐       │
│  │Trainer  │      │    │Dashboards │     │    │Correlation │       │
│  │1,625 L  │      │    │+ KPI Cards│     │    │+ Sizing    │       │
│  └─────────┘      │    └───────────┘     │    └────────────┘       │
│                    │                      │                          │
│  ┌─────────────────┴──────────────────────┴──────────────────┐      │
│  │            pt_exchanges.py (Multi-Exchange Layer)         │      │
│  │   KuCoin ──→ Binance ──→ Coinbase ──→ Robinhood (exec)    │      │
│  └──────────────────────────────────────────────────────────┘      │
└──────────────────────────────────────────────────────────────────────┘
```

### 2.2 Core Trading Pipeline — How It All Works

#### Phase 1: Training (`pt_trainer.py` — 1,625 lines)

The AI "memorizes" historical market patterns:

1. **Loads historical OHLCV data** from KuCoin API for each configured coin
2. **Processes 7 timeframes**: 1h, 2h, 4h, 6h, 12h, 1d, 1w
3. **Extracts pattern windows** — sliding windows of N consecutive candles
4. **Stores each pattern** as: feature vector + outcome (what the next candle did)
5. **Saves to JSON memory files** per coin per timeframe (e.g., `BTC/1h_memory.json`)
6. **Training is required once** before the bot can run — click "Train All" in the Hub

#### Phase 2: Live Prediction (`pt_thinker.py` — 1,381 lines)

At runtime, the Thinker continuously predicts where prices will go:

1. **Matches current market state** against stored patterns using k-Nearest Neighbors
2. **Produces a weighted average** of the closest matches, weighted by past reliability
3. **Outputs 7 LONG levels** (predicted low prices) + **7 SHORT levels** (predicted high prices)
4. **Levels = signal strengths** (1-7), one per timeframe from 1hr to 1wk
5. **Online learning**: After each candle closes, compares prediction vs reality and adjusts pattern weights — patterns that were accurate get higher weight, inaccurate ones get demoted
6. **Integrates optional layers**:
   - `pt_sentiment.py` → news/social sentiment scores via MCP
   - `pt_ml_ensemble.py` → XGBoost/Random Forest model blending
   - `pt_regime_detection.py` → trending/ranging/volatile/quiet classification

**Key insight**: This is NOT a neural network or LLM. It's instance-based learning — the system literally remembers every pattern it's seen and votes on what will happen next based on similarity to the current state.

#### Phase 3: Trade Execution (`pt_trader.py` — 2,309 lines)

The Trader makes buy/sell decisions based on the Thinker's signals:

**Entry Rule**: `LONG ≥ 3 AND SHORT == 0` → Start trade
- LONG level 3 means "at least 3 of the 7 timeframe predictions show the price dropping below their predicted lows"
- SHORT == 0 means "no timeframe predicts the price will go above its predicted high"
- This combination signals: "price is stretched to the downside across multiple timeframes"

**DCA (Dollar Cost Averaging)** — When price drops after entry:
- **Neural level trigger**: Buy more when price hits the next predicted low level
- **OR hardcoded trigger**: Buy at -2.5%, -5%, -10%, -20%, -30%, -40%, -50% drawdown
- **Whichever comes first** fires the DCA
- **Safety limit**: Max 2 DCA buys per 24h rolling window (prevents dumping money into freefall)
- **DCA multiplier**: Each DCA buy is 2× larger than the previous position size
- **Start allocation**: 0.5% of account value for the initial buy

**Exit Rule**: Trailing profit margin
- **Activation**: When profit > 5% (no DCA) or > 2.5% (after DCA occurred)
- **Trailing**: Once activated, the sell line follows price upward, staying 0.5% behind the peak
- **Execution**: Sells immediately when price drops below the trailing line
- **Purpose**: Captures as much upside as possible while locking in gains

**Risk controls built into the Trader**:
- `pt_risk_management.py` → Position limits, portfolio-level risk checks
- `pt_position_sizing.py` → ATR-based volatility-adjusted position sizing
- `pt_rebalancer.py` → Drift detection from target allocations
- Hot-reloadable config via `pt_config.py` (trading params update without restart)

#### Phase 4: Exchange Execution (`pt_exchanges.py` — 1,006 lines)

| Exchange | Role | API |
|----------|------|-----|
| KuCoin | Primary price data source | REST API |
| Binance | Price data fallback #1 | REST API |
| Coinbase | Price data fallback #2 | REST API |
| Robinhood | **Trading execution** | Crypto API (robin_stocks) |

- **Price aggregation**: Median/VWAP across all available exchanges
- **Fallback chain**: KuCoin → Binance → Coinbase (automatic on failure)
- **Arbitrage monitoring**: Detects price spreads between exchanges
- **Verification**: Cross-exchange price check before placing any trade

#### Phase 5: Analytics & Journaling

| Module | Function | Storage |
|--------|----------|---------|
| `pt_analytics.py` (770 L) | Trade journal, performance tracking | SQLite |
| `pt_analytics_dashboard.py` (262 L) | KPI cards + period comparison | GUI widgets |
| `pt_advanced_analytics.py` | Extended analytics | SQLite |
| `pt_volume.py` (1,026 L) | Volume analysis, anomaly detection | In-memory |
| `pt_correlation.py` (447 L) | Cross-asset correlation matrix | SQLite |

**TradeJournal** tracks:
- Every entry, DCA, and exit with trade group IDs linking them together
- Win rate, total PnL, profit factor, Sharpe ratio, max drawdown
- Period comparisons: today, 7 days, 30 days, all time

#### Phase 6: Notifications (`pt_notifications.py` — 1,180 lines)

| Platform | Method | Rate Limit |
|----------|--------|------------|
| Email (Gmail) | SMTP via yagmail | 5/min |
| Discord | Webhook with embed colors | 10/min |
| Telegram | Bot API with MarkdownV2 | 10/min |

- **Async non-blocking** — notifications never block the trading loop
- **4 severity levels**: INFO (blue), WARNING (orange), ERROR (red), CRITICAL (dark red)
- **Per-level routing**: e.g., only send CRITICAL to email, everything to Discord
- **SQLite audit trail**: every notification attempt logged with success/failure
- **Convenience methods**: `NotifyTrade()` auto-detects loss vs profit

### 2.3 GUI System — The Hub (`pt_hub.py` — 5,835 lines)

The single-window Tkinter application that orchestrates everything:

| Panel | What It Shows |
|-------|---------------|
| **Left Panel** | Start/Stop buttons for AI + Trader, account status ($$$), Neural Level bars (0-7 per coin), live console logs |
| **Charts Tab** | Real-time candlestick chart with: blue LONG prediction lines, orange SHORT lines, green trailing profit margin, red DCA trigger, yellow cost basis, colored trade dots (red=buy, purple=DCA, green=sell) |
| **Analytics Tab** | KPI Cards (win rate, PnL, profit factor, Sharpe), performance tables across time periods |
| **Volume Tab** | Volume profile stats, volume ratio vs average, trend direction, Z-score anomalies |
| **Risk Tab** | Correlation matrix heatmap, portfolio diversification score, ATR-based position sizing recommendations |
| **Alerts Tab** | Custom alert rule builder |
| **Heatmap Tab** | Visual correlation heatmap |
| **Performance Tab** | Performance attribution analysis |
| **Replay Tab** | Historical trade replay tool |
| **Settings** | Tabbed config: trading params, notifications, exchanges, analytics, risk management |

### 2.4 Configuration System (`pt_config.py` — 628 lines)

- **YAML-based** (`config.yaml`) with automatic migration from legacy `gui_settings.json`
- **Environment variable overrides** with `POWERTRADER_` prefix
- **Hot-reload** via file watcher — trading params update without restart
- **Dataclasses**: TradingConfig, NotificationConfig, ExchangeConfig, AnalyticsConfig, SystemConfig
- **Singleton ConfigManager** with `get_config()` global access
- **Callback system** for GUI integration when config changes

### 2.5 Additional Python Modules

| Module | Lines | What It Does |
|--------|-------|-------------|
| `pt_config.py` | 628 | Unified YAML config with hot-reload + env var overrides |
| `pt_logging.py` | 538 | Structured JSON logging with rotation + color console |
| `pt_enterprise.py` | 370 | RBAC (role-based access control) + audit logging |
| `pt_defi.py` | 340 | DeFi integration stubs (future PancakeSwap, etc.) |
| `pt_ml_ensemble.py` | 431 | XGBoost/Random Forest ensemble model blending |
| `pt_model_registry.py` | 368 | ML model version tracking + comparison |
| `pt_nlp_strategy.py` | 341 | Regex-based natural language strategy parsing |
| `pt_rl_optimizer.py` | 450 | Q-learning RL agent for parameter optimization |
| `pt_marketplace.py` | — | Strategy marketplace catalog |
| `pt_web_dashboard.py` | 413 | Flask/FastAPI web-based dashboard |
| `pt_gui_alerts.py` | — | Customizable alert rule builder UI |
| `pt_gui_heatmap.py` | — | Correlation heatmap visualization |
| `pt_gui_performance.py` | — | Performance attribution charts |
| `pt_gui_replay.py` | — | Trade history replay tool |
| `pt_backtester.py` | 876 | Historical strategy testing |
| `pt_thinker_exchanges.py` | 100 | Exchange wrapper for the Thinker |
| `pt_risk_dashboard.py` | 157 | Risk management GUI |
| `pt_volume_dashboard.py` | 117 | Volume analysis GUI |

---

## 3. Go Ultra-Project: `ultratrader-go/` (20,565 lines)

### 3.1 Architecture Overview

The Go system is a **clean-room reimplementation** that takes the best architecture ideas from:
- **TraderAlice/OpenAlice** → Platform architecture, domain boundaries, event logging
- **c9s/bbgo** → Go trading kernel, exchange/session/strategy abstractions
- **ccxt/ccxt** → Exchange capability realism (not all exchanges support everything)
- **Ekliptor/WolfBot** → Advanced execution features

**Zero external dependencies** — uses only Go standard library + `sync` primitives.

### 3.2 Package Map

```
cmd/ultratrader/            → Entry point (loads config, creates App, calls Start)

internal/
  core/
    app/                    → Composition root — wires ALL dependencies, manages lifecycle
    config/                 → YAML config loading + secrets management (AES-GCM)
    eventlog/               → Durable append-only event log
    logging/                → Structured JSON logger with file rotation
    utils/                  → ParseFloat, misc helpers

  exchange/
    types.go                → Adapter interface: Name(), Capabilities(), ListMarkets(), Balances(), PlaceOrder()
    registry.go             → Factory registry: Register("paper", ...), Register("binance", ...)
    capabilities.go         → Capability constants (Spot, Margin, Futures, etc.)
    ratelimit.go            → Exchange-side rate limit tracking
    paper/adapter.go        → Paper exchange: simulated order fills, deterministic testing
    binance/adapter.go      → Binance REST adapter: GetTickerPrice, PlaceOrder, WebSocket feed
    aggregator/             → Multi-exchange price aggregation (median/VWAP/mean/best-bid-ask)
    ratelimit/limiter.go    → Token bucket rate limiter with concurrent refill

  marketdata/
    types.go                → Tick, Candle, Subscription, CandleSubscription interfaces
    feed.go                 → StreamFeed interface: Subscribe(), GetTickerPrice()
    aggregator.go           → Market data aggregation across sources
    paper/                  → Simulated market streams (random walk, configurable drift)

  trading/
    account/                → Account management: multi-account support, CRUD operations
    execution/              → ExecutionService: order routing through risk pipeline → exchange
    execution/repository.go → Order history, execution summary, block rate tracking
    portfolio/              → Portfolio tracker: positions, realized/unrealized PnL, concentration
    orders/                 → Advanced orders: stop-loss, take-profit, trailing-stop, bracket, OCO
    rebalancer/             → Portfolio rebalancing: drift detection, order generation

  strategy/
    runtime.go              → Strategy interface: OnTick(), OnMarketTick(), OnMarketCandle()
    scheduler/              → 3 modes: timer (periodic), stream (tick-driven), candle-stream
    composite/              → Signal voting: unanimous, majority, any, weighted resolution
    regime/                 → Market regime detection: volatility, trend, Bollinger bandwidth
    sizing/                 → Position sizing: fixed, %risk, Kelly criterion, volatility-target, equal-weight
    demo/                   → Demo strategies: PriceThreshold, EMACrossover, TickMomentumBurst,
                               TickMeanReversion, MACDCrossover, BollingerReversion, CandleSMACross

  risk/
    guard.go                → Guard interface + Pipeline (sequential check-all-before-execute)
    symbol_whitelist.go     → Only allowed symbols pass
    max_notional.go         → Max total notional per account
    max_notional_per_symbol.go → Max notional for any single symbol (uses live market value)
    max_open_positions.go   → Max concurrent open positions
    max_concentration.go    → Max % concentration in any symbol
    cooldown.go             → Minimum time between orders
    duplicate_symbol.go     → Prevent same-symbol orders within time window
    duplicate_side.go       → Prevent same-side (buy/buy or sell/sell) orders within window
    circuitbreaker/         → Circuit breaker: CLOSED→OPEN→HALF_OPEN lifecycle

  backtest/
    engine.go               → Candle-driven backtesting engine with market emulation
    multisymbol.go          → Multi-symbol chronological feed alignment
    optimizer/
      walkforward.go        → Walk-forward optimization (out-of-sample validation)
      gridsearch.go         → Parallel grid search parameter tuning
      montecarlo.go         → Monte Carlo ruin probability simulation

  analytics/
    correlation/matrix.go   → Rolling Pearson correlation + diversification score + heatmap data
    journal/journal.go      → Trade journal + PerformanceStats (win rate, Sharpe, profit factor, max DD)
    features/extractor.go   → Feature extraction for ML: returns, RSI, SMA ratio, volume ratio, etc.
    ml/ensemble.go          → ML ensemble predictor
    rl/qlearning.go         → Q-learning RL agent
    patterns.go             → Pattern recognition engine
    arbitrage.go            → Arbitrage detection
    orderflow.go            → Order flow analysis

  indicator/
    indicators.go           → SMA, EMA, RSI, MACD, Bollinger Bands, ATR
    volume.go               → VWAP, OBV, Volume SMA, Volume Ratio, MFI, Chaikin Money Flow

  notification/
    notifier.go             → Email (SMTP), Discord (webhook), Telegram (Bot API)
                              Severity levels, per-channel minimum level filtering

  metrics/
    tracker.go              → Rolling window metrics collection + persistence

  persistence/
    db.go                   → SQLite database (parameterized queries, SQL injection prevented)
    orders/                 → Order journal persistence
    snapshot/               → Account snapshot persistence
    reports/                → Runtime report persistence

  reporting/
    api/                    → REST API handlers + rate limiting middleware + input validation
    runtime/                → Report generation (timer/stream/candle-stream modes)
    analysis/               → Trend analysis across historical reports

  enterprise/
    rbac.go                 → Multi-account RBAC with roles + permissions
    audit.go                → Cryptographic audit logging (HMAC-signed entries)

  connectors/
    httpapi/                → HTTP server with 16 endpoints + WebSocket support
```

### 3.3 Application Lifecycle

```go
// Simplified startup flow in cmd/ultratrader/main.go:
cfg := config.Load("config/development-timer.json")
app, _ := app.New(cfg)
app.Start(ctx)   // → starts HTTP server, scheduler, reports
app.Shutdown(ctx) // → graceful HTTP shutdown, logger close
```

**What happens on `Start()`**:
1. Logs "app startup initiated"
2. Appends `app.started` event to event log
3. Creates bootstrap snapshots for each account
4. Starts HTTP runtime (binds to `127.0.0.1:0` by default)
5. Starts scheduler service (timer/stream/candle-stream based on config)
6. Runs one strategy cycle immediately
7. Persists startup summary report
8. Logs "app startup completed" with full state snapshot

### 3.4 Risk Guard Pipeline — Order Verification

Every order passes through **9 sequential guards** before execution:

```
Order Request
    │
    ▼
┌─────────────────────┐
│ SymbolWhitelistGuard │  → Only allowed symbols (e.g., BTCUSDT, ETHUSDT)
└─────────┬───────────┘
          ▼
┌─────────────────────┐
│ MaxNotionalGuard    │  → Total notional across all orders < limit
└─────────┬───────────┘
          ▼
┌─────────────────────┐
│ MaxNotionalPerSymbol│  → Per-symbol notional < limit (uses live market value)
└─────────┬───────────┘
          ▼
┌─────────────────────┐
│ CooldownGuard       │  → Min time between any two orders
└─────────┬───────────┘
          ▼
┌─────────────────────┐
│ DuplicateSymbolGuard│  → No same-symbol orders within time window
└─────────┬───────────┘
          ▼
┌─────────────────────┐
│ DuplicateSideGuard  │  → No same-side (buy+buy) within time window
└─────────┬───────────┘
          ▼
┌─────────────────────┐
│ MaxOpenPositionsGuard│ → Max concurrent open positions
└─────────┬───────────┘
          ▼
┌─────────────────────┐
│ MaxConcentrationGuard│ → No symbol exceeds concentration % of portfolio
└─────────┬───────────┘
          ▼
┌─────────────────────┐
│ CircuitBreaker      │  → If failure threshold reached, blocks ALL orders
└─────────┬───────────┘
          ▼
      ✅ APPROVED → Execute via Exchange Adapter
```

If ANY guard rejects, the order is blocked with a `GuardError` containing the guard name and reason.

### 3.5 Strategy Runtime

Three scheduling modes determine how strategies are invoked:

| Mode | Trigger | Use Case |
|------|---------|----------|
| `timer` | Periodic interval (e.g., every 5s) | Pull-based strategies (PriceThreshold, EMACrossover) |
| `stream` | Market data tick event | Tick-reactive strategies (MomentumBurst, MeanReversion) |
| `candle-stream` | Candle completion event | Candle-based strategies (SMACross) |

**Signal flow**:
```
MarketData Feed → Tick/Candle → Strategy.OnMarketTick() → []Signal
                                                       ↓
                                              Risk Pipeline
                                                       ↓
                                            ExecutionService
                                                       ↓
                                            Exchange.PlaceOrder()
```

**Composite Strategy** — combines multiple strategies via voting:
- **Unanimous**: All strategies must agree
- **Majority**: More than half must agree
- **Any**: Any strategy fires (highest confidence wins)
- **Weighted**: Weighted vote (e.g., strategy A gets 2x weight over B)

### 3.6 Advanced Order Types

| Type | Trigger | Example |
|------|---------|---------|
| **StopLoss** | Price crosses trigger on adverse side | Buy position: triggers when price drops ≤ trigger |
| **TakeProfit** | Price crosses trigger on favorable side | Buy position: triggers when price rises ≥ trigger |
| **TrailingStop** | Percentage trail from peak | 2% trail: sells when price drops 2% from highest seen |
| **StopLimit** | Price hits stop, then limit order placed | Stop at $95k, limit buy at $94.5k |
| **Bracket** | StopLoss + TakeProfit pair | Auto-creates both protective orders |
| **OCO** | One-Cancels-Other group | When stop-loss triggers, take-profit is cancelled |

### 3.7 Backtesting Suite

| Component | Capability |
|-----------|------------|
| `CandleEngine` | Replay historical candles with configurable spread, fees, slippage |
| `MultiSymbolFeed` | Chronological alignment of multiple symbols for cross-asset testing |
| `MarketEmulator` | Configurable: maker/taker fees, slippage %, latency, fill probability |
| `WalkForwardOptimizer` | Train on window N, test on window N+1 (prevents overfitting) |
| `GridSearchOptimizer` | Parallel parameter grid search across strategy configs |
| `MonteCarlo` | Randomized shuffle of trade results → ruin probability distribution |

### 3.8 Market Regime Detection

| Detector | Method | Output |
|----------|--------|--------|
| `VolatilityDetector` | ATR/price ratio | QUIET / TRENDING / VOLATILE |
| `TrendDetector` | ADX-like directional movement | TRENDING / RANGING |
| `BollingerBandwidthDetector` | Band width percentile | VOLATILE / QUIET |
| `CompositeDetector` | Majority vote across all detectors | Final regime classification |

### 3.9 Position Sizing Library

| Sizer | Method | Best For |
|-------|--------|----------|
| `FixedSizer` | Constant lot size | Simple strategies, testing |
| `PercentRiskSizer` | Risk X% of portfolio per trade using ATR for stop distance | Conservative risk management |
| `KellySizer` | Kelly Criterion: f* = (bp-q)/b, capped at 25% | Aggressive optimal growth |
| `VolatilityTargetSizer` | Normalize all positions to target portfolio volatility | Stable risk-adjusted returns |
| `EqualWeightSizer` | Divide portfolio equally across N positions | Diversified portfolios |

### 3.10 Technical Indicators (11 total)

| Indicator | Category | Use Case |
|-----------|----------|----------|
| SMA | Trend | Simple moving average crossover |
| EMA | Trend | Faster-moving crossover signals |
| RSI | Momentum | Overbought/oversold (0-100) |
| MACD | Trend+Momentum | Signal line crossovers, divergence |
| Bollinger Bands | Volatility | Squeeze detection, mean reversion |
| ATR | Volatility | Position sizing, stop distance |
| VWAP | Volume+Price | Institutional price benchmark |
| OBV | Volume | Buying/selling pressure |
| Volume SMA | Volume | Volume trend confirmation |
| MFI | Volume+Momentum | Money flow in/out (0-100) |
| Chaikin Money Flow | Volume | Accumulation/distribution |

### 3.11 HTTP API Surface (16 endpoints)

```
GET  /                           → HTML dashboard
GET  /dashboard                  → HTML dashboard
GET  /healthz                    → Health check (for load balancers)
GET  /readyz                     → Readiness check
GET  /api/status                 → Runtime name, ready state, account count
GET  /api/portfolio              → All positions with live market values
GET  /api/portfolio-summary      → Open positions, concentration, total value, PnL
GET  /api/orders                 → Order history
GET  /api/execution-summary      → Execution stats (total, filled, blocked, block rate)
GET  /api/execution-diagnostics  → Execution summary + metrics snapshot
GET  /api/exposure-diagnostics   → Concentration map, top symbol, total exposure
GET  /api/guard-diagnostics      → Guard names + per-guard trigger counts
GET  /api/metrics                → Rolling window metrics (fill rate, latency, etc.)
GET  /api/guards                 → Active guard names
GET  /api/runtime-reports/latest → Latest report per type
GET  /api/runtime-reports/history → Historical reports by type
GET  /api/runtime-reports/trends  → Trend analysis across reports
```

### 3.12 Enterprise Features

| Feature | Implementation |
|---------|---------------|
| **RBAC** | Multi-account with roles (admin/trader/viewer) + permission checks |
| **Audit Logging** | HMAC-signed audit entries (tamper detection) |
| **Secrets Management** | AES-GCM encryption for API keys in config |
| **Input Validation** | Request body schema validation on all API endpoints |
| **API Rate Limiting** | Token bucket per-IP rate limiting middleware |

---

## 4. FreeLLM Proxy Integration (`../litellm_control_panel/`)

The project includes a **LiteLLM-based proxy** that routes LLM API calls to free/cheap providers:

### 4.1 How It Works

The `freellm.exe` binary at `../litellm_control_panel/` is a custom LiteLLM proxy that:
- Discovers **316+ LLM models** across multiple providers
- Routes requests labeled as one model name to the best available free/cheap model
- Maintains **provider rankings** and performance metrics in `provider_metrics.db`
- Auto-benchmarks and scores models on quality + latency

### 4.2 Model Aliases

The proxy maps familiar model names to free alternatives:

| Alias | Actual Routing |
|-------|---------------|
| `claude-sonnet-4-20250514` | deepseek-v4-flash-free (opencode.ai), mistral-large (Mistral API) |
| `claude-opus-4-20250514` | Multiple free high-capability models |
| `free-llm` | Best available free model (auto-selected) |
| `webai-gemini` | Gemini models via WebAI browser engine |
| 31 total aliases | Each with primary + fallback routing |

### 4.3 Providers

| Provider | Models Available | Auth |
|----------|-----------------|------|
| opencode.ai/zen | 32 free models | No API key |
| OpenRouter | 233 models | API key |
| NVIDIA/NIM | 22 models | API key |
| Groq | 2 models | API key |
| Cerebras | 1 model | API key |
| Fireworks | 4 models | API key |
| Hyperbolic | 4 models | API key |
| SambaNova | 3 models | API key |
| Mistral | Via direct API | MISTRAL_API_KEY |
| WebAI (Gemini) | Browser-based | No key |
| LM Studio | 6 local models | Local |

### 4.4 Potential Trading AI Integration

The proxy could power:
- **Sentiment analysis**: Route `pt_sentiment.py` LLM calls through freellm for free news analysis
- **NLP strategy parsing**: `pt_nlp_strategy.py` could use free LLMs to interpret natural language trading rules
- **RL reward shaping**: `pt_rl_optimizer.py` could use LLM feedback for reward function design
- **Report generation**: The Go reporting layer could call free LLMs for narrative analytics

---

## 5. Cross-Reference: Python → Go Mapping

| Python Module | Go Package | Status |
|--------------|------------|--------|
| `pt_hub.py` | `connectors/httpapi/` | ✅ Ported (REST API + HTML dashboard replaces Tkinter) |
| `pt_thinker.py` | `strategy/` + `analytics/ml/` | ⚠️ Partial (kNN core still Python-only) |
| `pt_trader.py` | `trading/execution/` + `trading/orders/` | ✅ Ported |
| `pt_trainer.py` | `backtest/` | ⚠️ Partial (training concept mapped to backtest) |
| `pt_backtester.py` | `backtest/engine.go` + `optimizer/` | ✅ Ported (substantially upgraded) |
| `pt_analytics.py` | `analytics/journal/` | ✅ Ported |
| `pt_exchanges.py` | `exchange/` (registry, binance, paper) | ✅ Ported |
| `pt_multi_exchange.py` | `exchange/aggregator/` | ✅ Ported |
| `pt_notifications.py` | `notification/notifier.go` | ✅ Ported |
| `pt_volume.py` | `indicator/volume.go` | ✅ Ported |
| `pt_correlation.py` | `analytics/correlation/` | ✅ Ported |
| `pt_position_sizing.py` | `strategy/sizing/` | ✅ Ported (5 sizers vs Python's 1) |
| `pt_rebalancer.py` | `trading/rebalancer/` | ✅ Ported |
| `pt_regime_detection.py` | `strategy/regime/` | ✅ Ported (4 detectors vs Python's 1) |
| `pt_ml_ensemble.py` | `analytics/ml/ensemble.go` | ✅ Ported |
| `pt_rl_optimizer.py` | `analytics/rl/qlearning.go` | ✅ Ported |
| `pt_sentiment.py` | — | ❌ Not ported yet |
| `pt_nlp_strategy.py` | — | ❌ Not ported yet |
| `pt_feature_engine.py` | `analytics/features/` | ✅ Ported |
| `pt_risk_management.py` | `risk/` (9 guards + circuit breaker) | ✅ Ported (substantially upgraded) |
| `pt_enterprise.py` | `enterprise/` (RBAC + audit) | ✅ Ported |
| `pt_config.py` | `core/config/` | ✅ Ported |
| `pt_logging.py` | `core/logging/` | ✅ Ported |

### Go-Only Features (No Python Equivalent)

| Package | Feature |
|---------|---------|
| `risk/circuitbreaker/` | Circuit breaker pattern (CLOSED→OPEN→HALF_OPEN) |
| `strategy/composite/` | Signal voting with 4 resolution modes |
| `strategy/scheduler/` | 3 scheduling modes (timer/stream/candle-stream) |
| `marketdata/` | Stream subscription system with tick + candle events |
| `metrics/tracker/` | Rolling window metrics + persistence |
| `persistence/` | Durable order journal, snapshot store, report store |
| `reporting/api/` | REST API with rate limiting + input validation |
| `reporting/runtime/` | Scheduled report generation + trend analysis |
| `exchange/binance/` | Binance REST + WebSocket market data feed |
| `exchange/ratelimit/` | Token bucket rate limiter |
| `backtest/optimizer/` | Walk-forward, grid search, Monte Carlo |

---

## 6. Submodule Inventory

47+ external repositories organized by page:

| Page | Count | Key References | Purpose |
|------|-------|---------------|---------|
| 02 | 10 | **OpenAlice**, **catalyst**, **Krypto-trading-bot** | Architecture patterns, bot frameworks |
| 03 | 8 | **WolfBot**, **golang-crypto-trading-bot**, **TradeBot** | Advanced features, Go implementations |
| 04 | 9 | **ccxt**, **bbgo**, **CryptoTradingFramework** | Exchange libraries, Go kernels |
| 05 | 10 | **FinRL_Crypto**, **intelligent-trading-bot** | AI/ML trading research |
| 06 | 11 | **LLMAgentCrypto**, **pycryptobot**, **awesome-systematic-trading** | LLM agents, systematic trading research |

**Removed submodules** (dead links or broken internal deps):
- `Krypto-Hashers-Community/polymarket-crypto-sports-arbitrage-trading-bot` (404)
- `RobertMarcellos/polymarket-copy-trading-bot` (404)
- `SFCQuantX/polymarket-trading-agent` (404)
- `ericjang/cryptocurrency_arbitrage` (broken internal `fxbtc` submodule)
- `hello2all/gamma-ray` (nested submodules using dead `git://` protocol)
- `fluidex/dingir-exchange` (broken nested submodule)

---

## 7. Test Coverage

| Layer | Files | Status |
|-------|-------|--------|
| Python | `tests/test_all_modules.py` | Basic module import validation |
| Go | 91 `_test.go` files | 35 packages pass, 3 have no tests (utils, marketdata, account) |
| Go build | `go build ./...` | ✅ Clean, zero errors |
| Go deps | `go.sum` | Zero external dependencies beyond stdlib |

---

## 8. Deployment Options

| Method | Command | Notes |
|--------|---------|-------|
| Python local | `python pt_hub.py` | Requires Tkinter + display server |
| Go local | `go run ./cmd/ultratrader` | Default: paper exchange + paper market data |
| Go with config | `go run ./cmd/ultratrader --config config/development-timer.json` | Timer-based scheduler |
| Go stream mode | `--config config/development-stream.json` | Tick-driven scheduler |
| Go paper service | `--config config/paper-service.json` | Daemon-ready paper trading |
| Docker | `docker build -t ultratrader-go .` | Alpine-based image |
| Docker Compose | `docker compose up --build` | Multi-service |

---

## 9. Key Architectural Insights

### What Makes This Project Unique

1. **The kNN AI is genuinely different** — not a neural net, not an LLM, but a pattern-matching engine that literally memorizes every historical pattern and votes on the future based on similarity. Online weight adjustment means it adapts without retraining.

2. **The "no stop-loss" philosophy is intentional and well-reasoned** — the operator believes stop-losses are a futures trading concept blindly applied to spot trading where they cause unnecessary realized losses. DCA + patience is the risk management strategy.

3. **The Go port is a clean-room architectural upgrade** — same behavior goals, radically better engineering: interface-driven, concurrent-safe, zero external deps, comprehensive guard pipeline, multiple scheduling modes, advanced order types, walk-forward backtesting.

4. **The FreeLLM proxy is a cost multiplier** — routes to 316+ free models, enabling AI-augmented trading features (sentiment, NLP strategy, RL reward) at zero marginal cost.

### Current Gaps

| Gap | Impact |
|-----|--------|
| kNN core not yet in Go | Go can't run the main Thinker AI natively |
| Python test coverage | Only 1 test file for 25K lines |
| Sentiment + NLP not in Go | Two Python modules without Go equivalents |
| No live exchange trading in Go | Paper mode only (Binance adapter is data-only) |
| Web dashboard not feature-complete | Go HTML dashboard < Python Tkinter GUI |

### Recommended Next Steps

1. **Port the kNN engine to Go** — the single highest-impact gap
2. **Wire FreeLLM into Go** — use the proxy for AI-augmented features
3. **Add a real exchange trading adapter** — Binance spot trading in Go
4. **Complete the web dashboard** — port remaining Python GUI features to React/Vite
5. **Expand Go test coverage** — close the 3 packages without test files
