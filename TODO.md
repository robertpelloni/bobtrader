# TODO

## Immediate Priorities for the Go Ultra-Project

### Runtime control and diagnostics
- [x] Add richer guard diagnostics beyond guard names
- [x] Expose block reasons and guard-trigger counts
- [x] Add coordinated app shutdown tests spanning runtime + scheduler + logger + stream subscriptions
- [x] Add richer execution diagnostics including success/block rates and symbol concentration summaries

### Risk management
- [x] Fully wire `max-concentration` guard using live market value rather than cost-basis fallback
- [x] Add exposure / concentration diagnostics endpoints
- [ ] Add max-open-position and concentration policy tuning docs/examples
- [x] Add additional guards for duplicate side suppression / max notional per symbol / account exposure

### Market data
- [x] Add stream-driven strategy consumption path
- [x] Add event/subscription interfaces beyond simple tick subscription
- [x] Add richer paper stream simulation patterns
- [x] Add integration of stream-fed strategies into scheduler/runtime lifecycle
- [x] Add live candle streaming with `CandleSubscription` and `CandleStreamService`

### Analytics and reporting
- [x] Add persistent metrics history
- [x] Add persistent valuation / PnL history
- [x] Add runtime analytics/reporting modules on top of report storage
- [x] Add execution summary history over time

### Backtesting and Simulation
- [x] Add backtesting subsystem
- [x] Add candle/multi-timeframe strategy support
- [x] Add advanced market emulation (fees, slippage, maker/taker)
- [x] Add optimization subsystem (Parallel execution)
- [x] Add live candle streaming to scheduler (candle-stream mode)
- [x] Add MACD, Bollinger Bands, ATR indicators
- [x] Add real exchange adapters beyond paper mode (Binance REST adapter)
- [x] Add Binance market data feed implementing StreamFeed
- [x] Add walk-forward optimization for out-of-sample parameter validation
- [ ] Add Binance WebSocket adapter for real-time streaming
- [ ] Add reconciliation/fill state improvements

### Operator surfaces
- [x] Add richer operator-facing diagnostics APIs
- [x] Add portfolio summary endpoint distinct from raw positions
- [x] Add UI/dashboard layer for Go runtime
- [x] Add deployment packaging and environment profiles

## Legacy PowerTrader AI Documentation / Cleanup
- [ ] Reconcile stale roadmap/module inventory docs with actual repo state
- [ ] Add clearer distinction between legacy Python runtime and Go ultra-project workstream
- [ ] Audit and document partially integrated Python features more systematically

## Repo-wide Documentation
- [ ] Keep `VISION.md`, `MEMORY.md`, `DEPLOY.md`, `ROADMAP.md`, `TODO.md`, `HANDOFF.md`, `CHANGELOG.md`, `VERSION.md` synchronized
- [ ] Continue expanding submodule/reference documentation as the Go runtime assimilates new ideas
