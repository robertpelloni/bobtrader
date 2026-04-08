# Go Phase-34 Optimization Subsystem

## Summary
This phase introduces a hyper-parameter optimization framework directly coupled to the `backtest.Engine`. It allows quantitative operators to programmatically search for the most profitable strategy configurations.

## Context & Motivation
Submodules like `WolfBot` and the Machine Learning pipelines in the legacy Python `PowerTrader AI` extensively utilize parameter tuning to find optimal indicator lengths, threshold values, and trigger bounds. Doing this in Python can be slow. By building this natively in Go, we leverage rapid in-memory execution to iterate over thousands of parameter combinations in seconds.

## Delivered

### Optimizer Architecture
- Created the `internal/backtest/optimizer` package.
- Designed `ParameterMap` to hold abstract sets of variables and `StrategyBuilder` as a factory pattern for dynamically instantiating strategies based on those maps.
- Created `RunResult` containing the parameters, the final `backtest.Result`, and a unified `Score`.
- Built `ScoringFunction` interfaces, defaulting to `DefaultScorer` which simply extracts `RealizedPnL`. This makes it trivial to swap to a Sharpe Ratio or Sortino Ratio scorer in the future.

### Grid Search Engine
- Implemented `generateGrid()` utilizing a recursive Cartesian product algorithm to unroll defined parameter boundaries (e.g., `map[string][]interface{}`) into a flat array of combinations.
- Implemented `GridSearchCandles()` which loops over the generated parameter grids, spawns the strategy, spins up a fresh `backtest.Engine` with predefined `EmulatorOptions` (friction), runs the simulation, and returns a globally sorted list of outcomes from best to worst.

### Validation & Testing
- Developed `grid_test.go` asserting the Cartesian algorithm's accuracy (producing 4 distinct runs from 2 options for `fast` and 2 options for `slow` SMAs).
- Validated that the results array correctly sorted the outcomes descending by their PnL scores.

## Architectural Significance
This phase completes the fundamental "Quant Pipeline" of the Go ultra-project. A developer can now write a strategy (Phase 30/32), evaluate it realistically with fees (Phase 33), and tune its parameters programmatically (Phase 34). Because the execution occurs sequentially against in-memory interfaces, it runs entirely CPU-bound with zero network or I/O bottlenecks.

## Recommended next steps
1.  **Concurrency Optimization:** Currently, `GridSearchCandles` iterates sequentially. Introduce a `sync.WaitGroup` worker pool or `errgroup` to allow parallel grid execution, fully exploiting multi-core architectures.
2.  **Live Candle Streaming:** Build out the `marketdata` feed and `scheduler` to subscribe to live exchange websocket K-Line feeds and dispatch them via `runtime.CandleEvent`.
3.  **Real Exchange Adapters:** Start integrating CCXT or Binance-specific REST/Websocket connections beyond the `paper` adapter to bridge the optimized strategies into live execution.
