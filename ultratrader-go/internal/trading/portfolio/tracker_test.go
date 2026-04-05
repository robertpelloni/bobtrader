package portfolio

import (
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

func TestTrackerApply(t *testing.T) {
	tracker := NewTracker()
	tracker.Apply(exchange.Order{Symbol: "BTCUSDT", Side: exchange.Buy, Quantity: "0.50"})
	tracker.Apply(exchange.Order{Symbol: "BTCUSDT", Side: exchange.Sell, Quantity: "0.10"})

	positions := tracker.Positions()
	if len(positions) != 1 {
		t.Fatalf("expected 1 position, got %d", len(positions))
	}
	if positions[0].Quantity != 0.4 {
		t.Fatalf("expected quantity 0.4, got %v", positions[0].Quantity)
	}
}
