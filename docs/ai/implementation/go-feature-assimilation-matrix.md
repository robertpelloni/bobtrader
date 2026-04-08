# Go Feature Assimilation Matrix

## Purpose
Track how the emerging `ultratrader-go/` codebase maps to the audited source projects and to the long-term Go ultra-project plan.

## Current Status
The project now has a policy-aware paper trading loop, in-memory runtime state, structured logging, market-value estimation, PnL tracking, runtime metrics, operator API surfaces, explicit runtime lifecycle control, persistent runtime summary reports, market-data streaming abstractions, block-reason-aware diagnostics, and app-level shutdown validation. It increasingly resembles a supervised service kernel rather than a bootstrap harness.

## Matrix

| Target subsystem | Current status | Primary source inspiration | Secondary source inspiration | Notes |
|---|---|---|---|---|
| App runtime / composition root | Implemented | OpenAlice | BBGO | Platform-style orchestration remains the backbone |
| Config loading | Implemented | OpenAlice | PyCryptoBot, PowerTrader | Config now covers persistence, logging, server, scheduler, risk, and report history |
| Structured logging | Implemented | platform ops patterns | OpenAlice, PowerTrader | Context-driven correlation IDs and JSON logging now exist |
| Runtime metrics | Implemented | operator observability patterns | PowerTrader dashboard mentality | Attempts/successes/blocks tracked in memory |
| Block-reason metrics | Implemented | operator diagnostics | policy-aware runtime supervision | Metrics now track which guards blocked executions |
| Runtime report persistence | Implemented | OpenAlice durable state mindset | PowerTrader reporting mentality | Startup summaries, metrics snapshots, and valuation snapshots now persist |
| Event log | Implemented | OpenAlice | PowerTrader analytics mindset | JSONL append-only event durability remains central |
| Unified account model | Implemented | OpenAlice UTA | PowerTrader account-centric operation | Accounts remain the main execution boundary |
| Exchange capability vocabulary | Implemented | CCXT | BBGO | Capability-driven contract remains intact |
| Exchange registry | Implemented | BBGO | OpenAlice broker registry concept | Factory-based registration in place |
| Paper exchange adapter | Implemented | BBGO | PyCryptoBot paper/safe-first mindset | First safe adapter remains the execution target |
| Guard pipeline framework | Implemented | OpenAlice | PowerTrader risk controls | General pipeline in place with diagnosable names and structured guard errors |
| Symbol whitelist guard | Implemented | OpenAlice | PowerTrader practical safeguards | Concrete symbol policy enforcement |
| Max notional guard | Implemented | OpenAlice | PowerTrader position/risk bounds | Concrete monetary policy enforcement |
| Cooldown guard | Implemented | OpenAlice | WolfBot/PowerTrader temporal control ideas | Prevents immediate repeated symbol execution per account |
| Duplicate symbol guard | Implemented | OpenAlice | runtime safety patterns | Uses recent repository history to block repeated symbol execution |
| Max open positions guard | Implemented | OpenAlice | portfolio-aware runtime control | Uses live portfolio state to constrain expansion |
| Concentration guard | Implemented scaffold + live-valued exposure wiring | PowerTrader analytics/risk thinking | portfolio-aware control patterns | Now wired through live-valued exposure view |
| Execution service | Implemented | BBGO | OpenAlice service composition | Real account -> guard -> adapter -> persistence flow exists |
| Correlation-aware execution logs | Implemented | platform observability patterns | OpenAlice runtime introspection | Execution flow carries correlation IDs into logs and journals |
| Execution repository | Implemented | platform state management patterns | BBGO runtime direction | In-memory order state and summary data exist |
| Execution summary | Implemented | operator diagnostics patterns | PowerTrader dashboard thinking | Order counts and latest-symbol summary are available |
| Order journal | Implemented | PowerTrader journal mindset | OpenAlice durability | Orders persist independently of events |
| Portfolio tracker | Implemented | BBGO runtime direction | PowerTrader analytics intuition | Internal position state exists |
| Portfolio valuation | Implemented | PowerTrader dashboard mentality | BBGO market-aware runtime direction | Total market value derived from paper market data |
| Portfolio PnL | Implemented | PowerTrader analytics direction | BBGO runtime accounting direction | Realized and unrealized PnL tracked |
| Snapshot persistence | Implemented | OpenAlice | PowerTrader journaling/dashboards | Bootstrap snapshots continue to persist |
| Health/readiness endpoints | Implemented | cloud-native ops conventions | PowerTrader operator UX | Minimal operator-facing API surface exists |
| Status API | Implemented | OpenAlice platform introspection | PowerTrader dashboard ideas | `/api/status` exists |
| Portfolio API | Implemented | PowerTrader dashboard ideas | OpenAlice runtime state exposure | `/api/portfolio` exposes valued positions and PnL |
| Orders API | Implemented | OpenAlice state exposure | PowerTrader trade visibility | `/api/orders` exposes in-memory order state |
| Execution summary API | Implemented | operator diagnostics | PowerTrader state visibility | `/api/execution-summary` exposes order summary data |
| Metrics API | Implemented | operator observability | PowerTrader-style runtime dashboards | `/api/metrics` exposes attempt/success/block counts |
| Guards API | Implemented | operator diagnostics | OpenAlice introspection mindset | `/api/guards` exposes configured guard names |
| Guard diagnostics API | Implemented | operator diagnostics | policy-aware supervision | `/api/guard-diagnostics` exposes active guards plus block reason metrics |
| Runtime reports API | Implemented | durable state introspection | reporting-oriented runtime governance | `/api/runtime-reports/latest` exposes latest report-by-type snapshots |
| HTTP runtime wrapper | Implemented | platform ops patterns | OpenAlice connector/runtime thinking | Server lifecycle shell exists |
| HTTP runtime shutdown control | Implemented | service lifecycle patterns | OpenAlice runtime discipline | Runtime start/shutdown is test-covered |
| App-level shutdown integration | Implemented | daemon readiness patterns | OpenAlice runtime discipline | App startup + health + shutdown path now integration-tested |
| Market data interface | Implemented | BBGO | CCXT, WolfBot | Abstraction exists |
| Market data streaming interface | Implemented | BBGO stream-first thinking | WolfBot runtime event flow | Subscription abstraction now exists |
| Paper market data feed | Implemented | BBGO | PowerTrader practical bootstrap needs | Deterministic local feed supports strategy development |
| Paper tick subscription | Implemented | BBGO stream model | daemon/runtime evolution path | `SubscribeTicks()` exists for paper market data |
| Strategy runtime | Implemented | BBGO | WolfBot signal/event chaining | Signal aggregation runtime exists |
| Core Indicator Library | Implemented | BBGO | WolfBot | Added fundamental technical indicators (SMA, EMA, RSI) to support complex strategy logic. |
| Market-data-aware strategy | Implemented | BBGO | WolfBot, PowerTrader | Strategy consumes feed data |
| Strategy scheduler | Implemented | WolfBot event flow | BBGO runtime thinking | Converts signals into execution requests |
| Recurring scheduler service | Implemented | WolfBot event loop direction | BBGO daemon trajectory | Repeated execution behavior is test-covered |
| Backtesting | Implemented | BBGO | WolfBot, PowerTrader | Basic event-driven simulation engine over historical ticks with portfolio tracking |
| Multi-Timeframe Logic | Implemented | BBGO | WolfBot | Strategies can implement `CandleStrategy` to consume historical and live k-line/candle arrays. |
| Market Emulation | Implemented | BBGO | PowerTrader | Configurable slippage and maker/taker fees integrated natively into simulated backtest `Engine.processSignals` |
| Optimization | Not yet implemented | BBGO | WolfBot | Deferred |
| Arbitrage engine | Not yet implemented | WolfBot | kelvinau, ericjang, polymarket repos | Later advanced module |
| Notifications | Not yet implemented in Go | PowerTrader | BBGO, OpenAlice | Still reference-only |
| Dashboard / operator UI | Not yet implemented in Go | PowerTrader | BBGO, OpenAlice | Operator APIs exist, but full UI remains deferred |
| AI / research layer | Not yet implemented in Go | OpenAlice | PowerTrader, AI-specific repos | Must remain optional |

## Interpretation
### Architectural identity is getting sharper
The project now visibly favors:
- account-centered orchestration,
- capability-driven exchange abstraction,
- policy-before-execution,
- event and journal durability,
- execution plus internal in-memory state,
- strategy/runtime separation,
- deterministic paper-first development,
- operator-readable state surfaces,
- structured observability,
- runtime state over time,
- explicit service lifecycle control,
- gradual movement toward event-driven data flow,
- durable runtime summary history.

### Why this matters
This matrix continues to protect the project from random feature sprawl. New work should still be justified against:
1. the target architecture,
2. audited inspirations,
3. convergence value for the unified Go platform.

## Recommended use
Update this matrix whenever a major subsystem is added or materially strengthened.
