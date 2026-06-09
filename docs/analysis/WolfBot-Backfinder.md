# Architectural Analysis: Ekliptor/WolfBot (Backfinder)

## Overview
The `Backfinder` in WolfBot is a specialized component for optimizing strategy parameters. It systematically explores a range of configuration values (numerical, string, or boolean) to find the optimal setup for a given market.

## Key Architectural Patterns

### 1. Parallel Process Forking
- **Efficiency through Parallelism:** `Backfinder` uses `child_process.fork` to run multiple backtests concurrently. This is essential given the high computational cost of running hundreds or thousands of historical simulations.
- **Async Flow Control:** Uses `async.parallelLimit` to manage the pool of workers and avoid overwhelming the system resources (memory/CPU).
- **Assimilation Strategy for Go:** In `ultratrader-go`, we should leverage Go's native concurrency (`goroutines` and `worker pools`) instead of process forking, which is significantly more efficient and easier to coordinate.

### 2. Config Range Exploration
- **Range Definition:** Uses `ConfigRange` classes to define start, stop, and step values for parameters (e.g., EMA period from 10 to 100 with step 10).
- **Cartesian Product:** The system generates all possible permutations of the specified ranges to build a complete test grid.
- **Assimilation Strategy for Go:** Implement a `ParameterGrid` generator that produces slices of strategy configurations.

### 3. Result Management and Heap
- **Top Performers:** WolfBot uses a heap to keep track of the best-performing parameter sets during the run.
- **Persistent Reports:** Results are written to disk for operator analysis after the optimization completes.
- **Assimilation Strategy for Go:** The `backtest.Engine` should return a `Result` struct that includes realized PnL, drawdowns, and trade counts, which the optimizer can then rank and store.

## Implementation Takeaways for UltraTrader Go
1. **Parallel Optimizer:** Implement a Go-native parallel runner in `internal/backtest/optimizer` that distributes backtest jobs across a pool of goroutines.
2. **Flexible Scoring:** Allow the optimizer to rank results based on different metrics (e.g., Total PnL, Sharpe Ratio, or Max Drawdown).
3. **Reproducibility:** Ensure that optimized parameters can be easily exported and loaded into a live `App` configuration.
