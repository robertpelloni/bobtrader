package demo

import (
	"context"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

type TickMomentumBurst struct {
	accountID        string
	symbol           string
	quantity         string
	lookbackTicks    int
	buyThresholdPct  float64
	sellThresholdPct float64
	prices           []float64
	lastSignalAction string
}

func NewTickMomentumBurst(accountID, symbol, quantity string, lookbackTicks int, buyThresholdPct, sellThresholdPct float64) *TickMomentumBurst {
	if lookbackTicks < 2 {
		lookbackTicks = 2
	}
	return &TickMomentumBurst{
		accountID:        accountID,
		symbol:           symbol,
		quantity:         quantity,
		lookbackTicks:    lookbackTicks,
		buyThresholdPct:  buyThresholdPct,
		sellThresholdPct: sellThresholdPct,
	}
}

func (s *TickMomentumBurst) Name() string                                        { return "demo-tick-momentum-burst" }
func (s *TickMomentumBurst) OnTick(_ context.Context) ([]strategy.Signal, error) { return nil, nil }

func (s *TickMomentumBurst) OnMarketTick(_ context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
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

	oldest := s.prices[0]
	latest := s.prices[len(s.prices)-1]
	if oldest <= 0 {
		return nil, nil
	}
	changePct := ((latest - oldest) / oldest) * 100

	if changePct >= s.buyThresholdPct && s.lastSignalAction != "buy" {
		s.lastSignalAction = "buy"
		return []strategy.Signal{{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "buy",
			Reason:    "tick momentum burst above threshold",
			Quantity:  s.quantity,
			OrderType: "market",
		}}, nil
	}
	if changePct <= -s.sellThresholdPct && s.lastSignalAction != "sell" {
		s.lastSignalAction = "sell"
		return []strategy.Signal{{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "sell",
			Reason:    "tick momentum burst below threshold",
			Quantity:  s.quantity,
			OrderType: "market",
		}}, nil
	}
	return nil, nil
}
