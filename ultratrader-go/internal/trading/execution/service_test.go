package execution

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/eventlog"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange/paper"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/persistence/orders"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/risk"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

func TestExecutePlacesOrder(t *testing.T) {
	accounts, err := account.NewService([]config.AccountConfig{{
		ID:           "paper-main",
		Name:         "Paper Main",
		Enabled:      true,
		Exchange:     "paper",
		Capabilities: []string{"spot", "paper", "orders"},
	}})
	if err != nil {
		t.Fatalf("NewService returned error: %v", err)
	}

	registry := exchange.NewRegistry()
	if err := registry.Register("paper", func() exchange.Adapter { return paper.New() }); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	eventPath := filepath.Join(t.TempDir(), "events.jsonl")
	events, err := eventlog.New(eventPath)
	if err != nil {
		t.Fatalf("eventlog.New returned error: %v", err)
	}
	orderPath := filepath.Join(t.TempDir(), "orders.jsonl")
	orderStore, err := orders.NewStore(orderPath)
	if err != nil {
		t.Fatalf("orders.NewStore returned error: %v", err)
	}

	service := NewService(accounts, registry, risk.NewPipeline(), events, orderStore)
	order, err := service.Execute(context.Background(), "paper-main", exchange.OrderRequest{
		Symbol:   "BTCUSDT",
		Side:     exchange.Buy,
		Type:     exchange.MarketOrder,
		Quantity: "0.01",
	}, risk.OrderIntent{AccountID: "paper-main", Symbol: "BTCUSDT", Notional: 100})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if order.ID == "" {
		t.Fatal("expected order id")
	}

	data, err := os.ReadFile(orderPath)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if !strings.Contains(string(data), order.ID) {
		t.Fatalf("expected order id in order journal, got %q", string(data))
	}
}
