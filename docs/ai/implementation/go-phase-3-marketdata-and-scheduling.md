# Go Phase-3 Market Data, Journaling, and Scheduling

## Summary
This phase pushes the Go ultra-project from a kernel-service skeleton toward a minimally self-driving platform shape by adding order journaling, market-data interfaces, a paper market-data feed, strategy scheduling, and a first demo strategy that executes through the real execution service.

## Delivered
### Persistence
- append-only order journal store under `internal/persistence/orders`
- execution service now persists placed orders into the order journal

### Market data
- market-data types and feed interface under `internal/marketdata`
- first paper market-data feed with deterministic prices for bootstrap symbols

### Strategy system
- enriched strategy signal model to include quantity and order type
- first demo strategy: `demo-buy-once`
- first scheduler package that converts signals into executable order intents

### App integration
- app now wires:
  - order journal,
  - paper market-data feed,
  - demo strategy,
  - scheduler,
  - optional HTTP runtime
- startup now runs the scheduler once, causing the demo strategy to place a paper order and persist it

### Tests
- order journal tests
- paper market-data feed tests
- demo strategy tests
- scheduler tests
- execution test now validates order journaling
- app integration test now validates event, snapshot, and order persistence together

## Architectural significance
This is the first phase where the Go project stops being purely infrastructural and starts behaving like a primitive trading application. The system now has a complete bootstrap loop:

1. app starts
2. account snapshot is persisted
3. demo strategy emits a signal
4. scheduler converts signal to order request
5. execution service routes it through the exchange registry and paper adapter
6. event log and order journal are updated

That loop is intentionally simple, but it is the first full closed path from strategy to persisted trade artifact.

## Influence mapping
### BBGO influence
- strategy-to-execution kernel framing
- adapter-mediated execution
- Go-native service modularity

### OpenAlice influence
- orchestrated service graph
- account-centered boundaries
- event durability as a first-class concern

### PowerTrader influence
- journaling and operator visibility
- practical incremental buildout from a retail-operator perspective

### WolfBot influence
- strategy-event pipeline direction
- preserving a path toward richer scheduler and multi-timeframe logic

## Why the demo strategy matters
The `demo-buy-once` strategy is intentionally tiny, but it proves an important implementation truth:
- strategies can be modeled independently of execution,
- schedulers can translate signals into orders,
- execution can remain behind account/risk/exchange abstractions,
- durable artifacts can be produced without external infrastructure.

## What remains deferred
- real market-data streaming
- order reconciliation
- portfolio and PnL engine
- recurring strategy scheduler/timer loop
- exchange-backed market adapters
- risk guard implementations beyond empty pipeline
- full HTTP server lifecycle tests

## Recommended next steps
1. Add market-data subscription/event interfaces.
2. Add order reconciliation and in-memory execution repository.
3. Add recurring scheduler loop with tick cadence.
4. Add portfolio state and position tracking.
5. Add first real risk guards.
6. Add a richer demo strategy that consumes the market-data feed.
