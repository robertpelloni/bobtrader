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
	guard := NewDuplicateSideGuard(stubRecentSideChecker{hit: true}, time.Minute)

	// Buy side check
	if err := guard.Check(context.Background(), account.Account{}, OrderIntent{Symbol: "BTCUSDT", Side: BuySide}); err == nil {
		t.Fatal("expected duplicate-side guard to fail for buy")
	}

	// Sell side check
	if err := guard.Check(context.Background(), account.Account{}, OrderIntent{Symbol: "BTCUSDT", Side: SellSide}); err == nil {
		t.Fatal("expected duplicate-side guard to fail for sell")
	}

	// No hit should pass
	guardOK := NewDuplicateSideGuard(stubRecentSideChecker{hit: false}, time.Minute)
	if err := guardOK.Check(context.Background(), account.Account{}, OrderIntent{Symbol: "BTCUSDT", Side: BuySide}); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
