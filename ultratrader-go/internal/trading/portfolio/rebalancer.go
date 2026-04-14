package portfolio

import (
	"context"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/logging"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

// RebalanceConfig defines the settings for portfolio rebalancing.
type RebalanceConfig struct {
	Enabled                bool
	TargetAllocations      map[string]float64 // e.g., "BTC": 50.0, "ETH": 30.0
	DriftThresholdPct      float64            // Minimum deviation % required to trigger a trade
	AvoidWashSales         bool
	RebalanceIntervalHours int
	TriggerMode            string // "time", "threshold", or "both"
}

// Adjustment describes a required trade to bring an asset back to its target allocation.
type Adjustment struct {
	Symbol     string
	CurrentPct float64
	TargetPct  float64
	DriftPct   float64
	DiffUSD    float64
	Action     exchange.OrderSide
}

// Rebalancer computes the orders necessary to bring a portfolio back to target allocations.
type Rebalancer struct {
	config RebalanceConfig
	logger *logging.Logger
}

// NewRebalancer creates a new Rebalancer with the given configuration.
func NewRebalancer(config RebalanceConfig, logger *logging.Logger) *Rebalancer {
	if logger == nil {
		logger, _ = logging.New(logging.Config{Stdout: true})
	}
	return &Rebalancer{
		config: config,
		logger: logger,
	}
}

// CheckDrift calculates the drift of the current portfolio from target allocations.
// It returns a list of adjustments needed for assets that have drifted beyond the threshold.
func (r *Rebalancer) CheckDrift(ctx context.Context, tracker *Tracker, feed marketdata.Feed) []Adjustment {
	if len(r.config.TargetAllocations) == 0 {
		return nil
	}

	totalValue := tracker.TotalMarketValue(ctx, feed)
	if totalValue <= 0 {
		return nil
	}

	targets := make(map[string]float64)
	var targetSum float64
	for k, v := range r.config.TargetAllocations {
		sym := strings.ToUpper(strings.TrimSpace(k))
		targets[sym] = v
		targetSum += v
	}

	// Normalize if targets sum > 100%
	if targetSum > 100.0 {
		r.logger.WithContext(ctx).Info("Target allocations sum > 100%. Normalizing.", map[string]any{"sum": targetSum})
		for k, v := range targets {
			targets[k] = (v / targetSum) * 100.0
		}
	}

	var adjustments []Adjustment
	threshold := r.config.DriftThresholdPct
	if threshold <= 0 {
		threshold = 5.0
	}

	positions := tracker.ValuedPositions(ctx, feed)
	currentMap := make(map[string]Position)
	for _, p := range positions {
		currentMap[strings.ToUpper(strings.TrimSpace(p.Symbol))] = p
	}

	for symbol, targetPct := range targets {
		var currentValue float64
		if p, ok := currentMap[symbol]; ok {
			currentValue = p.MarketValue
		}

		currentPct := (currentValue / totalValue) * 100.0
		drift := currentPct - targetPct
		driftAbs := drift
		if driftAbs < 0 {
			driftAbs = -driftAbs
		}

		if driftAbs >= threshold {
			targetValue := totalValue * (targetPct / 100.0)
			diffUSD := targetValue - currentValue

			action := exchange.Sell
			if diffUSD > 0 {
				action = exchange.Buy
			}

			diffAbs := diffUSD
			if diffAbs < 0 {
				diffAbs = -diffAbs
			}

			adjustments = append(adjustments, Adjustment{
				Symbol:     symbol,
				CurrentPct: currentPct,
				TargetPct:  targetPct,
				DriftPct:   drift,
				DiffUSD:    diffAbs,
				Action:     action,
			})
		}
	}

	return adjustments
}

// OrderHistoryItem is used to represent past executed orders for wash sale tracking since exchange.Order doesn't have a CreatedAt timestamp.
type OrderHistoryItem struct {
	Symbol    string
	Side      exchange.OrderSide
	Timestamp time.Time
}

// IsWashSale checks if selling the asset at the current price would constitute a wash sale.
// It returns true if the current price is less than the average cost AND there was a buy in the last 30 days.
func (r *Rebalancer) IsWashSale(symbol string, currentPrice, avgCost float64, recentOrders []OrderHistoryItem) bool {
	if currentPrice >= avgCost {
		return false // Selling at a profit is not a wash sale
	}

	now := time.Now()
	thirtyDays := 30 * 24 * time.Hour

	for _, o := range recentOrders {
		if strings.EqualFold(o.Symbol, symbol) && o.Side == exchange.Buy {
			if now.Sub(o.Timestamp) <= thirtyDays {
				return true
			}
		}
	}
	return false
}

// GenerateOrders calculates the required orders to rebalance the portfolio, factoring in wash sale rules.
func (r *Rebalancer) GenerateOrders(ctx context.Context, tracker *Tracker, feed marketdata.Feed, recentOrders []OrderHistoryItem) []exchange.OrderRequest {
	adjustments := r.CheckDrift(ctx, tracker, feed)
	var orders []exchange.OrderRequest

	positions := tracker.ValuedPositions(ctx, feed)
	posMap := make(map[string]Position)
	for _, p := range positions {
		posMap[strings.ToUpper(strings.TrimSpace(p.Symbol))] = p
	}

	for _, adj := range adjustments {
		if adj.Action == exchange.Sell && r.config.AvoidWashSales {
			p, ok := posMap[adj.Symbol]
			if ok && p.MarketPrice > 0 && p.AverageEntryPrice > 0 {
				if r.IsWashSale(adj.Symbol, p.MarketPrice, p.AverageEntryPrice, recentOrders) {
					r.logger.WithContext(ctx).Info("Wash sale prevented", map[string]any{
						"symbol":   adj.Symbol,
						"price":    p.MarketPrice,
						"avg_cost": p.AverageEntryPrice,
					})
					continue
				}
			}
		}

		// Since our tracker works with USD amounts we need to calculate the quantity to buy/sell
		// based on the current market price.
		var qty float64
		p, ok := posMap[adj.Symbol]
		if ok && p.MarketPrice > 0 {
			qty = adj.DiffUSD / p.MarketPrice
		} else {
			// Fallback: fetch price from feed directly if not in tracker
			tick, err := feed.LatestTick(ctx, adj.Symbol)
			if err == nil && tick.Price != "" {
				if parsedPrice, err := strconv.ParseFloat(tick.Price, 64); err == nil && parsedPrice > 0 {
					qty = adj.DiffUSD / parsedPrice
				}
			}
			if qty == 0 {
				r.logger.WithContext(ctx).Error("Cannot generate rebalance order: no market price available", map[string]any{"symbol": adj.Symbol})
				continue
			}
		}

		orders = append(orders, exchange.OrderRequest{
			Symbol:   adj.Symbol,
			Side:     adj.Action,
			Type:     exchange.MarketOrder,
			Quantity: formatFloat(qty),
		})
	}

	// Sort so SELLs happen before BUYs to free up capital
	sort.Slice(orders, func(i, j int) bool {
		return orders[i].Side == exchange.Sell && orders[j].Side == exchange.Buy
	})

	return orders
}

// IsRebalanceDue checks if a rebalance should be triggered based on time interval.
func (r *Rebalancer) IsRebalanceDue(lastRebalance time.Time) bool {
	if !r.config.Enabled {
		return false
	}
	mode := strings.ToLower(r.config.TriggerMode)
	if mode != "time" && mode != "both" {
		return false
	}
	interval := time.Duration(r.config.RebalanceIntervalHours) * time.Hour
	return time.Since(lastRebalance) >= interval
}

// formatFloat is a simple helper to convert float to string.
func formatFloat(f float64) string {
	return string(strconv.FormatFloat(f, 'f', -1, 64))
}
