package analytics_test

import (
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/analytics"
)

func generatePrices() []float64 {
	// Let's create an ideal head and shoulders pattern
	// Left shoulder at idx 5, Head at idx 15, Right shoulder at idx 25
	prices := make([]float64, 40)

	// Baseline
	for i := range prices {
		prices[i] = 100.0
	}

	// Left shoulder (110)
	prices[4] = 105.0
	prices[5] = 110.0
	prices[6] = 105.0

	// Head (130)
	prices[14] = 115.0
	prices[15] = 130.0
	prices[16] = 115.0

	// Right shoulder (108) - slightly off from left but within 2% tolerance
	prices[24] = 104.0
	prices[25] = 108.0
	prices[26] = 104.0

	return prices
}

func TestPatternRecognizer_HeadShoulders(t *testing.T) {
	prices := generatePrices()
	recognizer := analytics.NewPatternRecognizer(0.03) // 3% tolerance for shoulders (110 vs 108 is ~1.8% diff)

	results := recognizer.Scan(prices)

	foundHS := false
	for _, r := range results {
		if r.Type == analytics.HeadShoulders {
			foundHS = true
			if r.EndIndex != 25 {
				t.Errorf("Expected H&S to end at index 25 (Right Shoulder peak), got %d", r.EndIndex)
			}
			if r.Confidence < 0.1 || r.Confidence > 1.0 {
				t.Errorf("Expected confidence between 0.1 and 1.0, got %f", r.Confidence)
			}
		}
	}

	if !foundHS {
		t.Errorf("Failed to detect Head and Shoulders pattern")
	}
}

func TestPatternRecognizer_DoubleTop(t *testing.T) {
	// Create an ideal double top
	prices := make([]float64, 30)
	for i := range prices {
		prices[i] = 50.0
	}

	// Peak 1
	prices[9] = 95.0
	prices[10] = 100.0
	prices[11] = 95.0

	// Valley
	prices[14] = 70.0
	prices[15] = 60.0
	prices[16] = 70.0

	// Peak 2
	prices[19] = 95.0
	prices[20] = 100.0
	prices[21] = 95.0

	recognizer := analytics.NewPatternRecognizer(0.01) // 1% tolerance
	results := recognizer.Scan(prices)

	foundDT := false
	for _, r := range results {
		if r.Type == analytics.DoubleTop {
			foundDT = true
			if r.EndIndex != 20 {
				t.Errorf("Expected Double Top to end at index 20, got %d", r.EndIndex)
			}
			// Exact match should have very high confidence
			if r.Confidence < 0.99 {
				t.Errorf("Expected high confidence for identical peaks, got %f", r.Confidence)
			}
		}
	}

	if !foundDT {
		t.Errorf("Failed to detect Double Top pattern")
	}
}
