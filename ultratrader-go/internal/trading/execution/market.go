package execution

import (
	"context"
	"fmt"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

// MarketStrategy implements a simple market order execution strategy.
type MarketStrategy struct {
	adapter exchange.Adapter
}

// NewMarketStrategy creates a new market order execution strategy.
func NewMarketStrategy(adapter exchange.Adapter) *MarketStrategy {
	return &MarketStrategy{
		adapter: adapter,
	}
}

// Name returns the name of the strategy.
func (s *MarketStrategy) Name() string {
	return "market"
}

// Execute executes the order as a market order.
func (s *MarketStrategy) Execute(ctx context.Context, order exchange.Order) error {
	request := exchange.OrderRequest{
		Symbol:   order.Symbol,
		Side:     order.Side,
		Type:     exchange.MarketOrder,
		Quantity: order.Quantity,
	}

	_, err := s.adapter.PlaceOrder(ctx, request)
	if err != nil {
		return fmt.Errorf("failed to place market order: %w", err)
	}

	return nil
}
