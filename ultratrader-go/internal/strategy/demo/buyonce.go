package demo

import (
	"context"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

type BuyOnce struct {
	accountID string
	symbol    string
	quantity  string
	once      sync.Once
	emitted   bool
}

func NewBuyOnce(accountID, symbol, quantity string) *BuyOnce {
	return &BuyOnce{accountID: accountID, symbol: symbol, quantity: quantity}
}

func (s *BuyOnce) Name() string { return "demo-buy-once" }

func (s *BuyOnce) OnTick(_ context.Context) ([]strategy.Signal, error) {
	s.once.Do(func() { s.emitted = true })
	if !s.emitted {
		return nil, nil
	}
	s.emitted = false
	return []strategy.Signal{{
		AccountID: s.accountID,
		Symbol:    s.symbol,
		Action:    "buy",
		Reason:    "bootstrap demo strategy",
		Quantity:  s.quantity,
		OrderType: "market",
	}}, nil
}
