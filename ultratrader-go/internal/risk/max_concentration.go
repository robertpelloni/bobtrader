package risk

import (
	"context"
	"fmt"
	"strings"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

type concentrationReader interface {
	CurrentValue(symbol string) float64
	TotalValue() float64
}

type MaxConcentrationGuard struct {
	MaxPct    float64
	Portfolio concentrationReader
}

func NewMaxConcentrationGuard(maxPct float64, portfolio concentrationReader) MaxConcentrationGuard {
	return MaxConcentrationGuard{MaxPct: maxPct, Portfolio: portfolio}
}

func (g MaxConcentrationGuard) Name() string { return "max-concentration" }

func (g MaxConcentrationGuard) Check(_ context.Context, _ account.Account, intent OrderIntent) error {
	if g.MaxPct <= 0 || g.Portfolio == nil {
		return nil
	}
	total := g.Portfolio.TotalValue()
	if total <= 0 {
		return nil
	}
	symbol := strings.ToUpper(strings.TrimSpace(intent.Symbol))
	next := g.Portfolio.CurrentValue(symbol) + intent.Notional
	pct := (next / total) * 100
	if pct > g.MaxPct {
		return fmt.Errorf("concentration %.2f%% exceeds max %.2f%% for %s", pct, g.MaxPct, symbol)
	}
	return nil
}
