# UltraTrader Go

Phase-1 and Phase-2 scaffold for the planned unified Go trading platform.

## Current scope
This scaffold now establishes the first stable foundation for the future system:
- application runtime
- config loading
- append-only event log
- unified trading account model
- exchange capability interfaces
- exchange registry
- paper exchange adapter
- guard pipeline contracts
- execution service
- account snapshot store
- health/readiness HTTP handlers
- strategy runtime skeleton

## Planned role in the repo
`ultratrader-go/` is the clean-room destination for the long-term Go ultra-project documented in:
- `docs/ai/requirements/go-ultra-project-requirements.md`
- `docs/ai/design/go-ultra-project-architecture.md`
- `docs/ai/planning/go-ultra-project-program-plan.md`

## Run
```bash
go run ./cmd/ultratrader
```

## Test
```bash
go test ./...
```
