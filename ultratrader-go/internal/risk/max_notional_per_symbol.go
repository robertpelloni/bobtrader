package risk

import (
	"context"
	"fmt"
	"strings"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

type symbolValueReader interface {
	CurrentValue(symbol string) float64
}

type MaxNotionalPerSymbolGuard struct {
	Limit     float64
	Portfolio symbolValueReader
}

func NewMaxNotionalPerSymbolGuard(limit float64, portfolio symbolValueReader) MaxNotionalPerSymbolGuard {
	return MaxNotionalPerSymbolGuard{Limit: limit, Portfolio: portfolio}
}

func (g MaxNotionalPerSymbolGuard) Name() string { return "max-notional-per-symbol" }

func (g MaxNotionalPerSymbolGuard) Check(_ context.Context, _ account.Account, intent OrderIntent) error {
	if g.Limit <= 0 || g.Portfolio == nil {
		return nil
	}
	symbol := strings.ToUpper(strings.TrimSpace(intent.Symbol))
	next := g.Portfolio.CurrentValue(symbol) + intent.Notional
	if next > g.Limit {
		return fmt.Errorf("symbol %s projected notional %.2f exceeds limit %.2f", symbol, next, g.Limit)
	}
	return nil
}
