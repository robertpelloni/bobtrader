package execution

import (
	"context"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

// mockExecutionAdapter tracks orders without balance logic for simple flow verification.
type mockExecutionAdapter struct {
	lastOrder exchange.Order
}

func (m *mockExecutionAdapter) Name() string                                               { return "mock" }
func (m *mockExecutionAdapter) Capabilities() []exchange.Capability                        { return nil }
func (m *mockExecutionAdapter) ListMarkets(ctx context.Context) ([]exchange.Market, error) { return nil, nil }
func (m *mockExecutionAdapter) Balances(ctx context.Context) ([]exchange.Balance, error)   { return nil, nil }
func (m *mockExecutionAdapter) PlaceOrder(ctx context.Context, r exchange.OrderRequest) (exchange.Order, error) {
	m.lastOrder = exchange.Order{
		ID:       "mock-1",
		Symbol:   r.Symbol,
		Side:     r.Side,
		Quantity: r.Quantity,
		Status:   exchange.StatusClosed,
	}
	return m.lastOrder, nil
}

func TestTradeExecutionFlowIntegration(t *testing.T) {
	ctx := context.Background()
	mock := &mockExecutionAdapter{}

	// Wrap with live safety
	marketStrat := NewMarketStrategy(mock)
	liveWrapper := NewLiveStrategyWrapper(marketStrat, 1.0) // 1% slippage limit

	t.Run("End_to_End_Execution", func(t *testing.T) {
		order := exchange.Order{
			Symbol:   "BTCUSDT",
			Side:     exchange.Buy,
			Quantity: "0.01",
		}

		err := liveWrapper.Execute(ctx, order)
		if err != nil {
			t.Fatalf("Live wrapper failed to execute: %v", err)
		}

		if mock.lastOrder.Symbol != "BTCUSDT" {
			t.Errorf("Expected BTCUSDT, got %s", mock.lastOrder.Symbol)
		}
		if mock.lastOrder.Quantity != "0.01" {
			t.Errorf("Expected 0.01, got %s", mock.lastOrder.Quantity)
		}
	})
}
