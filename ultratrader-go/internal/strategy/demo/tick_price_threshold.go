package demo

import (
	"context"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

type TickPriceThreshold struct {
	accountID   string
	symbol      string
	quantity    string
	maxBuyPrice string
	emitted     bool
}

func NewTickPriceThreshold(accountID, symbol, quantity, maxBuyPrice string) *TickPriceThreshold {
	return &TickPriceThreshold{accountID: accountID, symbol: symbol, quantity: quantity, maxBuyPrice: maxBuyPrice}
}

func (s *TickPriceThreshold) Name() string                                        { return "demo-tick-price-threshold" }
func (s *TickPriceThreshold) OnTick(_ context.Context) ([]strategy.Signal, error) { return nil, nil }

func (s *TickPriceThreshold) OnMarketTick(_ context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	if s.emitted || tick.Symbol != s.symbol {
		return nil, nil
	}
	if compareDecimalStrings(tick.Price, s.maxBuyPrice) <= 0 {
		s.emitted = true
		return []strategy.Signal{{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "buy",
			Reason:    "stream price at or below threshold",
			Quantity:  s.quantity,
			OrderType: "market",
		}}, nil
	}
	return nil, nil
}
