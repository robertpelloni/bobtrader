# TODO

## Current Sprint (v2.1.x)

### Hierarchical Strategies (v2.3.0)
- [x] Bump version to 2.3.0
- [x] Implement ADX indicator for trend strength
- [x] Implement MacroRegime strategy (EMA + ADX)
- [x] Implement MicroScalper (Tick-volatility)
- [x] Implement RegimeFilter (Macro -> Micro signal suppression)
- [x] Implement ATR reusable sizer
- [x] Implement SiphoningManager (Micro-to-Macro profit redirection)
- [x] Wire 'hierarchical_scalper' suite in App container
- [x] Implement DrawdownGuard for ruin protection
- [x] Implement GridTrading strategy for volatile ranges
- [x] Implement GoldenCross strategy for major trends
- [x] Implement BollingerBreakout strategy
- [x] Integrate Arbitrage signals into strategy runtime
- [x] Implement MLAwareStrategy with kNN Ensemble
- [x] Implement AdaptiveKellySizer using live strategy stats
- [x] Implement Diversified Siphoning (Multi-asset accumulation)
- [x] Implement OrderflowScalper strategy
- [x] Add "Siphoned Wealth" KPI to dashboard
- [x] Implement StatisticalArbitrage strategy (Pairs trading)
- [x] Implement WhaleFlow macro strategy
- [x] Integrate Rebalancer into Siphoning flow
- [x] Implement GeneticOptimizer for evolutionary tuning
- [x] Implement RLFilter for reinforcement learning entries
- [x] Add Siphoning Trend charts to Dashboard

### WebSocket Feed (v2.3.0)
- [x] Debug goroutine-to-channel delivery (Migrated to gorilla/websocket)
- [x] Add auto-reconnection with exponential backoff
- [x] Add WebSocket health monitoring endpoint
- [x] Add Candle History API

### Strategy Enhancement
- [ ] Kelly criterion / volatility-adjusted position sizing
- [ ] Walk-forward parameter optimization on historical data
- [ ] MACD strategy in stream mode
- [ ] ATR-based dynamic sizing

### Real Exchange Integration
- [ ] Wire execution to real Binance spot API
- [ ] Order reconciliation service
- [ ] Trade history sync from exchange
- [ ] Circuit breaker for API resilience

### Dashboard
- [ ] React/Vite SPA dashboard
- [ ] Real-time WebSocket streaming to frontend
- [ ] Interactive charts (TradingView lightweight)
- [ ] Strategy parameter tuning UI

---

## Completed (v2.0.54 and earlier)

### Runtime control and diagnostics
- [x] Add richer guard diagnostics beyond guard names
- [x] Expose block reasons and guard-trigger counts
- [x] Add coordinated app shutdown tests spanning runtime + scheduler + logger + stream subscriptions
- [x] Add richer execution diagnostics including success/block rates and symbol concentration summaries

### Risk management
- [x] Fully wire `max-concentration` guard using live market value rather than cost-basis fallback
- [x] Add exposure / concentration diagnostics endpoints
- [x] Add max-open-position and concentration policy tuning docs/examples
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
- [x] Add Binance WebSocket adapter for real-time streaming
- [x] Add reconciliation/fill state improvements (Order Reconciliation)
- [x] Add periodic reconciliation service (background auto-sync)
- [x] Add trade history sync from exchange
- [x] Add notification system (Email/Discord/Telegram)
- [x] Add position sizing library
- [x] Add circuit breaker for API resilience
- [x] Add correlation analysis engine
- [x] Add trade journal and performance analytics
- [x] Add market regime detection
- [x] Add volume indicators (VWAP, OBV, MFI, CMF)
- [x] Add advanced order types (Stop/Limit/Trailing/Bracket/OCO)
- [x] Add strategy composition and signal voting
- [x] Add portfolio rebalancer
- [x] Add multi-exchange price aggregation
- [x] Add sentiment analysis integration
- [x] Add NLP strategy parsing

### Operator surfaces
- [x] Add richer operator-facing diagnostics APIs
- [x] Add portfolio summary endpoint distinct from raw positions
- [x] Add UI/dashboard layer for Go runtime
- [x] Add deployment packaging and environment profiles

### Autonomous Paper Trading (v2.0.50–v2.0.53)
- [x] Signal-to-execution pipeline with real Binance.US data
- [x] Config-driven strategy params (no recompilation)
- [x] TrailingTakeProfit with functional option pattern
- [x] PortfolioSizer for risk-adjusted position sizing
- [x] Sell-aware risk pipeline (guards exempt sell orders)
- [x] Fee-corrected paper execution (0.1% taker fee)
- [x] Duplicate-sell prevention (dust threshold, position re-check)
- [x] Signal log persistence (JSONL, auto-flush)
- [x] Verified 20-min live test: 47 trades, 80% WR

### Assimilation Program (v2.0.51–v2.0.54)
- [x] Assimilate `TraderAlice/OpenAlice` architectural patterns (ExecutionManager)
- [x] Assimilate `c9s/bbgo` exchange abstractions (Binance Adapter)
- [x] Assimilate `Ekliptor/WolfBot` advanced features (WolfBotBollingerStrategy)
- [x] Assimilate `ccxt/ccxt` exchange abstraction realism (TypedError mapping)
- [x] Assimilate `ctubio/Krypto-trading-bot` market-making (initial MarketMaker)
- [x] Assimilate `whittlem/pycryptobot` risk management (DynamicTSL, ProfitBank)
- [x] Assimilate `freqtrade/freqtrade` strategy patterns (DoubleEMATrend)
- [x] Execute System Test Phase and verify trading functionality
- [x] Execute Sandbox Test Phase and verify risk controls
- [x] Integrate and verify live market feed performance
- [x] Execute Integration Test Phase and verify market data/execution
- [x] Conduct final live integration test and verify real-time performance

### Repository Maintenance (v2.0.54)
- [x] Remove 6 orphaned submodule references from git index
- [x] Merge assimilation feature branch into main
- [x] Update build/start scripts for Go-first workflow
- [x] Update .gitignore for build artifacts and orphaned dirs
- [x] Reconcile ROADMAP.md and TODO.md with actual project state

### Legacy PowerTrader AI Documentation / Cleanup
- [x] Reconcile stale roadmap/module inventory docs with actual repo state
- [x] Add clearer distinction between legacy Python runtime and Go ultra-project workstream
- [x] Audit and document partially integrated Python features more systematically

### Repo-wide Documentation
- [x] Keep `VISION.md`, `MEMORY.md`, `DEPLOY.md`, `ROADMAP.md`, `TODO.md`, `HANDOFF.md`, `CHANGELOG.md`, `VERSION.md` synchronized
- [x] Continue expanding submodule/reference documentation as the Go runtime assimilates new ideas

---

## Remaining Backlog

- [ ] Deploy to live market conditions (real capital, not paper)
- [ ] Search and categorize next 43 candidates in `ASSIMILATION_CANDIDATES.md`
- [ ] Multi-exchange price aggregation (KuCoin, Coinbase adapters)
- [ ] Portfolio rebalancer with wash-sale prevention
- [ ] Drawdown monitoring with auto-shutdown
- [ ] Compliance reporting (risk flags, recommendations)
