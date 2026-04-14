package optimizer_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest/optimizer"
)

type mockOptEvaluator struct{}

func (m *mockOptEvaluator) Evaluate(ctx context.Context, params optimizer.ParameterSet, start, end time.Time) (float64, error) {
	// Let's pretend the optimal is fast=10, slow=20
	// We'll calculate a score based on distance from optimal
	fast := params["fast"]
	slow := params["slow"]

	distFast := fast - 10
	if distFast < 0 {
		distFast = -distFast
	}

	distSlow := slow - 20
	if distSlow < 0 {
		distSlow = -distSlow
	}

	// Max score is 100, drops by distance
	score := 100.0 - (distFast * 5.0) - (distSlow * 5.0)
	return score, nil
}

func TestGenerateGrid(t *testing.T) {
	ranges := map[string]optimizer.ParamRange{
		"fast": {Min: 5, Max: 15, Step: 5},   // 5, 10, 15 (3 vals)
		"slow": {Min: 10, Max: 30, Step: 10}, // 10, 20, 30 (3 vals)
	}

	grid, err := optimizer.GenerateGrid(ranges)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(grid) != 9 {
		t.Fatalf("Expected 9 combinations, got %d", len(grid))
	}

	// Make sure 10, 20 is in there
	found := false
	for _, p := range grid {
		if p["fast"] == 10 && p["slow"] == 20 {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected to find fast=10, slow=20 combination")
	}
}

func TestGenerateGrid_Invalid(t *testing.T) {
	ranges := map[string]optimizer.ParamRange{
		"fast": {Min: 15, Max: 5, Step: 5}, // min > max
	}
	_, err := optimizer.GenerateGrid(ranges)
	if err == nil {
		t.Errorf("Expected error for min > max")
	}

	ranges = map[string]optimizer.ParamRange{
		"fast": {Min: 5, Max: 15, Step: 0}, // step = 0
	}
	_, err = optimizer.GenerateGrid(ranges)
	if err == nil {
		t.Errorf("Expected error for step <= 0")
	}
}

func TestGridSearchOptimizer_Optimize(t *testing.T) {
	eval := &mockOptEvaluator{}
	opt := optimizer.NewGridSearchOptimizer(eval)

	ranges := map[string]optimizer.ParamRange{
		"fast": {Min: 5, Max: 15, Step: 5},
		"slow": {Min: 10, Max: 30, Step: 10},
	}
	grid, _ := optimizer.GenerateGrid(ranges)

	ctx := context.Background()
	results, err := opt.Optimize(ctx, grid, time.Now(), time.Now(), 4)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) != 9 {
		t.Fatalf("Expected 9 results, got %d", len(results))
	}

	// Sort manually here just to double check, but it should already be sorted by Optimize
	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})

	best := results[0]
	if best.Score != 100.0 {
		t.Errorf("Expected top score 100.0, got %f", best.Score)
	}

	if best.Parameters["fast"] != 10 || best.Parameters["slow"] != 20 {
		t.Errorf("Expected best params fast=10, slow=20, got fast=%f, slow=%f", best.Parameters["fast"], best.Parameters["slow"])
	}
}
