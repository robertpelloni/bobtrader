package optimizer_test

import (
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest/optimizer"
)

func TestRunMonteCarlo_Basic(t *testing.T) {
	// 5 wins of 10%, 5 losses of 5%
	trades := []optimizer.TradeResult{
		{PnLPct: 0.10}, {PnLPct: 0.10}, {PnLPct: 0.10}, {PnLPct: 0.10}, {PnLPct: 0.10},
		{PnLPct: -0.05}, {PnLPct: -0.05}, {PnLPct: -0.05}, {PnLPct: -0.05}, {PnLPct: -0.05},
	}

	config := optimizer.MonteCarloConfig{
		Simulations:     1000,
		InitialCapital:  10000.0,
		PositionSizePct: 1.0, // Risk full account size on every trade (compound entirely)
	}

	result := optimizer.RunMonteCarlo(trades, config)

	// Since we are compounding, the order matters a bit, but overall equity should be highly positive
	if result.MedianFinalEquity <= 10000 {
		t.Errorf("Expected median final equity to be > 10000, got %f", result.MedianFinalEquity)
	}

	// Max drawdown depends on ordering. If all 5 losses happen first, max DD is around 22.6%
	if result.WorstMaxDrawdown < 0.2 || result.WorstMaxDrawdown > 0.3 {
		t.Errorf("Expected worst max drawdown around 0.22, got %f", result.WorstMaxDrawdown)
	}

	if result.RuinProbability > 0 {
		t.Errorf("Expected 0 ruin probability for this sequence, got %f", result.RuinProbability)
	}
}

func TestRunMonteCarlo_Ruin(t *testing.T) {
	// High risk of ruin: 1 win of 50%, 4 losses of 30%
	trades := []optimizer.TradeResult{
		{PnLPct: 0.50},
		{PnLPct: -0.30}, {PnLPct: -0.30}, {PnLPct: -0.30}, {PnLPct: -0.30},
	}

	config := optimizer.MonteCarloConfig{
		Simulations:     100,
		InitialCapital:  10000.0,
		PositionSizePct: 1.0,
	}

	result := optimizer.RunMonteCarlo(trades, config)

	if result.RuinProbability == 0 {
		t.Errorf("Expected non-zero ruin probability")
	}

	if result.WorstMaxDrawdown != 1.0 {
		t.Errorf("Expected worst max drawdown to be 1.0 (ruin), got %f", result.WorstMaxDrawdown)
	}
}
