package marketmaking

import (
	"context"
	"fmt"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

// QuotingStyle defines how the market maker places quotes.
type QuotingStyle string

const (
	PingPong QuotingStyle = "ping-pong"
)

// MarketMaker implements a basic market-making strategy inspired by Krypto-trading-bot.
type MarketMaker struct {
	mu           sync.Mutex
	adapter      exchange.Adapter
	symbol       string
	style        QuotingStyle
	bidSpread    float64 // percentage
	askSpread    float64 // percentage
	quantity     string
	activeOrders map[string]exchange.Order
}

// NewMarketMaker creates a new market making strategy.
func NewMarketMaker(adapter exchange.Adapter, symbol string, quantity string) *MarketMaker {
	return &MarketMaker{
		adapter:      adapter,
		symbol:       symbol,
		style:        PingPong,
		bidSpread:    0.01, // 0.1%
		askSpread:    0.01,
		quantity:     quantity,
		activeOrders: make(map[string]exchange.Order),
	}
}

func (m *MarketMaker) Name() string {
	return "market-maker"
}

// OnPriceUpdate is called when the market price changes.
func (m *MarketMaker) OnPriceUpdate(ctx context.Context, midPrice float64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Simplistic PingPong logic:
	// If no orders, place one on each side.
	// If one side fills, replace it.

	if len(m.activeOrders) == 0 {
		// Place initial quotes
		return m.placeQuotes(ctx, midPrice)
	}

	return nil
}

func (m *MarketMaker) placeQuotes(ctx context.Context, midPrice float64) error {
	bidPrice := midPrice * (1 - m.bidSpread/100)
	askPrice := midPrice * (1 + m.askSpread/100)

	// Place Bid
	bidReq := exchange.OrderRequest{
		Symbol:   m.symbol,
		Side:     exchange.Buy,
		Type:     exchange.LimitOrder,
		Quantity: m.quantity,
		Price:    fmt.Sprintf("%.2f", bidPrice),
	}
	bid, err := m.adapter.PlaceOrder(ctx, bidReq)
	if err == nil {
		m.activeOrders[bid.ID] = bid
	}

	// Place Ask
	askReq := exchange.OrderRequest{
		Symbol:   m.symbol,
		Side:     exchange.Sell,
		Type:     exchange.LimitOrder,
		Quantity: m.quantity,
		Price:    fmt.Sprintf("%.2f", askPrice),
	}
	ask, err := m.adapter.PlaceOrder(ctx, askReq)
	if err == nil {
		m.activeOrders[ask.ID] = ask
	}

	return nil
}

// Execute handles the Strategy interface.
func (m *MarketMaker) Execute(ctx context.Context, order exchange.Order) error {
	// Market maker is usually self-driven by price updates, but we can support external triggers.
	return nil
}
