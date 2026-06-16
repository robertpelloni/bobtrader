# Autonomous Dual-Bot Competition Framework

**System:** `fully_automated_gay_luxuxy_communism`
**Version:** 1.0.0
**Status:** Active Development

---

## 1. Core Concept

Two independent trading bot implementations (Python legacy stack and Go ultratrader) run **simultaneously with live Binance capital**, competing to maximize profitability. An AI agent monitors both, modifies their source code, rebuilds, restarts, and iterates — constantly searching for the optimal trading strategy and implementation.

### Asset Allocation

| Recipient | Allocation | Purpose |
|-----------|-----------|---------|
| **Python Bot** | 25% of portfolio | Legacy PowerTrader AI — proven but frozen |
| **Go Bot** | 25% of portfolio | UltraTrader Go — modern, actively developed |
| **Reserve** | 50% of portfolio | Held as USDT, deployed strategically |

**Allocation Enforcement:** The supervising AI agent rebalances assets via Binance API. Positions and balances are checked before each trading cycle. Reserve can be deployed to either bot as a performance "bonus" when one bot proves superior.

### Design Principles

1. **Evolutionary Competition** — Bots compete; the better performer gets more capital
2. **Safe Protocol** — Neither bot can lose >30% of its allocation before automatic shutdown
3. **Observability** — Every trade, signal, and decision is logged
4. **Self-Modifying** — The AI agent rewrites, rebuilds, and restarts both bots autonomously
5. **No Interference** — Bots do not share positions or signals; each is an independent agent

---

## 2. System Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                 fully_automated_gay_luxuxy_communism             │
│                    (Supervising AI Agent)                        │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  ┌─────────────────────┐         ┌─────────────────────┐         │
│  │     Python Bot      │         │       Go Bot        │         │
│  │     (25% cap)       │         │     (25% cap)       │         │
│  │                     │         │                     │         │
│  │  pt_hub.py          │         │  ultratrader-go     │         │
│  │  pt_thinker.py      │         │  cmd/ultratrader    │         │
│  │  pt_trader.py       │         │  internal/core/     │         │
│  │  pt_exchanges.py    │         │  internal/strategy/ │         │
│  │  pt_notifications.py│         │  internal/trading/  │         │
│  │  pt_analytics.py    │         │                     │         │
│  └────────┬────────────┘         └────────┬────────────┘         │
│           │                               │                       │
│           └───────────┬───────────────────┘                      │
│                       │                                          │
│              ┌────────▼────────┐                                  │
│              │  Binance.US API  │                                  │
│              │  (separate keys) │                                  │
│              └────────┬────────┘                                  │
│                       │                                           │
│              ┌────────▼────────┐                                  │
│              │     Ledger      │                                  │
│              │  (shared PnL    │                                  │
│              │   tracking)     │                                  │
│              └─────────────────┘                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 2.1 Component Responsibilities

**Supervising AI Agent:**
- Monitors both bots' performance metrics every N seconds/minutes
- Detects underperformance, crashes, or anomalous behavior
- Generates and applies code patches to either bot
- Rebuilds and restarts modified bots
- Rebalances capital between bots based on performance
- Maintains the shared ledger and performance database

**Python Bot:**
- Runs the legacy PowerTrader AI system
- Controls 25% of portfolio (will trade within this limit)
- Uses its own independent strategy set (kNN prediction, technical indicators)
- Outputs structured logs for performance monitoring

**Go Bot:**
- Runs the UltraTrader Go system
- Controls 25% of portfolio (will trade within this limit)
- Uses its own independent strategy set (14+ strategies currently)
- Exposes health/performance API endpoints

### 2.2 Data Flow

```
Binance.US Market Data
    │
    ├──→ Python Bot ──→ Python Signals ──→ Python Trades ──→ Python PnL
    │
    ├──→ Go Bot ──────→ Go Signals ──────→ Go Trades ──────→ Go PnL
    │
    └──→ Supervisor ──→ Aggregate Analysis ──→ Decision Engine ──→ Code Patches
```

---

## 3. Asset Management Protocol

### 3.1 Initial Allocation

1. Read total portfolio value from Binance.US account
2. Compute: `pythonAllocation = total * 0.25`, `goAllocation = total * 0.25`, `reserve = total * 0.50`
3. Transfer the calculated amount of USDT to each bot's operational wallet/trading pair
4. Record baseline in the ledger

