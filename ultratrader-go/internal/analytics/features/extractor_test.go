package features

import (
	"math"
	"testing"
)

func generateCandles(n int) []CandleData {
	candles := make([]CandleData, n)
	for i := 0; i < n; i++ {
		price := 100 + float64(i)*0.5
		candles[i] = CandleData{
			Open:   price - 0.5,
			High:   price + 1.0,
			Low:    price - 1.0,
			Close:  price,
			Volume: 1000 + float64(i)*10,
		}
	}
	return candles
}

func TestExtractor_BasicFeatures(t *testing.T) {
	ext := NewExtractor(10, 5)
	candles := generateCandles(20)

	var features FeatureMap
	for _, c := range candles {
		features = ext.Update(c)
	}

	if features["close"] <= 0 {
		t.Error("expected positive close")
	}
	if features["volume"] <= 0 {
		t.Error("expected positive volume")
	}
}

func TestExtractor_Returns(t *testing.T) {
	ext := NewExtractor(10, 5)
	candles := generateCandles(15)

	var features FeatureMap
	for _, c := range candles {
		features = ext.Update(c)
	}

	// Should have 1-period return
	if _, ok := features["return_1"]; !ok {
		t.Error("expected return_1 feature")
	}
	// Should have 5-period return
	if _, ok := features["return_5"]; !ok {
		t.Error("expected return_5 feature")
	}
}

func TestExtractor_RSI(t *testing.T) {
	ext := NewExtractor(14, 5)

	// Rising prices should push RSI up
	for i := 0; i < 20; i++ {
		ext.Update(CandleData{Close: 100 + float64(i)*2, Volume: 1000})
	}

	features := ext.Update(CandleData{Close: 150, Volume: 1000})
	if features["rsi"] < 50 {
		t.Errorf("expected RSI > 50 for rising prices, got %f", features["rsi"])
	}
}

func TestExtractor_Volatility(t *testing.T) {
	ext := NewExtractor(10, 5)

	// Create volatile candles
	for i := 0; i < 15; i++ {
		swing := 5.0
		if i%2 == 0 {
			swing = -5.0
		}
		ext.Update(CandleData{Close: 100 + swing, Volume: 1000})
	}

	features := ext.Update(CandleData{Close: 100, Volume: 1000})
	if features["volatility"] <= 0 {
		t.Error("expected positive volatility for swingy prices")
	}
}

func TestExtractor_VolumeRatio(t *testing.T) {
	ext := NewExtractor(10, 5)

	// Normal volume
	for i := 0; i < 5; i++ {
		ext.Update(CandleData{Close: 100, Volume: 1000})
	}

	// Spike volume
	features := ext.Update(CandleData{Close: 100, Volume: 5000})
	if features["volume_ratio"] < 2.0 {
		t.Errorf("expected volume_ratio > 2.0 for spike, got %f", features["volume_ratio"])
	}
}

func TestExtractor_CandleShape(t *testing.T) {
	ext := NewExtractor(10, 5)

	features := ext.Update(CandleData{
		Open:   100,
		High:   105,
		Low:    95,
		Close:  103,
		Volume: 1000,
	})

	if features["range_pct"] <= 0 {
		t.Error("expected positive range_pct")
	}
	if features["body_pct"] <= 0 {
		t.Error("expected positive body_pct")
	}
	if features["upper_shadow"] < 0 || features["lower_shadow"] < 0 {
		t.Error("expected non-negative shadow values")
	}
}

func TestExtractor_Names(t *testing.T) {
	ext := NewExtractor(10, 5)
	names := ext.Names()
	if len(names) < 10 {
		t.Errorf("expected at least 10 feature names, got %d", len(names))
	}
}

func TestBuildMatrix(t *testing.T) {
	candles := generateCandles(20)
	matrix := BuildMatrix(candles, 10, 5)

	if len(matrix.Headers) < 10 {
		t.Errorf("expected at least 10 headers, got %d", len(matrix.Headers))
	}
	if len(matrix.Rows) != 20 {
		t.Errorf("expected 20 rows, got %d", len(matrix.Rows))
	}
	if len(matrix.Rows[0]) != len(matrix.Headers) {
		t.Errorf("row width %d != header count %d", len(matrix.Rows[0]), len(matrix.Headers))
	}
}

func TestBuildMatrix_SufficientData(t *testing.T) {
	candles := generateCandles(30)
	matrix := BuildMatrix(candles, 14, 5)

	// Last row should have all features populated
	lastRow := matrix.Rows[len(matrix.Rows)-1]
	emptyCount := 0
	for _, v := range lastRow {
		if math.IsNaN(v) || v == 0 {
			emptyCount++
		}
	}
	// Most features should be populated after 30 candles
	if emptyCount > len(matrix.Headers)/2 {
		t.Errorf("too many empty features in last row: %d/%d", emptyCount, len(matrix.Headers))
	}
}
