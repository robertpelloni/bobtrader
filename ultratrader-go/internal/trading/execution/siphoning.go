package execution

import (
	"context"
	"fmt"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/risk"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/portfolio"
)

// SiphoningManager monitors realized profits from high-frequency (micro) strategies
// and siphons a portion into long-term (macro) trend positions.
type SiphoningManager struct {
	mu               sync.Mutex
	portfolioTracker *portfolio.Tracker
	feed             marketdata.Feed
	service          *Service
	siphonPct        float64 // Percentage of profit to siphon (0.1 = 10%)
	macroSymbol      string  // Target symbol for macro positions (e.g., BTCUSDT)
	accountID        string
}

func NewSiphoningManager(
	portfolio *portfolio.Tracker,
	feed marketdata.Feed,
	service *Service,
	accountID string,
	macroSymbol string,
	siphonPct float64,
) *SiphoningManager {
	return &SiphoningManager{
		portfolioTracker: portfolio,
		feed:             feed,
		service:          service,
		accountID:        accountID,
		macroSymbol:      macroSymbol,
		siphonPct:        siphonPct,
	}
}

// OnTradeExit is called when a position is closed.
// It calculates the realized PnL and triggers a macro buy if profit is positive.
func (m *SiphoningManager) OnTradeExit(ctx context.Context, symbol string, pnl float64) error {
	if pnl <= 0 {
		return nil
	}

	siphonAmount := pnl * m.siphonPct
	if siphonAmount <= 0.1 { // Dust threshold
		return nil
	}

	tick, err := m.feed.LatestTick(ctx, m.macroSymbol)
	if err != nil {
		return fmt.Errorf("get macro price: %w", err)
	}

	price := utils.ParseFloat(tick.Price)
	if price <= 0 {
		return fmt.Errorf("invalid macro price: %s", tick.Price)
	}

	// Calculate quantity based on siphoned amount and current price
	quantity := siphonAmount / price

	// Place a macro buy order using siphoned funds
	req := exchange.OrderRequest{
		Symbol:   m.macroSymbol,
		Side:     exchange.Buy,
		Type:     exchange.MarketOrder,
		Quantity: fmt.Sprintf("%.8f", quantity),
	}

	intent := risk.OrderIntent{
		AccountID: m.accountID,
		Symbol:    m.macroSymbol,
		Side:      risk.BuySide,
		Notional:  siphonAmount,
		IsExit:    false,
		Metadata:  map[string]any{"source": "siphoning"},
	}

	_, err = m.service.Execute(ctx, m.accountID, req, intent)
	return err
}
