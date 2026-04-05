# Go Ultra-Project Program Plan

## Program Objective
Create a single Go-based trading platform by systematically assimilating the strongest ideas and subsystem capabilities from the audited submodule set while preserving legal cleanliness, architectural coherence, and maintainability.

## Program Strategy
This will be executed as a long-running phased program, not a one-shot rewrite.

## Phase Map

### Phase 0 — Audit and Decision Framing
**Status:** In progress

Deliverables:
- submodule inventory,
- architecture audit,
- best-architecture decision,
- best-kernel decision,
- licensing constraints,
- target architecture,
- migration plan.

Exit criteria:
- documented decision to use BBGO-style Go kernel + OpenAlice-style architecture.

### Phase 1 — New Go Repository Skeleton
Deliverables:
- `cmd/`, `internal/`, `web/`, `api/` scaffolding,
- config loader,
- logger,
- event log,
- app runtime,
- account abstraction,
- exchange capability interfaces,
- empty strategy/risk/execution module boundaries.

Exit criteria:
- runnable Go service with tests and docs.

### Phase 2 — Trading Kernel Foundation
Deliverables:
- balances,
- positions,
- orders,
- account lifecycle,
- exchange adapter contract,
- at least one real adapter,
- paper broker.

Primary inspirations:
- BBGO,
- OpenAlice UTA,
- PowerTrader trader state.

Exit criteria:
- account can place paper orders through the unified execution path.

### Phase 3 — Market Data and Persistence
Deliverables:
- ticker/trade/candle/orderbook streams,
- historical storage,
- append-only event journal,
- trade journal,
- snapshots,
- cache/replay support.

Primary inspirations:
- BBGO streams,
- CCXT normalization,
- OpenAlice event log,
- PowerTrader analytics journal.

Exit criteria:
- deterministic local replay and persistent audit trail.

### Phase 4 — Strategy Engine and Indicators
Deliverables:
- strategy runtime,
- built-in strategies,
- indicator framework,
- multi-timeframe support,
- event chaining,
- simulation support.

Primary inspirations:
- BBGO strategies,
- WolfBot event-based chaining,
- PowerTrader signal logic.

Exit criteria:
- at least 3 built-in strategies running in both paper and backtest modes.

### Phase 5 — Risk and Guard Pipeline
Deliverables:
- max position size guard,
- cooldown guard,
- whitelist guard,
- exposure/concentration guard,
- execution preflight pipeline,
- policy config.

Primary inspirations:
- OpenAlice guards,
- PowerTrader risk manager.

Exit criteria:
- all executions flow through auditable guards.

### Phase 6 — Backtesting and Optimization
Deliverables:
- candle replay engine,
- slippage/fee model,
- optimization framework,
- reporting,
- parameter search.

Primary inspirations:
- BBGO optimizer,
- WolfBot parameter optimization,
- PowerTrader backtester.

Exit criteria:
- strategy can be backtested and optimized reproducibly.

### Phase 7 — Advanced Trading Modules
Deliverables:
- arbitrage engine,
- TWAP/VWAP execution,
- portfolio rebalance,
- optional lending/funding,
- optional prediction-market adapters.

Primary inspirations:
- WolfBot,
- BBGO,
- selected specialized submodules.

Exit criteria:
- advanced modules are isolated, optional, and production-testable.

### Phase 8 — Operator Experience
Deliverables:
- web dashboard,
- CLI management,
- notifications,
- account views,
- logs and charts.

Primary inspirations:
- PowerTrader hub/dashboard concepts,
- BBGO dashboard,
- OpenAlice connectors and UI organization.

Exit criteria:
- one operator can configure and supervise the full system from a single interface.

### Phase 9 — AI and Research Modules
Deliverables:
- research assistant layer,
- news ingestion,
- sentiment hooks,
- optional tool/agent runtime,
- explanation workflows.

Primary inspirations:
- OpenAlice,
- PowerTrader AI overlays,
- AI-related submodules where legally safe.

Exit criteria:
- AI assists the operator but remains optional to the core runtime.

## Assimilation Order by Source Project

### Group A — Immediate reference systems
1. `c9s/bbgo`
2. `TraderAlice/OpenAlice`
3. `ccxt/ccxt`
4. `Ekliptor/WolfBot`
5. `whittlem/pycryptobot`

### Group B — Secondary feature mines
6. `ArsenAbazian/CryptoTradingFramework`
7. `Bohr1005/xcrypto`
8. `saniales/golang-crypto-trading-bot`
9. `RobertMarcellos/polymarket-copy-trading-bot`
10. `kelvinau/crypto-arbitrage`
11. `AdeelMufti/CryptoBot`
12. `asavinov/intelligent-trading-bot`

### Group C — Narrow or specialized references
- polymarket agents/bots,
- TradingView webhook/API tools,
- copy-trading bots,
- exchange-platform/full-stack systems,
- tutorial and strategy repositories.

These should be assimilated only after the kernel is stable.

## Ranking Logic
Each source project should be evaluated for assimilation on five axes:
- architecture quality,
- implementation maturity,
- feature value,
- portability to Go,
- licensing safety.

### Assimilation policy
- High architecture + high portability + acceptable license → implement early.
- High feature value + weak architecture → mine ideas, re-home into strong modules.
- No license / restrictive license → use only as behavioral reference unless explicit legal clearance is obtained.

## Major Risks
- accidental multi-license contamination,
- reproducing multiple overlapping subsystems instead of converging,
- dragging in exchange-specific abstractions too early,
- overbuilding AI before the trading kernel is stable,
- allowing the project to become a monolith.

## Risk Mitigations
- maintain explicit subsystem boundaries,
- implement one feature family at a time,
- require tests and docs per assimilation wave,
- prefer interface-first design,
- keep a feature matrix to avoid duplicate implementation.

## Definition of Done for the Program
The Go ultra-project is successful when it has:
- a coherent Go codebase,
- one account model,
- one exchange abstraction model,
- one strategy engine,
- one risk pipeline,
- one backtest engine,
- one persistence/event model,
- one dashboard/operator story,
- a documented feature matrix showing which legacy/submodule capabilities were absorbed.

## Immediate Next Steps
1. Finalize comparative submodule audit document.
2. Commit current documentation and submodule organization.
3. Create the new Go project scaffold in a separate top-level directory.
4. Start Phase 1 with config, runtime, event log, and account abstractions.
