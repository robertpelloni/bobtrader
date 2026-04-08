package optimizer

import (
	"context"
	"fmt"
	"sort"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

// WalkForwardConfig controls walk-forward optimization behavior.
type WalkForwardConfig struct {
	WindowCandles  int // Number of candles per training window
	StepCandles    int // How many candles to advance per step (validation period)
	MinTrades      int // Minimum number of trades required for a result to be valid
	OptimizationConfig
}

// DefaultWalkForwardConfig provides sensible defaults.
func DefaultWalkForwardConfig() WalkForwardConfig {
	return WalkForwardConfig{
		WindowCandles:      100,
		StepCandles:        20,
		MinTrades:          1,
		OptimizationConfig: DefaultOptimizationConfig(),
	}
}

// WalkForwardWindow represents a single training + validation pair.
type WalkForwardWindow struct {
	TrainStart int // Start index in the candle array for training
	TrainEnd   int // End index for training (exclusive)
	ValidStart int // Start index for validation
	ValidEnd   int // End index for validation (exclusive)
}

// WalkForwardStepResult holds the results of a single walk-forward step.
type WalkForwardStepResult struct {
	Window          WalkForwardWindow
	BestParams      ParameterMap
	TrainScore      float64
	TrainResult     backtest.Result
	ValidationScore float64
	ValidationResult backtest.Result
}

// WalkForwardResult aggregates all walk-forward step results.
type WalkForwardResult struct {
	Steps       []WalkForwardStepResult
	AvgValScore float64
	BestStep    int // Index of the step with highest validation score
	TotalSteps  int
}

// WalkForwardCandles performs rolling-window walk-forward optimization.
// It splits the candle history into overlapping windows, runs grid search
// on each training window, and validates the best parameters on the next window.
func WalkForwardCandles(
	ctx context.Context,
	builder StrategyBuilder,
	candles []marketdata.Candle,
	initialCapital float64,
	opts backtest.EmulatorOptions,
	paramGrid map[string][]interface{},
	scorer ScoringFunction,
	wfConfig WalkForwardConfig,
) (*WalkForwardResult, error) {
	if len(candles) < wfConfig.WindowCandles+wfConfig.StepCandles {
		return nil, fmt.Errorf("insufficient candles (%d) for window=%d step=%d",
			len(candles), wfConfig.WindowCandles, wfConfig.StepCandles)
	}

	if scorer == nil {
		scorer = DefaultScorer
	}

	// Generate rolling windows
	windows := generateWindows(len(candles), wfConfig.WindowCandles, wfConfig.StepCandles)

	result := &WalkForwardResult{
		Steps:      make([]WalkForwardStepResult, 0, len(windows)),
		TotalSteps: len(windows),
	}

	var totalValScore float64
	bestValScore := -1e18
	bestStepIdx := 0

	for _, window := range windows {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		// Extract training candles
		trainCandles := candles[window.TrainStart:window.TrainEnd]
		trainHistory := &sliceCandleHistory{candles: trainCandles}

		// Run grid search on training window
		trainResults, err := GridSearchCandles(
			ctx, builder, trainHistory, initialCapital, opts,
			paramGrid, scorer, wfConfig.OptimizationConfig,
		)
		if err != nil {
			continue
		}

		// Filter results with minimum trades
		var validResults []RunResult
		for _, r := range trainResults {
			if r.Result.TotalTrades >= wfConfig.MinTrades {
				validResults = append(validResults, r)
			}
		}
		if len(validResults) == 0 {
			continue
		}

		bestTrainResult := validResults[0] // Already sorted by score descending

		// Validate best params on out-of-sample window
		validCandles := candles[window.ValidStart:window.ValidEnd]
		validHistory := &sliceCandleHistory{candles: validCandles}

		bestStrat, err := builder(bestTrainResult.Params)
		if err != nil {
			continue
		}

		eng := backtest.NewEngineWithOptions(bestStrat, initialCapital, opts)
		validRes, err := eng.RunCandles(ctx, validHistory)
		if err != nil {
			continue
		}

		valScore := scorer(validRes)

		stepResult := WalkForwardStepResult{
			Window:           window,
			BestParams:       bestTrainResult.Params,
			TrainScore:       bestTrainResult.Score,
			TrainResult:      bestTrainResult.Result,
			ValidationScore:  valScore,
			ValidationResult: validRes,
		}

		result.Steps = append(result.Steps, stepResult)
		totalValScore += valScore

		if valScore > bestValScore {
			bestValScore = valScore
			bestStepIdx = len(result.Steps) - 1
		}
	}

	if len(result.Steps) > 0 {
		result.AvgValScore = totalValScore / float64(len(result.Steps))
		result.BestStep = bestStepIdx
	}

	return result, nil
}

// generateWindows creates rolling training + validation window pairs.
func generateWindows(totalCandles, windowSize, stepSize int) []WalkForwardWindow {
	var windows []WalkForwardWindow
	trainStart := 0
	for {
		trainEnd := trainStart + windowSize
		validEnd := trainEnd + stepSize
		if validEnd > totalCandles {
			break
		}
		windows = append(windows, WalkForwardWindow{
			TrainStart: trainStart,
			TrainEnd:   trainEnd,
			ValidStart: trainEnd,
			ValidEnd:   validEnd,
		})
		trainStart += stepSize
	}
	return windows
}

// sliceCandleHistory is a simple CandleHistoryProvider backed by a slice.
type sliceCandleHistory struct {
	candles []marketdata.Candle
}

func (s *sliceCandleHistory) Candles() []marketdata.Candle {
	return s.candles
}

// WindowSplitResult holds the result of a single walk-forward split.
type WindowSplitResult struct {
	WindowIndex int
	TrainScore  float64
	ValidScore  float64
	Overfit     float64 // TrainScore - ValidScore (positive means overfitting)
}

// AnalyzeOverfitting computes overfitting metrics across all walk-forward steps.
func AnalyzeOverfitting(wfResult *WalkForwardResult) []WindowSplitResult {
	var analysis []WindowSplitResult
	for i, step := range wfResult.Steps {
		analysis = append(analysis, WindowSplitResult{
			WindowIndex: i,
			TrainScore:  step.TrainScore,
			ValidScore:  step.ValidationScore,
			Overfit:     step.TrainScore - step.ValidationScore,
		})
	}
	sort.SliceStable(analysis, func(i, j int) bool {
		return analysis[i].Overfit < analysis[j].Overfit // Least overfit first
	})
	return analysis
}
