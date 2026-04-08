package demo

import (
	"context"
	"fmt"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/indicator"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

type EMACrossover struct {
	accountID string
	symbol    string
	quantity  string
	fast      *indicator.EMA
	slow      *indicator.EMA
	feed      marketdata.Feed
	lastState string // "above" or "below"
}

func NewEMACrossover(accountID, symbol, quantity string, fastPeriod, slowPeriod int, feed marketdata.Feed) *EMACrossover {
	return &EMACrossover{
		accountID: accountID,
		symbol:    symbol,
		quantity:  quantity,
		fast:      indicator.NewEMA(fastPeriod),
		slow:      indicator.NewEMA(slowPeriod),
		feed:      feed,
	}
}

func (s *EMACrossover) Name() string { return "demo-ema-crossover" }

func (s *EMACrossover) OnTick(ctx context.Context) ([]strategy.Signal, error) {
	tick, err := s.feed.LatestTick(ctx, s.symbol)
	if err != nil {
		return nil, err
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
				Reason:    fmt.Sprintf("EMA crossover: fast(%v) crossed above slow(%v)", f, sl),
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
				Reason:    fmt.Sprintf("EMA crossover: fast(%v) crossed below slow(%v)", f, sl),
				Quantity:  s.quantity,
				OrderType: "market",
			}}, nil
		}
		s.lastState = "below"
	}

	return nil, nil
}
