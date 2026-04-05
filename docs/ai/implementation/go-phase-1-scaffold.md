# Go Phase-1 Scaffold Implementation Notes

## Summary
This phase creates the first executable scaffold for the planned unified Go trading platform in `ultratrader-go/`.

## Delivered
- independent Go module: `ultratrader-go/`
- CLI entrypoint: `cmd/ultratrader`
- config package with defaults + JSON loading
- append-only JSONL event log package
- unified trading account model and account service
- exchange capability and adapter contracts
- guard pipeline contract
- unit tests for config loading, event log append, and guard pipeline behavior

## Why this matters
The previous documentation phase identified the target architecture and recommended BBGO-style kernel design with OpenAlice-style system organization. This scaffold turns that recommendation into a concrete project root that can now be expanded safely.

## Design choices
### Separate top-level module
A dedicated top-level Go module avoids coupling the new system to the legacy Python runtime and keeps the migration clean.

### Event log first
An append-only event log was prioritized because it is foundational for:
- observability,
- auditability,
- replay,
- future connector delivery,
- workflow/event orchestration.

### Account abstraction early
A unified account type was introduced immediately because account isolation is central to the long-term architecture.

### Capability-based exchange contracts
The exchange package starts with capabilities instead of assuming all venues support all features. This aligns with both the audit findings and the target architecture.

## Current limitations
This is intentionally a scaffold, not yet a trading engine. It does not yet include:
- live exchange adapters,
- order execution routing,
- market data streaming,
- persistence beyond event log,
- strategy runtime,
- dashboard/API surfaces.

## Next recommended implementation steps
1. add exchange registry + paper adapter,
2. add persistent account repository and snapshot model,
3. add execution service and order intents,
4. add market data interfaces,
5. add structured logging package,
6. add HTTP health/readiness server,
7. begin strategy engine skeleton.
