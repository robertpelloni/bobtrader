package risk

import (
	"context"
	"fmt"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/portfolio"
)

// DrawdownGuard monitors the portfolio's peak-to-trough decline and blocks
// new trades if the decline exceeds a configured threshold.
type DrawdownGuard struct {
	mu           sync.RWMutex
	tracker      *portfolio.Tracker
	feed         marketdata.Feed
	maxDrawdown  float64 // Percentage (e.g., 0.10 = 10% maximum decline)
	peakValue    float64
	currentDD    float64
}

func NewDrawdownGuard(tracker *portfolio.Tracker, feed marketdata.Feed, maxDrawdown float64) *DrawdownGuard {
	return &DrawdownGuard{
		tracker:     tracker,
		feed:        feed,
		maxDrawdown: maxDrawdown,
	}
}

func (g *DrawdownGuard) Name() string { return "drawdown-limit" }

func (g *DrawdownGuard) Check(ctx context.Context, _ account.Account, intent OrderIntent) error {
	// Exits are usually allowed to reduce exposure
	if intent.IsExit {
		return nil
	}

	g.mu.Lock()
	defer g.mu.Unlock()

	// Update portfolio valuation
	currentValue := g.tracker.TotalMarketValue(ctx, g.feed)
	if currentValue > g.peakValue {
		g.peakValue = currentValue
	}

	if g.peakValue > 0 {
		g.currentDD = (g.peakValue - currentValue) / g.peakValue
	}

	if g.currentDD > g.maxDrawdown {
		return fmt.Errorf("portfolio drawdown too high: %.2f%% (limit %.2f%%)", g.currentDD*100, g.maxDrawdown*100)
	}

	return nil
}

// CurrentDrawdown returns the last calculated drawdown percentage.
func (g *DrawdownGuard) CurrentDrawdown() float64 {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.currentDD
}
