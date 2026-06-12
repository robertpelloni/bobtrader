package demo

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

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
// 4. If price drops stopLossPct% below entry, cut losses (stop-loss)
// 5. If position is held longer than maxHoldDuration, sell (time-based exit)
type TrailingTakeProfit struct {
	accountID    string
	symbol       string
	quantity     string
	activatePct  float64       // profit % above entry to activate trailing
	trailPct     float64       // trail gap % below peak
	stopLossPct  float64       // max loss % before cutting (0 = disabled)
	maxHold      time.Duration // max time to hold a position (0 = disabled)
	portfolio    PositionReader
	feed         marketdata.Feed
	entryPrice   float64
	highWaterMark float64
	activated    bool
	positionKnown bool
	positionStart time.Time // when we first detected the position
}

// PositionReader reads current position data for a symbol.
type PositionReader interface {
	HasOpenPosition(symbol string) bool
	PositionQuantity(symbol string) float64
	AverageEntryPrice(symbol string) float64
}

// TrailingOption configures a TrailingTakeProfit strategy.
type TrailingOption func(*TrailingTakeProfit)

// WithStopLossPct sets the stop-loss percentage. Set to 0 to disable.
func WithStopLossPct(pct float64) TrailingOption {
	return func(s *TrailingTakeProfit) { s.stopLossPct = pct }
}

// WithMaxHoldMinutes sets the maximum hold duration in minutes. Set to 0 to disable.
func WithMaxHoldMinutes(minutes int) TrailingOption {
	return func(s *TrailingTakeProfit) {
		if minutes > 0 {
			s.maxHold = time.Duration(minutes) * time.Minute
		}
	}
}

// WithPortfolioEntry sets the portfolio reader for entry price lookup.
func WithPortfolioEntry(portfolio PositionReader) TrailingOption {
	return func(s *TrailingTakeProfit) { s.portfolio = portfolio }
}

// WithFeed sets the market data feed for price lookups.
func WithFeed(feed marketdata.Feed) TrailingOption {
	return func(s *TrailingTakeProfit) { s.feed = feed }
}

func NewTrailingTakeProfit(
	accountID, symbol, quantity string,
	activatePct, trailPct float64,
	opts ...TrailingOption,
) *TrailingTakeProfit {
	if activatePct <= 0 {
		activatePct = 1.0
	}
	if trailPct <= 0 {
		trailPct = 0.3
	}
	s := &TrailingTakeProfit{
		accountID:   accountID,
		symbol:      symbol,
		quantity:    quantity,
		activatePct: activatePct,
		trailPct:    trailPct,
		stopLossPct: 3.0,
		maxHold:     5 * time.Minute,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
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

	// Reset state when no position held
	if s.portfolio == nil || !s.portfolio.HasOpenPosition(s.symbol) {
		s.resetState()
		return nil, nil
	}

	price := utils.ParseFloat(tick.Price)
	if price <= 0 {
		return nil, nil
	}

	// Record entry price from portfolio when we first detect a position
	if !s.positionKnown {
		s.entryPrice = s.portfolio.AverageEntryPrice(s.symbol)
		if s.entryPrice <= 0 {
			s.entryPrice = price
		}
		s.positionKnown = true
		s.highWaterMark = price
		s.positionStart = time.Now().UTC()
		return nil, nil // wait one tick before evaluating
	}

	// Update high water mark
	if price > s.highWaterMark {
		s.highWaterMark = price
	}

	// Check time-based exit: if held too long without reaching activation
	if s.maxHold > 0 && !s.positionStart.IsZero() {
		held := time.Since(s.positionStart)
		if held >= s.maxHold {
			heldQty := s.portfolio.PositionQuantity(s.symbol)
			qty := formatQuantity(s.symbol, heldQty)
			if heldQty <= 0 {
				qty = s.quantity
			}
			profitPct := 0.0
			if s.entryPrice > 0 {
				profitPct = ((price - s.entryPrice) / s.entryPrice) * 100
			}
			reason := "max-hold-exit"
			if profitPct > 0 {
				reason = fmt.Sprintf("max-hold %.0fs: price %.2f (+%.2f%% from entry %.2f)",
					s.maxHold.Seconds(), price, profitPct, s.entryPrice)
			} else {
				reason = fmt.Sprintf("max-hold %.0fs: price %.2f (%.2f%% from entry %.2f)",
					s.maxHold.Seconds(), price, profitPct, s.entryPrice)
			}
			s.resetState()
			return []strategy.Signal{{
				AccountID: s.accountID,
				Symbol:    s.symbol,
				Action:    "sell",
				Reason:    reason,
				Quantity:  qty,
				OrderType: "market",
			}}, nil
		}
	}

	// Check stop-loss first: if price dropped too far below entry, cut losses
	if s.stopLossPct > 0 && s.entryPrice > 0 {
		lossPct := ((s.entryPrice - price) / s.entryPrice) * 100
		if lossPct >= s.stopLossPct {
			heldQty := s.portfolio.PositionQuantity(s.symbol)
			qty := formatQuantity(s.symbol, heldQty)
			if heldQty <= 0 {
				qty = s.quantity
			}
			s.resetState()
			return []strategy.Signal{{
				AccountID: s.accountID,
				Symbol:    s.symbol,
				Action:    "sell",
				Reason:    fmt.Sprintf("stop-loss: price %.2f dropped %.1f%% below entry %.2f", price, lossPct, s.entryPrice),
				Quantity:  qty,
				OrderType: "market",
			}}, nil
		}
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
		heldQty := s.portfolio.PositionQuantity(s.symbol)
		qty := formatQuantity(s.symbol, heldQty)
		if heldQty <= 0 {
			qty = s.quantity
		}
		s.resetState()
		return []strategy.Signal{{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "sell",
			Reason:    fmt.Sprintf("trailing TP: price %.2f <= trail %.2f (entry %.2f, high %.2f)",
				price, trailStop, s.entryPrice, s.highWaterMark),
			Quantity:  qty,
			OrderType: "market",
		}}, nil
	}

	return nil, nil
}

func (s *TrailingTakeProfit) resetState() {
	s.activated = false
	s.highWaterMark = 0
	s.entryPrice = 0
	s.positionKnown = false
	s.positionStart = time.Time{}
}

// formatQuantity formats a float64 quantity to a string with appropriate precision for the symbol.
func formatQuantity(symbol string, qty float64) string {
	precision := 4 // default fallback to 4 decimals (e.g. ETHUSDT)
	upperSymbol := strings.ToUpper(symbol)

	if strings.Contains(upperSymbol, "BTC") {
		precision = 5
	} else if strings.Contains(upperSymbol, "ETH") {
		precision = 4
	} else if strings.Contains(upperSymbol, "BNB") {
		precision = 3
	} else if strings.Contains(upperSymbol, "SOL") {
		precision = 3
	} else if strings.Contains(upperSymbol, "XRP") {
		precision = 1
	} else if strings.Contains(upperSymbol, "ADA") {
		precision = 1
	} else if strings.Contains(upperSymbol, "DOGE") {
		precision = 0
	} else {
		// General fallback based on quantity range if unknown
		if qty >= 1 {
			precision = 4
		} else if qty >= 0.01 {
			precision = 6
		} else {
			precision = 8
		}
	}

	pow := math.Pow(10, float64(precision))
	// Add a tiny epsilon (1e-9) to prevent float64 representation precision loss from truncating down a value like 0.0297 to 0.0296.
	truncatedQty := math.Floor(qty*pow + 1e-9) / pow

	return fmt.Sprintf("%.*f", precision, truncatedQty)
}
