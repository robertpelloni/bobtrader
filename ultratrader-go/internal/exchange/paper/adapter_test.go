package paper

import (
	"context"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

func TestPaperAdapterPlacesOrder(t *testing.T) {
	adapter := New()
	order, err := adapter.PlaceOrder(context.Background(), exchange.OrderRequest{
		Symbol:   "BTCUSDT",
		Side:     exchange.Buy,
		Type:     exchange.MarketOrder,
		Quantity: "0.01",
	})
	if err != nil {
		t.Fatalf("PlaceOrder returned error: %v", err)
	}
	if order.ID == "" {
		t.Fatal("expected generated order id")
	}
	if order.Status != "filled" {
		t.Fatalf("expected filled status, got %q", order.Status)
	}
}
