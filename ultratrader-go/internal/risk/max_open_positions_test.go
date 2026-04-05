package risk

import (
	"context"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

type stubPortfolio struct {
	hasPosition bool
	count       int
}

func (s stubPortfolio) HasOpenPosition(symbol string) bool { return s.hasPosition }
func (s stubPortfolio) OpenPositionCount() int             { return s.count }

func TestMaxOpenPositionsGuard(t *testing.T) {
	guard := NewMaxOpenPositionsGuard(1, stubPortfolio{hasPosition: false, count: 1})
	if err := guard.Check(context.Background(), account.Account{}, OrderIntent{Symbol: "ETHUSDT"}); err == nil {
		t.Fatal("expected open position guard to fail")
	}

	guard = NewMaxOpenPositionsGuard(1, stubPortfolio{hasPosition: true, count: 1})
	if err := guard.Check(context.Background(), account.Account{}, OrderIntent{Symbol: "BTCUSDT"}); err != nil {
		t.Fatalf("expected existing-position symbol to pass: %v", err)
	}
}
