package execution

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/notifications"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/risk"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/portfolio"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/rebalancer"
)

// SiphoningManager monitors realized profits from high-frequency (micro) strategies
// and siphons a portion into a diversified basket of long-term (macro) trend positions.
type SiphoningManager struct {
	mu               sync.Mutex
	portfolioTracker *portfolio.Tracker
	feed             marketdata.Feed
	service          *Service
	rebalancer       *rebalancer.Rebalancer
	notifications    *notifications.Manager
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

	allocations := make([]rebalancer.Allocation, 0, len(weights))
	for sym, w := range weights {
		allocations = append(allocations, rebalancer.Allocation{Symbol: sym, Weight: w})
	}

	return &SiphoningManager{
		portfolioTracker: portfolio,
		feed:             feed,
		service:          service,
		rebalancer:       rebalancer.New(allocations, 0.01), // tight threshold for siphoning
		accountID:        accountID,
		macroWeights:     weights,
		siphonPct:        siphonPct,
	}
}

// OnTradeExit is called when a position is closed.
// It calculates the realized PnL and triggers macro buys according to weights if profit is positive.
func (m *SiphoningManager) SetNotificationManager(n *notifications.Manager) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.notifications = n
}

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
	notifManager := m.notifications
	m.mu.Unlock()

	if notifManager != nil {
		notifManager.Notify(ctx, notifications.Notification{
			Level:   notifications.Info,
			Source:  "SiphoningManager",
			Message: fmt.Sprintf("Siphoning realized profit: $%.2f (Origin: %s)", siphonAmount, symbol),
		})
	}

	// Use rebalancer to determine where the siphon should go based on current drift.
	// This ensures we always buy the most "underweight" asset first.
	positions := m.portfolioTracker.ValuedPositions(ctx, m.feed)
	holdings := make([]rebalancer.Holding, len(positions))
	for i, p := range positions {
		holdings[i] = rebalancer.Holding{Symbol: p.Symbol, Quantity: p.Quantity, Value: p.MarketValue}
	}

	result := m.rebalancer.Compute(holdings)

	// If the portfolio is drifted, allocate siphoned amount to buy underweight assets.
	// (Simplified: we use macroWeights if rebalancer doesn't suggest a clear path)
	if !result.NeedsRebalance {
		for macroSymbol, weight := range m.macroWeights {
			m.executeSiphonBuy(ctx, macroSymbol, siphonAmount*weight, symbol)
		}
		return nil
	}

	// Filter only for buys (siphoning only adds to positions)
	var buyOrders []rebalancer.RebalanceOrder
	for _, o := range result.Orders {
		if o.Side == "buy" {
			buyOrders = append(buyOrders, o)
		}
	}

	if len(buyOrders) == 0 {
		// Fallback to equal siphoning if no clear buys needed
		for macroSymbol, weight := range m.macroWeights {
			m.executeSiphonBuy(ctx, macroSymbol, siphonAmount*weight, symbol)
		}
		return nil
	}

	// Distribute siphoned amount among underweight assets
	for _, o := range buyOrders {
		// Weight based on absolute drift magnitude
		allocation := (math.Abs(o.DriftPct) / result.MaxDrift) * siphonAmount
		m.executeSiphonBuy(ctx, o.Symbol, allocation, symbol)
	}

	return nil
}

func (m *SiphoningManager) executeSiphonBuy(ctx context.Context, macroSymbol string, amount float64, origin string) {
	if amount <= 0.05 {
		return
	}

	tick, err := m.feed.LatestTick(ctx, macroSymbol)
	if err != nil {
		return
	}

	price := utils.ParseFloat(tick.Price)
	if price <= 0 {
		return
	}

	quantity := amount / price

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
		Notional:  amount,
		IsExit:    false,
		Metadata:  map[string]any{"source": "siphoning", "origin": origin},
	}

	_, _ = m.service.Execute(ctx, m.accountID, req, intent)
}

// Stats returns the total amount siphoned.
func (m *SiphoningManager) Stats() float64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.totalSiphoned
}
