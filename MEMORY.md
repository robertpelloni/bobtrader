# MEMORY

## Ongoing Observations and Learned Preferences

### 1. Architectural selection memory
- Best architecture among imported projects: `TraderAlice/OpenAlice`
- Best practical Go kernel reference: `c9s/bbgo`
- Best exchange abstraction reference: `ccxt/ccxt`
- Best advanced feature mine: `Ekliptor/WolfBot`

### 2. Implementation strategy memory
- Do **not** blindly merge code from imported projects.
- Prefer **clean-room reimplementation** guided by behavior, architecture, and interface ideas.
- Use phased assimilation rather than giant rewrites.

### 3. Licensing memory
Imported repos contain mixed licenses including AGPL, GPL, CC-BY-NC, shareware-like, MIT, Apache, and no-license cases.
Therefore:
- architecture and behavior can be studied broadly,
- direct source reuse must be treated conservatively,
- the Go ultra-project should be developed as an original codebase.

### 4. Operator preference memory
The user strongly prefers:
- continuous autonomous progress,
- frequent commit/push cadence,
- deep documentation,
- strong changelog/version discipline,
- broad roadmap and TODO visibility,
- no process-killing behavior,
- preserving all existing progress.

### 5. Runtime safety memory
Do **not** mass-kill node or other processes.
Do not use global destructive commands that could terminate the coding session or unrelated services.

### 6. Repo hygiene memory
The repo often contains unrelated generated/untracked runtime artifacts.
When committing, prefer tightly scoped commits for:
- `ultratrader-go/`
- documentation files
- version/changelog/handoff files

Avoid sweeping in unrelated runtime debris.

### 7. Current Go ultra-project state memory
The Go project already has:
- config
- structured logging
- event log
- exchange registry
- paper exchange adapter
- market-data interfaces + paper stream/feed
- risk pipeline with multiple guards
- execution service
- execution repository
- order journal
- snapshot store
- runtime report store
- portfolio valuation + PnL
- metrics
- diagnostics APIs
- runtime lifecycle controls

### 8. Immediate next recommended directions memory
Strong next candidates:
- persistent metrics / valuation history
- richer guard diagnostics (details/reasons)
- exposure/concentration enforcement with live market values
- stream-driven strategy consumption
- full coordinated app shutdown tests
- reporting/analytics layers

### 9. Documentation preference memory
Maintain and update:
- `CHANGELOG.md`
- `VERSION.md`
- `ROADMAP.md`
- `TODO.md`
- `HANDOFF.md`
- `VISION.md`
- `MEMORY.md`
- `DEPLOY.md`

### 10. Versioning memory
Every meaningful build/session should increment the version and mention the bump in the commit message.

### 11. Runtime operations memory
- Using `127.0.0.1:0` as the default Go HTTP bind address avoids frequent local port-collision failures during repeated development/test runs.
- Logger/file-backed components need explicit cleanup paths to keep tests reliable on Windows.
- Runtime summary persistence is now useful as a bridge between raw journals and future analytics/reporting layers.
- Report readback (`Latest`, `LatestByType`) is the next key step after simple persistence because it turns durable artifacts into something the runtime and operator APIs can actually consume.
- Live-valued exposure views are preferable to cost-basis-only exposure estimates when preparing concentration controls.
