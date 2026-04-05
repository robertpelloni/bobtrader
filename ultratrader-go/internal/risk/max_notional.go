package risk

import (
	"context"
	"fmt"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

type MaxNotionalGuard struct {
	Limit float64
}

func NewMaxNotionalGuard(limit float64) MaxNotionalGuard {
	return MaxNotionalGuard{Limit: limit}
}

func (g MaxNotionalGuard) Name() string { return "max-notional" }

func (g MaxNotionalGuard) Check(_ context.Context, _ account.Account, intent OrderIntent) error {
	if g.Limit <= 0 {
		return nil
	}
	if intent.Notional > g.Limit {
		return fmt.Errorf("notional %.4f exceeds limit %.4f", intent.Notional, g.Limit)
	}
	return nil
}
