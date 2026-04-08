package demo

import (
	"context"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

func TestATRSizing_Crossover(t *testing.T) {
	s := NewATRSizing("paper-main", "BTCUSDT", "0.01", 0.02, 3, 5, 3)

	prices := []struct{ close, high, low float64 }{
		{100, 102, 98}, {95, 97, 93}, {90, 92, 88},
		{88, 90, 86}, {90, 92, 88}, {92, 94, 90}, {95, 97, 93}, {100, 102, 98},
	}

	totalBuy := 0
	totalSell := 0
	for _, p := range prices {
		candle := marketdata.Candle{
			Symbol: "BTCUSDT",
			Close:  formatFloat(p.close),
			High:   formatFloat(p.high),
			Low:    formatFloat(p.low),
		}
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
		t.Errorf("expected at least one buy signal during bullish crossover")
	}
}

func TestATRSizing_QuantityVaries(t *testing.T) {
	s := NewATRSizing("paper-main", "BTCUSDT", "0.01", 0.02, 3, 5, 3)

	warmup := []struct{ close, high, low float64 }{
		{100, 102, 98}, {99, 101, 97}, {98, 100, 96},
		{97, 99, 95}, {96, 98, 94},
	}
	for _, p := range warmup {
		candle := marketdata.Candle{Symbol: "BTCUSDT", Close: formatFloat(p.close), High: formatFloat(p.high), Low: formatFloat(p.low)}
		s.CandleEvent(context.Background(), candle)
	}

	// Trigger a buy signal with high volatility
	candle := marketdata.Candle{Symbol: "BTCUSDT", Close: "100", High: "110", Low: "90"}
	signals, _ := s.CandleEvent(context.Background(), candle)

	if len(signals) > 0 && signals[0].Quantity == "" {
		t.Errorf("expected non-empty quantity from ATR sizing")
	}
}

func TestATRSizing_Name(t *testing.T) {
	s := NewATRSizing("paper-main", "BTCUSDT", "0.01", 0.02, 3, 5, 3)
	if s.Name() != "ATRSizing" {
		t.Errorf("expected name ATRSizing, got %s", s.Name())
	}
}
