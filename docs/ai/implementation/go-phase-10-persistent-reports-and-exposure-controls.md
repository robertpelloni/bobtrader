# Go Phase-10 Persistent Reports and Exposure Controls

## Summary
This phase expands the Go ultra-project in three strategic directions:
- persistent runtime reporting,
- deeper portfolio-aware exposure controls,
- first market-data stream subscription abstraction.

## Delivered
### Persistent runtime reporting
- new append-only runtime report store under `internal/persistence/reports`
- app now persists a startup summary report containing:
  - order count,
  - portfolio value,
  - realized/unrealized PnL,
  - metrics snapshot,
  - active guards

### Exposure controls
- added `max-concentration` guard scaffold and runtime integration support through portfolio value methods
- app now wires `max-open-positions` and concentration-aware primitives from config and portfolio state

### Market-data evolution
- market-data package now defines streaming subscription concepts
- paper market-data feed now supports `SubscribeTicks()` for deterministic local tick streaming

## Architectural significance
This phase is important because the Go runtime now produces durable summaries of its own operating state. That is a meaningful step toward historical analytics and operational replay.

The most meaningful upgrades are:
- runtime state is no longer only in memory or logs; it is also summarized durably,
- portfolio state is increasingly reusable as a first-class risk-control input,
- the market-data layer now contains the first path toward event-driven rather than only pull-driven behavior.

## Influence mapping
### OpenAlice influence
- durable operational history and introspectable runtime state
- platform-like service lifecycle discipline

### PowerTrader influence
- operator-valued reporting and practical summary persistence
- portfolio-aware runtime diagnostics

### BBGO influence
- movement toward event-driven market-data flow and daemon-grade runtime behavior

### WolfBot influence
- repeated evaluation and event-driven runtime direction

## Validation
Inside `ultratrader-go/` the following checks passed:
- `gofmt -w ./cmd ./internal`
- `go test ./...`
- `go run ./cmd/ultratrader`

## What remains deferred
- persistent metrics and valuation history beyond startup summaries
- block-reason diagnostics and concentration-rate summaries
- richer stream consumers and strategy subscription hooks
- full coordinated lifecycle tests across app + scheduler + HTTP runtime + stream consumers

## Recommended next steps
1. Add richer block-reason diagnostics.
2. Add stream-driven strategy consumption path.
3. Add persistent metrics and valuation time-series.
4. Add coordinated app lifecycle integration tests.
5. Add deeper exposure/concentration enforcement wiring.
