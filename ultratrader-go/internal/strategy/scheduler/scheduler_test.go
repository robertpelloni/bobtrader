package scheduler

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/eventlog"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	exchangepaper "github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange/paper"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/persistence/orders"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/risk"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/execution"
)

type oneShotStrategy struct{}

func (oneShotStrategy) Name() string { return "one-shot" }
func (oneShotStrategy) OnTick(_ context.Context) ([]strategy.Signal, error) {
	return []strategy.Signal{{AccountID: "paper-main", Symbol: "BTCUSDT", Action: "buy", Quantity: "0.01", OrderType: "market"}}, nil
}

func TestRunOnceExecutesSignals(t *testing.T) {
	accounts, err := account.NewService([]config.AccountConfig{{ID: "paper-main", Name: "Paper Main", Enabled: true, Exchange: "paper", Capabilities: []string{"spot", "paper", "orders"}}})
	if err != nil {
		t.Fatalf("account.NewService error: %v", err)
	}
	registry := exchange.NewRegistry()
	if err := registry.Register("paper", func() exchange.Adapter { return exchangepaper.New() }); err != nil {
		t.Fatalf("registry.Register error: %v", err)
	}
	events, _ := eventlog.New(filepath.Join(t.TempDir(), "events.jsonl"))
	orderStore, _ := orders.NewStore(filepath.Join(t.TempDir(), "orders.jsonl"))
	executor := execution.NewService(accounts, registry, risk.NewPipeline(), events, orderStore)
	runtime := strategy.NewRuntime(oneShotStrategy{})
	scheduler := New(runtime, executor)

	if err := scheduler.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
}
