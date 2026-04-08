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
