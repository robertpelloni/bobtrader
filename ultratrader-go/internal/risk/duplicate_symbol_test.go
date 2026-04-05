package risk

import (
	"context"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

type stubRecentChecker struct{ hit bool }

func (s stubRecentChecker) HasRecentSymbol(symbol string, within time.Duration) bool {
	return s.hit
}

func TestDuplicateSymbolGuard(t *testing.T) {
	guard := NewDuplicateSymbolGuard(stubRecentChecker{hit: true}, time.Minute)
	if err := guard.Check(context.Background(), account.Account{}, OrderIntent{Symbol: "BTCUSDT"}); err == nil {
		t.Fatal("expected duplicate symbol guard to fail")
	}
}
