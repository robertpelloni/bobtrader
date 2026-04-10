package rebalancer

import (
	"fmt"
	"math"
	"sort"
)

// Allocation defines a target weight for a symbol.
type Allocation struct {
	Symbol string
	Weight float64 // 0.0 to 1.0, all weights should sum to ~1.0
}

// Holding represents a current portfolio position.
type Holding struct {
	Symbol  string
	Quantity float64
	Value   float64 // Current market value
}

// RebalanceOrder represents a trade needed to rebalance.
type RebalanceOrder struct {
	Symbol    string  `json:"symbol"`
	Side      string  `json:"side"` // "buy" or "sell"
	Quantity  float64 `json:"quantity"`
	EstValue  float64 `json:"est_value"` // Estimated dollar value
	DriftPct  float64 `json:"drift_pct"` // Current drift from target
}

// RebalanceResult contains the rebalancing analysis.
type RebalanceResult struct {
	Orders      []RebalanceOrder `json:"orders"`
	TotalBuys   float64          `json:"total_buys"`
	TotalSells  float64          `json:"total_sells"`
	MaxDrift    float64          `json:"max_drift"`
	NeedsRebalance bool          `json:"needs_rebalance"`
}

// Rebalancer manages portfolio target allocations and drift detection.
type Rebalancer struct {
	targets   map[string]float64
	threshold float64 // Drift threshold to trigger rebalance (e.g., 0.05 = 5%)
}

// New creates a new rebalancer with target allocations and drift threshold.
func New(allocations []Allocation, driftThreshold float64) *Rebalancer {
	targets := make(map[string]float64)
	for _, a := range allocations {
		targets[a.Symbol] = a.Weight
	}
	if driftThreshold <= 0 {
		driftThreshold = 0.05 // Default 5%
	}
	return &Rebalancer{
		targets:   targets,
		threshold: driftThreshold,
	}
}

// Compute analyzes current holdings and generates rebalance orders.
func (r *Rebalancer) Compute(holdings []Holding) RebalanceResult {
	totalValue := 0.0
	for _, h := range holdings {
		totalValue += h.Value
	}
	if totalValue <= 0 {
		return RebalanceResult{}
	}

	// Calculate current weights
	currentWeights := make(map[string]float64)
	for _, h := range holdings {
		currentWeights[h.Symbol] = h.Value / totalValue
	}

	// Generate rebalance orders
	var orders []RebalanceOrder
	var totalBuys, totalSells float64
	var maxDrift float64

	// Process all tracked symbols (including those with zero holdings)
	allSymbols := make(map[string]bool)
	for s := range r.targets {
		allSymbols[s] = true
	}
	for _, h := range holdings {
		allSymbols[h.Symbol] = true
	}

	for symbol := range allSymbols {
		targetWeight := r.targets[symbol]
		currentWeight := currentWeights[symbol]
		drift := currentWeight - targetWeight

		if math.Abs(drift) > maxDrift {
			maxDrift = math.Abs(drift)
		}

		// Only generate orders if drift exceeds threshold
		if math.Abs(drift) <= r.threshold {
			continue
		}

		driftValue := drift * totalValue

		if drift > 0 {
			// Overweight — need to sell
			// Find the holding to get quantity
			var qty float64
			for _, h := range holdings {
				if h.Symbol == symbol && h.Value > 0 {
					qty = driftValue / (h.Value / h.Quantity)
					break
				}
			}
			orders = append(orders, RebalanceOrder{
				Symbol:   symbol,
				Side:     "sell",
				Quantity: qty,
				EstValue: driftValue,
				DriftPct: drift,
			})
			totalSells += driftValue
		} else {
			// Underweight — need to buy
			// Estimate quantity from current price if available
			var price float64
			for _, h := range holdings {
				if h.Symbol == symbol && h.Quantity > 0 {
					price = h.Value / h.Quantity
					break
				}
			}
			var qty float64
			if price > 0 {
				qty = -driftValue / price
			}

			orders = append(orders, RebalanceOrder{
				Symbol:   symbol,
				Side:     "buy",
				Quantity: qty,
				EstValue: -driftValue,
				DriftPct: drift,
			})
			totalBuys += -driftValue
		}
	}

	// Sort: sells first (free up capital), then buys
	sort.Slice(orders, func(i, j int) bool {
		if orders[i].Side != orders[j].Side {
			return orders[i].Side == "sell"
		}
		return math.Abs(orders[i].DriftPct) > math.Abs(orders[j].DriftPct)
	})

	return RebalanceResult{
		Orders:          orders,
		TotalBuys:      totalBuys,
		TotalSells:     totalSells,
		MaxDrift:       maxDrift,
		NeedsRebalance: maxDrift > r.threshold,
	}
}

// Drift returns the current drift for each tracked symbol.
func (r *Rebalancer) Drift(holdings []Holding) map[string]float64 {
	totalValue := 0.0
	for _, h := range holdings {
		totalValue += h.Value
	}
	if totalValue <= 0 {
		return nil
	}

	drifts := make(map[string]float64)
	for _, h := range holdings {
		currentWeight := h.Value / totalValue
		targetWeight := r.targets[h.Symbol]
		drifts[h.Symbol] = currentWeight - targetWeight
	}

	// Include symbols with target but no holding
	for symbol, target := range r.targets {
		if _, exists := drifts[symbol]; !exists {
			drifts[symbol] = -target // Fully underweight
		}
	}

	return drifts
}

// Summary returns a human-readable rebalance summary.
func (rr *RebalanceResult) Summary() string {
	if !rr.NeedsRebalance {
		return "Portfolio is balanced (no rebalance needed)"
	}
	return fmt.Sprintf("Rebalance needed: %d orders, $%.2f buys, $%.2f sells, max drift %.1f%%",
		len(rr.Orders), rr.TotalBuys, rr.TotalSells, rr.MaxDrift*100)
}
