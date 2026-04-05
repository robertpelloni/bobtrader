package risk

import (
	"context"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

type stubConcentration struct {
	current float64
	total   float64
}

func (s stubConcentration) CurrentValue(symbol string) float64 { return s.current }
func (s stubConcentration) TotalValue() float64                { return s.total }

func TestMaxConcentrationGuard(t *testing.T) {
	guard := NewMaxConcentrationGuard(25, stubConcentration{current: 10, total: 100})
	if err := guard.Check(context.Background(), account.Account{}, OrderIntent{Symbol: "BTCUSDT", Notional: 10}); err != nil {
		t.Fatalf("expected pass, got %v", err)
	}
	if err := guard.Check(context.Background(), account.Account{}, OrderIntent{Symbol: "BTCUSDT", Notional: 20}); err == nil {
		t.Fatal("expected concentration guard to fail")
	}
}
