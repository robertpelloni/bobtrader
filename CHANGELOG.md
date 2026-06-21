# Changelog

All notable changes to PowerTrader AI will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v3.0.0.html).

## [2.1.5] - 2026-06-21
### Added
- **Drawdown Monitor and Auto-Shutdown** — Added `DrawdownMonitor` to `internal/risk` to track peak portfolio value and calculate current drawdown. If the configured `MaxDrawdownPct` is exceeded, the application triggers an auto-shutdown (`os.Exit(1)`) to prevent further losses. This is wired into the main background loop in `app.go`.

## [2.1.4] - 2026-06-20
### Added
- **Real Exchange Execution Wiring** — The `ExecutionManager` and `Reconciler` are now dynamically instantiated with the `binance.Adapter` when live trading accounts are active, correctly routing live execution outside of the paper fallback.
- **Intelligent Order Reconciliation** — The background order `Reconciler` loop now actively resolves local state drift by updating the `executionRepo` against queried live exchange data.
- **Trade History Boot-Sync** — On startup, the system now queries the exchange via the new `TradeHistoryQuerier` interface (`/api/v3/myTrades`) to fetch recent trades and perfectly sync the `PortfolioTracker` state.
- **API Circuit Breaker** — Hardened the Binance adapter with a `circuitbreaker.Breaker` wrapping all HTTP execution pathways to prevent runaway failure loops during exchange instability.

## [2.1.3] - 2026-06-18
### Added
- **WebSocket Feed Hardening** — Fixed goroutine-to-channel delivery bugs in the Binance WebSocket stream feed caused by json parsing of large numbers, added auto-reconnection with exponential backoff on disconnects.
- **WebSocket Health API & Dashboard** — Exposed WS health (connection state and staleness ms) via `/api/ws-health` and added visual indicators to the web dashboard UI.

### Changed
- **Default Feed** — Switched `market_data.source` default from `rest` to `websocket` in primary paper trading config files.

## [2.1.2] - 2026-06-11
### Changed
- **Live Risk Settings** — Optimized `config/eth-live.json` by reducing `cooldown_ms` to 30s and `duplicate_window_ms` to 1s to prevent over-blocking strategy transitions.

## [2.1.1] - 2026-06-11
### Added
- **CGO-Free Persistence Layer** — Replaced `github.com/mattn/go-sqlite3` with `modernc.org/sqlite` to eliminate CGO dependencies and enable native cross-compilation on Windows.
- **BNB Fee Optimization** — Purchased BNB balance to pay for live trading transaction fees at a 25% discount.

### Fixed
- **Windows Compilation Blockers** — Resolved cgo compiler compatibility issues with GCC 15/w64devkit.

## [2.1.0] - 2026-06-10
### Added
- **Paper trading with real Binance.US data** — `config/paper-live-data.json` uses real market prices for simulated execution ($10k starting balance)
- **API key verification tool** — `cmd/apikeys-check/main.go` validates Binance API keys with read-only checks (prices, candles, balances)
- **Aggressive trading config** — `config/paper-aggressive.json` with 5% risk, 5s cooldown, tighter Bollinger bands
- **All-strategies config** — `config/paper-all-strategies.json` with 14 strategies across 9 symbols
- **Tick Momentum Burst strategy** — Buys/sells on 0.15% price spikes
- **Tick Mean Reversion strategy** — Trades when price deviates 0.10% from recent average
- **Double EMA Trend strategy** — Triple EMA trend following (5/13/50)
- **USDT Stablecoin Scalp strategy** — Trades 0.9991-1.0000 range with stop loss at 0.98
- **USDC Stablecoin Scalp strategy** — Wider thresholds for more volatile USDC
- **Weekly Cycle strategy** — Buys Monday dip, sells Sunday peak based on historical weekly patterns
- **China Session strategy** — Exploits Asian session volatility (pre-Asia buy, Asia spike sell)
- **Sentiment Engine** — Aggregates multiple sentiment providers:
  - CryptoPanic news API
  - Fear & Greed Index (live)
  - Market Events (BTC halving, FOMC, ETF decisions, tax season)
  - Stock Market Correlation (SPY as risk indicator)
  - YouTube Sentiment (monitors 8 crypto channels: Arcane Bear, Benjamin Cowen, Coin Bureau, etc.)
  - Whale Alert (tracks large exchange inflows/outflows)
- **Cross-Exchange Arbitrage** — Detects price differences between exchanges
- **Sentiment-Aware Strategy** — Combines all sentiment sources with technical analysis
- **Whale Alert Strategy** — Trades based on large whale movements (exchange inflow = bearish, outflow = bullish)
- **9 trading pairs** — BTC, ETH, SOL, XLM, ADA, DOGE, XRP, USDT, USDC
- **Test suite additions** — `live_backtest_test.go`, `stress_test.go`, `accuracy_test.go` from merged feature branch

### Fixed
- **primaryAccountID selection** — Now correctly prefers `paper` and `paper-market-aware` accounts
- **Balance reader routing** — Paper accounts use simulated balance, Binance accounts use real balance
- **Market-aware paper adapter** — Fill prices now use real Binance.US ticker data
- **BollingerReversion dual-signal bug** — Added else-if guard and lastSignal dedup

### Changed
- **Strategy count** — Increased from 4 to 14 strategies per symbol
- **Config-driven strategy params** — All tuning in JSON config (no recompilation needed)
- **Faster polling** — 3s interval (was 5s) for aggressive mode
- **Lower guard thresholds** — 5s cooldown (was 15s), 10s duplicate window (was 30s)

## [2.0.54] - 2026-06-08
### Added
- **Submodule sanitization** — Removed 6 orphaned submodule references missing from .gitmodules
- **Feature branch merge** — Assimilated `assimilate-top-crypto-bots-phase-1` branch with:
  - `ExecutionManager` — Modular strategy coordination
  - `WolfBotBollingerStrategy` — Breakout-aware Bollinger from WolfBot patterns
  - `DoubleEMATrendStrategy` — Trend-following from freqtrade patterns
  - `DynamicTrailingStop` — Trailing stop with high-price tracking (from pycryptobot)
  - `ProfitBank` / `PreventLoss` — Multi-layered exit safety strategies
  - `MarketMakerStrategy` — Ping-pong quoting from Krypto-trading-bot
  - `LiveStrategyWrapper` — Production safety wrapper for live strategies
  - `LiveHistoryProvider` — High-fidelity backtesting with real market data
  - CCXT-inspired unified error mapping (`internal/exchange/errors.go`)
