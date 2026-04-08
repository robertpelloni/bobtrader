# Go Phase-37 Expanded Indicator Library

## Summary
This phase significantly expands the technical indicator library from 3 indicators (SMA, EMA, RSI) to 6 by adding MACD, Bollinger Bands, and ATR — the three most commonly used indicators in algorithmic crypto trading.

## Context & Motivation
Phases 30 and 32 established a minimal indicator foundation (SMA, EMA, RSI) sufficient for basic crossover strategies. However, real quantitative trading systems require a richer indicator toolkit. WolfBot and BBGO both implement extensive indicator libraries. The three indicators added in this phase were selected because:

1. **MACD** — The gold standard momentum indicator. Used in virtually every trading system for trend confirmation and divergence detection.
2. **Bollinger Bands** — Essential volatility indicator. Critical for mean-reversion strategies and breakout detection. PowerTrader's Python codebase uses Bollinger Bands extensively.
3. **ATR** — Volatility measure used for position sizing and stop-loss calculation. BBGO uses ATR for dynamic risk management.

## Delivered

### MACD (Moving Average Convergence Divergence)
- Returns a `MACDResult` struct with `MACD`, `Signal`, and `Histogram` fields.
- Built compositionally from three `EMA` instances (fast, slow, signal).
- Configurable periods (default: 12/26/9).
- `Update(value)` and `Last()` methods consistent with existing indicator pattern.

### Bollinger Bands
- Returns a `BollingerBandsResult` struct with `Upper`, `Middle`, `Lower`, and `Bandwidth` fields.
- Standard deviation calculated from the SMA window.
- Configurable period and multiplier (standard: 20 period, 2.0 multiplier).
- Bandwidth metric enables volatility-based strategy logic (squeeze detection).

### ATR (Average True Range)
- Accepts `(high, low, close)` triple — consistent with candle data.
- Computes True Range as max of three standard measures.
- Uses Wilder smoothing (exponential) after warmup period.
- Essential for volatility-adjusted position sizing.

### Testing
All 10 indicator tests pass:
- Original: SMA, EMA, RSI
- New: MACD crossover, MACD Last(), Bollinger constant, Bollinger varying, Bollinger insufficient data, ATR warmup, ATR insufficient data

## Architecture

```
internal/indicator/
├── indicators.go       # All 6 indicators: SMA, EMA, RSI, MACD, BollingerBands, ATR
└── indicators_test.go  # 10 comprehensive test cases
```

All indicators follow the same streaming `Update() → Result` / `Last() → Result` pattern, making them trivially composable in strategies.

## Next Steps
1. **Indicator-Based Strategies** — Build new demo strategies using MACD crossovers, Bollinger Band mean-reversion, and ATR position sizing.
2. **Real Exchange Adapters** — Connect the now-mature strategy pipeline to live market data.
3. **Walk-Forward Optimization** — Run the concurrent optimizer against expanded parameter grids spanning multiple indicator parameters.
