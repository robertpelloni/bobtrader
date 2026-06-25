package multiexchange

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

// ArbitrageOrderRequest represents a coordinated multi-exchange trade.
type ArbitrageOrderRequest struct {
	Symbol      string
	Quantity    string
	BuyExchange string
	BuyPrice    string
	SellExchange string
	SellPrice    string
}

// ArbitrageExecutor handles the atomic execution of cross-exchange trades.
type ArbitrageExecutorV2 struct {
	registry *exchange.Registry
	mu       sync.Mutex
	results  []map[string]interface{}
}

// NewArbitrageExecutorV2 creates a new ArbitrageExecutorV2.
func NewArbitrageExecutorV2(registry *exchange.Registry) *ArbitrageExecutorV2 {
	return &ArbitrageExecutorV2{
		registry: registry,
		results:  make([]map[string]interface{}, 0),
	}
}

// ExecuteAtomic performs a coordinated buy and sell across two exchanges.
// In a true atomic system, we would use a more sophisticated transaction manager,
// but here we use concurrent execution with error handling.
func (a *ArbitrageExecutorV2) ExecuteAtomic(ctx context.Context, req ArbitrageOrderRequest) (map[string]interface{}, error) {
	buyAdapter, err := a.registry.Create(req.BuyExchange)
	if err != nil {
		return nil, fmt.Errorf("buy adapter not found: %w", err)
	}

	sellAdapter, err := a.registry.Create(req.SellExchange)
	if err != nil {
		return nil, fmt.Errorf("sell adapter not found: %w", err)
	}

	log.Printf("[ArbExecutor] Executing atomic arbitrage for %s: Buy on %s, Sell on %s",
		req.Symbol, req.BuyExchange, req.SellExchange)

	var buyOrder, sellOrder exchange.Order
	var buyErr, sellErr error
	var wg sync.WaitGroup

	wg.Add(2)

	// Execute Buy
	go func() {
		defer wg.Done()
		buyOrder, buyErr = buyAdapter.PlaceOrder(ctx, exchange.OrderRequest{
			Symbol:   req.Symbol,
			Side:     exchange.Buy,
			Type:     exchange.MarketOrder,
			Quantity: req.Quantity,
		})
	}()

	// Execute Sell
	go func() {
		defer wg.Done()
		sellOrder, sellErr = sellAdapter.PlaceOrder(ctx, exchange.OrderRequest{
			Symbol:   req.Symbol,
			Side:     exchange.Sell,
			Type:     exchange.MarketOrder,
			Quantity: req.Quantity,
		})
	}()

	wg.Wait()

	status := "success"
	if buyErr != nil || sellErr != nil {
		status = "partial_failure"
		if buyErr != nil && sellErr != nil {
			status = "failed"
		}
	}

	result := map[string]interface{}{
		"type":          "atomic_arbitrage",
		"symbol":        req.Symbol,
		"quantity":      req.Quantity,
		"buy_exchange":  req.BuyExchange,
		"buy_order_id":  buyOrder.ID,
		"buy_error":     fmt.Sprintf("%v", buyErr),
		"sell_exchange": req.SellExchange,
		"sell_order_id": sellOrder.ID,
		"sell_error":    fmt.Sprintf("%v", sellErr),
		"status":        status,
		"timestamp":     time.Now().Format(time.RFC3339),
	}

	a.mu.Lock()
	a.results = append(a.results, result)
	a.mu.Unlock()

	if status == "failed" {
		return result, fmt.Errorf("arbitrage failed: buy_err=%v, sell_err=%v", buyErr, sellErr)
	}

	return result, nil
}

// ChainLeg represents a single trade in a multi-hop sequence.
type ChainLeg struct {
	Exchange string
	Symbol   string
	Side     exchange.OrderSide
	Quantity string // If empty, uses output from previous leg
}

// ExecuteChain performs a sequence of trades where each trade depends on the completion of the previous.
func (a *ArbitrageExecutorV2) ExecuteChain(ctx context.Context, chain []ChainLeg) ([]exchange.Order, error) {
	log.Printf("[ArbExecutor] Executing multi-hop chain with %d legs", len(chain))

	var orders []exchange.Order
	var lastQuantity string

	for i, leg := range chain {
		adapter, err := a.registry.Create(leg.Exchange)
		if err != nil {
			return orders, fmt.Errorf("leg %d: adapter %s not found: %w", i, leg.Exchange, err)
		}

		qty := leg.Quantity
		if qty == "" {
			qty = lastQuantity
		}

		log.Printf("[ArbExecutor] Leg %d: %s %s on %s (Qty: %s)", i, leg.Side, leg.Symbol, leg.Exchange, qty)

		order, err := adapter.PlaceOrder(ctx, exchange.OrderRequest{
			Symbol:   leg.Symbol,
			Side:     leg.Side,
			Type:     exchange.MarketOrder,
			Quantity: qty,
		})

		if err != nil {
			return orders, fmt.Errorf("leg %d failed: %w", i, err)
		}

		orders = append(orders, order)
		lastQuantity = order.ExecutedQty // Pass resulting quantity to next leg

		// Optional: small delay to allow exchange state to propagate
		time.Sleep(100 * time.Millisecond)
	}

	return orders, nil
}

// GetResults returns history of executed arbitrage trades.
func (a *ArbitrageExecutorV2) GetResults() []map[string]interface{} {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.results
}
