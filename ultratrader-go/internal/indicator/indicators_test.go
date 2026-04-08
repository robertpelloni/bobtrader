package indicator

import (
	"math"
	"testing"
)

func TestSMA(t *testing.T) {
	sma := NewSMA(3)
	sma.Update(10)
	sma.Update(20)
	val := sma.Update(30)
	if val != 20 {
		t.Errorf("expected 20, got %v", val)
	}
	val = sma.Update(40)
	if val != 30 {
		t.Errorf("expected 30, got %v", val)
	}
}

func TestEMA(t *testing.T) {
	ema := NewEMA(3)
	val := ema.Update(10)
	if val != 10 {
		t.Errorf("expected 10, got %v", val)
	}
	val = ema.Update(20)
	// k = 2/(3+1) = 0.5
	// ema = (20-10)*0.5 + 10 = 15
	if val != 15 {
		t.Errorf("expected 15, got %v", val)
	}
}

func TestRSI(t *testing.T) {
	rsi := NewRSI(2)
	rsi.Update(10)
	rsi.Update(12) // gain 2
	rsi.Update(11) // loss 1
	val := rsi.Last()
	// avgGain = 2/2 = 1, avgLoss = 1/2 = 0.5
	// rs = 1/0.5 = 2
	// rsi = 100 - 100/3 = 66.66
	if math.Abs(val-66.666666) > 0.01 {
		t.Errorf("expected approx 66.66, got %v", val)
	}
}

func TestMACD(t *testing.T) {
	macd := NewMACD(3, 6, 3)

	// Feed enough data to warm up both fast and slow EMAs
	prices := []float64{10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}
	var result MACDResult
	for _, p := range prices {
		result = macd.Update(p)
	}

	// After a steady uptrend, MACD line should be positive
	if result.MACD <= 0 {
		t.Errorf("expected positive MACD line after uptrend, got %v", result.MACD)
	}
	// Histogram should equal MACD - Signal
	if math.Abs(result.Histogram-(result.MACD-result.Signal)) > 0.0001 {
		t.Errorf("histogram should equal MACD - Signal, got histogram=%v macd=%v signal=%v",
				result.Histogram, result.MACD, result.Signal)
	}
}

func TestMACDLast(t *testing.T) {
	macd := NewMACD(3, 6, 3)
	prices := []float64{10, 20, 30, 40, 50}
	var lastResult MACDResult
	for _, p := range prices {
		lastResult = macd.Update(p)
	}
	checkResult := macd.Last()
	if math.Abs(checkResult.MACD-lastResult.MACD) > 0.0001 {
		t.Errorf("Last() MACD should match last Update(), got %v vs %v", checkResult.MACD, lastResult.MACD)
	}
}

func TestBollingerBands(t *testing.T) {
	bb := NewBollingerBands(5, 2.0)

	// Feed constant prices — all bands should equal the price, bandwidth should be 0
	for i := 0; i < 5; i++ {
		bb.Update(100)
	}
	result := bb.Last()
	if math.Abs(result.Upper-100) > 0.01 {
		t.Errorf("expected upper band 100 with constant prices, got %v", result.Upper)
	}
	if math.Abs(result.Lower-100) > 0.01 {
		t.Errorf("expected lower band 100 with constant prices, got %v", result.Lower)
	}
	if math.Abs(result.Middle-100) > 0.01 {
		t.Errorf("expected middle band 100 with constant prices, got %v", result.Middle)
	}
	if result.Bandwidth > 0.01 {
		t.Errorf("expected zero bandwidth with constant prices, got %v", result.Bandwidth)
	}
}

func TestBollingerBandsVarying(t *testing.T) {
	bb := NewBollingerBands(4, 2.0)
	// prices: 10, 20, 30, 40
	// mean = 25, stddev = sqrt((225+25+25+225)/4) = sqrt(125) ≈ 11.18
	bb.Update(10)
	bb.Update(20)
	bb.Update(30)
	result := bb.Update(40)

	expectedMiddle := 25.0
	if math.Abs(result.Middle-expectedMiddle) > 0.01 {
		t.Errorf("expected middle %v, got %v", expectedMiddle, result.Middle)
	}
	// Upper should be > Middle, Lower should be < Middle
	if result.Upper <= result.Middle {
		t.Errorf("upper band should be above middle, got upper=%v middle=%v", result.Upper, result.Middle)
	}
	if result.Lower >= result.Middle {
		t.Errorf("lower band should be below middle, got lower=%v middle=%v", result.Lower, result.Middle)
	}
	if result.Bandwidth <= 0 {
		t.Errorf("bandwidth should be positive with varying prices, got %v", result.Bandwidth)
	}
}

func TestBollingerBandsInsufficientData(t *testing.T) {
	bb := NewBollingerBands(5, 2.0)
	result := bb.Update(100) // only 1 data point, period is 5
	if result.Upper != 0 || result.Middle != 0 || result.Lower != 0 {
		t.Errorf("expected zero result with insufficient data, got upper=%v middle=%v lower=%v",
				result.Upper, result.Middle, result.Lower)
	}
}

func TestATR(t *testing.T) {
	atr := NewATR(3)

	// First update just initializes
	atr.Update(10, 8, 9)

	// Second: high=12, low=10, close=11, prev=9
	// TR = max(12-10, |12-9|, |10-9|) = max(2, 3, 1) = 3
	atr.Update(12, 10, 11)

	// Third: high=15, low=11, close=13, prev=11
	// TR = max(15-11, |15-11|, |11-11|) = max(4, 4, 0) = 4
	atr.Update(15, 11, 13)

	// Fourth: high=14, low=12, close=13, prev=13
	// TR = max(14-12, |14-13|, |12-13|) = max(2, 1, 1) = 2
	val := atr.Update(14, 12, 13)

	// After 3 periods: ATR = (3 + 4 + 2) / 3 = 3.0
	if math.Abs(val-3.0) > 0.01 {
		t.Errorf("expected ATR 3.0, got %v", val)
	}
}

func TestATRInsufficientData(t *testing.T) {
	atr := NewATR(5)
	atr.Update(10, 8, 9)
	val := atr.Update(12, 10, 11)
	// Only 2 true ranges collected, not yet 5 periods
	if val <= 0 {
		t.Errorf("expected partial ATR > 0, got %v", val)
	}
}

