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
