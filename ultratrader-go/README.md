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

With a config profile:
```bash
go run ./cmd/ultratrader --config config/development-timer.json
```

## Test
```bash
go test ./...
```

## Config Profiles
- `config/development-timer.json`
- `config/development-stream.json`
- `config/paper-service.json`

## Container
Build and run:
```bash
docker build -t ultratrader-go .
docker run --rm -p 8080:8080 ultratrader-go
```

Or with Compose:
```bash
docker compose up --build
```