### 3.2 Rebalancing Rules

| Condition | Action |
|-----------|--------|
| One bot's PnL > other by 10%+ over 24h | Award 5% of reserve to winning bot |
| One bot's drawdown > 20% | Reduce allocation by half, return to reserve |
| Both bots negative for 48h | Halt both, switch to paper-only mode |
| Reserve hits zero | Stop rebalancing, both bots at current allocation |

### 3.3 Drawdown Protection

```
┌─────────────────────────────────────────────────────────────────┐
│                    Drawdown Protection System                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Drawdown < 10%  →  Normal operation                             │
│  Drawdown < 20%  →  Reduce position size by 50%                  │
│  Drawdown < 30%  →  Close all positions, halt bot                │
│  Drawdown ≥ 30%  →  Emergency shutdown (transfer all to reserve) │
│                                                                  │
│  Recovery: After 24h above 10% drawdown, allow restart           │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### 3.4 Balance Check Frequency

- Every 5 minutes: Check open positions, free balance, locked balance
- Every 15 minutes: Compute PnL for each bot
- Every hour: Full portfolio reconciliation
- Every 24 hours: Rebalance decision

---

## 4. Performance Monitoring & Comparison

### 4.1 Key Performance Indicators (KPIs)

| Metric | Python Bot | Go Bot | Notes |
|--------|-----------|--------|-------|
| **Total PnL** | $ value | $ value | Absolute profit/loss |
| **Win Rate** | % | % | Profitable trades / total trades |
| **Sharpe Ratio** | computed | computed | Risk-adjusted return |
| **Max Drawdown** | % | % | Peak-to-trough decline |
| **Trade Frequency** | trades/hour | trades/hour | How active each bot is |
| **Avg Trade Duration** | minutes | minutes | Holding period |
| **Avg Profit per Trade** | $ | $ | Dollar per winning trade |
| **Avg Loss per Trade** | $ | $ | Dollar per losing trade |
| **Profit Factor** | ratio | ratio | Gross profit / gross loss |
| **Strategy Utilization** | strategies active | strategies active | How many strategies fire |
| **Signal-to-Execution Rate** | % executed | % executed | Signals that become trades |
| **Capital Utilization** | % of allocation | % of allocation | How much is deployed vs idle |

### 4.2 Comparison Dashboard

The supervising agent should expose a live comparison endpoint (e.g., HTTP on port 8300):

```
GET /compare

{
  "timestamp": "...",
  "python": { KPI values },
  "go": { KPI values },
  "winner": "go|python|tie",
  "allocation": {
    "python": { "current": 123.45, "initial": 100.00, "pct": 25 },
    "go": { "current": 134.56, "initial": 100.00, "pct": 25 },
    "reserve": { "current": 212.99, "pct": 50 },
    "total": 471.00
  },
  "recommendation": "transfer 5% reserve to Go bot"
}
```

### 4.3 Performance History

Every KPI is persisted to a time-series database (JSONL files in `data/competition/`):

```
data/competition/
├── performance.jsonl     # Timestamped snapshots of both bots' KPIs
├── allocations.jsonl     # Rebalancing events with amounts
├── patches.jsonl         # Every code patch applied to either bot
├── restarts.jsonl        # Every restart event with reason
└── decisions.jsonl       # Every supervisor decision with rationale
```

---

## 5. Code Modification Pipeline

This is the core of the autonomous system — the ability to modify source code, rebuild, and restart without human intervention.

### 5.1 Trigger Events

The AI agent initiates a code change when any of these occur:

| Trigger | Threshold | Action |
|---------|-----------|--------|
| **Underperformance** | Bot's PnL < other bot for 24h | Tweak strategy params or add new strategy |
| **Drawdown spike** | DD > 15% in 1h | Tighten risk guards, reduce position sizes |
| **Stagnation** | No trades for 6h | Adjust signal thresholds or add market regime |
| **Competitive gap** | >20% PnL gap between bots | Port winning strategies to losing bot |
| **Market regime shift** | F&G changes category | Activate/deactivate market-sensitive strategies |
| **Error rate** | >10% of signals fail | Fix error handling, retry logic |
| **Opportunity detection** | New pattern observed | Write new strategy implementation |

### 5.2 Modification Workflow

```
1. DETECT   →  Identify trigger event from KPI comparison
2. ANALYZE  →  Read relevant source files, understand current logic
3. DESIGN   →  Formulate code change (new param, new strategy, fix)
4. EDIT     →  Apply change to source file(s)
5. BUILD    →  Compile (Go: go build, Python: syntax check)
6. VALIDATE →  Run unit tests (if any exist), check for compilation errors
7. BACKUP   →  Backup original file before applying change
8. DEPLOY   →  Stop old process, start new process
9. MONITOR  →  Observe for N minutes; if crash or error, rollback
10. ROLLBACK →  Restore backup and restart if validation or monitoring fails
```

### 5.3 Backup & Rollback

Every modification creates a backup:

```
.git/autonomous_backups/
├── python/
│   ├── 2026-06-14_15-30-00_pt_trader.py
│   └── 2026-06-14_16-00-00_pt_strategies.py
└── go/
    ├── 2026-06-14_15-35-00_ws_feed.go
    └── 2026-06-14_16-05-00_scheduler.go
