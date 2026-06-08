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

// DuplicateSideGuard prevents rapid repeated orders of the same side
// (buy or sell) for the same symbol. It checks the intent's Side field
// against recent execution history.
type DuplicateSideGuard struct {
	Repository recentSymbolSideChecker
	Within     time.Duration
}

func NewDuplicateSideGuard(repo recentSymbolSideChecker, within time.Duration) DuplicateSideGuard {
	return DuplicateSideGuard{Repository: repo, Within: within}
}

func (g DuplicateSideGuard) Name() string {
	return "duplicate-side"
}

func (g DuplicateSideGuard) Check(_ context.Context, _ account.Account, intent OrderIntent) error {
	if g.Repository == nil || g.Within <= 0 {
		return nil
	}

	// Exit orders bypass duplicate check
	if intent.IsExit {
		return nil
	}

	symbol := strings.ToUpper(strings.TrimSpace(intent.Symbol))

	var side exchange.OrderSide
	switch intent.Side {
	case BuySide:
		side = exchange.Buy
	case SellSide:
		side = exchange.Sell
	default:
		side = exchange.Buy
	}

	if g.Repository.HasRecentSymbolSide(symbol, side, g.Within) {
		return fmt.Errorf("recent %s order already exists for %s", side, symbol)
	}
	return nil
}
