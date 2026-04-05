# TODO

## Immediate Priorities for the Go Ultra-Project

### Runtime control and diagnostics
- [ ] Add richer guard diagnostics beyond guard names
- [ ] Expose block reasons and guard-trigger counts
- [ ] Add coordinated app shutdown tests spanning runtime + scheduler + logger + stream subscriptions
- [ ] Add richer execution diagnostics including success/block rates and symbol concentration summaries

### Risk management
- [ ] Fully wire `max-concentration` guard using live market value rather than cost-basis fallback
- [ ] Add exposure / concentration diagnostics endpoints
- [ ] Add max-open-position and concentration policy tuning docs/examples
- [ ] Add additional guards for duplicate side suppression / max notional per symbol / account exposure

### Market data
- [ ] Add stream-driven strategy consumption path
- [ ] Add event/subscription interfaces beyond simple tick subscription
- [ ] Add richer paper stream simulation patterns
- [ ] Add integration of stream-fed strategies into scheduler/runtime lifecycle

### Analytics and reporting
- [ ] Add persistent metrics history
- [ ] Add persistent valuation / PnL history
- [ ] Add runtime analytics/reporting modules on top of report storage
- [ ] Add execution summary history over time

### Trading engine expansion
- [ ] Add richer strategy library
- [ ] Add backtesting subsystem
- [ ] Add optimization subsystem
- [ ] Add real exchange adapters beyond paper mode
- [ ] Add reconciliation/fill state improvements

### Operator surfaces
- [ ] Add richer operator-facing diagnostics APIs
- [ ] Add portfolio summary endpoint distinct from raw positions
- [ ] Add UI/dashboard layer for Go runtime
- [ ] Add deployment packaging and environment profiles

## Legacy PowerTrader AI Documentation / Cleanup
- [ ] Reconcile stale roadmap/module inventory docs with actual repo state
- [ ] Add clearer distinction between legacy Python runtime and Go ultra-project workstream
- [ ] Audit and document partially integrated Python features more systematically

## Repo-wide Documentation
- [ ] Keep `VISION.md`, `MEMORY.md`, `DEPLOY.md`, `ROADMAP.md`, `TODO.md`, `HANDOFF.md`, `CHANGELOG.md`, `VERSION.md` synchronized
- [ ] Continue expanding submodule/reference documentation as the Go runtime assimilates new ideas
