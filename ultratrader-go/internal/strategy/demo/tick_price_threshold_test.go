package demo

import (
	"context"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

func TestTickPriceThresholdEmitsOnMatchingTick(t *testing.T) {
	strategy := NewTickPriceThreshold("paper-main", "BTCUSDT", "0.01", "70000.00")
	signals, err := strategy.OnMarketTick(context.Background(), marketdata.Tick{Symbol: "BTCUSDT", Price: "65000.00", Timestamp: time.Now()})
	if err != nil {
		t.Fatalf("OnMarketTick returned error: %v", err)
	}
	if len(signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(signals))
	}
}
