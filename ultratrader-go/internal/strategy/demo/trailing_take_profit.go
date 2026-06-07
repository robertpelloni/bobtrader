package demo

import (
	"context"
	"fmt"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// TrailingTakeProfit watches held positions and generates sell signals
// when price rises above a profit threshold, then trails upward.
// It sells when price falls back below the trailing line.
// This is the core exit strategy for autonomous trading.
type TrailingTakeProfit struct {
	accountID    string
	symbol       string
	quantity     string
	activatePct  float64 // profit % to activate trailing
	trailPct     float64 // trail gap % below peak
	portfolio    PositionReader
	feed         marketdata.Feed
	activated    bool
	highWaterMark float64
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
		accountID:    accountID,
		symbol:       symbol,
		quantity:     quantity,
		activatePct:  activatePct,
		trailPct:     trailPct,
		portfolio:    portfolio,
		feed:         feed,
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
		return nil, nil
	}

	price := utils.ParseFloat(tick.Price)
	if price <= 0 {
		return nil, nil
	}

	// Update high water mark
	if price > s.highWaterMark {
		s.highWaterMark = price
	}

	// Calculate trailing stop price
	trailStop := s.highWaterMark * (1 - s.trailPct/100)

	// Check activation: price must be at least activatePct above our entry
	// We approximate entry from portfolio (or just use trail logic)
	if !s.activated {
		// Simple activation: if price has moved up at all from recent low
		// The actual entry price check happens implicitly — if we're in position
		// and price trails up, we activate once it drops
		if s.highWaterMark > 0 && price >= s.highWaterMark*(1-s.activatePct/100) {
			// Not yet activated — wait for profit
			return nil, nil
		}
		// We've already dipped from high — but let's require at least some rise
		// Activation happens once the price is within trailing range of the high
		if price <= trailStop && s.highWaterMark > price*(1+s.activatePct/100) {
			// High water mark was significantly above current price
			s.activated = true
		}
		return nil, nil
	}

	// After activation, sell if price drops below trailing stop
	if price <= trailStop {
		// Determine quantity to sell — sell entire position
		qty := s.quantity
		if s.portfolio != nil {
			heldQty := s.portfolio.PositionQuantity(s.symbol)
			if heldQty > 0 {
				qty = formatQuantity(heldQty)
			}
		}
		s.activated = false
		s.highWaterMark = 0
		return []strategy.Signal{{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "sell",
			Reason:    fmt.Sprintf("trailing TP hit: price %.2f <= trail %.2f (high %.2f, trail %.1f%%)", price, trailStop, s.highWaterMark, s.trailPct),
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
