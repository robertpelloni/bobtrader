package optimizer

import (
	"context"
	"fmt"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest"
)

// BacktestEvaluator implements Evaluator by running a real backtest simulation.
type BacktestEvaluator struct {
	builder StrategyBuilder
	history backtest.CandleHistoryProvider
	scorer  ScoringFunction
	capital float64
}

func NewBacktestEvaluator(builder StrategyBuilder, history backtest.CandleHistoryProvider, scorer ScoringFunction, capital float64) *BacktestEvaluator {
	if scorer == nil {
		scorer = DefaultScorer
	}
	if capital <= 0 {
		capital = 10000
	}
	return &BacktestEvaluator{
		builder: builder,
		history: history,
		scorer:  scorer,
		capital: capital,
	}
}

func (e *BacktestEvaluator) Evaluate(ctx context.Context, params ParameterSet, start, end time.Time) (float64, error) {
	// 1. Filter history for the window
	var windowCandles []any // placeholder for filtered candles
	_ = windowCandles

	// Converting ParameterSet to ParameterMap
	pMap := make(ParameterMap)
	for k, v := range params {
		pMap[k] = v
	}

	// 2. Build strategy
	strat, err := e.builder(pMap)
	if err != nil {
		return 0, fmt.Errorf("failed to build strategy: %w", err)
	}

	// 3. Run backtest (Simplified for v3.1.0, assumes history is already filtered or engine handles it)
	engine := backtest.NewEngine(strat, e.capital)
	res, err := engine.RunCandles(ctx, e.history)
	if err != nil {
		return 0, fmt.Errorf("backtest execution failed: %w", err)
	}

	// 4. Score the result
	return e.scorer(res), nil
}
