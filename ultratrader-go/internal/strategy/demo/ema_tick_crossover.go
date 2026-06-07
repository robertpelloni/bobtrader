package demo

import (
	"context"
	"fmt"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/indicator"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// EMATickCrossover generates signals when fast EMA crosses slow EMA.
// Unlike the polling-based EMACrossover, this works tick-by-tick
// for real-time signal generation.
type EMATickCrossover struct {
	accountID string
	symbol    string
	quantity  string
	fast      *indicator.EMA
	slow      *indicator.EMA
	lastState string // "above" or "below"
}

func NewEMATickCrossover(accountID, symbol, quantity string, fastPeriod, slowPeriod int) *EMATickCrossover {
	return &EMATickCrossover{
		accountID: accountID,
		symbol:    symbol,
		quantity:  quantity,
		fast:      indicator.NewEMA(fastPeriod),
		slow:      indicator.NewEMA(slowPeriod),
	}
}

func (s *EMATickCrossover) Name() string {
	return "ema-tick-crossover"
}

func (s *EMATickCrossover) OnTick(_ context.Context) ([]strategy.Signal, error) {
	return nil, nil
}

func (s *EMATickCrossover) OnMarketTick(_ context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	if tick.Symbol != s.symbol {
		return nil, nil
	}
	price := utils.ParseFloat(tick.Price)
	if price <= 0 {
		return nil, nil
	}

	f := s.fast.Update(price)
	sl := s.slow.Update(price)

	if f > sl {
		if s.lastState == "below" {
			s.lastState = "above"
			return []strategy.Signal{{
				AccountID: s.accountID,
				Symbol:    s.symbol,
				Action:    "buy",
				Reason:    fmt.Sprintf("EMA crossover up: fast(%.2f) > slow(%.2f)", f, sl),
				Quantity:  s.quantity,
				OrderType: "market",
			}}, nil
		}
		s.lastState = "above"
	} else if f < sl {
		if s.lastState == "above" {
			s.lastState = "below"
			return []strategy.Signal{{
				AccountID: s.accountID,
				Symbol:    s.symbol,
				Action:    "sell",
				Reason:    fmt.Sprintf("EMA crossover down: fast(%.2f) < slow(%.2f)", f, sl),
				Quantity:  s.quantity,
				OrderType: "market",
			}}, nil
		}
		s.lastState = "below"
	}

	return nil, nil
}
