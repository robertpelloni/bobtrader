# Go Phase-41 Walk-Forward Optimization

## Summary
This phase implements walk-forward optimization — the gold standard for validating that strategy parameters are robust and not overfitted to historical data.

## Context & Motivation
Grid search (Phase 34) and concurrent optimization (Phase 35) find the best parameters for a given dataset. However, there's a critical risk: **overfitting**. A strategy that looks perfect in-sample may fail catastrophically on new data. Walk-forward analysis addresses this by:

1. Splitting historical data into rolling training + validation windows
2. Running grid search on each training window to find optimal parameters
3. Testing those parameters on the *next* unseen window (out-of-sample)
4. Measuring the gap between in-sample and out-of-sample performance

This is the standard validation approach used by professional quantitative trading firms and is recommended by CCXT, WolfBot, and BBGO documentation.

## Delivered

### Walk-Forward Engine (`optimizer/walkforward.go`)

#### Core Algorithm
1. `generateWindows()` creates rolling window pairs from the candle history
2. For each window:
   - Extract training candles → run `GridSearchCandles` → find best params
   - Extract validation candles → run single backtest with best params → score
3. Aggregate results with `WalkForwardResult`

#### Configuration
- `WindowCandles` — Size of each training window (default: 100)
- `StepCandles` — How far to advance per step / validation window size (default: 20)
- `MinTrades` — Minimum trades required for a valid result (default: 1)
- `OptimizationConfig` — Reuses the concurrent worker pool settings

#### Overfitting Analysis
`AnalyzeOverfitting()` computes:
- **Overfit = TrainScore - ValidationScore** for each window
- Sorted from least to most overfit
- Positive overfit means the strategy performed better in-sample than out-of-sample

### Testing
- `TestWalkForwardCandles` — 200 candles, 7 rolling windows, 2x2 parameter grid
- `TestWalkForwardOverfitting` — Validates overfit analysis output
- `TestWalkForwardInsufficientData` — Error handling for insufficient history
- `TestGenerateWindows` — Validates window generation boundaries

## Architecture

```
optimizer/
├── optimizer.go         # Types: ParameterMap, StrategyBuilder, ScoringFunction, RunResult
├── grid.go              # GridSearchCandles (concurrent worker pool)
├── walkforward.go       # WalkForwardCandles (rolling window optimization)
├── grid_test.go         # Grid search tests
└── walkforward_test.go  # Walk-forward tests
```

Walk-forward reuses the concurrent grid search internally, so every training window evaluation is parallelized across CPU cores.

## Next Steps
1. **Sharpe Ratio Scorer** — Add risk-adjusted scoring functions beyond raw PnL
2. **Monte Carlo Validation** — Random permutation testing for statistical significance
3. **Parameter Stability Analysis** — Track how optimal parameters change across windows
