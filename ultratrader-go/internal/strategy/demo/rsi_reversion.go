package demo

import (
	"context"
	"fmt"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/indicator"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// RSIReversion generates buy signals when RSI drops below oversold
// and sell signals when RSI rises above overbought. It is a mean-reversion
// strategy that works well in ranging markets.
type RSIReversion struct {
	accountID    string
	symbol       string
	quantity     string
	period       int
	oversold     float64
	overbought   float64
	rsi          *indicator.RSI
	lastSignal   string // "buy" or "sell" — prevents repeated signals
}

func NewRSIReversion(accountID, symbol, quantity string, period int, oversold, overbought float64) *RSIReversion {
	if period < 2 {
		period = 14
	}
	if oversold <= 0 {
		oversold = 30
	}
	if overbought <= 0 {
		overbought = 70
	}
	return &RSIReversion{
		accountID:  accountID,
		symbol:     symbol,
		quantity:   quantity,
		period:     period,
		oversold:   oversold,
		overbought: overbought,
		rsi:        indicator.NewRSI(period),
	}
}

func (s *RSIReversion) Name() string {
	return fmt.Sprintf("rsi-reversion-%d", s.period)
}

func (s *RSIReversion) OnTick(_ context.Context) ([]strategy.Signal, error) {
	return nil, nil
}

func (s *RSIReversion) OnMarketTick(_ context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	if tick.Symbol != s.symbol {
		return nil, nil
	}
	price := utils.ParseFloat(tick.Price)
	if price <= 0 {
		return nil, nil
	}

	rsiVal := s.rsi.Update(price)

	// Don't signal until RSI is fully warmed up
	if !s.rsi.Ready() {
		return nil, nil
	}

	// Buy when RSI crosses back above oversold from below
	// (only after RSI has had enough data to be meaningful)
	if rsiVal <= s.oversold && s.lastSignal != "buy" && rsiVal > 0 && rsiVal < 100 {
		s.lastSignal = "buy"
		return []strategy.Signal{{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "buy",
			Reason:    fmt.Sprintf("RSI(%d) oversold at %.1f", s.period, rsiVal),
			Quantity:  s.quantity,
			OrderType: "market",
		}}, nil
	}

	// Sell when RSI crosses back below overbought from above
	if rsiVal >= s.overbought && s.lastSignal != "sell" && rsiVal > 0 && rsiVal < 100 {
		s.lastSignal = "sell"
		return []strategy.Signal{{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "sell",
			Reason:    fmt.Sprintf("RSI(%d) overbought at %.1f", s.period, rsiVal),
			Quantity:  s.quantity,
			OrderType: "market",
		}}, nil
	}

	// Reset signal state when RSI returns to neutral zone (40-60)
	if rsiVal >= 40 && rsiVal <= 60 {
		s.lastSignal = ""
	}

	return nil, nil
}
