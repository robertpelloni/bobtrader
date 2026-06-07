package paper

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

// MarketAwareAdapter is a paper trading adapter that fills orders at real
// market prices obtained from a marketdata.Feed. This allows simulated trades
// to use live Binance data while never placing real orders.
type MarketAwareAdapter struct {
	mu      sync.Mutex
	feed    marketdata.Feed
	balance float64 // USDT balance
	orders  []exchange.Order
	positions map[string]float64 // symbol -> quantity held
}

func NewMarketAwareAdapter(feed marketdata.Feed, initialUSDT float64) *MarketAwareAdapter {
	if initialUSDT <= 0 {
		initialUSDT = 10000
	}
	return &MarketAwareAdapter{
		feed:      feed,
		balance:   initialUSDT,
		positions: make(map[string]float64),
	}
}

func (a *MarketAwareAdapter) Name() string { return "paper-market-aware" }

func (a *MarketAwareAdapter) Capabilities() []exchange.Capability {
	return []exchange.Capability{
		exchange.CapabilitySpot,
		exchange.CapabilityPaper,
		exchange.CapabilityBalances,
		exchange.CapabilityOrders,
		exchange.CapabilityCandles,
		exchange.CapabilityTickers,
	}
}

func (a *MarketAwareAdapter) ListMarkets(_ context.Context) ([]exchange.Market, error) {
	return []exchange.Market{
		{Symbol: "BTCUSDT", BaseAsset: "BTC", QuoteAsset: "USDT", PriceScale: 2, QuantityScale: 6},
		{Symbol: "ETHUSDT", BaseAsset: "ETH", QuoteAsset: "USDT", PriceScale: 2, QuantityScale: 6},
		{Symbol: "SOLUSDT", BaseAsset: "SOL", QuoteAsset: "USDT", PriceScale: 2, QuantityScale: 3},
		{Symbol: "BNBUSDT", BaseAsset: "BNB", QuoteAsset: "USDT", PriceScale: 2, QuantityScale: 3},
	}, nil
}

func (a *MarketAwareAdapter) Balances(_ context.Context) ([]exchange.Balance, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := []exchange.Balance{
		{Asset: "USDT", Free: fmt.Sprintf("%.2f", a.balance), Locked: "0"},
	}
	for asset, qty := range a.positions {
		if qty > 0 {
			out = append(out, exchange.Balance{
				Asset: asset,
				Free:  fmt.Sprintf("%.8f", qty),
				Locked: "0",
			})
		}
	}
	return out, nil
}

func (a *MarketAwareAdapter) PlaceOrder(ctx context.Context, request exchange.OrderRequest) (exchange.Order, error) {
	if request.Symbol == "" {
		return exchange.Order{}, fmt.Errorf("symbol is required")
	}
	if request.Quantity == "" {
		return exchange.Order{}, fmt.Errorf("quantity is required")
	}

	// Get real market price for fill
	priceStr := request.Price
	if priceStr == "" {
		tick, err := a.feed.LatestTick(ctx, request.Symbol)
		if err != nil {
			return exchange.Order{}, fmt.Errorf("get market price for fill: %w", err)
		}
		priceStr = tick.Price
		if priceStr == "" {
			return exchange.Order{}, fmt.Errorf("empty market price for %s", request.Symbol)
		}
	}

	price := parseFloat(priceStr)
	qty := parseFloat(request.Quantity)
	notional := price * qty

	a.mu.Lock()
	defer a.mu.Unlock()

	// Simulate balance changes
	baseAsset := baseFromSymbol(request.Symbol)
	switch request.Side {
	case exchange.Buy:
		if notional > a.balance {
			return exchange.Order{}, fmt.Errorf("insufficient USDT balance: need %.2f, have %.2f", notional, a.balance)
		}
		a.balance -= notional
		a.positions[baseAsset] += qty
	case exchange.Sell:
		held := a.positions[baseAsset]
		if qty > held {
			return exchange.Order{}, fmt.Errorf("insufficient %s balance: need %.8f, have %.8f", baseAsset, qty, held)
		}
		a.positions[baseAsset] -= qty
		a.balance += notional
	}

	// Add 0.1% taker fee
	fee := notional * 0.001
	if request.Side == exchange.Buy {
		// Fee taken from received base asset (reduce position slightly)
		a.positions[baseAsset] -= qty * 0.001
	} else {
		// Fee taken from received USDT
		a.balance -= fee
	}

	order := exchange.Order{
		ID:       fmt.Sprintf("paper-%d", time.Now().UnixNano()),
		Symbol:   request.Symbol,
		Side:     request.Side,
		Type:     request.Type,
		Status:   "filled",
		Quantity: request.Quantity,
		Price:    priceStr,
	}
	a.orders = append(a.orders, order)
	return order, nil
}

// Balance returns current USDT balance (for position sizing).
func (a *MarketAwareAdapter) USDTBalance() float64 {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.balance
}

// SetFeed allows swapping the feed at runtime (for testing).
func (a *MarketAwareAdapter) SetFeed(feed marketdata.Feed) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.feed = feed
}

func baseFromSymbol(symbol string) string {
	// Simple extraction: BTCUSDT -> BTC, ETHUSDT -> ETH, etc.
	quote := "USDT"
	if len(symbol) > len(quote) && symbol[len(symbol)-len(quote):] == quote {
		return symbol[:len(symbol)-len(quote)]
	}
	return symbol
}

func parseFloat(s string) float64 {
	var whole, frac float64
	var fracDiv float64 = 1
	seenDot := false
	for _, ch := range s {
		switch {
		case ch == '.':
			seenDot = true
		case ch >= '0' && ch <= '9':
			digit := float64(ch - '0')
			if !seenDot {
				whole = whole*10 + digit
			} else {
				fracDiv *= 10
				frac += digit / fracDiv
			}
		}
	}
	return whole + frac
}
