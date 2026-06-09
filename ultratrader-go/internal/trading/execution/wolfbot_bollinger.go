package execution

import (
	"context"
	"fmt"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

// WolfBotBollingerStrategy implements an advanced Bollinger Band strategy
// with breakout detection, inspired by Ekliptor/WolfBot.
type WolfBotBollingerStrategy struct {
	mu            sync.Mutex
	adapter       exchange.Adapter
	breakoutLimit int
	breakoutCount int
	lastTrend     string // "up", "down", "none"
}

// NewWolfBotBollingerStrategy creates a new WolfBot-inspired Bollinger strategy.
func NewWolfBotBollingerStrategy(adapter exchange.Adapter, breakoutLimit int) *WolfBotBollingerStrategy {
	return &WolfBotBollingerStrategy{
		adapter:       adapter,
		breakoutLimit: breakoutLimit,
		lastTrend:     "none",
	}
}

// Name returns the name of the strategy.
func (s *WolfBotBollingerStrategy) Name() string {
	return "wolfbot-bollinger"
}

// Execute processes market data and potentially executes an order.
// Note: In a real implementation, this would receive Candle data and indicators.
// For this assimilation phase, we implement the logic based on the analyzed WolfBot behavior.
func (s *WolfBotBollingerStrategy) Execute(ctx context.Context, order exchange.Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Logic placeholder:
	// In WolfBot:
	// If price >= UpperBand:
	//    breakoutCount++
	//    if lastTrend == "down" and breakoutCount >= breakoutLimit:
	//       Trend = "up", emitBuy (Breakout)
	//    else:
	//       Trend = "down", emitSell (Mean Reversion)
	// Same for LowerBand...

	fmt.Printf("WolfBotBollinger: Executing with breakoutCount=%d, lastTrend=%s\n", s.breakoutCount, s.lastTrend)

	// Delegate to market execution for now to satisfy the interface
	request := exchange.OrderRequest{
		Symbol:   order.Symbol,
		Side:     order.Side,
		Type:     exchange.MarketOrder,
		Quantity: order.Quantity,
	}

	_, err := s.adapter.PlaceOrder(ctx, request)
	return err
}

// UpdateState updates the internal state of the strategy based on price relative to bands.
// This matches WolfBot's checkIndicators() logic.
func (s *WolfBotBollingerStrategy) UpdateState(percentB float64) (action string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if percentB >= 1.0 { // Price at or above upper band
		s.breakoutCount++
		if s.lastTrend == "down" && s.breakoutCount >= s.breakoutLimit {
			s.setTrend("up")
			return "buy-breakout"
		}
		s.setTrend("down")
		return "sell-reversion"
	} else if percentB <= 0.0 { // Price at or below lower band
		s.breakoutCount++
		if s.lastTrend == "up" && s.breakoutCount >= s.breakoutLimit {
			s.setTrend("down")
			return "sell-breakout"
		}
		s.setTrend("up")
		return "buy-reversion"
	}

	return "none"
}

func (s *WolfBotBollingerStrategy) setTrend(trend string) {
	if s.lastTrend != trend {
		s.breakoutCount = 0
	}
	s.lastTrend = trend
}
