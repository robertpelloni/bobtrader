package risk

import (
	"context"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

func TestSymbolWhitelistGuard(t *testing.T) {
	guard := NewSymbolWhitelistGuard([]string{"BTCUSDT", "ETHUSDT"})
	if err := guard.Check(context.Background(), account.Account{}, OrderIntent{Symbol: "BTCUSDT"}); err != nil {
		t.Fatalf("expected pass, got error: %v", err)
	}
	if err := guard.Check(context.Background(), account.Account{}, OrderIntent{Symbol: "DOGEUSDT"}); err == nil {
		t.Fatal("expected failure for non-whitelisted symbol")
	}
}
