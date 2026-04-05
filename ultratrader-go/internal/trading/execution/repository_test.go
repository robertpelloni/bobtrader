package execution

import (
	"testing"
	"time"

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

func TestRepositorySummaryAndRecentSymbol(t *testing.T) {
	repo := NewRepository()
	repo.Save(exchange.Order{ID: "1", Symbol: "BTCUSDT"})
	repo.Save(exchange.Order{ID: "2", Symbol: "BTCUSDT"})
	summary := repo.Summary()
	if summary.TotalOrders != 2 || summary.OrdersBySymbol["BTCUSDT"] != 2 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if !repo.HasRecentSymbol("BTCUSDT", time.Minute) {
		t.Fatal("expected recent BTCUSDT order")
	}
}
