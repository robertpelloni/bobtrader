package optimizer_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest/optimizer"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/demo"
)

func TestGridSearchCandles(t *testing.T) {
	now := time.Now()
	// Create an artificial price wave
	prices := []string{"10", "12", "14", "16", "14", "12", "10", "8", "6", "8", "10", "12", "14"}

	var candles []marketdata.Candle
	for i, p := range prices {
		candles = append(candles, marketdata.Candle{
			Symbol:    "BTCUSDT",
			Close:     p,
			Timestamp: now.Add(time.Duration(i) * time.Hour),
		})
	}
	history := backtest.NewMemoryCandleHistory(candles)

	// Define parameter ranges
	paramGrid := map[string][]interface{}{
		"fast_period": {2, 3},
		"slow_period": {4, 5},
	}

	// Define Strategy Builder
	builder := func(params optimizer.ParameterMap) (strategy.Strategy, error) {
		fast, ok := params["fast_period"].(int)
		if !ok {
			return nil, fmt.Errorf("invalid fast_period")
		}
		slow, ok := params["slow_period"].(int)
		if !ok {
			return nil, fmt.Errorf("invalid slow_period")
		}
		// Skip invalid combinations (fast must be < slow)
		if fast >= slow {
			return nil, fmt.Errorf("fast period must be less than slow period")
		}

		return demo.NewCandleSMACross("opt-test", "BTCUSDT", "1.0", fast, slow), nil
	}

	// Zero friction for pure mathematical comparison
	opts := backtest.EmulatorOptions{MakerFeeRate: 0, TakerFeeRate: 0, SlippageRate: 0}

	results, err := optimizer.GridSearchCandles(
		context.Background(),
		builder,
		history,
		1000.0,
		opts,
		paramGrid,
		optimizer.DefaultScorer, // Maximizing Realized PnL
		optimizer.DefaultOptimizationConfig(),
	)

	if err != nil {
		t.Fatalf("unexpected error during grid search: %v", err)
	}

	// We expect 4 permutations: (2,4), (2,5), (3,4), (3,5)
	if len(results) != 4 {
		t.Fatalf("expected 4 valid results, got %d", len(results))
	}

	// Results should be sorted by score descending
	for i := 0; i < len(results)-1; i++ {
		if results[i].Score < results[i+1].Score {
			t.Errorf("results not sorted correctly: index %d score %f < index %d score %f", i, results[i].Score, i+1, results[i+1].Score)
		}
	}

	// The best parameter set should be first.
	// We just ensure we got parameters back properly.
	bestParams := results[0].Params
	if _, ok := bestParams["fast_period"]; !ok {
		t.Errorf("best parameters missing fast_period")
	}
}

func TestGridSearchConcurrentStress(t *testing.T) {
	now := time.Now()
	// Larger artificial wave
	var candles []marketdata.Candle
	for i := 0; i < 100; i++ {
		price := 10.0 + float64(i%10) // Simple cyclical pattern
		candles = append(candles, marketdata.Candle{
			Symbol:    "BTCUSDT",
			Close:     fmt.Sprintf("%f", price),
			Timestamp: now.Add(time.Duration(i) * time.Hour),
		})
	}
	history := backtest.NewMemoryCandleHistory(candles)

	// 10x10 Grid = 100 permutations
	paramGrid := map[string][]interface{}{}
	var fasts []interface{}
	var slows []interface{}
	for i := 2; i <= 11; i++ {
		fasts = append(fasts, i)
		slows = append(slows, i+10) // ensures fast < slow always
	}
	paramGrid["fast_period"] = fasts
	paramGrid["slow_period"] = slows

	builder := func(params optimizer.ParameterMap) (strategy.Strategy, error) {
		fast := params["fast_period"].(int)
		slow := params["slow_period"].(int)
		return demo.NewCandleSMACross("stress", "BTCUSDT", "1.0", fast, slow), nil
	}

	opts := backtest.EmulatorOptions{MakerFeeRate: 0, TakerFeeRate: 0, SlippageRate: 0}

	// Force concurrent execution
	optConfig := optimizer.OptimizationConfig{MaxWorkers: 8}

	results, err := optimizer.GridSearchCandles(
		context.Background(),
		builder,
		history,
		1000.0,
		opts,
		paramGrid,
		optimizer.DefaultScorer,
		optConfig,
	)

	if err != nil {
		t.Fatalf("unexpected error during stress test: %v", err)
	}

	// Expect exactly 100 results from 100 permutations
	if len(results) != 100 {
		t.Fatalf("expected 100 results, got %d", len(results))
	}
}
