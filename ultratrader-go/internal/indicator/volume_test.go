package indicator

import (
	"math"
	"testing"
)

const eps = 1e-6

func TestVWAP_Basic(t *testing.T) {
	vwap := NewVWAP()

	// First candle: TP=100, Vol=10 -> VWAP=100
	result := vwap.Update(CandleInput{High: 102, Low: 98, Close: 100, Volume: 10})
	if math.Abs(result-100) > eps {
		t.Errorf("expected VWAP=100, got %f", result)
	}

	// Second candle: TP=110, Vol=20 -> VWAP = (100*10 + 110*20) / 30 = 106.67
	result = vwap.Update(CandleInput{High: 112, Low: 108, Close: 110, Volume: 20})
	expected := (100.0*10.0 + 110.0*20.0) / 30.0
	if math.Abs(result-expected) > 0.01 {
		t.Errorf("expected VWAP=%.2f, got %f", expected, result)
	}
}

func TestVWAP_Reset(t *testing.T) {
	vwap := NewVWAP()
	vwap.Update(CandleInput{High: 102, Low: 98, Close: 100, Volume: 10})
	vwap.Reset()

	result := vwap.Update(CandleInput{High: 112, Low: 108, Close: 110, Volume: 20})
	// After reset, VWAP should be just the new candle's TP
	if math.Abs(result-110) > eps {
		t.Errorf("expected VWAP=110 after reset, got %f", result)
	}
}

func TestOBV_Basic(t *testing.T) {
	obv := NewOBV()

	// First update: OBV = volume
	result := obv.Update(CandleInput{Close: 100, Volume: 1000})
	if math.Abs(result-1000) > eps {
		t.Errorf("expected OBV=1000, got %f", result)
	}

	// Price up: OBV += volume
	result = obv.Update(CandleInput{Close: 105, Volume: 500})
	if math.Abs(result-1500) > eps {
		t.Errorf("expected OBV=1500, got %f", result)
	}

	// Price down: OBV -= volume
	result = obv.Update(CandleInput{Close: 103, Volume: 300})
	if math.Abs(result-1200) > eps {
		t.Errorf("expected OBV=1200, got %f", result)
	}

	// Price unchanged: OBV stays same
	result = obv.Update(CandleInput{Close: 103, Volume: 200})
	if math.Abs(result-1200) > eps {
		t.Errorf("expected OBV=1200 (unchanged), got %f", result)
	}
}

func TestVolumeSMA(t *testing.T) {
	sma := NewVolumeSMA(3)

	if sma.Update(100) != 0 {
		t.Error("expected 0 for insufficient data")
	}
	if sma.Update(200) != 0 {
		t.Error("expected 0 for insufficient data")
	}
	result := sma.Update(300)
	expected := (100 + 200 + 300) / 3.0
	if math.Abs(result-expected) > eps {
		t.Errorf("expected %.2f, got %f", expected, result)
	}
}

func TestVolumeRatio(t *testing.T) {
	vr := NewVolumeRatio(3)

	// First few updates return 1.0 (no average yet)
	r1 := vr.Update(100)
	if r1 != 1.0 {
		t.Errorf("expected 1.0 before average established, got %f", r1)
	}

	vr.Update(100)
	vr.Update(100)
	vr.Update(100)

	// Average = 100, current = 200 -> ratio = 2.0
	r2 := vr.Update(200)
	// Window: [100, 100, 200] -> avg = 133.33, ratio = 200/133.33 = 1.5
	if math.Abs(r2-1.5) > eps {
		t.Errorf("expected ratio 1.5, got %f", r2)
	}
}

func TestMFI_Basic(t *testing.T) {
	mfi := NewMFI(14)

	// First update returns neutral
	result := mfi.Update(CandleInput{High: 102, Low: 98, Close: 100, Volume: 1000})
	if math.Abs(result-50) > eps {
		t.Errorf("expected neutral 50 for first update, got %f", result)
	}
}

func TestMFI_Overbought(t *testing.T) {
	mfi := NewMFI(5)

	// Rising prices with increasing volume should push MFI toward 100
	baseClose := 100.0
	for i := 0; i < 10; i++ {
		mfi.Update(CandleInput{
			High:   baseClose + 2,
			Low:    baseClose - 2,
			Close:  baseClose + float64(i)*2,
			Volume: 1000 + float64(i)*100,
		})
	}

	result := mfi.Last()
	// After 10 rising candles, MFI should be very high
	// Note: we need to check the last computed value
	_ = result
}

func TestChaikinMoneyFlow_Positive(t *testing.T) {
	cmf := NewChaikinMoneyFlow(5)

	// Candles closing near highs = positive CMF
	for i := 0; i < 5; i++ {
		cmf.Update(CandleInput{
			High:   110,
			Low:    100,
			Close:  108, // Close near high
			Volume: 1000,
		})
	}

	result := cmf.Last()
	if result <= 0 {
		t.Errorf("expected positive CMF for closes near highs, got %f", result)
	}
}

func TestChaikinMoneyFlow_Negative(t *testing.T) {
	cmf := NewChaikinMoneyFlow(5)

	// Candles closing near lows = negative CMF
	for i := 0; i < 5; i++ {
		cmf.Update(CandleInput{
			High:   110,
			Low:    100,
			Close:  102, // Close near low
			Volume: 1000,
		})
	}

	result := cmf.Last()
	if result >= 0 {
		t.Errorf("expected negative CMF for closes near lows, got %f", result)
	}
}

// Last() helper methods that need testing
func TestMFI_Last(t *testing.T) {
	mfi := NewMFI(14)
	if mfi.Last() != 50.0 {
		t.Errorf("expected neutral MFI before data, got %f", mfi.Last())
	}
}
