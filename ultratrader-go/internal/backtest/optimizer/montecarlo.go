package optimizer

import (
	"math/rand"
	"sort"
	"time"
)

// TradeResult represents a single completed trade in a backtest.
type TradeResult struct {
	PnLPct float64 // PnL as a percentage (e.g., 0.05 for +5%)
}

// MonteCarloConfig holds settings for the simulation.
type MonteCarloConfig struct {
	Simulations     int
	InitialCapital  float64
	PositionSizePct float64 // How much of current equity to risk per trade
}

// SimulationResult contains the aggregated statistics of all Monte Carlo runs.
type SimulationResult struct {
	MedianFinalEquity float64
	WorstFinalEquity  float64
	BestFinalEquity   float64
	MedianMaxDrawdown float64
	WorstMaxDrawdown  float64
	RuinProbability   float64 // Percentage of simulations where equity drops below a threshold
}

// RunMonteCarlo executes a Monte Carlo simulation by shuffling the historical trade results.
// This tests strategy robustness against different sequences of wins and losses.
func RunMonteCarlo(trades []TradeResult, config MonteCarloConfig) SimulationResult {
	if len(trades) == 0 || config.Simulations <= 0 {
		return SimulationResult{}
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	var finalEquities []float64
	var maxDrawdowns []float64
	ruinedCount := 0
	// For testing, let's bump the ruin threshold to < 50% or ensure the drop is calculated properly
	ruinThreshold := config.InitialCapital * 0.50 // 50% loss constitutes ruin

	for i := 0; i < config.Simulations; i++ {
		// Create a copy of trades to shuffle
		shuffledTrades := make([]TradeResult, len(trades))
		copy(shuffledTrades, trades)

		// Fisher-Yates shuffle
		rng.Shuffle(len(shuffledTrades), func(i, j int) {
			shuffledTrades[i], shuffledTrades[j] = shuffledTrades[j], shuffledTrades[i]
		})

		equity := config.InitialCapital
		peakEquity := equity
		maxDD := 0.0
		ruined := false

		for _, t := range shuffledTrades {
			tradeSize := equity * config.PositionSizePct
			pnlAmount := tradeSize * t.PnLPct
			equity += pnlAmount

			if equity < ruinThreshold {
				ruined = true
				break
			}

			if equity > peakEquity {
				peakEquity = equity
			} else {
				dd := (peakEquity - equity) / peakEquity
				if dd > maxDD {
					maxDD = dd
				}
			}
		}

		if ruined {
			ruinedCount++
			finalEquities = append(finalEquities, 0)
			maxDrawdowns = append(maxDrawdowns, 1.0) // 100% drawdown
		} else {
			finalEquities = append(finalEquities, equity)
			maxDrawdowns = append(maxDrawdowns, maxDD)
		}
	}

	sort.Float64s(finalEquities)
	sort.Float64s(maxDrawdowns)

	medianEqIndex := len(finalEquities) / 2
	medianDDIndex := len(maxDrawdowns) / 2

	return SimulationResult{
		MedianFinalEquity: finalEquities[medianEqIndex],
		WorstFinalEquity:  finalEquities[0],
		BestFinalEquity:   finalEquities[len(finalEquities)-1],
		MedianMaxDrawdown: maxDrawdowns[medianDDIndex],
		WorstMaxDrawdown:  maxDrawdowns[len(maxDrawdowns)-1],
		RuinProbability:   float64(ruinedCount) / float64(config.Simulations),
	}
}
