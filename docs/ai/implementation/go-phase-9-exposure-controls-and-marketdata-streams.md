# Go Phase-9 Exposure Controls and Market-Data Streams

## Summary
This phase expands the Go ultra-project in three important directions:
- portfolio-aware exposure controls,
- operator-visible guard diagnostics,
- first market-data streaming/subscription abstraction for future daemon behavior.

## Delivered
### Exposure and portfolio-aware risk
- added `max-concentration` guard skeleton infrastructure support through portfolio value methods
- added `max-open-positions` guard using live portfolio state
- portfolio tracker now exposes:
  - `HasOpenPosition()`
  - `OpenPositionCount()`
  - `CurrentValue()`
  - `TotalValue()`

### Guard diagnostics
- `/api/guards` endpoint added
- app now exposes active guard names through the diagnostics API and startup logs

### Market-data streams
- market-data package now defines streaming subscription concepts
- paper market-data feed implements `SubscribeTicks()` for local deterministic tick streaming

### Runtime lifecycle
- HTTP runtime now binds via an explicit listener and exposes the resolved address
- runtime shutdown remains test-covered and development-safe with ephemeral ports

## Architectural significance
This phase is important because it further shifts the Go system from request/response bootstrap execution toward daemon-ready behavior.

The most meaningful upgrades are:
- risk can now reason about existing portfolio structure,
- operators can inspect which guards are live,
- market-data abstractions now include a pathway for push-style/event-driven evolution.

## Influence mapping
### OpenAlice influence
- introspectable runtime behavior
- explicit service lifecycle and policy visibility

### PowerTrader influence
- practical operator-facing diagnostics
- portfolio-aware guardrails

### BBGO influence
- movement toward event-driven market-data consumption and daemon operation

### WolfBot influence
- stronger runtime sensitivity to state over time and repeated evaluation paths

## Validation
Inside `ultratrader-go/` the following checks passed:
- `gofmt -w ./cmd ./internal`
- `go test ./...`
- `go run ./cmd/ultratrader`

## What remains deferred
- full concentration guard wiring into the app pipeline
- persistent metrics/valuation history
- richer execution diagnostics such as block reasons and rates over time
- full app shutdown integration tests across runtime + scheduler + logging + background stream consumers

## Recommended next steps
1. Wire max-concentration guard into app config and runtime.
2. Add persistent runtime report history for metrics and valuation snapshots.
3. Add block-reason diagnostics and rates.
4. Add app shutdown integration tests spanning all lifecycle-managed components.
5. Add richer market-data event consumers and strategy subscription hooks.
