package optimizer_test

import (
	"context"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest/optimizer"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// mockStrategy implements strategy.CandleStrategy and strategy.Strategy
type mockStrategy struct {
	p1 float64
}

func (s *mockStrategy) Name() string { return "mock" }
func (s *mockStrategy) OnTick(_ context.Context) ([]strategy.Signal, error) { return nil, nil }
func (s *mockStrategy) OnMarketCandle(_ context.Context, candle marketdata.Candle) ([]strategy.Signal, error) {
	// Dummy logic: buy if price > p1
	price := 0.0
	// For simplicity, just read the first char of Close if not empty
	if len(candle.Close) > 0 {
		price = float64(candle.Close[0] - '0')
	}

	if price > s.p1 {
		return []strategy.Signal{{Action: "buy", Quantity: "1", Symbol: "BTCUSDT"}}, nil
	}
	return nil, nil
}

func mockBuilder(params optimizer.ParameterMap) (strategy.Strategy, error) {
	p1, ok := params["p1"].(float64)
	if !ok {
		return nil, nil // Return error in real app
	}
	return &mockStrategy{p1: p1}, nil
}

func TestBacktestEvaluator(t *testing.T) {
	start := time.Now()
	end := start.Add(time.Hour)

	candles := []marketdata.Candle{
		{Timestamp: start.Add(10 * time.Minute), Close: "50000"},
		{Timestamp: start.Add(20 * time.Minute), Close: "51000"},
	}

	historyBuilder := func(s, e time.Time) (backtest.CandleHistoryProvider, error) {
		sliced := optimizer.SliceHistory(candles, s, e)
		return optimizer.NewSlicedCandleHistory(sliced), nil
	}

	eval := optimizer.NewBacktestEvaluator(
		historyBuilder,
		mockBuilder,
		optimizer.DefaultScorer,
		10000,
		backtest.DefaultEmulatorOptions(),
	)

	params := optimizer.ParameterMap{"p1": 4.0} // Will trigger buy on both (5 > 4)

	score, err := eval.Evaluate(context.Background(), params, start, end)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Because we just buy and never sell, and price doesn't change after buy in this simple mock
	// Realized PnL is 0.
	if score != 0 {
		t.Errorf("expected score 0 (no sells), got %f", score)
	}
}

func TestSliceHistory(t *testing.T) {
	start := time.Now()
	candles := []marketdata.Candle{
		{Timestamp: start.Add(1 * time.Minute)},
		{Timestamp: start.Add(2 * time.Minute)},
		{Timestamp: start.Add(3 * time.Minute)},
	}

	sub := optimizer.SliceHistory(candles, start, start.Add(2*time.Minute+time.Second))
	if len(sub) != 2 {
		t.Errorf("expected 2 candles, got %d", len(sub))
	}
}
