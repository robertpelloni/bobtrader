package demo

import (
	"context"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

type PriceThreshold struct {
	accountID   string
	symbol      string
	quantity    string
	maxBuyPrice string
	feed        marketdata.Feed
	emitted     bool
}

func NewPriceThreshold(accountID, symbol, quantity, maxBuyPrice string, feed marketdata.Feed) *PriceThreshold {
	return &PriceThreshold{accountID: accountID, symbol: symbol, quantity: quantity, maxBuyPrice: maxBuyPrice, feed: feed}
}

func (s *PriceThreshold) Name() string { return "demo-price-threshold" }

func (s *PriceThreshold) OnTick(ctx context.Context) ([]strategy.Signal, error) {
	if s.emitted {
		return nil, nil
	}
	tick, err := s.feed.LatestTick(ctx, s.symbol)
	if err != nil {
		return nil, err
	}
	if compareDecimalStrings(tick.Price, s.maxBuyPrice) <= 0 {
		s.emitted = true
		return []strategy.Signal{{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "buy",
			Reason:    "price at or below threshold",
			Quantity:  s.quantity,
			OrderType: "market",
		}}, nil
	}
	return nil, nil
}

func compareDecimalStrings(left, right string) int {
	l := utils.ParseFloat(left)
	r := utils.ParseFloat(right)
	if l == r {
		return 0
	}
	if l < r {
		return -1
	}
	return 1
}
