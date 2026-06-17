# Changelog

All notable changes to BobTrader (UltraTrader Go) will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v3.0.0.html).

## [3.3.0-alpha] - 2026-06-17
### Added
- **VWAP Execution Strategy** — Implemented institutional-grade Volume Weighted Average Price execution in `internal/trading/execution/vwap.go`.
- **Atomic Multi-Venue Arbitrage** — Coordinated, concurrent execution of cross-exchange trades with `ArbitrageExecutorV2` in `internal/trading/multiexchange/chain_executor.go`.
- **Liquidity Depth Visualization** — Added `DepthVisualizer` React component and `/api/marketdata/depth` endpoint to visualize global order book liquidity.
- **Embedded Dashboard serving** — Backend now automatically serves the React production build from `web/dist` with SPA routing support.

### Fixed
- **CorrelationGuard Build Regression** — Fixed type mismatch in `internal/risk/correlation_guard.go`.
- **Primary Account Selection** — Improved fallback logic for identifying the main trading account in `App.go`.

## [3.2.0] - 2026-06-15
### Added
- **Correlation Guards** — Risk diversification layer to prevent over-concentration in highly correlated assets.
- **Liquidity-Aware Smart Router** — Selects optimal exchange based on real-time BBO, fees, and depth.
- **InfluxDB Integration** — High-performance time-series metrics persistence.

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
