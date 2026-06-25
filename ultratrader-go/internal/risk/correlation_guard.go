package risk

import (
	"context"
	"fmt"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/analytics/correlation"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/portfolio"
)

// CorrelationGuard prevents adding positions that are highly correlated with
// existing heavy positions, ensuring effective portfolio diversification.
type CorrelationGuard struct {
	matrix         *correlation.CorrelationMatrix
	tracker        *portfolio.Tracker
	maxCorrelation float64
}

func NewCorrelationGuard(matrix *correlation.CorrelationMatrix, tracker *portfolio.Tracker, maxCorr float64) *CorrelationGuard {
	if maxCorr <= 0 {
		maxCorr = 0.85
	}
	return &CorrelationGuard{
		matrix:         matrix,
		tracker:        tracker,
		maxCorrelation: maxCorr,
	}
}

func (g *CorrelationGuard) Name() string { return "correlation-diversifier" }

func (g *CorrelationGuard) Check(ctx context.Context, acct account.Account, intent OrderIntent) error {
	// Only apply to entry orders
	if intent.IsExit {
		return nil
	}

	correlations := g.matrix.Compute()
	positions := g.tracker.Positions()

	for _, p := range positions {
		if p.Symbol == intent.Symbol {
			continue
		}

		// Keys are alphabetical "SYM1:SYM2"
		s1, s2 := intent.Symbol, p.Symbol
		if s2 < s1 {
			s1, s2 = s2, s1
		}
		key := s1 + ":" + s2

		if corr, ok := correlations[key]; ok {
			if corr > g.maxCorrelation {
				return fmt.Errorf("symbol %s is too highly correlated (%.2f) with existing position %s (limit: %.2f)",
					intent.Symbol, corr, p.Symbol, g.maxCorrelation)
			}
		}
	}

	return nil
}
