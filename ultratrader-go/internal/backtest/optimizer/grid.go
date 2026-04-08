package optimizer

import (
	"context"
	"sort"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest"
)

// generateGrid computes the Cartesian product of all provided parameter arrays.
func generateGrid(paramGrid map[string][]interface{}) []ParameterMap {
	var keys []string
	for k := range paramGrid {
		keys = append(keys, k)
	}

	var results []ParameterMap
	var helper func(keyIndex int, currentMap ParameterMap)

	helper = func(keyIndex int, currentMap ParameterMap) {
		if keyIndex == len(keys) {
			// Copy map to avoid referencing the same map across iterations
			copyMap := make(ParameterMap)
			for k, v := range currentMap {
				copyMap[k] = v
			}
			results = append(results, copyMap)
			return
		}

		key := keys[keyIndex]
		values := paramGrid[key]
		for _, v := range values {
			currentMap[key] = v
			helper(keyIndex+1, currentMap)
		}
	}

	helper(0, make(ParameterMap))
	return results
}

// GridSearchCandles systematically evaluates all parameter combinations over a candle history
// and returns the results sorted by their score (highest first).
func GridSearchCandles(
	ctx context.Context,
	builder StrategyBuilder,
	history backtest.CandleHistoryProvider,
	initialCapital float64,
	opts backtest.EmulatorOptions,
	paramGrid map[string][]interface{},
	scorer ScoringFunction,
) ([]RunResult, error) {
	if scorer == nil {
		scorer = DefaultScorer
	}

	perms := generateGrid(paramGrid)
	var results []RunResult

	// Currently running sequentially. Could be optimized with goroutines for massive grids.
	for _, p := range perms {
		strat, err := builder(p)
		if err != nil {
			continue // Skip invalid parameter combinations
		}

		eng := backtest.NewEngineWithOptions(strat, initialCapital, opts)
		res, err := eng.RunCandles(ctx, history)
		if err != nil {
			continue // Skip runs that fail
		}

		results = append(results, RunResult{
			Params: p,
			Result: res,
			Score:  scorer(res),
		})
	}

	// Sort descending by score
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}
