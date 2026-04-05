package risk

import (
	"context"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

func TestMaxNotionalGuard(t *testing.T) {
	guard := NewMaxNotionalGuard(100)
	if err := guard.Check(context.Background(), account.Account{}, OrderIntent{Notional: 50}); err != nil {
		t.Fatalf("expected pass, got error: %v", err)
	}
	if err := guard.Check(context.Background(), account.Account{}, OrderIntent{Notional: 150}); err == nil {
		t.Fatal("expected failure when notional exceeds limit")
	}
}
