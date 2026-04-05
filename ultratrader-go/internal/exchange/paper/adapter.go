package paper

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

type Adapter struct {
	mu       sync.Mutex
	balances []exchange.Balance
	markets  []exchange.Market
	orders   []exchange.Order
}

func New() *Adapter {
	return &Adapter{
		balances: []exchange.Balance{{Asset: "USDT", Free: "10000", Locked: "0"}, {Asset: "BTC", Free: "0", Locked: "0"}},
		markets:  []exchange.Market{{Symbol: "BTCUSDT", BaseAsset: "BTC", QuoteAsset: "USDT", PriceScale: 2, QuantityScale: 6}, {Symbol: "ETHUSDT", BaseAsset: "ETH", QuoteAsset: "USDT", PriceScale: 2, QuantityScale: 6}},
	}
}

func (a *Adapter) Name() string { return "paper" }

func (a *Adapter) Capabilities() []exchange.Capability {
	return []exchange.Capability{exchange.CapabilitySpot, exchange.CapabilityPaper, exchange.CapabilityBalances, exchange.CapabilityOrders, exchange.CapabilityCandles, exchange.CapabilityTickers}
}

func (a *Adapter) ListMarkets(_ context.Context) ([]exchange.Market, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]exchange.Market, len(a.markets))
	copy(out, a.markets)
	return out, nil
}

func (a *Adapter) Balances(_ context.Context) ([]exchange.Balance, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]exchange.Balance, len(a.balances))
	copy(out, a.balances)
	return out, nil
}

func (a *Adapter) PlaceOrder(_ context.Context, request exchange.OrderRequest) (exchange.Order, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if request.Symbol == "" {
		return exchange.Order{}, fmt.Errorf("symbol is required")
	}
	if request.Quantity == "" {
		return exchange.Order{}, fmt.Errorf("quantity is required")
	}
	price := request.Price
	if price == "" {
		price = defaultPrice(request.Symbol)
	}
	order := exchange.Order{ID: fmt.Sprintf("paper-%d", time.Now().UnixNano()), Symbol: request.Symbol, Side: request.Side, Type: request.Type, Status: "filled", Quantity: request.Quantity, Price: price}
	a.orders = append(a.orders, order)
	return order, nil
}

func defaultPrice(symbol string) string {
	switch symbol {
	case "BTCUSDT":
		return "65000.00"
	case "ETHUSDT":
		return "3200.00"
	default:
		return "1.00"
	}
}
