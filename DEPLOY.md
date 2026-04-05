# DEPLOY

## Scope
This file captures the latest known deployment and execution instructions for the current repository, including:
- the legacy Python PowerTrader AI system,
- the new Go ultra-project scaffold/runtime under `ultratrader-go/`.

## 1. Legacy Python PowerTrader AI
### Local start
From the repo root:
```bash
python pt_hub.py
```

### Known runtime requirements
- Python environment with dependencies installed
- GUI-capable environment for Tkinter hub
- additional runtime files may be required for live trading, such as API credentials and config files

### Current notes
The legacy Python system is feature-rich but unevenly integrated. Refer to existing docs such as:
- `README.md`
- `MANUAL.md`
- `NOTIFICATIONS_README.md`
- `ROADMAP.md`

## 2. Go Ultra-Project (`ultratrader-go/`)
### Purpose
The Go runtime is the clean-room destination for the future unified trading platform.

### Run
```bash
cd ultratrader-go
go run ./cmd/ultratrader
```

### Test
```bash
cd ultratrader-go
go test ./...
```

### Current default behavior
By default the Go runtime:
- starts with development config defaults,
- uses paper exchange + paper market data,
- uses an ephemeral localhost HTTP bind (`127.0.0.1:0`),
- performs one startup scheduler pass,
- writes startup artifacts to `data/` under the Go module.

### Current persisted files
Under `ultratrader-go/data/` the runtime may create:
- event log
- snapshot log
- order journal
- runtime report log
- structured app log

### Current diagnostics APIs
The runtime currently exposes handler routes including:
- `/healthz`
- `/readyz`
- `/api/status`
- `/api/portfolio`
- `/api/orders`
- `/api/execution-summary`
- `/api/metrics`
- `/api/guards`

### Config behavior
If no config file is provided, defaults are used.
To run with a config file:
```bash
cd ultratrader-go
go run ./cmd/ultratrader --config path/to/config.json
```

## 3. Deployment Recommendations
### Current best practice
For now, treat `ultratrader-go/` as:
- a local development runtime,
- a paper-trading validation environment,
- an evolving foundation for later daemon/service deployment.

### Not yet recommended for
- real-money production deployment
- unsupervised live trading
- external public exposure without further hardening

## 4. Pre-Deployment Checklist
Before broader deployment, verify:
- `go test ./...` passes
- structured logs are being written as expected
- diagnostics APIs return valid data
- runtime reports are being persisted
- guard configuration matches intended risk posture
- scheduler settings are appropriate
- server bind address is intentional

## 5. Future Deployment Work
Still needed before stronger deployment guidance:
- coordinated app shutdown tests
- persistent metrics/valuation history
- deeper diagnostics and reporting
- stronger risk policies
- real exchange adapters beyond paper mode
- deployment packaging and environment-specific configs

## 6. Notes
Do **not** use destructive process-kill commands that could terminate the coding session or unrelated local services.
