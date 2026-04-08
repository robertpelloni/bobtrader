package optimizer

import (
	"math"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest"
)

// SharpeScorer returns a scoring function that computes a Sharpe-like ratio.
// Higher score = better risk-adjusted return.
func SharpeScorer() ScoringFunction {
	return func(result backtest.Result) float64 {
		if result.TotalTrades == 0 {
			return 0
		}
		avgReturn := result.RealizedPnL / float64(result.TotalTrades)
		if result.RealizedPnL <= 0 {
			return result.RealizedPnL
		}
		sharpe := avgReturn / (math.Sqrt(math.Abs(avgReturn)) + 1e-10)
		return sharpe
	}
}

// ProfitFactorScorer rewards strategies with high returns relative to capital.
func ProfitFactorScorer() ScoringFunction {
	return func(result backtest.Result) float64 {
		if result.TotalTrades == 0 {
			return 0
		}
		if result.RealizedPnL > 0 {
			return 1 + result.RealizedPnL/1000
		}
		return result.RealizedPnL / 1000
	}
}

// WinRateScorer scores based on proportion of winning trades.
func WinRateScorer() ScoringFunction {
	return func(result backtest.Result) float64 {
		if result.TotalTrades == 0 || len(result.Orders) == 0 {
			return 0
		}

		wins := 0
		losses := 0
		var buyPrice float64

		for _, order := range result.Orders {
			if order.Side == "buy" || order.Side == "Buy" {
				buyPrice = simpleParseFloat(order.Price)
			} else if order.Side == "sell" || order.Side == "Sell" {
				sellPrice := simpleParseFloat(order.Price)
				if buyPrice > 0 {
					if sellPrice > buyPrice {
						wins++
					} else {
						losses++
					}
					buyPrice = 0
				}
			}
		}

		total := wins + losses
		if total == 0 {
			return 0
		}
		return float64(wins) / float64(total)
	}
}

// CompositeScorer combines multiple scoring functions with weights.
type CompositeScorer struct {
	scorers []weightedScorer
}

type weightedScorer struct {
	scorer ScoringFunction
	weight float64
}

// NewCompositeScorer creates a new composite scorer.
func NewCompositeScorer() *CompositeScorer {
	return &CompositeScorer{}
}

// Add adds a scoring function with a weight.
func (c *CompositeScorer) Add(scorer ScoringFunction, weight float64) *CompositeScorer {
	c.scorers = append(c.scorers, weightedScorer{scorer: scorer, weight: weight})
	return c
}

// Score returns the weighted composite score.
func (c *CompositeScorer) Score(result backtest.Result) float64 {
	var total float64
	for _, ws := range c.scorers {
		total += ws.scorer(result) * ws.weight
	}
	return total
}

// Scorer returns the composite as a ScoringFunction.
func (c *CompositeScorer) Scorer() ScoringFunction {
	return c.Score
}

// simpleParseFloat parses a string to float64 without importing strconv.
func simpleParseFloat(s string) float64 {
	var result, frac float64
	var fracDiv float64 = 1
	var neg bool
	var hasDigit bool
	var pastDot bool
	i := 0
	if i < len(s) && s[i] == '-' {
		neg = true
		i++
	}
	for ; i < len(s); i++ {
		if s[i] == '.' {
			pastDot = true
			continue
		}
		if s[i] >= '0' && s[i] <= '9' {
			hasDigit = true
			if !pastDot {
				result = result*10 + float64(s[i]-'0')
			} else {
				frac = frac*10 + float64(s[i]-'0')
				fracDiv *= 10
			}
		}
	}
	if !hasDigit {
		return 0
	}
	result += frac / fracDiv
	if neg {
		result = -result
	}
	return result
}
