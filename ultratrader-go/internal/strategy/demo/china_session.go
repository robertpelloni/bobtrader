package demo

import (
	"context"
	"fmt"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// ChinaSessionStrategy exploits the Asian session volatility pattern.
// Key insight: China's population comes online around 9AM Beijing time (1AM UTC)
// and drives significant volume and price movement.
//
// Historical patterns:
// - 00:00-01:00 UTC: Pre-Asia quiet period (low volume)
// - 01:00-03:00 UTC: Asia wakes up (China, Japan, Korea) - VOLATILITY SPIKE
// - 03:00-05:00 UTC: Asian morning session - trend continuation
// - 08:00-10:00 UTC: European overlap - another volatility window
//
// Strategy:
// - Buy BEFORE the Asia session (00:00-01:00 UTC) when quiet
// - Sell INTO the Asia session spike (01:30-03:00 UTC)
// - Also catches the "China lunch break" pattern (04:00-05:00 UTC = 12-1PM Beijing)
type ChinaSessionStrategy struct {
	accountID      string
	symbol         string
	quantity       string
	lastSignalHour int
	inPosition     bool
	entryPrice     float64
	priceHistory   []float64
	maxHistory     int
}

func NewChinaSessionStrategy(
	accountID, symbol, quantity string,
) *ChinaSessionStrategy {
	return &ChinaSessionStrategy{
		accountID:    accountID,
		symbol:       symbol,
		quantity:     quantity,
		priceHistory: make([]float64, 0, 60),
		maxHistory:   60,
	}
}

func (s *ChinaSessionStrategy) Name() string {
	return fmt.Sprintf("china-session-%s", s.symbol)
}

func (s *ChinaSessionStrategy) OnTick(_ context.Context) ([]strategy.Signal, error) {
	return nil, nil
}

func (s *ChinaSessionStrategy) OnMarketTick(_ context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	if tick.Symbol != s.symbol {
		return nil, nil
	}

	price := utils.ParseFloat(tick.Price)
	if price <= 0 {
		return nil, nil
	}

	// Track price history
	s.priceHistory = append(s.priceHistory, price)
	if len(s.priceHistory) > s.maxHistory {
		s.priceHistory = s.priceHistory[1:]
	}

	now := time.Now().UTC()
	hour := now.Hour()
	minute := now.Minute()

	// Prevent duplicate signals in same hour
	if hour == s.lastSignalHour {
		return nil, nil
	}

	// Need some price history
	if len(s.priceHistory) < 10 {
		return nil, nil
	}

	var signals []strategy.Signal

	// ── PHASE 1: Pre-Asia Accumulation ──────────────
	// Buy during the quiet window before Asia wakes
	// 00:00-00:59 UTC (8PM-9PM EST / before midnight Beijing)
	// Volume is low, spreads are tight, good entry point
	if hour == 0 && !s.inPosition {
		// Check if price is stable (not crashing)
		recentVolatility := s.calculateVolatility(10)
		if recentVolatility < 2.0 { // Less than 2% volatility = quiet
			signals = append(signals, strategy.Signal{
				StrategyName: s.Name(),
				AccountID:    s.accountID,
				Symbol:       s.symbol,
				Action:       "buy",
				Quantity:     s.quantity,
				Reason: fmt.Sprintf("CHINA SESSION PRE-LOAD: %d:%02d UTC - buying quiet window before Asia wakes (volatility: %.2f%%)",
					hour, minute, recentVolatility),
			})
			s.inPosition = true
			s.entryPrice = price
			s.lastSignalHour = hour
		}
	}

	// ── PHASE 2: Asia Session Spike ─────────────────
	// Sell into the volatility spike when China/Asia comes online
	// 01:30-03:00 UTC (9:30AM-11AM Beijing)
	// This is when volume spikes and price moves
	if hour >= 1 && hour <= 3 && minute >= 30 && s.inPosition {
		profitPct := (price - s.entryPrice) / s.entryPrice * 100

		// Only sell if we have profit or if it's getting late in the window
		if profitPct > 0 || hour >= 2 {
			signals = append(signals, strategy.Signal{
				StrategyName: s.Name(),
				AccountID:    s.accountID,
				Symbol:       s.symbol,
				Action:       "sell",
				Quantity:     s.quantity,
				Reason: fmt.Sprintf("CHINA SESSION SELL: %d:%02d UTC - Asia volatility spike (bought at %.2f, %.2f%% change)",
					hour, minute, s.entryPrice, profitPct),
			})
			s.inPosition = false
			s.entryPrice = 0
			s.lastSignalHour = hour
		}
	}

	// ── PHASE 3: Asia Lunch Break ───────────────────
	// China lunch break = 12:00-13:00 Beijing = 04:00-05:00 UTC
	// Often a mini-dip before afternoon session
	if hour == 4 && !s.inPosition {
		signals = append(signals, strategy.Signal{
			StrategyName: s.Name(),
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "buy",
			Quantity:     s.quantity,
			Reason: fmt.Sprintf("CHINA LUNCH BREAK: %d:%02d UTC - buying Asia lunch dip (12PM Beijing)",
				hour, minute),
		})
		s.inPosition = true
		s.entryPrice = price
		s.lastSignalHour = hour
	}

	// ── PHASE 4: Europe Overlap ─────────────────────
	// European session overlap (08:00-10:00 UTC) often has another spike
	// Sell if we're still holding from lunch break
	if hour >= 8 && hour <= 10 && s.inPosition {
		profitPct := (price - s.entryPrice) / s.entryPrice * 100
		signals = append(signals, strategy.Signal{
			StrategyName: s.Name(),
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "sell",
			Quantity:     s.quantity,
			Reason: fmt.Sprintf("EUROPE OVERLAP SELL: %d:%02d UTC - European session spike (bought at %.2f, %.2f%% change)",
				hour, minute, s.entryPrice, profitPct),
		})
		s.inPosition = false
		s.entryPrice = 0
		s.lastSignalHour = hour
	}

	return signals, nil
}

func (s *ChinaSessionStrategy) calculateVolatility(period int) float64 {
	if len(s.priceHistory) < period {
		period = len(s.priceHistory)
	}
	if period < 2 {
		return 0
	}

	// Calculate simple price range as volatility measure
	high := 0.0
	low := 1e18
	for _, p := range s.priceHistory[len(s.priceHistory)-period:] {
		if p > high {
			high = p
		}
		if p < low {
			low = p
		}
	}

	if low == 0 {
		return 0
	}
	return (high - low) / low * 100
}

// GetSessionInfo returns current session state for diagnostics
func (s *ChinaSessionStrategy) GetSessionInfo() map[string]interface{} {
	now := time.Now().UTC()
	beijing := now.Add(8 * time.Hour)

	session := "Unknown"
	hour := now.Hour()
	switch {
	case hour >= 0 && hour < 1:
		session = "Pre-Asia (quiet)"
	case hour >= 1 && hour < 5:
		session = "Asian Morning (volatile)"
	case hour >= 5 && hour < 8:
		session = "Asian Afternoon"
	case hour >= 8 && hour < 12:
		session = "European Session"
	case hour >= 12 && hour < 16:
		session = "US Morning"
	case hour >= 16 && hour < 20:
		session = "US Afternoon"
	case hour >= 20 && hour < 24:
		session = "US Evening / Pre-Asia"
	}

	return map[string]interface{}{
		"utc_time":     now.Format("15:04"),
		"beijing_time": beijing.Format("15:04"),
		"session":      session,
		"in_position":  s.inPosition,
		"entry_price":  s.entryPrice,
		"windows": map[string]string{
			"pre_asia_buy":    "00:00-01:00 UTC (8-9PM Beijing)",
			"asia_spike_sell": "01:30-03:00 UTC (9:30-11AM Beijing)",
			"lunch_break_buy": "04:00-05:00 UTC (12-1PM Beijing)",
			"europe_overlap":  "08:00-10:00 UTC (4-6PM Beijing)",
		},
	}
}
