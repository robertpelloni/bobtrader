package app

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest/optimizer"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	strategydemo "github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/demo"
)

// mockEvaluator bridges the Optimizer with the Backtest Engine.
type mockEvaluator struct {
	symbol  string
	history backtest.CandleHistoryProvider
}

func (e *mockEvaluator) Evaluate(ctx context.Context, params optimizer.ParameterSet, start, end time.Time) (float64, error) {
	fast := int(params["fast"])
	slow := int(params["slow"])
	trend := int(params["trend"])

	strat := strategydemo.NewDoubleEMATrendStrategy("opt-acct", e.symbol, "1.0", fast, slow, trend)
	engine := backtest.NewEngine(strat, 10000.0)

	result, err := engine.RunCandles(ctx, e.history)
	if err != nil {
		return 0, err
	}

	return result.RealizedPnL, nil
}

func TestOptimizerRun(t *testing.T) {
	ctx := context.Background()
	symbol := "BTCUSDT"

	// 1. Setup Synthetic Data with volatility to trigger signals
	now := time.Now().Round(time.Hour)
	var candles []marketdata.Candle
	for i := 0; i < 500; i++ {
		price := 60000.0
		if i < 300 {
			price += float64(i) * 10.0 // Uptrend
		} else if i < 400 {
			price += 3000.0 - float64(i-300)*20.0 // Downtrend
		} else {
			price += 1000.0 + float64(i-400)*30.0 // Second Uptrend
		}

		candles = append(candles, marketdata.Candle{
			Symbol:    symbol,
			Close:     fmt.Sprintf("%.2f", price),
			Timestamp: now.Add(time.Duration(i) * time.Hour),
		})
	}
	history := backtest.NewMemoryCandleHistory(candles)

	// 2. Define Parameter Grid
	ranges := map[string]optimizer.ParamRange{
		"fast":  {Min: 5, Max: 15, Step: 5},
		"slow":  {Min: 20, Max: 30, Step: 10},
		"trend": {Min: 50, Max: 100, Step: 50},
	}
	grid, _ := optimizer.GenerateGrid(ranges)
	t.Logf("Generated grid with %d permutations", len(grid))

	// 3. Run Parallel Optimization
	evaluator := &mockEvaluator{symbol: symbol, history: history}
	opt := optimizer.NewGridSearchOptimizer(evaluator)

	results, err := opt.Optimize(ctx, grid, time.Time{}, time.Time{}, 4)
	if err != nil {
		t.Fatalf("Optimization failed: %v", err)
	}

	// 4. Validate Top Result
	best := results[0]
	t.Logf("Best Parameters: %v Score=%.2f", best.Parameters, best.Score)

	if best.Score != 0 {
		t.Logf("Successfully executed trades and optimized parameters. PnL: %.2f", best.Score)
	} else {
		t.Log("Warning: Best score is 0. Strategies did not execute any trades.")
	}
}
