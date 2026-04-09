package regime

import (
	"math"
	"testing"
)

func generateTrendingCandles(n int, startPrice, trend float64) []CandleData {
	candles := make([]CandleData, n)
	for i := 0; i < n; i++ {
		price := startPrice + trend*float64(i)
		candles[i] = CandleData{
			Open:   price,
			High:   price + math.Abs(trend),
			Low:    price - math.Abs(trend),
			Close:  price + trend*0.5,
			Volume: 1000,
		}
	}
	return candles
}

func generateRangingCandles(n int, basePrice, rangeSize float64) []CandleData {
	candles := make([]CandleData, n)
	for i := 0; i < n; i++ {
		// Sinusoidal movement within range
		offset := math.Sin(float64(i)*0.5) * rangeSize
		price := basePrice + offset
		candles[i] = CandleData{
			Open:   price,
			High:   price + rangeSize*0.1,
			Low:    price - rangeSize*0.1,
			Close:  price,
			Volume: 1000,
		}
	}
	return candles
}

func generateVolatileCandles(n int, basePrice float64) []CandleData {
	candles := make([]CandleData, n)
	for i := 0; i < n; i++ {
		swing := basePrice * 0.08 // 8% swings = volatile
		if i%2 == 0 {
			swing = -swing
		}
		price := basePrice + swing
		candles[i] = CandleData{
			Open:   basePrice,
			High:   price + basePrice*0.04,
			Low:    price - basePrice*0.04,
			Close:  price,
			Volume: 5000,
		}
	}
	return candles
}

func TestRegime_String(t *testing.T) {
	tests := []struct {
		r    Regime
		want string
	}{
		{RegimeTrending, "TRENDING"},
		{RegimeRanging, "RANGING"},
		{RegimeVolatile, "VOLATILE"},
		{RegimeQuiet, "QUIET"},
		{Regime(99), "UNKNOWN"},
	}
	for _, tt := range tests {
		if got := tt.r.String(); got != tt.want {
			t.Errorf("Regime(%d).String() = %q, want %q", tt.r, got, tt.want)
		}
	}
}

func TestVolatilityDetector_Trending(t *testing.T) {
	d := NewVolatilityDetector(0.05, 0.01, 14)
	candles := generateTrendingCandles(30, 100, 1.5) // Strong uptrend

	regime := d.Detect(candles)
	if regime != RegimeTrending {
		t.Errorf("expected TRENDING for strong trend, got %s", regime)
	}
}

func TestVolatilityDetector_Volatile(t *testing.T) {
	d := NewVolatilityDetector(0.05, 0.01, 14)
	candles := generateVolatileCandles(30, 100)

	regime := d.Detect(candles)
	if regime != RegimeVolatile {
		t.Errorf("expected VOLATILE for large swings, got %s", regime)
	}
}

func TestVolatilityDetector_Quiet(t *testing.T) {
	d := NewVolatilityDetector(0.05, 0.01, 14)
	candles := generateRangingCandles(30, 100, 0.5) // Very tight range

	regime := d.Detect(candles)
	if regime != RegimeQuiet {
		t.Errorf("expected QUIET for tight range, got %s", regime)
	}
}

func TestVolatilityDetector_InsufficientData(t *testing.T) {
	d := NewVolatilityDetector(0.05, 0.01, 14)
	regime := d.Detect([]CandleData{{Close: 100}})
	if regime != RegimeQuiet {
		t.Errorf("expected QUIET for insufficient data, got %s", regime)
	}
}

func TestTrendDetector_Trending(t *testing.T) {
	d := NewTrendDetector(14, 20, 40)
	candles := generateTrendingCandles(30, 100, 2.0)

	regime := d.Detect(candles)
	if regime != RegimeTrending {
		t.Errorf("expected TRENDING for strong directional move, got %s", regime)
	}
}

func TestTrendDetector_QuietMarket(t *testing.T) {
	d := NewTrendDetector(14, 20, 40)
	// Generate perfectly flat candles
	candles := make([]CandleData, 30)
	for i := range candles {
		candles[i] = CandleData{
			High:  100.1,
			Low:   99.9,
			Close: 100,
		}
	}

	regime := d.Detect(candles)
	if regime == RegimeTrending {
		t.Errorf("expected non-trending for flat market, got %s", regime)
	}
}

func TestBollingerBandwidthDetector_Volatile(t *testing.T) {
	d := NewBollingerBandwidthDetector(20, 2.0, 0.02, 0.05)
	candles := generateVolatileCandles(30, 100)

	regime := d.Detect(candles)
	if regime != RegimeVolatile {
		t.Errorf("expected VOLATILE for wide Bollinger bands, got %s", regime)
	}
}

func TestBollingerBandwidthDetector_InsufficientData(t *testing.T) {
	d := NewBollingerBandwidthDetector(20, 2.0, 0.02, 0.05)
	regime := d.Detect([]CandleData{{Close: 100}})
	if regime != RegimeQuiet {
		t.Errorf("expected QUIET for insufficient data, got %s", regime)
	}
}

func TestCompositeDetector_Majority(t *testing.T) {
	cd := NewCompositeDetector(
		&mockDetector{regime: RegimeTrending},
		&mockDetector{regime: RegimeTrending},
		&mockDetector{regime: RegimeVolatile},
	)

	regime := cd.Detect(nil)
	if regime != RegimeTrending {
		t.Errorf("expected TRENDING from majority vote, got %s", regime)
	}
}

func TestCompositeDetector_Unanimous(t *testing.T) {
	cd := NewCompositeDetector(
		&mockDetector{regime: RegimeQuiet},
		&mockDetector{regime: RegimeQuiet},
	)

	regime := cd.Detect(nil)
	if regime != RegimeQuiet {
		t.Errorf("expected QUIET from unanimous vote, got %s", regime)
	}
}

func TestCompositeDetector_Empty(t *testing.T) {
	cd := NewCompositeDetector()
	regime := cd.Detect(nil)
	if regime != RegimeQuiet {
		t.Errorf("expected QUIET for no detectors, got %s", regime)
	}
}

func TestCalculateATR(t *testing.T) {
	candles := []CandleData{
		{High: 102, Low: 98, Close: 100},
		{High: 105, Low: 99, Close: 104},
		{High: 107, Low: 103, Close: 106},
	}
	atr := calculateATR(candles, 3)
	if atr <= 0 {
		t.Errorf("expected positive ATR, got %f", atr)
	}
}

// mockDetector for testing composite
type mockDetector struct {
	regime Regime
}

func (m *mockDetector) Name() string           { return "mock" }
func (m *mockDetector) Detect(_ []CandleData) Regime { return m.regime }
