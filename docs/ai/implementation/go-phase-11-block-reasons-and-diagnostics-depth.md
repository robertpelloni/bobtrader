# Go Phase-11 Block Reasons and Diagnostics Depth

## Summary
This phase deepens the Go ultra-project's diagnostics model by teaching the runtime not only that an execution was blocked, but also *why* it was blocked. It also extends the project with richer exposure-control scaffolding and a first durable reporting layer suitable for future analytics.

## Delivered
### Block-reason diagnostics
- risk pipeline now returns structured `GuardError` values preserving the guard name
- execution metrics now retain block-reason counts by guard
- `/api/guard-diagnostics` now exposes active guards plus metrics-backed guard/block context

### Persistent reporting baseline
- runtime report storage remains in place and is now clearly established as the bridge between raw journals and future analytics modules
- app startup persists a durable `startup-summary` report containing core runtime summary data

### Exposure-control groundwork
- `max-concentration` guard added as a dedicated risk primitive
- portfolio tracker now exposes current value and total value helpers for portfolio-aware controls

## Architectural significance
This phase is important because operator trust depends on understanding not just what the system did, but why it refused to act. A blocked trade without a reason is noise; a blocked trade with a classified guard reason is an operationally meaningful signal.

The runtime is therefore getting closer to a system that can:
- explain policy outcomes,
- report supervisory state clearly,
- support future analytics over execution and risk behavior.

## Validation
Inside `ultratrader-go/` the following checks passed:
- `gofmt -w ./cmd ./internal`
- `go test ./...`
- `go run ./cmd/ultratrader`

## Recommended next steps
1. Add persistent metrics and valuation time-series history.
2. Add richer guard diagnostics endpoint content beyond names + counts.
3. Add stream-driven strategies that consume paper tick subscriptions.
4. Add coordinated full app shutdown integration tests.
5. Add concentration enforcement using live valued positions in the runtime loop.
