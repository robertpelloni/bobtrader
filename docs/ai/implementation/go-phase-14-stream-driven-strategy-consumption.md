# Go Phase-14 Stream-Driven Strategy Consumption

## Summary
This phase begins the transition from purely timer-driven evaluation toward event-driven runtime behavior.

## Delivered
- Added scheduler stream service that subscribes to market-data ticks and triggers strategy evaluation on incoming events.
- Extended scheduler configuration with a `mode` field so the runtime can choose between:
  - `timer`
  - `stream`
- Wired the app to start either the timer scheduler service or the stream scheduler service based on config.

## Architectural significance
This is an important milestone because it moves the Go runtime closer to the kind of event-driven behavior seen in more advanced trading systems. Instead of only evaluating strategies on a fixed wall-clock schedule, the runtime can now evolve toward reacting directly to incoming market-data events.

This does not replace the timer path; it complements it.

## Validation
Inside `ultratrader-go/` the following checks passed:
- `gofmt -w ./cmd ./internal`
- `go test ./...`
- `go run ./cmd/ultratrader`

## Recommended next steps
1. Add richer paper stream simulation patterns.
2. Add stream-driven strategies that consume tick streams more directly.
3. Add recurring lifecycle integration tests with active stream mode.
4. Add backpressure / debounce / tick-coalescing considerations if stream intensity increases.
