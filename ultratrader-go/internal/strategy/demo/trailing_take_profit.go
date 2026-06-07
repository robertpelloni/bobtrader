package demo

import (
	"context"
	"fmt"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// TrailingTakeProfit watches held positions and generates sell signals
// when the price rises then falls back past a trailing stop.
// This is the primary exit strategy for autonomous trading.
//
// Logic:
// 1. Wait until price is at least activatePct% above the entry price
// 2. Track the high water mark as price rises
// 3. When price drops trailPct% below the high water mark, sell
type TrailingTakeProfit struct {
	accountID     string
	symbol        string
	quantity      string
	activatePct   float64 // profit % above entry to activate
	trailPct      float64 // trail gap % below peak
	portfolio     PositionReader
	feed          marketdata.Feed
	entryPrice    float64 // recorded when position detected
	activated     bool
	highWaterMark float64
	positionKnown bool
}

// PositionReader reads the current held quantity for a symbol.
type PositionReader interface {
	HasOpenPosition(symbol string) bool
	PositionQuantity(symbol string) float64
}

func NewTrailingTakeProfit(
	accountID, symbol, quantity string,
	activatePct, trailPct float64,
	portfolio PositionReader,
	feed marketdata.Feed,
) *TrailingTakeProfit {
	if activatePct <= 0 {
		activatePct = 2.0
	}
	if trailPct <= 0 {
		trailPct = 0.5
	}
	return &TrailingTakeProfit{
		accountID:   accountID,
		symbol:      symbol,
		quantity:    quantity,
		activatePct: activatePct,
		trailPct:    trailPct,
		portfolio:   portfolio,
		feed:        feed,
	}
}

func (s *TrailingTakeProfit) Name() string {
	return fmt.Sprintf("trailing-tp-%s", s.symbol)
}

func (s *TrailingTakeProfit) OnTick(_ context.Context) ([]strategy.Signal, error) {
	return nil, nil
}

func (s *TrailingTakeProfit) OnMarketTick(_ context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	if tick.Symbol != s.symbol {
		return nil, nil
	}

	// Only sell if we have a position
	if s.portfolio == nil || !s.portfolio.HasOpenPosition(s.symbol) {
		s.activated = false
		s.highWaterMark = 0
		s.entryPrice = 0
		s.positionKnown = false
		return nil, nil
	}

	price := utils.ParseFloat(tick.Price)
	if price <= 0 {
		return nil, nil
	}

	// Record entry price when we first detect a position
	if !s.positionKnown {
		s.entryPrice = price
		s.positionKnown = true
		s.highWaterMark = price
		return nil, nil // wait one tick before evaluating
	}

	// Update high water mark
	if price > s.highWaterMark {
		s.highWaterMark = price
	}

	// Check activation: price must be at least activatePct% above entry
	if !s.activated {
		if s.entryPrice > 0 {
			profitPct := ((price - s.entryPrice) / s.entryPrice) * 100
			if profitPct >= s.activatePct {
				s.activated = true
			}
		}
		if !s.activated {
			return nil, nil // not yet profitable enough
		}
	}

	// After activation, sell if price drops below trailing stop
	trailStop := s.highWaterMark * (1 - s.trailPct/100)
	if price <= trailStop {
		// Sell entire position
		heldQty := s.portfolio.PositionQuantity(s.symbol)
		qty := formatQuantity(heldQty)
		if heldQty <= 0 {
			qty = s.quantity // fallback
		}
		s.activated = false
		s.highWaterMark = 0
		s.entryPrice = 0
		s.positionKnown = false
		return []strategy.Signal{{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "sell",
			Reason:    fmt.Sprintf("trailing TP: price %.2f <= trail %.2f (entry %.2f, high %.2f)", price, trailStop, s.entryPrice, s.highWaterMark),
			Quantity:  qty,
			OrderType: "market",
		}}, nil
	}

	return nil, nil
}

// formatQuantity formats a float64 quantity to a string with appropriate precision.
func formatQuantity(qty float64) string {
	if qty >= 1 {
		return fmt.Sprintf("%.4f", qty)
	}
	if qty >= 0.01 {
		return fmt.Sprintf("%.6f", qty)
	}
	return fmt.Sprintf("%.8f", qty)
}