```

**Auto-rollback triggers:**
- Process crashes within 5 minutes of restart
- Error rate > 20% after modification
- PnL gap > 25% in favor of the non-modified bot within 1 hour
- Any Binance API auth failure after code change

### 5.4 Git Workflow

The supervising AI agent should commit every change:

```bash
git add -A
git commit -m "auto: [python|go] brief description of change

Trigger: [reason]
Metrics before: [KPIs]
Metrics after: [expected improvement]"
git push origin main
```

---

## 6. Process Lifecycle Management

### 6.1 Startup Sequence

```
1. Supervisor starts
2. Supervisor reads config (allocations, API keys, pairs)
3. Supervisor starts Python bot as subprocess
   └─ python3 pt_hub.py --config config/autonomous-paper.json &
4. Supervisor starts Go bot as subprocess
   └─ go run -buildvcs=false ./cmd/ultratrader --config config/paper-live-data.json &
5. Supervisor waits for both health endpoints to respond 200
6. Supervisor verifies Binance API connectivity for both bots
7. Supervisor marks system as "OPERATIONAL"
8. Supervisor begins monitoring loop
```

### 6.2 Monitoring Loop

```
Every N seconds (configurable, default 30):

1. Check Python bot: process alive? health endpoint? API accessible?
2. Check Go bot: process alive? health endpoint? API accessible?
3. Fetch both bots' current PnL, open positions, balances
4. Compute KPIs for both
5. Log all metrics to performance.jsonl
6. Check for trigger events (see 5.1)
7. If trigger detected → enter modification workflow (see 5.2)
8. If no trigger → continue monitoring
```

### 6.3 Restart Protocol

**Restart triggers:**
- Process crash detected
- Health check fails 3 consecutive times
- After source code modification
- Memory or CPU exceeded thresholds

**Restart procedure:**
1. Send SIGTERM to process
2. Wait 10 seconds
3. If process still alive, send SIGKILL
4. Verify port is free
5. Start new process
6. Wait for health check (up to 30 seconds)
7. If health check fails, rollback and restart previous version

### 6.4 Crash Recovery

If a bot crashes:
1. Immediately freeze the crashed bot's positions (no new trades)
2. Check if the other bot can manage the extra capital temporarily
3. Restart the crashed bot with its backup config
4. If crash repeats 3 times, mark bot as "BROKEN" and investigate deeper

---

## 7. Strategy Competition & Knowledge Transfer

### 7.1 Strategy Pool

Both bots maintain independent strategy sets, but the supervisor can port strategies between them:

| Strategy | Python | Go | Notes |
|----------|--------|----|-------|
| EMA Crossover | ✅ | ✅ | Both |
| Bollinger Bands | ✅ | ✅ | Both |
| RSI Reversion | ✅ | ✅ | Both |
| Trailing Take Profit | ❌ | ✅ | Go-only (portable) |
| Tick Momentum Burst | ❌ | ✅ | Go-only (portable) |
| Tick Mean Reversion | ❌ | ✅ | Go-only (portable) |
| kNN Prediction | ✅ | ❌ | Python-only (port to Go) |
| Whale Alert | ❌ | ✅ | Go-only (portable) |
| Sentiment Analysis | ❌ | ✅ | Go-only (portable) |
| Market Making | ❌ | ✅ | Go-only (portable) |
| Weekly Cycle | ❌ | ✅ | Go-only (portable) |
| China Session | ❌ | ✅ | Go-only (portable) |

**Porting priority:** Strategies that outperform on one platform should be ported to the other platform within 24 hours of proven success.

### 7.2 Competition Rules

1. **Fair fight** — Both bots use the same Binance.US market data feed
2. **Equal starting capital** — Both start with 25% of portfolio (adjusted for decimal constraints)
3. **Same trading pairs** — Both trade the same symbols (ETHUSDT, BTCUSDT, etc.)
4. **Independent execution** — Neither bot can see or influence the other's trades
5. **Time-bound evaluation** — Performance is evaluated in 24h windows
6. **No collusion** — Bots do not coordinate strategies or hedge each other

### 7.3 Winner Determination

At each 24h evaluation point:

```python
if python_pnl > go_pnl * 1.10:  # 10% better
    winner = "python"
    transfer 5% reserve to Python allocation
