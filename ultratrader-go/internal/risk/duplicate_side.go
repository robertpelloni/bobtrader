package risk

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

type recentSymbolSideChecker interface {
	HasRecentSymbolSide(symbol string, side exchange.OrderSide, within time.Duration) bool
}

type DuplicateSideGuard struct {
	Repository recentSymbolSideChecker
	Side       exchange.OrderSide
	Within     time.Duration
}

func NewDuplicateSideGuard(repo recentSymbolSideChecker, side exchange.OrderSide, within time.Duration) DuplicateSideGuard {
	return DuplicateSideGuard{Repository: repo, Side: side, Within: within}
}

func (g DuplicateSideGuard) Name() string { return "duplicate-side" }

func (g DuplicateSideGuard) Check(_ context.Context, _ account.Account, intent OrderIntent) error {
	if g.Repository == nil || g.Within <= 0 {
		return nil
	}
	symbol := strings.ToUpper(strings.TrimSpace(intent.Symbol))
	if g.Repository.HasRecentSymbolSide(symbol, g.Side, g.Within) {
		return fmt.Errorf("recent %s order already exists for %s", g.Side, symbol)
	}
	return nil
}
