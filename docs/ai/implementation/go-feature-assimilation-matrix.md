# Go Feature Assimilation Matrix

## Purpose
Track how the emerging `ultratrader-go/` codebase maps to the audited source projects and to the long-term Go ultra-project plan.

## Current Status
The project has moved from pure scaffold into an early self-driving kernel shape. It now includes the first persisted end-to-end strategy execution path using a paper exchange.

## Matrix

| Target subsystem | Current status | Primary source inspiration | Secondary source inspiration | Notes |
|---|---|---|---|---|
| App runtime / composition root | Implemented | OpenAlice | BBGO | Clear runtime assembly is modeled after platform-style orchestration rather than script startup |
| Config loading | Implemented | OpenAlice | PyCryptoBot, PowerTrader | Explicit defaults, file loading, and expanding subsystem config domains |
| Event log | Implemented | OpenAlice | PowerTrader analytics mindset | JSONL append-only event log remains a central durability primitive |
| Unified account model | Implemented | OpenAlice UTA | PowerTrader account-centric operation | Accounts remain the primary boundary of execution and policy |
| Exchange capability vocabulary | Implemented | CCXT | BBGO | Capability-driven contracts prevent false assumptions across venues |
| Exchange registry | Implemented | BBGO | OpenAlice broker registry concept | Factory-based registration provides a path for modular adapters |
| Paper exchange adapter | Implemented | BBGO | PyCryptoBot paper/safe-first mindset | First adapter enables safe end-to-end execution testing |
| Guard pipeline | Implemented scaffold | OpenAlice | PowerTrader risk controls | Framework exists; concrete guards remain to be added |
| Execution service | Implemented | BBGO | OpenAlice service composition | Real service path now exists and is tested |
| Order journal | Implemented | PowerTrader journal mindset | OpenAlice event durability | Orders are now persisted independently from events |
| Snapshot persistence | Implemented | OpenAlice | PowerTrader journaling/dashboards | Append-only bootstrap snapshots already work |
| Health/readiness endpoints | Implemented | cloud-native ops conventions | PowerTrader operator UX | Minimal HTTP status surface exists |
| HTTP runtime wrapper | Implemented | platform ops patterns | OpenAlice connector/service thinking | Runtime exists; richer server lifecycle can follow |
| Market data interface | Implemented scaffold | BBGO | CCXT, WolfBot | Feed abstraction now exists |
| Paper market data feed | Implemented | BBGO | PowerTrader practical bootstrap needs | Deterministic local feed enables strategy development without exchange dependencies |
| Strategy runtime | Implemented | BBGO | WolfBot signal/event chaining | Runtime aggregates signals from strategies |
| Demo strategy | Implemented | iterative bootstrap design | PowerTrader practical operator-first development | `demo-buy-once` proves strategy-to-execution integration |
| Strategy scheduler | Implemented | WolfBot event flow | BBGO runtime thinking | Converts strategy signals into execution requests |
| Backtesting | Not yet implemented | BBGO | WolfBot, PowerTrader | Deferred until market data and execution abstractions mature |
| Optimization | Not yet implemented | BBGO | WolfBot | Deferred |
| Arbitrage engine | Not yet implemented | WolfBot | kelvinau, ericjang, polymarket repos | Later advanced module |
| Notifications | Not yet implemented in Go | PowerTrader | BBGO, OpenAlice | Existing Python implementation is still reference-only |
| Dashboard / operator UI | Not yet implemented in Go | PowerTrader | BBGO, OpenAlice | Health/readiness is only the first operational foothold |
| AI / research layer | Not yet implemented in Go | OpenAlice | PowerTrader, AI-specific repos | Must remain optional and layered above the kernel |

## Interpretation
### Already established architectural identity
Even at this early stage, the new Go project demonstrates these chosen biases:
- account-centered orchestration,
- event-first durability,
- capability-driven exchange abstraction,
- policy-before-execution,
- modular service boundaries,
- scheduling separated from execution,
- persistence separated from adapters.

### Why this matters
This matrix exists to prevent the program from devolving into random feature accretion. Each new subsystem should still be justified against:
1. the target architecture,
2. the audited source inspirations,
3. the long-term convergence plan.

## Recommended use
Update this matrix whenever a major subsystem is introduced or materially revised.
