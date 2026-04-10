package multiexchange

import (
	"math"
	"time"
)

// ArbitrageOpportunity represents a detected cross-exchange spread.
type ArbitrageOpportunity struct {
	Coin            string
	BuyExchange     string
	BuyPrice        float64
	SellExchange    string
	SellPrice       float64
	SpreadPct       float64
	EstimatedProfit float64
	MaxQuantity     float64
	Timestamp       time.Time
}

// ArbitrageExecutor scans and tracks arbitrage opportunities.
type ArbitrageExecutor struct {
	manager       ExchangeManager
	fees          map[string]float64
	minSpreadPct  float64
	executed      []map[string]interface{}
}

// NewArbitrageExecutor creates a new ArbitrageExecutor.
func NewArbitrageExecutor(manager ExchangeManager, fees map[string]float64, minSpreadPct float64) *ArbitrageExecutor {
	if fees == nil {
		fees = make(map[string]float64)
	}
	return &ArbitrageExecutor{
		manager:      manager,
		fees:         fees,
		minSpreadPct: minSpreadPct,
		executed:     make([]map[string]interface{}, 0),
	}
}

// Scan looks for arbitrage opportunities for a list of coins.
func (a *ArbitrageExecutor) Scan(coins []string) []ArbitrageOpportunity {
	var opportunities []ArbitrageOpportunity
	exchanges := a.manager.GetExchanges()

	for _, coin := range coins {
		var bestBuyEx, bestSellEx string
		bestBuyPrice := math.MaxFloat64
		bestSellPrice := 0.0

		for _, ex := range exchanges {
			ticker, err := a.manager.GetTicker(coin, ex)
			if err != nil {
				continue
			}

			if ticker.Ask > 0 && ticker.Ask < bestBuyPrice {
				bestBuyPrice = ticker.Ask
				bestBuyEx = ex
			}

			if ticker.Bid > bestSellPrice {
				bestSellPrice = ticker.Bid
				bestSellEx = ex
			}
		}

		if bestBuyEx == "" || bestSellEx == "" || bestBuyEx == bestSellEx {
			continue
		}

		spreadPct := ((bestSellPrice - bestBuyPrice) / bestBuyPrice) * 100.0

		if spreadPct > a.minSpreadPct {
			buyFeePct := a.getFee(bestBuyEx)
			sellFeePct := a.getFee(bestSellEx)

			netSpreadPct := spreadPct - buyFeePct - sellFeePct

			if netSpreadPct > 0 {
				maxQty := 100.0 / bestBuyPrice // Example: using $100 as base
				buyCost := maxQty * bestBuyPrice * (1 + (buyFeePct / 100.0))
				sellRev := maxQty * bestSellPrice * (1 - (sellFeePct / 100.0))
				profit := sellRev - buyCost

				opportunities = append(opportunities, ArbitrageOpportunity{
					Coin:            coin,
					BuyExchange:     bestBuyEx,
					BuyPrice:        bestBuyPrice,
					SellExchange:    bestSellEx,
					SellPrice:       bestSellPrice,
					SpreadPct:       netSpreadPct,
					EstimatedProfit: profit,
					MaxQuantity:     maxQty,
					Timestamp:       time.Now(),
				})
			}
		}
	}

	return opportunities
}

// ExecuteOpportunity simulates or executes the arbitrage trade.
func (a *ArbitrageExecutor) ExecuteOpportunity(opp ArbitrageOpportunity) map[string]interface{} {
	buyFee := a.getFee(opp.BuyExchange) / 100.0
	sellFee := a.getFee(opp.SellExchange) / 100.0

	buyCost := opp.MaxQuantity * opp.BuyPrice * (1 + buyFee)
	sellRev := opp.MaxQuantity * opp.SellPrice * (1 - sellFee)
	profit := sellRev - buyCost

	result := map[string]interface{}{
		"type":          "arbitrage",
		"coin":          opp.Coin,
		"buy_exchange":  opp.BuyExchange,
		"buy_price":     opp.BuyPrice,
		"sell_exchange": opp.SellExchange,
		"sell_price":    opp.SellPrice,
		"quantity":      opp.MaxQuantity,
		"buy_cost":      buyCost,
		"sell_revenue":  sellRev,
		"net_profit":    profit,
		"spread_pct":    opp.SpreadPct,
		"status":        "simulated",
		"timestamp":     time.Now().Format(time.RFC3339),
	}

	a.executed = append(a.executed, result)
	return result
}

// ScanAndExecute scans and executes any profitable opportunities.
func (a *ArbitrageExecutor) ScanAndExecute(coins []string) []map[string]interface{} {
	opps := a.Scan(coins)
	var results []map[string]interface{}

	for _, opp := range opps {
		if opp.EstimatedProfit > 1.0 { // Min $1 profit
			res := a.ExecuteOpportunity(opp)
			results = append(results, res)
		}
	}

	return results
}

func (a *ArbitrageExecutor) getFee(ex string) float64 {
	if fee, ok := a.fees[ex]; ok {
		return fee
	}
	return 0.10
}