elif go_pnl > python_pnl * 1.10:
    winner = "go"
    transfer 5% reserve to Go allocation
else:
    winner = "tie"
    no change
```

**Monthly showdown:** The bot with the highest monthly PnL gets an additional 10% of reserve transferred. If one bot consistently underperforms for 30 days, consider deprecating it.

---

## 8. Safety Guards

### 8.1 Hard Limits (Cannot Be Overridden by Code Changes)

| Guard | Limit | Consequence |
|-------|-------|-------------|
| Max position size per symbol | 15% of bot's allocation | Trade rejected |
| Max open positions per bot | 5 at any time | Order rejected |
| Max daily loss per bot | 30% of bot's allocation | Bot halted |
| Max leverage | 1x (no margin) | API config enforcement |
| Min USDT reserve per bot | 5% of allocation | No new buys |
| Max trade frequency | 1 trade per 30 seconds per pair | Rate limited |
| Max concurrent modifications | 1 modification every 5 minutes | Blocked |
| Max rollbacks per day | 5 per bot | Supervisor intervention |

### 8.2 Soft Limits (Modifiable by Supervisor)

| Guard | Default | Adjustment Range |
|-------|---------|-----------------|
| Position sizing multiplier | 1.0 | 0.1 – 3.0 |
| Cooldown between trades | 60s | 10s – 600s |
| Max slippage tolerance | 0.5% | 0.1% – 2.0% |
| Stop-loss trigger | -5% | -2% – -15% |
| Take-profit trigger | +3% | +1% – +10% |

### 8.3 Emergency Shutdown

If ANY of these occur, both bots halt immediately:

1. Binance API returns "account locked" or authentication error
2. Total portfolio drawdown exceeds 25%
3. Either bot's PnL drops below -50% of its initial allocation
4. Python or Go process crashes 5+ times in 1 hour
5. Network connectivity lost for > 5 minutes
6. Unexplained balance discrepancy > 1% of portfolio

**Emergency shutdown procedure:**
```
1. Cancel all open orders for both bots
2. Do NOT close existing positions (to avoid locking in losses)
3. Set both bots to "HALTED" state
4. Log all trades and signals for audit
5. Notify operator (if configured)
6. Wait for manual intervention
```

---

## 9. Logging & Observability

### 9.1 Log Structure

Every event is logged in structured JSONL format:

```json
{"timestamp":"2026-06-14T15:30:00Z","type":"signal","bot":"go","symbol":"ETHUSDT","strategy":"bollinger","action":"buy","price":1665.49,"confidence":0.75}
{"timestamp":"2026-06-14T15:30:01Z","type":"trade","bot":"go","symbol":"ETHUSDT","side":"buy","quantity":0.001,"price":1665.49,"fee":0.000001665,"status":"filled"}
{"timestamp":"2026-06-14T15:30:05Z","type":"kpi","bot":"python","win_rate":0.8,"total_pnl":1.23,"drawdown":0.05}
{"timestamp":"2026-06-14T15:30:06Z","type":"modification","bot":"go","file":"ws_feed.go","change":"remove bufio.Reader","status":"committed"}
{"timestamp":"2026-06-14T15:31:00Z","type":"restart","bot":"python","reason":"code change","status":"success","duration_ms":2340}
```

### 9.2 Key Metrics Endpoint

Each bot exposes its own health/metrics endpoint:

| Bot | Endpoint | Port |
|-----|----------|------|
| Python | http://127.0.0.1:8299/health | 8299 |
| Go | http://127.0.0.1:8300/health | 8300 |
| Supervisor | http://127.0.0.1:8400/ | 8400 |

### 9.3 Supervisor Dashboard

The supervisor exposes a unified dashboard at port 8400:

```
GET /          → HTML dashboard with both bots' KPIs side-by-side
GET /api/comparison → JSON comparison data (for AI consumption)
GET /api/history    → Historical performance data
GET /api/modifications → Recent code changes
GET /api/logs       → Recent log entries
```

---

## 10. Implementation Roadmap

### Phase 1: Foundation (Week 1)
- [ ] Create supervisor agent entry point (`supervisor.py` or `supervisor.go`)
- [ ] Implement subprocess management (start, stop, restart both bots)
- [ ] Implement health check polling
- [ ] Implement basic PnL tracking
- [ ] Create comparison dashboard on port 8400

### Phase 2: Asset Management (Week 2)
- [ ] Implement Binance allocation API calls
- [ ] Build rebalancing logic
- [ ] Implement drawdown protection
- [ ] Set up performance history persistence
- [ ] Build KPI computation engine

### Phase 3: Code Modification (Week 3)
- [ ] Build backup/rollback system
- [ ] Implement git commit workflow
- [ ] Build modification trigger detection
- [ ] Create patch generation pipeline
- [ ] Implement validation & test running

### Phase 4: Competition (Week 4)
- [ ] Implement comparison scoring
- [ ] Build strategy porting pipeline
- [ ] Add performance alerts
- [ ] Fine-tune evaluation windows
- [ ] Add automated reporting

### Phase 5: Production Hardening (Week 5+)
- [ ] Emergency shutdown testing
- [ ] Network failure simulation
- [ ] Crash recovery testing
- [ ] Performance benchmarking
- [ ] Long-term stability runs

---

## 11. Quick Reference

### Commands

```bash
# Start both bots
python3 supervisor.py start

