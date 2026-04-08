package demo

import (
	"context"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

func TestCandleSMACross(t *testing.T) {
	strat := NewCandleSMACross("test", "BTCUSDT", "1.0", 2, 4)

	prices := []string{"10", "10", "10", "10", "12", "14", "10", "8"}

	var allSignals []string
	now := time.Now()

	for i, p := range prices {
		sigs, err := strat.OnMarketCandle(context.Background(), marketdata.Candle{
			Symbol:    "BTCUSDT",
			Close:     p,
			Timestamp: now.Add(time.Duration(i) * time.Hour),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, s := range sigs {
			allSignals = append(allSignals, s.Action)
		}
	}

	// We expect at least a buy (golden cross) on the way up, and a sell (death cross) on the way down.
	if len(allSignals) < 2 {
		t.Fatalf("expected at least 2 signals, got %v", allSignals)
	}

	if allSignals[0] != "buy" {
		t.Fatalf("expected first signal to be buy, got %s", allSignals[0])
	}
	if allSignals[len(allSignals)-1] != "sell" {
		t.Fatalf("expected last signal to be sell, got %s", allSignals[len(allSignals)-1])
	}
}
