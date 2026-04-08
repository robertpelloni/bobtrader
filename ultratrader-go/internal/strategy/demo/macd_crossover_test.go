package demo

import (
	"context"
	"fmt"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

func formatFloat(f float64) string {
	return fmt.Sprintf("%v", f)
}

func TestMACDCrossover_BullishCross(t *testing.T) {
	s := NewMACDCrossover("paper-main", "BTCUSDT", "0.01", 3, 6, 3)

	// Long downtrend followed by strong uptrend to force histogram crossover
	prices := []float64{
		100, 98, 96, 94, 92, 90, 88, 86, 84, 82, 80, // downtrend
		82, 85, 90, 96, 104, 114, 126, // strong uptrend
	}

	totalBuy := 0
	totalSell := 0
	for _, p := range prices {
		candle := marketdata.Candle{Symbol: "BTCUSDT", Close: formatFloat(p), High: formatFloat(p + 2), Low: formatFloat(p - 2)}
		signals, err := s.CandleEvent(context.Background(), candle)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, sig := range signals {
			if sig.Action == "buy" {
				totalBuy++
			}
			if sig.Action == "sell" {
				totalSell++
			}
		}
	}

	if totalBuy == 0 {
		t.Errorf("expected at least one buy signal during MACD bullish crossover, got buy=%d sell=%d", totalBuy, totalSell)
	}
}

func TestMACDCrossover_BearishCross(t *testing.T) {
	s := NewMACDCrossover("paper-main", "BTCUSDT", "0.01", 3, 6, 3)

	// Strong uptrend followed by sharp downtrend
	prices := []float64{
		80, 85, 90, 96, 104, 114, 126, 140, // uptrend
		130, 115, 100, 85, 70, // sharp downtrend
	}

	totalBuy := 0
	totalSell := 0
	for _, p := range prices {
		candle := marketdata.Candle{Symbol: "BTCUSDT", Close: formatFloat(p), High: formatFloat(p + 2), Low: formatFloat(p - 2)}
		signals, err := s.CandleEvent(context.Background(), candle)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, sig := range signals {
			if sig.Action == "buy" {
				totalBuy++
			}
			if sig.Action == "sell" {
				totalSell++
			}
		}
	}

	if totalSell == 0 {
		t.Errorf("expected at least one sell signal during MACD bearish crossover, got buy=%d sell=%d", totalBuy, totalSell)
	}
}

func TestMACDCrossover_Name(t *testing.T) {
	s := NewMACDCrossover("paper-main", "BTCUSDT", "0.01", 12, 26, 9)
	if s.Name() != "MACDCrossover" {
		t.Errorf("expected name MACDCrossover, got %s", s.Name())
	}
}
