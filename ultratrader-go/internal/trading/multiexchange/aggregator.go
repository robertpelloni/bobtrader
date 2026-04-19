package multiexchange

import (
	"fmt"
	"time"
)

// OrderBook represents the depth of market for an exchange.
type OrderBook struct {
	Bids [][2]float64 // [price, quantity]
	Asks [][2]float64
}

// ExchangeManagerV2 extends ExchangeManager with OrderBook capabilities.
type ExchangeManagerV2 interface {
	ExchangeManager
	GetOrderBook(symbol, exchange string, depth int) (OrderBook, error)
}

// ExecutionLeg represents a portion of a split order.
type ExecutionLeg struct {
	Exchange  string
	Side      string
	Coin      string
	Quantity  float64
	Price     float64
	Status    string
	Fee       float64
	Timestamp time.Time
}

// LiquidityAggregator splits large orders across exchanges to minimize market impact.
type LiquidityAggregator struct {
	manager      ExchangeManagerV2
	maxImpactPct float64
	fees         map[string]float64
}

// NewLiquidityAggregator creates a new LiquidityAggregator.
func NewLiquidityAggregator(manager ExchangeManagerV2, maxImpactPct float64, fees map[string]float64) *LiquidityAggregator {
	if fees == nil {
		fees = make(map[string]float64)
	}
	return &LiquidityAggregator{
		manager:      manager,
		maxImpactPct: maxImpactPct,
		fees:         fees,
	}
}

// GetLiquidityMap returns available liquidity (in USD/USDT) at each exchange.
func (l *LiquidityAggregator) GetLiquidityMap(coin, side string, depth int) map[string]float64 {
	liquidity := make(map[string]float64)
	exchanges := l.manager.GetExchanges()

	for _, ex := range exchanges {
		book, err := l.manager.GetOrderBook(coin, ex, depth)
		if err != nil {
			liquidity[ex] = 0.0
			continue
		}

		total := 0.0
		if side == "buy" {
			for _, level := range book.Asks {
				total += level[0] * level[1]
			}
		} else {
			for _, level := range book.Bids {
				total += level[0] * level[1]
			}
		}

		liquidity[ex] = total
	}

	return liquidity
}

// SplitOrder splits a large order proportionally across exchanges based on liquidity.
func (l *LiquidityAggregator) SplitOrder(coin, side string, totalQuantity float64) ([]ExecutionLeg, error) {
	liquidity := l.GetLiquidityMap(coin, side, 20)

	totalLiquidity := 0.0
	for _, liq := range liquidity {
		totalLiquidity += liq
	}

	if totalLiquidity == 0 {
		return nil, fmt.Errorf("no liquidity available for %s %s", coin, side)
	}

	var legs []ExecutionLeg

	for ex, liq := range liquidity {
		if liq <= 0 {
			continue
		}

		share := liq / totalLiquidity
		legQuantity := totalQuantity * share

		if legQuantity <= 0 {
			continue
		}

		ticker, err := l.manager.GetTicker(coin, ex)
		if err != nil {
			continue
		}

		price := ticker.Ask
		if side == "sell" {
			price = ticker.Bid
		}

		feePct, ok := l.fees[ex]
		if !ok {
			feePct = 0.10
		}

		fee := legQuantity * price * (feePct / 100.0)

		legs = append(legs, ExecutionLeg{
			Exchange:  ex,
			Side:      side,
			Coin:      coin,
			Quantity:  legQuantity, // Should ideally be rounded based on exchange rules
			Price:     price,
			Status:    "pending",
			Fee:       fee,
			Timestamp: time.Now(),
		})
	}

	return legs, nil
}
