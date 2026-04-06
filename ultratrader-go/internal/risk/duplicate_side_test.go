package risk

import (
	"context"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

type stubRecentSideChecker struct{ hit bool }

func (s stubRecentSideChecker) HasRecentSymbolSide(symbol string, side exchange.OrderSide, within time.Duration) bool {
	return s.hit
}

func TestDuplicateSideGuard(t *testing.T) {
	guard := NewDuplicateSideGuard(stubRecentSideChecker{hit: true}, exchange.Buy, time.Minute)
	if err := guard.Check(context.Background(), account.Account{}, OrderIntent{Symbol: "BTCUSDT"}); err == nil {
		t.Fatal("expected duplicate-side guard to fail")
	}
}
