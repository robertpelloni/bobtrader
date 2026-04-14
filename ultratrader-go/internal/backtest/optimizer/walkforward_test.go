package optimizer_test

import (
	"context"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest/optimizer"
)

type mockEvaluator struct {
	// A simple mock that scores based on sum of params + a bonus if it's the testing window
	// to simulate that some params overfit.
}

func (m *mockEvaluator) Evaluate(ctx context.Context, params optimizer.ParameterSet, start, end time.Time) (float64, error) {
	var score float64
	for _, v := range params {
		score += v
	}

	// Make out of sample slightly worse to simulate realistic degradation
	if end.Sub(start) <= 30*24*time.Hour { // Assume test windows are smaller
		score *= 0.8
	}

	return score, nil
}

func TestGenerateWindows(t *testing.T) {
	start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 12, 31, 0, 0, 0, 0, time.UTC)

	trainDur := 90 * 24 * time.Hour
	testDur := 30 * 24 * time.Hour

	windows := optimizer.GenerateWindows(start, end, trainDur, testDur)

	if len(windows) != 9 { // Jan-Mar(Train) -> Apr(Test) ... steps 1 month at a time
		t.Fatalf("Expected 9 windows, got %d", len(windows))
	}

	if windows[0].TrainStart != start {
		t.Errorf("First window should start at %v", start)
	}

	if windows[0].TestStart != windows[0].TrainEnd {
		t.Errorf("Test should start immediately after train")
	}
}

func TestWalkForwardOptimizer_Run(t *testing.T) {
	grid := []optimizer.ParameterSet{
		{"rsi_period": 14, "sma_period": 50},
		{"rsi_period": 21, "sma_period": 200},
	}

	eval := &mockEvaluator{}
	wfo := optimizer.NewWalkForwardOptimizer(eval, grid)

	start := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2023, 6, 1, 0, 0, 0, 0, time.UTC)
	windows := optimizer.GenerateWindows(start, end, 60*24*time.Hour, 30*24*time.Hour)

	ctx := context.Background()
	results, err := wfo.Run(ctx, windows)

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) != len(windows) {
		t.Fatalf("Expected %d results, got %d", len(windows), len(results))
	}

	// The second parameter set (21+200 = 221) should always win over the first (14+50 = 64)
	for _, r := range results {
		if r.Parameters["rsi_period"] != 21 {
			t.Errorf("Expected rsi_period=21 to win, got %f", r.Parameters["rsi_period"])
		}

		if r.InSample <= r.OutOfSample {
			t.Errorf("Expected InSample > OutOfSample due to our mock logic")
		}
	}

	agg := optimizer.Aggregate(results)
	if agg <= 0 {
		t.Errorf("Expected positive aggregate score")
	}
}
