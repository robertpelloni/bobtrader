package demo

import (
	"context"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

func TestBollingerReversion_BuyAtLowerBand(t *testing.T) {
	s := NewBollingerReversion("paper-main", "BTCUSDT", "0.01", 5, 2.0)

	prices := []float64{100, 100, 100, 100, 100, 80} // last one drops below lower band
	var signals []strategy.Signal
	var err error
	for _, p := range prices {
		candle := marketdata.Candle{Symbol: "BTCUSDT", Close: formatFloat(p), High: formatFloat(p + 1), Low: formatFloat(p - 1)}
		signals, err = s.CandleEvent(context.Background(), candle)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	buyCount := 0
	for _, sig := range signals {
		if sig.Action == "buy" {
			buyCount++
		}
	}
	if buyCount == 0 {
		t.Errorf("expected at least one buy signal when price drops to lower band")
	}
}

func TestBollingerReversion_SellAtUpperBand(t *testing.T) {
	s := NewBollingerReversion("paper-main", "BTCUSDT", "0.01", 5, 2.0)

	prices := []float64{100, 100, 100, 100, 100, 120} // last one spikes above upper band
	var signals []strategy.Signal
	var err error
	for _, p := range prices {
		candle := marketdata.Candle{Symbol: "BTCUSDT", Close: formatFloat(p), High: formatFloat(p + 1), Low: formatFloat(p - 1)}
		signals, err = s.CandleEvent(context.Background(), candle)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}

	sellCount := 0
	for _, sig := range signals {
		if sig.Action == "sell" {
			sellCount++
		}
	}
	if sellCount == 0 {
		t.Errorf("expected at least one sell signal when price spikes to upper band")
	}
}

func TestBollingerReversion_NoSignalInBand(t *testing.T) {
	s := NewBollingerReversion("paper-main", "BTCUSDT", "0.01", 5, 2.0)

	prices := []float64{100, 101, 99, 100, 100, 100}
	totalSignals := 0
	for _, p := range prices {
		candle := marketdata.Candle{Symbol: "BTCUSDT", Close: formatFloat(p), High: formatFloat(p + 1), Low: formatFloat(p - 1)}
		signals, err := s.CandleEvent(context.Background(), candle)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		totalSignals += len(signals)
	}
	if totalSignals != 0 {
		t.Errorf("expected no signals when price stays within bands, got %d", totalSignals)
	}
}

func TestBollingerReversion_Name(t *testing.T) {
	s := NewBollingerReversion("paper-main", "BTCUSDT", "0.01", 20, 2.0)
	if s.Name() != "BollingerReversion" {
		t.Errorf("expected name BollingerReversion, got %s", s.Name())
	}
}
