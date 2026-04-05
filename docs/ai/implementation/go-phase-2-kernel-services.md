# Go Phase-2 Kernel Services

## Summary
This phase extends the initial Go scaffold into a more recognizable trading-kernel skeleton by introducing the first real service graph around accounts, exchanges, execution, snapshots, health endpoints, and strategy runtime.

## Delivered
### Exchange subsystem
- exchange registry for adapter lookup and factory-based instantiation
- first clean-room paper trading adapter
- registry tests and paper-adapter tests

### Execution subsystem
- execution service that ties together:
  - accounts,
  - exchange registry,
  - guard pipeline,
  - event log
- successful execution path emits an event-log entry
- execution service test validates end-to-end paper order placement through the service layer

### Persistence
- append-only snapshot store for account-state journaling
- snapshot store test verifies durable write behavior

### Service interface layer
- health/readiness HTTP handler package with `/healthz` and `/readyz`
- endpoint test coverage added

### Strategy foundation
- strategy runtime skeleton with signal aggregation contract
- runtime test coverage added

### App integration
- `App` now wires:
  - event log,
  - snapshot store,
  - account service,
  - exchange registry,
  - execution service,
  - strategy runtime,
  - health/readiness handler
- startup now writes bootstrap account snapshots
- app test validates startup event and snapshot creation

## Architectural significance
This phase transforms the project from a static scaffold into the beginning of a service-oriented trading kernel. The important shift is that the system now has the first real path from:

account -> adapter lookup -> guard pipeline -> order placement -> event emission

That path is still intentionally small, but it is the first irreversible proof that the target architecture can be realized in Go with clean subsystem boundaries.

## Influence mapping
### BBGO influence
- exchange-oriented kernel thinking
- adapter-driven trading execution
- Go-native runtime shape
- early service modularization

### OpenAlice influence
- account-centric system boundary
- service layering and orchestration mindset
- future event-driven platform direction
- platform-like decomposition rather than bot-script decomposition

### PowerTrader influence
- journaling mindset
- risk/guard emphasis
- operator-facing health/readiness importance
- persistence-first observability

### CCXT influence
- capability vocabulary and adapter realism
- exchange heterogeneity acknowledged at the interface level

### WolfBot influence
- strategy runtime and signal flow awareness
- preserving a path toward richer event-driven strategy chaining later

## Validation
The following checks were run successfully inside `ultratrader-go/`:
- `gofmt -w ./cmd ./internal`
- `go test ./...`
- `go run ./cmd/ultratrader`

## What remains intentionally deferred
- real exchange adapters beyond paper mode
- order persistence model
- full account reconciliation
- portfolio engine
- websocket market data
- strategy scheduling
- broker capability enforcement
- dashboard/API beyond minimal health endpoints
- execution correlation IDs and structured logging

## Recommended next steps
1. Add execution repository and order journal.
2. Add account snapshot builder from adapter balances.
3. Add structured logger package and execution correlation IDs.
4. Add market data interfaces and paper ticker/candle feed.
5. Add strategy scheduler and binding of strategies to accounts.
6. Add first built-in demo strategy using the execution service.
7. Expose the health/readiness handler through a controllable HTTP server runtime.
