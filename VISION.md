# VISION

## Ultimate Vision
This repository is evolving from a single Python trading bot into a **unified trading systems workspace** with two simultaneous goals:

1. **Preserve and understand the existing PowerTrader AI system** in all of its practical, operator-facing detail.
2. **Build a next-generation Go ultra-project** that clean-room assimilates the strongest architecture, execution patterns, analytics ideas, and operator workflows from the imported reference corpus.

The end state is intended to be a highly capable, modular, observable, and operator-friendly trading platform that combines:
- robust exchange abstractions,
- safe execution pipelines,
- strategy runtime composition,
- rich analytics and journaling,
- operator-facing diagnostics and dashboards,
- long-term AI/research augmentation,
- strong documentation and maintainability discipline.

## Strategic Direction
The project is intentionally not pursuing a naive source merge of all imported systems.
Instead, the target direction is:
- **clean-room reimplementation**,
- **phased subsystem assimilation**,
- **architecture-first convergence**,
- **operator visibility and safety at every stage**.

## Architectural Thesis
The current strategic thesis remains:
- **Best architecture reference:** `TraderAlice/OpenAlice`
- **Best practical Go kernel reference:** `c9s/bbgo`
- **Best exchange abstraction reference:** `ccxt/ccxt`
- **Best feature mine for advanced capabilities:** `Ekliptor/WolfBot`

This means the intended final system should feel like:
- an **OpenAlice-style platform architecture**,
- powered by a **BBGO-style Go trading kernel**,
- informed by **CCXT-style capability realism**,
- enriched by **WolfBot-style advanced execution and strategy ideas**,
- while retaining the **practical operator sensibility of PowerTrader AI**.

## Product Characteristics the Final System Should Have
### Trading kernel
- exchange/broker abstraction
- spot first, extensible to margin/futures where safe
- reliable order placement and reconciliation
- market-data pull and stream support
- paper trading and deterministic simulation modes

### Runtime safety
- policy-before-execution
- multiple guard classes
- account-scoped controls
- portfolio-aware exposure limits
- temporal protections against duplicate/cascade execution

### State and analytics
- event log
- trade journal
- order journal
- account snapshots
- persistent runtime reports
- portfolio valuation
- realized/unrealized PnL
- metrics and diagnostics

### Strategy system
- modular strategies
- schedulers and eventually event-driven subscriptions
- multi-timeframe and signal composition over time
- backtesting and optimization as first-class subsystems

### Operator surfaces
- health/readiness
- structured logs
- diagnostics APIs
- eventually dashboards and rich UI
- clear config and help surfaces

### AI and research augmentation
- optional rather than mandatory
- tool-driven, domain-separated, explainable
- built atop a stable trading/runtime core instead of replacing it

## Current Implementation Reality
The Go ultra-project is currently in progressive staged construction under:
- `ultratrader-go/`

It already contains:
- runtime composition root,
- config loading,
- structured logging,
- event log,
- exchange registry,
- paper exchange adapter,
- risk pipeline and multiple concrete guards,
- execution service,
- execution repository,
- order and snapshot persistence,
- persistent runtime reports,
- market-data interfaces plus a paper stream/feed,
- strategy runtime and scheduler,
- portfolio state, valuation, and PnL,
- diagnostics APIs.

## High-Level Endgame
The desired endgame is a system where:
- operators can run safe paper or live trading workflows,
- all state transitions are inspectable,
- all core behaviors are documented,
- strategy execution is testable and replayable,
- exchange integration is modular,
- the service is daemon-ready and operationally transparent,
- AI/research layers can assist without destabilizing the kernel.

## Design Priorities
In priority order:
1. correctness and safety
2. observability
3. architecture quality
4. maintainability
5. composability
6. operator usefulness
7. feature breadth

## Anti-Goals
The project should avoid becoming:
- a monolithic bot script,
- an opaque black box,
- a license-confused source paste-up,
- a UI without trustworthy runtime state,
- an AI wrapper around an unstable execution core.

## Long-Term Success Definition
The project succeeds when it becomes:
- a robust Go trading platform,
- a well-documented operator system,
- a convergence point for the best ideas in the imported corpus,
- and a maintainable long-horizon foundation for future trading, analytics, and research work.
