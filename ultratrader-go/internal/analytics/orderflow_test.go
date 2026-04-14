package analytics_test

import (
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/analytics"
)

func TestOrderFlowAnalyzer_Analyze(t *testing.T) {
	// Bearish Divergence: Price goes UP, but net selling pressure (Delta) increases
	// 5 periods
	prices := []float64{100, 105, 110, 115, 120} // Up trend
	buyVols := []float64{10, 15, 5, 2, 0}
	sellVols := []float64{5, 10, 20, 25, 30} // Sell pressure building

	analyzer := analytics.NewOrderFlowAnalyzer(5)

	results, divType, err := analyzer.Analyze(prices, buyVols, sellVols)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(results) != 5 {
		t.Errorf("Expected 5 order flow results")
	}

	// Delta should be negative towards the end
	if results[4].Delta >= 0 {
		t.Errorf("Expected negative delta at end")
	}

	if divType != analytics.BearishDivergence {
		t.Errorf("Expected BearishDivergence, got %v", divType)
	}
}
