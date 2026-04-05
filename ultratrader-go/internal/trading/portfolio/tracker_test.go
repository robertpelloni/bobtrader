package portfolio

import (
	"context"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	marketdatapaper "github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata/paper"
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

func TestTotalMarketValue(t *testing.T) {
	tracker := NewTracker()
	tracker.Apply(exchange.Order{Symbol: "BTCUSDT", Side: exchange.Buy, Quantity: "0.50"})
	total := tracker.TotalMarketValue(context.Background(), marketdatapaper.New())
	if total != 32500 {
		t.Fatalf("expected total market value 32500, got %v", total)
	}
}
