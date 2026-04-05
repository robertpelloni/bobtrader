package portfolio

import (
	"context"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	marketdatapaper "github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata/paper"
)

func TestTrackerApply(t *testing.T) {
	tracker := NewTracker()
	tracker.Apply(exchange.Order{Symbol: "BTCUSDT", Side: exchange.Buy, Quantity: "0.50", Price: "60000"})
	tracker.Apply(exchange.Order{Symbol: "BTCUSDT", Side: exchange.Sell, Quantity: "0.10", Price: "65000"})
	positions := tracker.Positions()
	if len(positions) != 1 {
		t.Fatalf("expected 1 position, got %d", len(positions))
	}
	if positions[0].Quantity != 0.4 {
		t.Fatalf("expected quantity 0.4, got %v", positions[0].Quantity)
	}
	if positions[0].RealizedPnL != 500 {
		t.Fatalf("expected realized pnl 500, got %v", positions[0].RealizedPnL)
	}
}

func TestValuationAndPnL(t *testing.T) {
	tracker := NewTracker()
	tracker.Apply(exchange.Order{Symbol: "BTCUSDT", Side: exchange.Buy, Quantity: "0.50", Price: "60000"})
	total := tracker.TotalMarketValue(context.Background(), marketdatapaper.New())
	if total != 32500 {
		t.Fatalf("expected total market value 32500, got %v", total)
	}
	unrealized := tracker.TotalUnrealizedPnL(context.Background(), marketdatapaper.New())
	if unrealized != 2500 {
		t.Fatalf("expected unrealized pnl 2500, got %v", unrealized)
	}
	if !tracker.HasOpenPosition("BTCUSDT") {
		t.Fatal("expected open BTCUSDT position")
	}
	if tracker.OpenPositionCount() != 1 {
		t.Fatalf("expected one open position, got %d", tracker.OpenPositionCount())
	}
	if tracker.CurrentValue("BTCUSDT") != 30000 {
		t.Fatalf("expected cost-basis current value 30000, got %v", tracker.CurrentValue("BTCUSDT"))
	}
	if tracker.TotalValue() != 30000 {
		t.Fatalf("expected total value 30000, got %v", tracker.TotalValue())
	}
}
