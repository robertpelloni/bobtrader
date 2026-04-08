package optimizer

import (
	"context"
	"runtime"
	"sort"
	"sync"

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
	optConfig OptimizationConfig,
) ([]RunResult, error) {
	if scorer == nil {
		scorer = DefaultScorer
	}

	perms := generateGrid(paramGrid)
	if len(perms) == 0 {
		return nil, nil
	}

	workers := optConfig.MaxWorkers
	if workers <= 0 {
		workers = runtime.NumCPU()
	}
	if workers > len(perms) {
		workers = len(perms)
	}

	jobs := make(chan ParameterMap, len(perms))
	resultsCh := make(chan RunResult, len(perms))

	var wg sync.WaitGroup

	// Start worker pool
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for p := range jobs {
				// Check context cancellation
				select {
				case <-ctx.Done():
					return
				default:
				}

				strat, err := builder(p)
				if err != nil {
					continue // Skip invalid parameter combinations
				}

				eng := backtest.NewEngineWithOptions(strat, initialCapital, opts)
				res, err := eng.RunCandles(ctx, history)
				if err != nil {
					continue // Skip runs that fail
				}

				resultsCh <- RunResult{
					Params: p,
					Result: res,
					Score:  scorer(res),
				}
			}
		}()
	}

	// Dispatch jobs
	for _, p := range perms {
		jobs <- p
	}
	close(jobs)

	// Wait for workers to finish in a separate goroutine
	go func() {
		wg.Wait()
		close(resultsCh)
	}()

	// Collect results
	var results []RunResult
	for r := range resultsCh {
		results = append(results, r)
	}

	// Sort descending by score
	sort.SliceStable(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	return results, nil
}
