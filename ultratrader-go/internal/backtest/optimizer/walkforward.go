package optimizer

import (
	"context"
	"fmt"
	"sort"
	"time"
)

// ParameterSet represents a configuration of strategy parameters
type ParameterSet map[string]float64

// Result contains the outcome of an optimization run
type Result struct {
	Parameters  ParameterSet
	InSample    float64 // e.g. Sharpe Ratio or Total Return during training
	OutOfSample float64 // performance during testing window
}

// Window represents a single walk-forward iteration slice
type Window struct {
	TrainStart time.Time
	TrainEnd   time.Time
	TestStart  time.Time
	TestEnd    time.Time
}

// Evaluator is an interface for evaluating a set of parameters on a specific time window
type Evaluator interface {
	// Evaluate returns a fitness score for the given parameters over the specified time window
	Evaluate(ctx context.Context, params ParameterSet, start, end time.Time) (float64, error)
}

// WalkForwardOptimizer manages the windowed evaluation
type WalkForwardOptimizer struct {
	evaluator    Evaluator
	paramGrid    []ParameterSet
	trainWindows int // Duration in some arbitrary units, or we compute based on total timeframe
	testWindows  int
}

// NewWalkForwardOptimizer creates a new WFO instance
func NewWalkForwardOptimizer(evaluator Evaluator, grid []ParameterSet) *WalkForwardOptimizer {
	return &WalkForwardOptimizer{
		evaluator: evaluator,
		paramGrid: grid,
	}
}

// GenerateWindows slices a total timeframe into overlapping train/test windows.
func GenerateWindows(start, end time.Time, trainDuration, testDuration time.Duration) []Window {
	var windows []Window

	currentTrainStart := start
	for {
		trainEnd := currentTrainStart.Add(trainDuration)
		testStart := trainEnd
		testEnd := testStart.Add(testDuration)

		if testEnd.After(end) {
			break
		}

		windows = append(windows, Window{
			TrainStart: currentTrainStart,
			TrainEnd:   trainEnd,
			TestStart:  testStart,
			TestEnd:    testEnd,
		})

		// Step forward by the test duration
		currentTrainStart = currentTrainStart.Add(testDuration)
	}

	return windows
}

// Run executes the walk-forward optimization across all generated windows.
// For each window, it finds the best parameter set in the Train window and records its performance in the Test window.
func (o *WalkForwardOptimizer) Run(ctx context.Context, windows []Window) ([]Result, error) {
	if len(windows) == 0 {
		return nil, fmt.Errorf("no windows provided")
	}
	if len(o.paramGrid) == 0 {
		return nil, fmt.Errorf("parameter grid is empty")
	}

	var results []Result

	for i, w := range windows {
		_ = i
		var bestParams ParameterSet
		var bestInSampleScore float64 = -999999.0

		// 1. In-Sample Optimization
		for _, params := range o.paramGrid {
			score, err := o.evaluator.Evaluate(ctx, params, w.TrainStart, w.TrainEnd)
			if err != nil {
				continue
			}
			if score > bestInSampleScore {
				bestInSampleScore = score
				bestParams = params
			}
		}

		if bestParams == nil {
			// If all failed, skip this window
			continue
		}

		// 2. Out-Of-Sample Validation
		outOfSampleScore, err := o.evaluator.Evaluate(ctx, bestParams, w.TestStart, w.TestEnd)
		if err != nil {
			outOfSampleScore = 0 // Penalty or default
		}

		results = append(results, Result{
			Parameters:  bestParams,
			InSample:    bestInSampleScore,
			OutOfSample: outOfSampleScore,
		})
	}

	return results, nil
}

// Aggregate computes the overall average out-of-sample performance across all windows.
func Aggregate(results []Result) float64 {
	if len(results) == 0 {
		return 0
	}
	var sum float64
	for _, r := range results {
		sum += r.OutOfSample
	}
	return sum / float64(len(results))
}

// SortByOutOfSample sorts a slice of results descending by out-of-sample performance
func SortByOutOfSample(results []Result) {
	sort.Slice(results, func(i, j int) bool {
		return results[i].OutOfSample > results[j].OutOfSample
	})
}
