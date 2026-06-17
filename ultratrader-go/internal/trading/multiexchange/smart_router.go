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
	Exchange  string
	Price     float64
	FeePct    float64
	TotalCost float64 // Could be cost (buy) or revenue (sell)
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

		if side == "buy" {
			total = (quantity * price) + feeAmt
		} else { // sell
			total = (quantity * price) - feeAmt
		}

		routes = append(routes, Route{
			Exchange:  ex,
			Price:     price,
			FeePct:    feePct,
			TotalCost: total,
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
