package risk

import (
	"context"
	"fmt"
	"strings"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

type openPositionReader interface {
	HasOpenPosition(symbol string) bool
	OpenPositionCount() int
}

type MaxOpenPositionsGuard struct {
	Limit     int
	Portfolio openPositionReader
}

func NewMaxOpenPositionsGuard(limit int, portfolio openPositionReader) MaxOpenPositionsGuard {
	return MaxOpenPositionsGuard{Limit: limit, Portfolio: portfolio}
}

func (g MaxOpenPositionsGuard) Name() string { return "max-open-positions" }

func (g MaxOpenPositionsGuard) Check(_ context.Context, _ account.Account, intent OrderIntent) error {
	if g.Limit <= 0 || g.Portfolio == nil {
		return nil
	}
	symbol := strings.ToUpper(strings.TrimSpace(intent.Symbol))
	if g.Portfolio.HasOpenPosition(symbol) {
		return nil
	}
	if g.Portfolio.OpenPositionCount() >= g.Limit {
		return fmt.Errorf("open position limit %d reached", g.Limit)
	}
	return nil
}
