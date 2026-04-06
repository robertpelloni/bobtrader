package risk

import (
	"context"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

type stubSymbolValue struct{ current float64 }

func (s stubSymbolValue) CurrentValue(symbol string) float64 { return s.current }

func TestMaxNotionalPerSymbolGuard(t *testing.T) {
	guard := NewMaxNotionalPerSymbolGuard(100, stubSymbolValue{current: 40})
	if err := guard.Check(context.Background(), account.Account{}, OrderIntent{Symbol: "BTCUSDT", Notional: 30}); err != nil {
		t.Fatalf("expected pass, got %v", err)
	}
	if err := guard.Check(context.Background(), account.Account{}, OrderIntent{Symbol: "BTCUSDT", Notional: 80}); err == nil {
		t.Fatal("expected max-notional-per-symbol guard to fail")
	}
}