# Stop both bots
python3 supervisor.py stop

# Manual rebalance
python3 supervisor.py rebalance

# View status
curl http://127.0.0.1:8400/api/comparison

# Apply emergency halt
python3 supervisor.py halt

# Restart single bot
python3 supervisor.py restart --bot go
python3 supervisor.py restart --bot python
```

### Key Files

| File | Purpose |
|------|---------|
| `AUTONOMOUS_DUAL_BOT_STRATEGY.md` | This document |
| `supervisor.py` | Supervisor agent entry point |
| `ultratrader-go/cmd/ultratrader/main.go` | Go bot entry point |
| `pt_hub.py` | Python bot entry point |
| `config/competition.json` | Competition configuration |
| `data/competition/` | Competition metrics and history |
| `data/ledger/` | Shared PnL tracking |

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `BINANCE_API_KEY` | ✅ | Binance.US API key |
| `BINANCE_SECRET_KEY` | ✅ | Binance.US secret key |
| `SUPERVISOR_PORT` | ❌ | Supervisor dashboard port (default: 8400) |
| `GO_BOT_PORT` | ❌ | Go bot dashboard port (default: 8300) |
| `PYTHON_BOT_PORT` | ❌ | Python bot dashboard port (default: 8299) |
| `MAX_DRAWDOWN_PCT` | ❌ | Drawdown limit before halt (default: 30) |
| `EVALUATION_WINDOW_H` | ❌ | Performance evaluation window (default: 24) |
| `RESERVE_PCT` | ❌ | Reserve percentage (default: 50) |
| `PYTHON_ALLOCATION_PCT` | ❌ | Python allocation (default: 25) |
| `GO_ALLOCATION_PCT` | ❌ | Go allocation (default: 25) |

---

## 12. License & Attribution

This framework is part of the `fully_automated_gay_luxuxy_communism` project. The underlying bot implementations (PowerTrader AI Python stack and UltraTrader Go) are licensed under Apache 2.0.

**Warning:** Cryptocurrency trading carries significant financial risk. This system autonomously modifies trading software and deploys live capital. Use at your own risk. No guarantee of profitability is expressed or implied.
