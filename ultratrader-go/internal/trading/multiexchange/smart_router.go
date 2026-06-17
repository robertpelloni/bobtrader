package multiexchange

import (
	"fmt"
	"sort"
)

// ExchangeManager interface represents a basic facade over exchange APIs.
type ExchangeManager interface {
	GetTicker(coin, exchange string) (Ticker, error)
	GetExchanges() []string
}

// Ticker represents market ticker data.
type Ticker struct {
	Bid float64
	Ask float64
}

// SmartRouter finds the best exchange to execute an order based on price and fees.
type SmartRouter struct {
	manager ExchangeManager
	fees    map[string]float64
}

// Route represents a calculated execution route.
type Route struct {
	Exchange   string
	Price      float64
	FeePct     float64
	TotalCost  float64 // Could be cost (buy) or revenue (sell)
	Slippage   float64 // Estimated slippage based on liquidity
	Liquidity  float64 // Available depth at this venue
}

// NewSmartRouter creates a new SmartRouter.
func NewSmartRouter(manager ExchangeManager, fees map[string]float64) *SmartRouter {
	if fees == nil {
		fees = make(map[string]float64)
	}
	return &SmartRouter{
		manager: manager,
		fees:    fees,
	}
}

// CompareRoutes returns a sorted list of possible routes for execution.
// side is either "buy" or "sell".
func (s *SmartRouter) CompareRoutes(coin, side string, quantity float64) ([]Route, error) {
	// Upgrade manager to V2 if possible to get order book depth
	managerV2, hasV2 := s.manager.(ExchangeManagerV2)

	exchanges := s.manager.GetExchanges()
	if len(exchanges) == 0 {
		return nil, fmt.Errorf("no exchanges available")
	}

	var routes []Route
	for _, ex := range exchanges {
		ticker, err := s.manager.GetTicker(coin, ex)
		if err != nil {
			continue // Skip exchange if ticker fetch fails
		}

		price := ticker.Ask
		if side == "sell" {
			price = ticker.Bid
		}

		feePct, exists := s.fees[ex]
		if !exists {
			feePct = 0.10 // Default fee
		}

		feeAmt := quantity * price * (feePct / 100.0)
		total := 0.0
		slippage := 0.0
		liquidity := 0.0

		// Enhanced Liquidity-Aware Routing (v3.1.0)
		if hasV2 {
			book, err := managerV2.GetOrderBook(coin, ex, 20)
			if err == nil {
				// Simple slippage estimation: average price of the depth we need
				targetQty := quantity
				weightedPriceSum := 0.0
				filledQty := 0.0

				levels := book.Asks
				if side == "sell" {
					levels = book.Bids
				}

				for _, level := range levels {
					take := level[1]
					if filledQty + take > targetQty {
						take = targetQty - filledQty
					}
					weightedPriceSum += level[0] * take
					filledQty += take
					liquidity += level[0] * level[1]
					if filledQty >= targetQty {
						break
					}
				}

				if filledQty > 0 {
					avgFillPrice := weightedPriceSum / filledQty
					slippage = (avgFillPrice - price) / price
					if side == "sell" {
						slippage = (price - avgFillPrice) / price
					}
					// Update price to estimated fill price
					price = avgFillPrice
				}
			}
		}

		if side == "buy" {
			total = (quantity * price) + feeAmt
		} else { // sell
			total = (quantity * price) - feeAmt
		}

		routes = append(routes, Route{
			Exchange:   ex,
			Price:      price,
			FeePct:     feePct,
			TotalCost:  total,
			Slippage:   slippage,
			Liquidity:  liquidity,
		})
	}

	// Sort routes: best for buy is lowest cost, best for sell is highest revenue
	sort.Slice(routes, func(i, j int) bool {
		if side == "buy" {
			return routes[i].TotalCost < routes[j].TotalCost
		}
		return routes[i].TotalCost > routes[j].TotalCost
	})

	return routes, nil
}
