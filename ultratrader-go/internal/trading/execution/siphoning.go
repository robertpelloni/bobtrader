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
// and siphons a portion into a diversified basket of long-term (macro) trend positions.
type SiphoningManager struct {
	mu               sync.Mutex
	portfolioTracker *portfolio.Tracker
	feed             marketdata.Feed
	service          *Service
	siphonPct        float64            // Percentage of profit to siphon (0.1 = 10%)
	macroWeights     map[string]float64 // Symbol -> Weight (weights should sum to 1.0)
	accountID        string
	totalSiphoned    float64 // Total value siphoned in quote currency
}

func NewSiphoningManager(
	portfolio *portfolio.Tracker,
	feed marketdata.Feed,
	service *Service,
	accountID string,
	weights map[string]float64,
	siphonPct float64,
) *SiphoningManager {
	if weights == nil || len(weights) == 0 {
		weights = map[string]float64{"BTCUSDT": 1.0}
	}
	return &SiphoningManager{
		portfolioTracker: portfolio,
		feed:             feed,
		service:          service,
		accountID:        accountID,
		macroWeights:     weights,
		siphonPct:        siphonPct,
	}
}

// OnTradeExit is called when a position is closed.
// It calculates the realized PnL and triggers macro buys according to weights if profit is positive.
func (m *SiphoningManager) OnTradeExit(ctx context.Context, symbol string, pnl float64) error {
	if pnl <= 0 {
		return nil
	}

	siphonAmount := pnl * m.siphonPct
	if siphonAmount <= 0.1 { // Dust threshold
		return nil
	}

	m.mu.Lock()
	m.totalSiphoned += siphonAmount
	m.mu.Unlock()

	for macroSymbol, weight := range m.macroWeights {
		assetAmount := siphonAmount * weight
		if assetAmount <= 0.05 {
			continue
		}

		tick, err := m.feed.LatestTick(ctx, macroSymbol)
		if err != nil {
			continue // Skip this asset if price unavailable
		}

		price := utils.ParseFloat(tick.Price)
		if price <= 0 {
			continue
		}

		quantity := assetAmount / price

		req := exchange.OrderRequest{
			Symbol:   macroSymbol,
			Side:     exchange.Buy,
			Type:     exchange.MarketOrder,
			Quantity: fmt.Sprintf("%.8f", quantity),
		}

		intent := risk.OrderIntent{
			AccountID: m.accountID,
			Symbol:    macroSymbol,
			Side:      risk.BuySide,
			Notional:  assetAmount,
			IsExit:    false,
			Metadata:  map[string]any{"source": "siphoning", "origin": symbol},
		}

		_, _ = m.service.Execute(ctx, m.accountID, req, intent)
	}

	return nil
}

// Stats returns the total amount siphoned.
func (m *SiphoningManager) Stats() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.totalSiphoned
}
