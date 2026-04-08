package demo

import (
	"context"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

type TickMeanReversion struct {
	accountID        string
	symbol           string
	quantity         string
	lookbackTicks    int
	buyDeviationPct  float64
	sellDeviationPct float64
	prices           []float64
	lastSignalAction string
}

func NewTickMeanReversion(accountID, symbol, quantity string, lookbackTicks int, buyDeviationPct, sellDeviationPct float64) *TickMeanReversion {
	if lookbackTicks < 2 {
		lookbackTicks = 2
	}
	return &TickMeanReversion{
		accountID:        accountID,
		symbol:           symbol,
		quantity:         quantity,
		lookbackTicks:    lookbackTicks,
		buyDeviationPct:  buyDeviationPct,
		sellDeviationPct: sellDeviationPct,
	}
}

func (s *TickMeanReversion) Name() string                                        { return "demo-tick-mean-reversion" }
func (s *TickMeanReversion) OnTick(_ context.Context) ([]strategy.Signal, error) { return nil, nil }

func (s *TickMeanReversion) OnMarketTick(_ context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	if tick.Symbol != s.symbol {
		return nil, nil
	}
	price := utils.ParseFloat(tick.Price)
	if price <= 0 {
		return nil, nil
	}
	s.prices = append(s.prices, price)
	if len(s.prices) > s.lookbackTicks {
		s.prices = s.prices[len(s.prices)-s.lookbackTicks:]
	}
	if len(s.prices) < s.lookbackTicks {
		return nil, nil
	}
	avg := average(s.prices)
	if avg <= 0 {
		return nil, nil
	}
	deviationPct := ((price - avg) / avg) * 100

	if deviationPct <= -s.buyDeviationPct && s.lastSignalAction != "buy" {
		s.lastSignalAction = "buy"
		return []strategy.Signal{{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "buy",
			Reason:    "tick mean reversion below average threshold",
			Quantity:  s.quantity,
			OrderType: "market",
		}}, nil
	}
	if deviationPct >= s.sellDeviationPct && s.lastSignalAction != "sell" {
		s.lastSignalAction = "sell"
		return []strategy.Signal{{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "sell",
			Reason:    "tick mean reversion above average threshold",
			Quantity:  s.quantity,
			OrderType: "market",
		}}, nil
	}
	return nil, nil
}

func average(values []float64) float64 {
	if len(values) == 0 {
		return 0
	}
	total := 0.0
	for _, v := range values {
		total += v
	}
	return total / float64(len(values))
}
