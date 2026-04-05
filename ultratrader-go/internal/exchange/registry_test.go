package exchange

import (
	"context"
	"testing"
)

type testAdapter struct{}

func (testAdapter) Name() string                                    { return "test" }
func (testAdapter) Capabilities() []Capability                      { return []Capability{CapabilitySpot} }
func (testAdapter) ListMarkets(_ context.Context) ([]Market, error) { return nil, nil }
func (testAdapter) Balances(_ context.Context) ([]Balance, error)   { return nil, nil }
func (testAdapter) PlaceOrder(_ context.Context, _ OrderRequest) (Order, error) {
	return Order{ID: "1"}, nil
}

func TestRegistryRegisterAndCreate(t *testing.T) {
	registry := NewRegistry()
	if err := registry.Register("test", func() Adapter { return testAdapter{} }); err != nil {
		t.Fatalf("Register returned error: %v", err)
	}

	adapter, err := registry.Create("test")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if adapter.Name() != "test" {
		t.Fatalf("unexpected adapter name: %q", adapter.Name())
	}
}
