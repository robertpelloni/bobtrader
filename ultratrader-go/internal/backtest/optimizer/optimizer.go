package optimizer

import (
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// OptimizationConfig controls the execution behavior of the optimization run.
type OptimizationConfig struct {
	MaxWorkers int // Maximum number of concurrent backtest workers. 0 defaults to runtime.NumCPU().
}

// DefaultOptimizationConfig provides sensible defaults.
func DefaultOptimizationConfig() OptimizationConfig {
	return OptimizationConfig{
		MaxWorkers: 0, // 0 means use runtime.NumCPU() in the runner
	}
}

// ParameterMap holds a specific combination of strategy parameters.
type ParameterMap map[string]interface{}

// StrategyBuilder is a factory function that constructs a strategy given a set of parameters.
type StrategyBuilder func(params ParameterMap) (strategy.Strategy, error)

// RunResult encapsulates the outcome of a single backtest run within an optimization pass.
type RunResult struct {
	Params ParameterMap
	Result backtest.Result
	Score  float64
}

// ScoringFunction defines how an optimization run is evaluated (e.g., maximizing Realized PnL).
type ScoringFunction func(result backtest.Result) float64

// DefaultScorer maximizes Realized PnL.
func DefaultScorer(result backtest.Result) float64 {
	return result.RealizedPnL
}
