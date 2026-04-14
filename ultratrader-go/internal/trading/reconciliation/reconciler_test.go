package reconciliation

import (
	"context"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

// mockQuerier implements both exchange.Adapter and OrderQuerier for testing.
type mockQuerier struct {
	orders map[string]OrderStatus
}

func (m *mockQuerier) Name() string                                               { return "mock" }
func (m *mockQuerier) Capabilities() []exchange.Capability                        { return nil }
func (m *mockQuerier) ListMarkets(ctx context.Context) ([]exchange.Market, error) { return nil, nil }
func (m *mockQuerier) Balances(ctx context.Context) ([]exchange.Balance, error)   { return nil, nil }
func (m *mockQuerier) PlaceOrder(ctx context.Context, r exchange.OrderRequest) (exchange.Order, error) {
	return exchange.Order{}, nil
}

func (m *mockQuerier) QueryOrder(ctx context.Context, symbol, orderID string) (OrderStatus, error) {
	if status, ok := m.orders[orderID]; ok {
		return status, nil
	}
	return OrderStatus{}, nil
}

func TestReconcileOrders_Matched(t *testing.T) {
	querier := &mockQuerier{
		orders: map[string]OrderStatus{
			"123": {ID: "123", Symbol: "BTCUSDT", Status: "FILLED"},
		},
	}

	reconciler := NewReconciler(querier)
	localOrders := []exchange.Order{
		{ID: "123", Symbol: "BTCUSDT", Status: "filled"},
	}

	result, err := reconciler.ReconcileOrders(context.Background(), localOrders)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Matched != 1 {
		t.Errorf("expected 1 matched, got %d", result.Matched)
	}
	if result.Filled != 1 {
		t.Errorf("expected 1 filled, got %d", result.Filled)
	}
	if len(result.Discrepancies) != 0 {
		t.Errorf("expected 0 discrepancies, got %d", len(result.Discrepancies))
	}
}

func TestReconcileOrders_Discrepancy(t *testing.T) {
	querier := &mockQuerier{
		orders: map[string]OrderStatus{
			"456": {ID: "456", Symbol: "BTCUSDT", Status: "CANCELED"},
		},
	}

	reconciler := NewReconciler(querier)
	localOrders := []exchange.Order{
		{ID: "456", Symbol: "BTCUSDT", Status: "filled"},
	}

	result, err := reconciler.ReconcileOrders(context.Background(), localOrders)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Discrepancies) != 1 {
		t.Fatalf("expected 1 discrepancy, got %d", len(result.Discrepancies))
	}
	if result.Discrepancies[0].InternalStatus != "filled" {
		t.Errorf("expected internal status filled, got %s", result.Discrepancies[0].InternalStatus)
	}
	if result.Discrepancies[0].ExchangeStatus != "CANCELED" {
		t.Errorf("expected exchange status CANCELED, got %s", result.Discrepancies[0].ExchangeStatus)
	}
}

// stubAdapter implements exchange.Adapter but NOT OrderQuerier.
type stubAdapter struct{}

func (s *stubAdapter) Name() string                                               { return "stub" }
func (s *stubAdapter) Capabilities() []exchange.Capability                        { return nil }
func (s *stubAdapter) ListMarkets(ctx context.Context) ([]exchange.Market, error) { return nil, nil }
func (s *stubAdapter) Balances(ctx context.Context) ([]exchange.Balance, error)   { return nil, nil }
func (s *stubAdapter) PlaceOrder(ctx context.Context, r exchange.OrderRequest) (exchange.Order, error) {
	return exchange.Order{}, nil
}

func TestReconcileOrders_NoQuerier(t *testing.T) {
	reconciler := NewReconciler(&stubAdapter{})

	localOrders := []exchange.Order{
		{ID: "1", Symbol: "BTCUSDT", Status: "filled"},
		{ID: "2", Symbol: "ETHUSDT", Status: "canceled"},
	}

	result, err := reconciler.ReconcileOrders(context.Background(), localOrders)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Filled != 1 {
		t.Errorf("expected 1 filled, got %d", result.Filled)
	}
	if result.Canceled != 1 {
		t.Errorf("expected 1 canceled, got %d", result.Canceled)
	}
}

func TestReconcileResult_Summary(t *testing.T) {
	result := &ReconcileResult{
		TotalChecked:  10,
		Matched:       8,
		Filled:        7,
		Canceled:      1,
		Discrepancies: []Discrepancy{{OrderID: "99"}},
	}
	summary := result.Summary()
	if summary == "" {
		t.Errorf("expected non-empty summary")
	}
	t.Logf("Summary: %s", summary)
}
