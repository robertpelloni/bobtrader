package demo

import (
	"context"
	"fmt"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// WeeklyCycleStrategy exploits the Sunday peak pattern in crypto.
// Historically, crypto tends to:
// - Rally mid-week (Tue-Thu)
// - Peak on Sunday
// - Dip Monday-Tuesday (as weekend gains are taken)
//
// Strategy:
// - Buy on Monday/Tuesday dip
// - Hold through mid-week rally
// - Sell on Saturday/Sunday before the peak fades
type WeeklyCycleStrategy struct {
	accountID     string
	symbol        string
	quantity      string
	buyDays       []time.Weekday // Days to buy (accumulate)
	sellDays      []time.Weekday // Days to sell (take profit)
	lastSignalDay int
	inPosition    bool
	entryPrice    float64
}

func NewWeeklyCycleStrategy(
	accountID, symbol, quantity string,
) *WeeklyCycleStrategy {
	return &WeeklyCycleStrategy{
		accountID: accountID,
		symbol:    symbol,
		quantity:  quantity,
		// Buy on Sunday evening / Monday (accumulation phase)
		// The dip often happens Sunday night into Monday
		buyDays: []time.Weekday{time.Sunday, time.Monday},
		// Sell on Saturday / Sunday (distribution phase)
		// Peak often hits Saturday-Sunday
		sellDays: []time.Weekday{time.Friday, time.Saturday},
	}
}

func (s *WeeklyCycleStrategy) Name() string {
	return fmt.Sprintf("weekly-cycle-%s", s.symbol)
}

func (s *WeeklyCycleStrategy) OnTick(_ context.Context) ([]strategy.Signal, error) {
	return nil, nil
}

func (s *WeeklyCycleStrategy) OnMarketTick(_ context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	if tick.Symbol != s.symbol {
		return nil, nil
	}

	price := utils.ParseFloat(tick.Price)
	if price <= 0 {
		return nil, nil
	}

	// Use UTC time for consistency (crypto is 24/7 global market)
	now := time.Now().UTC()
	weekday := now.Weekday()
	hour := now.Hour()
	dayOfYear := now.YearDay()

	// Prevent duplicate signals on same day
	if dayOfYear == s.lastSignalDay {
		return nil, nil
	}

	var signals []strategy.Signal

	// BUY: Monday accumulation phase
	// Best time: Sunday night (22:00-23:59 UTC) or Monday early (00:00-06:00 UTC)
	// This catches the Sunday dip as US goes to sleep and before Asia wakes
	isBuyWindow := (weekday == time.Sunday && hour >= 22) || (weekday == time.Sunday && hour <= 6)
	if !isBuyWindow {
		isBuyWindow = (weekday == time.Monday && hour <= 6)
	}

	if isBuyWindow && !s.inPosition {
		signals = append(signals, strategy.Signal{
			StrategyName: s.Name(),
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "buy",
			Quantity:     s.quantity,
			Reason: fmt.Sprintf("WEEKLY CYCLE BUY: %s %d:00 UTC - Sunday/Monday accumulation window (historical dip)",
				weekday.String(), hour),
		})
		s.inPosition = true
		s.entryPrice = price
		s.lastSignalDay = dayOfYear
	}

	// SELL: Saturday/Sunday distribution phase
	// Best time: Saturday evening (18:00-23:59 UTC) or Sunday morning (00:00-12:00 UTC)
	// This captures the peak before Sunday sell-off
	isSellWindow := (weekday == time.Saturday && hour >= 18) || (weekday == time.Sunday && hour <= 12)

	if isSellWindow && s.inPosition {
		profitPct := (price - s.entryPrice) / s.entryPrice * 100
		signals = append(signals, strategy.Signal{
			StrategyName: s.Name(),
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "sell",
			Quantity:     s.quantity,
			Reason: fmt.Sprintf("WEEKLY CYCLE SELL: %s %d:00 UTC - Saturday/Sunday distribution (bought at %.2f, %.2f%% change)",
				weekday.String(), hour, s.entryPrice, profitPct),
		})
		s.inPosition = false
		s.entryPrice = 0
		s.lastSignalDay = dayOfYear
	}

	return signals, nil
}

// GetCycleInfo returns current cycle state for diagnostics
func (s *WeeklyCycleStrategy) GetCycleInfo() map[string]interface{} {
	now := time.Now().UTC()
	return map[string]interface{}{
		"current_day":    now.Weekday().String(),
		"current_hour":   now.Hour(),
		"in_position":    s.inPosition,
		"entry_price":    s.entryPrice,
		"buy_days":       []string{"Sunday", "Monday"},
		"sell_days":      []string{"Friday", "Saturday"},
		"best_buy_time":  "Sunday 22:00 - Monday 06:00 UTC",
		"best_sell_time": "Saturday 18:00 - Sunday 12:00 UTC",
	}
}
