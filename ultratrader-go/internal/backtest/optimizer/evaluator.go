package optimizer

import (
	"context"
	"fmt"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

// BacktestEvaluator implements Evaluator by running a full historical backtest
// over the specified time window.
type BacktestEvaluator struct {
	historyBuilder func(start, end time.Time) (backtest.CandleHistoryProvider, error)
	stratBuilder   StrategyBuilder
	scorer         ScoringFunction
	initialCapital float64
	opts           backtest.EmulatorOptions
}

// NewBacktestEvaluator creates an Evaluator that runs real backtests.
func NewBacktestEvaluator(
	historyBuilder func(start, end time.Time) (backtest.CandleHistoryProvider, error),
	stratBuilder StrategyBuilder,
	scorer ScoringFunction,
	initialCapital float64,
	opts backtest.EmulatorOptions,
) *BacktestEvaluator {
	if scorer == nil {
		scorer = DefaultScorer
	}
	return &BacktestEvaluator{
		historyBuilder: historyBuilder,
		stratBuilder:   stratBuilder,
		scorer:         scorer,
		initialCapital: initialCapital,
		opts:           opts,
	}
}

// Evaluate builds the strategy with params, gets the history for the window, runs the engine, and scores the result.
func (e *BacktestEvaluator) Evaluate(ctx context.Context, params ParameterMap, start, end time.Time) (float64, error) {
	// 1. Build history for this window
	history, err := e.historyBuilder(start, end)
	if err != nil {
		return 0, fmt.Errorf("failed to build history: %w", err)
	}

	candles := history.Candles()
	if len(candles) == 0 {
		return 0, fmt.Errorf("no candles in window %v to %v", start, end)
	}

	// 2. Build strategy with params
	strat, err := e.stratBuilder(params)
	if err != nil {
		return 0, fmt.Errorf("failed to build strategy: %w", err)
	}

	// 3. Run backtest
	engine := backtest.NewEngineWithOptions(strat, e.initialCapital, e.opts)
	result, err := engine.RunCandles(ctx, history)
	if err != nil {
		return 0, fmt.Errorf("backtest run failed: %w", err)
	}

	// 4. Score
	return e.scorer(result), nil
}

// SlicedCandleHistory is a helper to wrap a subset of candles as a provider.
type SlicedCandleHistory struct {
	candles []marketdata.Candle
}

func NewSlicedCandleHistory(candles []marketdata.Candle) *SlicedCandleHistory {
	return &SlicedCandleHistory{candles: candles}
}

func (s *SlicedCandleHistory) Candles() []marketdata.Candle {
	return s.candles
}

// SliceHistory is a helper function to filter a larger slice of candles down to a specific time window [start, end).
func SliceHistory(allCandles []marketdata.Candle, start, end time.Time) []marketdata.Candle {
	var subset []marketdata.Candle
	for _, c := range allCandles {
		if !c.Timestamp.Before(start) && c.Timestamp.Before(end) {
			subset = append(subset, c)
		}
	}
	return subset
}
