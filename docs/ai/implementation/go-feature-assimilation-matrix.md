# Go Feature Assimilation Matrix

## Purpose
Track how the emerging `ultratrader-go/` codebase maps to the audited source projects and to the long-term Go ultra-project plan.

## Current Status
The project is still in early kernel formation. The implemented code is intentionally foundational rather than feature-complete.

## Matrix

| Target subsystem | Current status | Primary source inspiration | Secondary source inspiration | Notes |
|---|---|---|---|---|
| App runtime / composition root | Implemented scaffold | OpenAlice | BBGO | Clear runtime assembly is modeled after platform-style orchestration rather than script startup |
| Config loading | Implemented scaffold | OpenAlice | PyCryptoBot, PowerTrader | Early emphasis on explicit configuration and defaults |
| Event log | Implemented scaffold | OpenAlice | PowerTrader analytics mindset | JSONL append-only design chosen as simplest durable first step |
| Unified account model | Implemented scaffold | OpenAlice UTA | PowerTrader account-centric operation | Accounts are the primary boundary of execution and policy |
| Exchange capability vocabulary | Implemented scaffold | CCXT | BBGO | Capability-driven contracts prevent false assumptions across venues |
| Exchange registry | Implemented scaffold | BBGO | OpenAlice broker registry concept | Factory-based registration is the starting point for adapters |
| Paper exchange adapter | Implemented scaffold | BBGO | PyCryptoBot paper/safe-first mindset | First adapter exists to unlock execution testing without external dependencies |
| Guard pipeline | Implemented scaffold | OpenAlice | PowerTrader risk controls | Execution must pass through ordered policy checks |
| Execution service | Implemented scaffold | BBGO | OpenAlice service composition | The first real service path now exists: account -> guard -> adapter -> event log |
| Snapshot persistence | Implemented scaffold | OpenAlice | PowerTrader journaling/dashboards | Current implementation is append-only JSONL; richer models can come later |
| Health/readiness endpoints | Implemented scaffold | cloud-native ops conventions | PowerTrader operator UX | Minimal HTTP status surface added before full API/dashboard work |
| Strategy runtime skeleton | Implemented scaffold | BBGO | WolfBot signal/event chaining | Current runtime only aggregates signals, but defines the extension seam |
| Market data | Not yet implemented | BBGO | CCXT, WolfBot | Planned next major subsystem |
| Backtesting | Not yet implemented | BBGO | WolfBot, PowerTrader | Deferred until market data and execution abstractions mature |
| Optimization | Not yet implemented | BBGO | WolfBot | Deferred |
| Arbitrage engine | Not yet implemented | WolfBot | kelvinau, ericjang, polymarket repos | Later advanced module |
| Notifications | Not yet implemented in Go | PowerTrader | BBGO, OpenAlice | Existing Python implementation is a reference only |
| Dashboard / operator UI | Not yet implemented in Go | PowerTrader | BBGO, OpenAlice | Health/readiness endpoints are the first operational foothold |
| AI / research layer | Not yet implemented in Go | OpenAlice | PowerTrader, AI-specific repos | Must stay optional and layered on top of the kernel |

## Interpretation
### Already established architectural identity
Even at this early stage, the new Go project is not arbitrary. It already demonstrates these chosen biases:
- account-centered orchestration,
- event-first durability,
- capability-driven exchange abstraction,
- policy-before-execution,
- modular service boundaries.

### Why this matters
This matrix is meant to prevent the program from degenerating into random feature accretion. Every new subsystem should be justified against:
1. the target architecture,
2. the relevant source inspirations,
3. the long-term convergence plan.

## Recommended use
Update this matrix whenever a major subsystem is introduced or materially revised.
