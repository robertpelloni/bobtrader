package analytics_test

import (
	"math"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/analytics"
)

func TestArbitrageDetector_AnalyzePair(t *testing.T) {
	// Two completely positively correlated series.
	// B is exactly half of A.
	pricesA := []float64{100, 105, 110, 115, 120, 115, 110, 105, 100, 95, 90}
	pricesB := []float64{50, 52.5, 55, 57.5, 60, 57.5, 55, 52.5, 50, 47.5, 45}

	detector := analytics.NewArbitrageDetector(10)

	// Truncate first element to fit window size of 10 exactly
	stats, err := detector.AnalyzePair("A", "B", pricesA[1:], pricesB[1:])
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should be perfectly correlated (Pearson ≈ 1.0)
	if math.Abs(stats.Correlation-1.0) > 0.001 {
		t.Errorf("Expected correlation 1.0, got %f", stats.Correlation)
	}

	// Spread is log(A)-log(B) = log(A/B) = log(2)
	expectedSpread := math.Log(2.0)
	if math.Abs(stats.SpreadMean-expectedSpread) > 0.001 {
		t.Errorf("Expected spread mean %f, got %f", expectedSpread, stats.SpreadMean)
	}

	// Because spread is perfectly constant, std dev is 0 and Z-Score is 0
	if stats.SpreadStdDev > 0.001 {
		t.Errorf("Expected 0 std dev, got %f", stats.SpreadStdDev)
	}

	if stats.ZScore != 0 {
		t.Errorf("Expected z-score 0, got %f", stats.ZScore)
	}
}

func TestArbitrageDetector_CheckSignal(t *testing.T) {
	detector := analytics.NewArbitrageDetector(10)

	// Simulate a z-score of 2.5 (overvalued spread)
	stats := analytics.PairStats{
		ZScore: 2.5,
	}

	action := detector.CheckSignal(stats, 2.0, 0.5)
	if action != "SHORT_A_LONG_B" {
		t.Errorf("Expected SHORT_A_LONG_B, got %s", action)
	}

	// Simulate mean reversion (z-score drops below 0.5)
	stats.ZScore = 0.4
	action = detector.CheckSignal(stats, 2.0, 0.5)
	if action != "FLAT" {
		t.Errorf("Expected FLAT, got %s", action)
	}

	// Simulate undervalued spread (z-score -2.1)
	stats.ZScore = -2.1
	action = detector.CheckSignal(stats, 2.0, 0.5)
	if action != "LONG_A_SHORT_B" {
		t.Errorf("Expected LONG_A_SHORT_B, got %s", action)
	}
}
