# BobTrader / UltraTrader

An autonomous cryptocurrency trading platform: a Go runtime that connects to live Binance market data, runs configurable strategies through a risk-guard pipeline, and executes via a paper-trading engine with fee simulation. It also preserves a legacy Python trading system (`pt_*.py`) for reference and a corpus of open-source trading bots as git submodules for architecture research.

> ⚠️ **This software can place real trades automatically when configured with live keys.** You are responsible for all financial outcomes. This is a personal experiment, not financial advice, and not a product — see the maintainer's note at the bottom.

## Key facts (from the code)

| | |
|---|---|
| Language | Go (primary), Python (legacy, frozen) |
| Market data | Binance.US REST; WebSocket experimental |
| Execution | Paper trading with 0.1% taker-fee simulation; live execution possible via API keys |
| Strategies | 8 built-in (Bollinger, RSI, EMA, Trailing TP, WolfBot, DoubleEMA, MarketMaker, Safety) |
| Risk guards | Whitelist, notional, concentration, cooldown, duplicate, max-positions, and others |
| API | REST endpoints for portfolio, orders, diagnostics, config |
| Config | JSON-driven — strategies tune without recompiling |
| Per-coin workspaces | `BNB/`, `DOGE/`, `ETH/`, `XRP/` directories carry per-asset state/config |

Docs: `MANUAL.md`, `MODULE_INDEX.md`, `AUTONOMOUS_DUAL_BOT_STRATEGY.md`, `docs/`.

## License

Apache 2.0 — see `LICENSE`.

## Maintainer's note

BobTrader is one of Robert Pelloni's money experiments. It is **not** a revenue source and is not maintained as one; it stays public as a foundation others can build a trading bot on, and as a record of the experiment. Trade with it at your own risk — that warning is the most important line in this README.
