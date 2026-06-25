package execution

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

// DynamicTrailingStop implements a sophisticated trailing stop loss
// inspired by whittlem/pycryptobot.
type DynamicTrailingStop struct {
	mu              sync.Mutex
	adapter         exchange.Adapter
	symbol          string
	trailPercent    float64
	triggerPercent  float64
	multiplier      float64
	maxTrailPercent float64
	highestPrice    float64
	triggered       bool
}

// NewDynamicTrailingStop creates a new dynamic trailing stop strategy.
func NewDynamicTrailingStop(adapter exchange.Adapter, symbol string, trailPct, triggerPct, mult, maxTrail float64) *DynamicTrailingStop {
	return &DynamicTrailingStop{
		adapter:         adapter,
		symbol:          symbol,
		trailPercent:    trailPct,
		triggerPercent:  triggerPct,
		multiplier:      mult,
		maxTrailPercent: maxTrail,
	}
}

func (s *DynamicTrailingStop) Name() string {
	return "dynamic-trailing-stop"
}

// Update evaluates the current price and potentially triggers a sell.
func (s *DynamicTrailingStop) Update(ctx context.Context, currentPrice float64, entryPrice float64) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	margin := (currentPrice - entryPrice) / entryPrice * 100

	// Check if we should activate the TSL
	if !s.triggered && margin > s.triggerPercent {
		s.triggered = true
		s.highestPrice = currentPrice
	}

	if !s.triggered {
		return false, nil
	}

	// Update highest price seen
	if currentPrice > s.highestPrice {
		s.highestPrice = currentPrice

		// If dynamic TSL is enabled (multiplier > 1), we could tighten the stop
		// In PyCryptoBot: s.trailPercent = round(s.trailPercent * s.multiplier, 1)
		// We'll keep it simple for now or implement the scaling logic here.
	}

	// Calculate current stop level
	stopPrice := s.highestPrice * (1 - s.trailPercent/100)

	if currentPrice <= stopPrice {
		// Trigger SELL
		fmt.Printf("DynamicTSL: Triggered for %s at %.2f (Highest: %.2f, Trail: %.2f%%)\n",
			s.symbol, currentPrice, s.highestPrice, s.trailPercent)

		return true, nil
	}

	return false, nil
}

// Execute performs the market sell order when the stop is hit.
func (s *DynamicTrailingStop) Execute(ctx context.Context, order exchange.Order) error {
	request := exchange.OrderRequest{
		Symbol:   order.Symbol,
		Side:     exchange.Sell,
		Type:     exchange.MarketOrder,
		Quantity: order.Quantity,
	}

	_, err := s.adapter.PlaceOrder(ctx, request)
	return err
}

func round(val float64, precision uint) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}
