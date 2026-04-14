package optimizer

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

// ParamRange defines a min, max, and step for a parameter search.
type ParamRange struct {
	Min  float64
	Max  float64
	Step float64
}

// GenerateGrid exhaustively generates all parameter combinations given a map of ranges.
func GenerateGrid(ranges map[string]ParamRange) ([]ParameterSet, error) {
	if len(ranges) == 0 {
		return nil, fmt.Errorf("no parameter ranges provided")
	}

	keys := make([]string, 0, len(ranges))
	for k := range ranges {
		if ranges[k].Step <= 0 {
			return nil, fmt.Errorf("parameter %s has invalid step %f", k, ranges[k].Step)
		}
		if ranges[k].Min > ranges[k].Max {
			return nil, fmt.Errorf("parameter %s has min > max", k)
		}
		keys = append(keys, k)
	}

	var grid []ParameterSet

	// Helper function for recursive combinatorics
	var combine func(int, ParameterSet)
	combine = func(idx int, currentSet ParameterSet) {
		if idx == len(keys) {
			// Deep copy
			cp := make(ParameterSet, len(currentSet))
			for k, v := range currentSet {
				cp[k] = v
			}
			grid = append(grid, cp)
			return
		}

		key := keys[idx]
		r := ranges[key]

		// Using a small epsilon to handle float precision issues in loop bounds
		for val := r.Min; val <= r.Max+0.000001; val += r.Step {
			currentSet[key] = val
			combine(idx+1, currentSet)
		}
	}

	combine(0, make(ParameterSet))

	return grid, nil
}

// SearchResult holds a parameter set and its evaluated score.
type SearchResult struct {
	Parameters ParameterSet
	Score      float64
}

// GridSearchOptimizer evaluates a grid of parameters concurrently.
type GridSearchOptimizer struct {
	evaluator Evaluator
}

// NewGridSearchOptimizer creates a new grid search engine.
func NewGridSearchOptimizer(evaluator Evaluator) *GridSearchOptimizer {
	return &GridSearchOptimizer{
		evaluator: evaluator,
	}
}

// Optimize evaluates the full grid concurrently and returns all results sorted descending by score.
func (o *GridSearchOptimizer) Optimize(ctx context.Context, grid []ParameterSet, start, end time.Time, concurrency int) ([]SearchResult, error) {
	if len(grid) == 0 {
		return nil, fmt.Errorf("grid is empty")
	}

	if concurrency <= 0 {
		concurrency = 1
	}

	jobs := make(chan ParameterSet, len(grid))
	results := make(chan SearchResult, len(grid))
	var wg sync.WaitGroup

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for params := range jobs {
				score, err := o.evaluator.Evaluate(ctx, params, start, end)
				if err != nil {
					// In a real system, we'd log this, but for the optimizer we assign worst possible score or skip.
					continue
				}
				results <- SearchResult{Parameters: params, Score: score}
			}
		}()
	}

	for _, p := range grid {
		jobs <- p
	}
	close(jobs)

	// Wait in background and close results
	go func() {
		wg.Wait()
		close(results)
	}()

	var allResults []SearchResult
	for r := range results {
		allResults = append(allResults, r)
	}

	if len(allResults) == 0 {
		return nil, fmt.Errorf("all evaluations failed")
	}

	// Sort descending
	sort.Slice(allResults, func(i, j int) bool {
		return allResults[i].Score > allResults[j].Score
	})

	return allResults, nil
}
