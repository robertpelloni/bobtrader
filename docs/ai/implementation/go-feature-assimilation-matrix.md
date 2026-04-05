# Go Feature Assimilation Matrix

## Purpose
Track how the emerging `ultratrader-go/` codebase maps to the audited source projects and to the long-term Go ultra-project plan.

## Current Status
The project now has a policy-aware paper trading bootstrap loop with internal order memory and portfolio state. It is still early, but the codebase now contains the first meaningful internal trading state beyond persistence files alone.

## Matrix

| Target subsystem | Current status | Primary source inspiration | Secondary source inspiration | Notes |
|---|---|---|---|---|
| App runtime / composition root | Implemented | OpenAlice | BBGO | Platform-style orchestration remains the backbone |
| Config loading | Implemented | OpenAlice | PyCryptoBot, PowerTrader | Config now covers persistence, server, scheduler, and risk |
| Event log | Implemented | OpenAlice | PowerTrader analytics mindset | JSONL append-only event durability remains central |
| Unified account model | Implemented | OpenAlice UTA | PowerTrader account-centric operation | Accounts remain the main execution boundary |
| Exchange capability vocabulary | Implemented | CCXT | BBGO | Capability-driven contract remains intact |
| Exchange registry | Implemented | BBGO | OpenAlice broker registry concept | Factory-based registration in place |
| Paper exchange adapter | Implemented | BBGO | PyCryptoBot paper/safe-first mindset | First safe adapter remains the execution target |
| Guard pipeline framework | Implemented | OpenAlice | PowerTrader risk controls | General pipeline in place |
| Symbol whitelist guard | Implemented | OpenAlice | PowerTrader practical safeguards | First concrete symbol policy enforcement |
| Max notional guard | Implemented | OpenAlice | PowerTrader position/risk bounds | First concrete monetary policy enforcement |
| Execution service | Implemented | BBGO | OpenAlice service composition | Real account -> guard -> adapter -> persistence flow exists |
| Execution repository | Implemented | platform state management patterns | BBGO runtime direction | In-memory order state now exists |
| Order journal | Implemented | PowerTrader journal mindset | OpenAlice durability | Orders persist independently of events |
| Portfolio tracker | Implemented | BBGO runtime direction | PowerTrader analytics intuition | First internal position state exists |
| Snapshot persistence | Implemented | OpenAlice | PowerTrader journaling/dashboards | Bootstrap snapshots continue to persist |
| Health/readiness endpoints | Implemented | cloud-native ops conventions | PowerTrader operator UX | Minimal operator-facing API surface exists |
| HTTP runtime wrapper | Implemented | platform ops patterns | OpenAlice connector/runtime thinking | Server lifecycle shell exists |
| Market data interface | Implemented | BBGO | CCXT, WolfBot | Abstraction exists |
| Paper market data feed | Implemented | BBGO | PowerTrader practical bootstrap needs | Deterministic local feed supports strategy development |
| Strategy runtime | Implemented | BBGO | WolfBot signal/event chaining | Signal aggregation runtime exists |
| Demo strategy (`buy-once`) | Implemented | iterative bootstrap design | PowerTrader practical development | Simple proof-of-life strategy |
| Market-data-aware strategy (`price-threshold`) | Implemented | BBGO | WolfBot, PowerTrader | First strategy that actually consumes feed data |
| Strategy scheduler | Implemented | WolfBot event flow | BBGO runtime thinking | Converts signals into execution requests |
| Recurring scheduler service | Implemented scaffold | WolfBot event loop direction | BBGO daemon trajectory | Available for future long-running mode |
| Backtesting | Not yet implemented | BBGO | WolfBot, PowerTrader | Deferred |
| Optimization | Not yet implemented | BBGO | WolfBot | Deferred |
| Arbitrage engine | Not yet implemented | WolfBot | kelvinau, ericjang, polymarket repos | Later advanced module |
| Notifications | Not yet implemented in Go | PowerTrader | BBGO, OpenAlice | Still reference-only |
| Dashboard / operator UI | Not yet implemented in Go | PowerTrader | BBGO, OpenAlice | Health/readiness only so far |
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
- deterministic paper-first development.

### Why this matters
This matrix continues to guard the project against random feature sprawl. New work should still be justified against:
1. the target architecture,
2. audited inspirations,
3. convergence value for the unified Go platform.

## Recommended use
Update this matrix whenever a major subsystem is added or materially strengthened.
