package risk

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

type recentSymbolChecker interface {
	HasRecentSymbol(symbol string, within time.Duration) bool
}

type DuplicateSymbolGuard struct {
	Repository recentSymbolChecker
	Within     time.Duration
}

func NewDuplicateSymbolGuard(repo recentSymbolChecker, within time.Duration) DuplicateSymbolGuard {
	return DuplicateSymbolGuard{Repository: repo, Within: within}
}

func (g DuplicateSymbolGuard) Name() string { return "duplicate-symbol" }

func (g DuplicateSymbolGuard) Check(_ context.Context, _ account.Account, intent OrderIntent) error {
	if g.Repository == nil || g.Within <= 0 {
		return nil
	}
	symbol := strings.ToUpper(strings.TrimSpace(intent.Symbol))
	if g.Repository.HasRecentSymbol(symbol, g.Within) {
		return fmt.Errorf("recent order already exists for %s", symbol)
	}
	return nil
}
