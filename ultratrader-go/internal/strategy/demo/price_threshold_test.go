package demo

import (
	"context"
	"testing"

	marketdatapaper "github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata/paper"
)

func TestPriceThresholdEmitsWhenPriceBelowThreshold(t *testing.T) {
	strategy := NewPriceThreshold("paper-main", "BTCUSDT", "0.01", "70000.00", marketdatapaper.New())
	signals, err := strategy.OnTick(context.Background())
	if err != nil {
		t.Fatalf("OnTick returned error: %v", err)
	}
	if len(signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(signals))
	}
}
