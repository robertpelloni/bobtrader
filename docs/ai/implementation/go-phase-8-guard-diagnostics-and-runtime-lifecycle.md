# Go Phase-8 Guard Diagnostics and Runtime Lifecycle

## Summary
This phase improves the Go ultra-project in two major operational areas:
- runtime lifecycle control and testability,
- richer diagnostics around guards, execution summaries, and portfolio-aware admission control.

## Delivered
### Runtime lifecycle
- HTTP runtime now manages its own listener and exposes:
  - `Address()`
  - `Shutdown()`
- runtime start/shutdown behavior is now integration-tested using a real ephemeral TCP listener
- default server address now uses `127.0.0.1:0` to avoid local port collisions during development and validation

### Guard and execution diagnostics
- risk pipeline now exposes configured guard names
- `/api/guards` endpoint added for runtime guard diagnostics
- max-open-positions guard introduced for portfolio-aware admission control
- execution summary surface remains available and now complements runtime metrics and PnL state

### Portfolio/risk integration
- portfolio tracker now exposes open-position checks and counts
- app uses portfolio-backed max-open-positions guard as part of the runtime pipeline

## Architectural significance
This phase is important because the system is now moving from “it can run” to “it can be controlled and interrogated like a real service.”

The most meaningful improvements are:
- HTTP runtime lifecycle is no longer implicit or opaque,
- guard configuration is now operator-visible,
- portfolio state is reusable as a direct risk-control input.

## Influence mapping
### OpenAlice influence
- service lifecycle discipline
- introspectable, controllable runtime behavior

### PowerTrader influence
- operator-centric diagnostics and runtime transparency

### BBGO influence
- daemon-grade service trajectory with explicit runtime control

## Validation
Inside `ultratrader-go/` the following checks passed:
- `gofmt -w ./cmd ./internal`
- `go test ./...`
- `go run ./cmd/ultratrader`

## What remains deferred
- persistent guard diagnostics history
- market-data event/subscription model
- advanced exposure/concentration policies beyond max-open-positions
- coordinated graceful shutdown tests spanning app + scheduler + HTTP runtime together
- richer execution rates and symbol concentration diagnostics over time

## Recommended next steps
1. Add exposure/concentration guards.
2. Add market-data event/subscription interfaces.
3. Add persistent valuation or metrics history.
4. Add full app shutdown integration tests.
5. Add richer execution diagnostics (rates, concentration, block reasons).
