# Go Phase-38 Indicator-Based Strategies

## Summary
This phase brings the expanded indicator library (Phase 37) to life by building three new candle-based demo strategies that leverage MACD, Bollinger Bands, and ATR for signal generation and position sizing.

## Delivered

### MACDCrossover
- Monitors the MACD histogram for zero-line crossovers.
- **Bullish crossover** (histogram crosses from negative to positive) → buy signal.
- **Bearish crossover** (histogram crosses from positive to negative) → sell signal.
- Configurable fast/slow/signal periods (default: 12/26/9).
- Test validates both bullish and bearish crossover detection with realistic price sequences.

### BollingerReversion
- Mean-reversion strategy operating within Bollinger Bands.
- **Buy** when price touches or drops below the lower band (oversold).
- **Sell** when price touches or exceeds the upper band (overbought).
- **No signal** when price remains within bands.
- Tests validate band-touch detection, spike detection, and in-band silence.

### ATRSizing
- Combines SMA crossover for directional signals with ATR-based position sizing.
- Calculates dynamic quantity: `(riskPerTrade × closePrice) / ATR`.
- Higher ATR (high volatility) → smaller position.
- Lower ATR (low volatility) → larger position.
- Floors quantity at 0.001 to prevent zero-size orders.

## Architecture

All three strategies implement the `CandleStrategy` interface via `CandleEvent(ctx, marketdata.Candle)`, making them compatible with:
- The backtesting engine (`backtest.Engine.RunCandles`)
- The live candle stream scheduler (`CandleStreamService`)
- The concurrent optimizer (`optimizer.GridSearchCandles`)

## Next Steps
1. **Real Exchange Adapters** — Connect strategies to live market data.
2. **Walk-Forward Optimization** — Optimize MACD/Bollinger/ATR parameters against historical data.
3. **Portfolio-Level Strategy** — Combine multiple indicator strategies into a composite portfolio signal.
