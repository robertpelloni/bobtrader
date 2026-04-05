package risk

import (
	"context"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

func TestCooldownGuard(t *testing.T) {
	guard := NewCooldownGuard(time.Minute)
	acct := account.Account{ID: "paper-main"}
	intent := OrderIntent{Symbol: "BTCUSDT"}
	if err := guard.Check(context.Background(), acct, intent); err != nil {
		t.Fatalf("first check should pass: %v", err)
	}
	if err := guard.Check(context.Background(), acct, intent); err == nil {
		t.Fatal("second check should fail due to cooldown")
	}
}
