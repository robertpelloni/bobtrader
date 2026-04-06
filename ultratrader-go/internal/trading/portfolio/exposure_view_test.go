package portfolio

import (
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	marketdatapaper "github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata/paper"
)

func TestExposureViewUsesMarketValue(t *testing.T) {
	tracker := NewTracker()
	tracker.Apply(exchange.Order{Symbol: "BTCUSDT", Side: exchange.Buy, Quantity: "0.5", Price: "60000"})
	view := NewExposureView(tracker, marketdatapaper.New())
	if view.CurrentValue("BTCUSDT") != 32500 {
		t.Fatalf("expected market value 32500, got %v", view.CurrentValue("BTCUSDT"))
	}
	if view.TotalValue() != 32500 {
		t.Fatalf("expected total value 32500, got %v", view.TotalValue())
	}
}
