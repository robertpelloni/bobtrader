# Go Phase-4 Risk, Portfolio, and Scheduler Loop

## Summary
This phase strengthens the Go ultra-project by adding real policy enforcement, in-memory execution/portfolio state, a market-data-aware demo strategy, and the first recurring scheduler service abstraction.

## Delivered
### Risk system
- concrete `max-notional` guard
- concrete `symbol-whitelist` guard
- app wiring now constructs the guard pipeline from configuration

### Execution state
- in-memory execution repository for recent order state
- in-memory portfolio tracker that updates positions from executions
- execution service now updates repository and portfolio tracker after successful order placement

### Strategy evolution
- market-data-aware `price-threshold` demo strategy that consumes the paper market-data feed
- scheduler loop service abstraction for future recurring execution cadence

### Configuration expansion
- scheduler configuration (`enabled`, `interval_ms`)
- risk configuration (`max_notional`, `allowed_symbols`)

### App integration
- app now wires:
  - risk guards from config,
  - execution repository,
  - portfolio tracker,
  - market-data-aware strategy,
  - scheduler service abstraction
- startup still runs a deterministic one-shot scheduling pass, while the recurring scheduler service is available for future long-running mode

## Architectural significance
This phase moves the project from “a paper trading loop exists” to “that loop is now policy-aware and starts accumulating internal state.”

The most important new idea is that the system no longer treats execution as fire-and-forget. It now has:
- policy enforcement before execution,
- internal order memory after execution,
- position/portfolio updates after execution.

That is the beginning of a real trading state model.

## Influence mapping
### OpenAlice influence
- policy/guard-first execution
- account-centric orchestration
- explicit service graph

### BBGO influence
- kernel-oriented stateful runtime trajectory
- exchange/strategy separation
- groundwork for recurrent scheduler operation

### PowerTrader influence
- practical operator-first persistence and visibility
- strong emphasis on safeguards before trade placement

### WolfBot influence
- richer strategy/event loop direction
- path toward recurring evaluation and stateful strategy operation

## Validation
The following checks succeeded inside `ultratrader-go/`:
- `gofmt -w ./cmd ./internal`
- `go test ./...`
- `go run ./cmd/ultratrader`

## What remains deferred
- persistent portfolio database/state store
- recurring loop integration tests
- richer market-data subscription model
- fill reconciliation
- position PnL and valuation
- structured logging with correlation IDs
- real exchange adapters beyond paper

## Recommended next steps
1. Add structured logger package.
2. Add recurring scheduler lifecycle tests and runtime mode selection.
3. Add valuation and PnL to the portfolio tracker using market data.
4. Add richer risk guards (cooldown, account exposure, duplicate suppression).
5. Add market-data subscription/events rather than only point-in-time reads.
6. Add a simple HTTP portfolio/status endpoint.
