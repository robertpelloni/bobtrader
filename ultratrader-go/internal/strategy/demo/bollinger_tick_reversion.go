package demo

import (
	"context"
	"fmt"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/indicator"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// BollingerTickReversion generates buy signals when price touches the lower
// Bollinger Band and sell signals when price touches the upper band.
// Unlike the candle-based BollingerReversion, this works tick-by-tick.
type BollingerTickReversion struct {
	accountID  string
	symbol     string
	quantity   string
	period     int
	multiplier float64
	bb         *indicator.BollingerBands
	lastSignal string
}

func NewBollingerTickReversion(accountID, symbol, quantity string, period int, multiplier float64) *BollingerTickReversion {
	if period < 5 {
		period = 20
	}
	if multiplier <= 0 {
		multiplier = 2.0
	}
	return &BollingerTickReversion{
		accountID:  accountID,
		symbol:     symbol,
		quantity:   quantity,
		period:     period,
		multiplier: multiplier,
		bb:         indicator.NewBollingerBands(period, multiplier),
	}
}

func (s *BollingerTickReversion) Name() string {
	return fmt.Sprintf("bollinger-tick-%d-%.1f", s.period, s.multiplier)
}

func (s *BollingerTickReversion) OnTick(_ context.Context) ([]strategy.Signal, error) {
	return nil, nil
}

func (s *BollingerTickReversion) OnMarketTick(_ context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	if tick.Symbol != s.symbol {
		return nil, nil
	}
	price := utils.ParseFloat(tick.Price)
	if price <= 0 {
		return nil, nil
	}

	result := s.bb.Update(price)

	// Need at least `period` prices to have valid bands
	if result.Upper == 0 && result.Lower == 0 {
		return nil, nil
	}

	// Buy when price drops to or below lower band
	if price <= result.Lower && s.lastSignal != "buy" {
		s.lastSignal = "buy"
		return []strategy.Signal{{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "buy",
			Reason:    fmt.Sprintf("bollinger lower band touch: price %.2f <= lower %.2f", price, result.Lower),
			Quantity:  s.quantity,
			OrderType: "market",
		}}, nil
	}

	// Sell when price rises to or above upper band
	// (only if we're already in a position — the TrailingTakeProfit also handles exits)
	if price >= result.Upper && s.lastSignal != "sell" {
		s.lastSignal = "sell"
		return []strategy.Signal{{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "sell",
			Reason:    fmt.Sprintf("bollinger upper band touch: price %.2f >= upper %.2f", price, result.Upper),
			Quantity:  s.quantity,
			OrderType: "market",
		}}, nil
	}

	return nil, nil
}
