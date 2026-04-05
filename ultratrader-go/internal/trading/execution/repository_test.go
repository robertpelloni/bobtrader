package execution

import (
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

func TestRepositorySaveAndList(t *testing.T) {
	repo := NewRepository()
	repo.Save(exchange.Order{ID: "b", Symbol: "ETHUSDT"})
	repo.Save(exchange.Order{ID: "a", Symbol: "BTCUSDT"})

	orders := repo.List()
	if len(orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(orders))
	}
	if orders[0].ID != "a" {
		t.Fatalf("expected sorted order list, got first id %q", orders[0].ID)
	}
}
