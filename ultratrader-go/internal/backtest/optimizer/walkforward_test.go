package optimizer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/backtest"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/indicator"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
)

func TestWalkForwardCandles(t *testing.T) {
	// Generate 200 candles with a sine wave pattern
	candles := generateSineCandles(200, "BTCUSDT", "1m")

	builder := func(params ParameterMap) (strategy.Strategy, error) {
		fast := params["fast_sma"].(int)
		slow := params["slow_sma"].(int)
		return &testSMAStrategy{
			accountID: "test",
			symbol:    "BTCUSDT",
			quantity:  "0.01",
			fastSMA:   indicator.NewSMA(fast),
			slowSMA:   indicator.NewSMA(slow),
		}, nil
	}

	paramGrid := map[string][]interface{}{
		"fast_sma": {3, 5},
		"slow_sma": {10, 15},
	}

	wfConfig := WalkForwardConfig{
		WindowCandles:      50,
		StepCandles:        20,
		MinTrades:          0,
		OptimizationConfig: OptimizationConfig{MaxWorkers: 2},
	}

	result, err := WalkForwardCandles(
		context.Background(),
		builder,
		candles,
		10000,
		backtest.DefaultEmulatorOptions(),
		paramGrid,
		nil, // use default scorer
		wfConfig,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TotalSteps == 0 {
		t.Errorf("expected at least one walk-forward step")
	}

	if len(result.Steps) == 0 {
		t.Fatalf("expected at least one step result")
	}

	// Verify that each step has valid window boundaries
	for i, step := range result.Steps {
		if step.Window.TrainStart >= step.Window.TrainEnd {
			t.Errorf("step %d: train start (%d) should be < train end (%d)", i, step.Window.TrainStart, step.Window.TrainEnd)
		}
		if step.Window.ValidStart >= step.Window.ValidEnd {
			t.Errorf("step %d: valid start (%d) should be < valid end (%d)", i, step.Window.ValidStart, step.Window.ValidEnd)
		}
		if step.BestParams == nil {
			t.Errorf("step %d: expected best params", i)
		}
	}

	t.Logf("Walk-forward completed: %d steps, avg validation score: %.2f, best step: %d",
		result.TotalSteps, result.AvgValScore, result.BestStep)
}

func TestWalkForwardOverfitting(t *testing.T) {
	candles := generateSineCandles(200, "BTCUSDT", "1m")

	builder := func(params ParameterMap) (strategy.Strategy, error) {
		fast := params["fast_sma"].(int)
		slow := params["slow_sma"].(int)
		return &testSMAStrategy{
			accountID: "test",
			symbol:    "BTCUSDT",
			quantity:  "0.01",
			fastSMA:   indicator.NewSMA(fast),
			slowSMA:   indicator.NewSMA(slow),
		}, nil
	}

	paramGrid := map[string][]interface{}{
		"fast_sma": {3, 5},
		"slow_sma": {10, 15},
	}

	wfConfig := WalkForwardConfig{
		WindowCandles:      50,
		StepCandles:        20,
		MinTrades:          0,
		OptimizationConfig: OptimizationConfig{MaxWorkers: 2},
	}

	result, _ := WalkForwardCandles(
		context.Background(), builder, candles, 10000,
		backtest.DefaultEmulatorOptions(), paramGrid, nil, wfConfig,
	)

	analysis := AnalyzeOverfitting(result)
	if len(analysis) == 0 {
		t.Errorf("expected overfitting analysis results")
	}

	for _, a := range analysis {
		t.Logf("Window %d: Train=%.2f Valid=%.2f Overfit=%.2f",
			a.WindowIndex, a.TrainScore, a.ValidScore, a.Overfit)
	}
}

func TestWalkForwardInsufficientData(t *testing.T) {
	candles := generateSineCandles(10, "BTCUSDT", "1m") // Too few candles

	builder := func(params ParameterMap) (strategy.Strategy, error) {
		return &testSMAStrategy{}, nil
	}

	wfConfig := WalkForwardConfig{
		WindowCandles:      100,
		StepCandles:        20,
		OptimizationConfig: OptimizationConfig{MaxWorkers: 1},
	}

	_, err := WalkForwardCandles(
		context.Background(), builder, candles, 10000,
		backtest.DefaultEmulatorOptions(), map[string][]interface{}{
			"p": {1},
		}, nil, wfConfig,
	)

	if err == nil {
		t.Errorf("expected error for insufficient data")
	}
}

func TestGenerateWindows(t *testing.T) {
	windows := generateWindows(100, 30, 10)
	if len(windows) == 0 {
		t.Fatalf("expected windows")
	}

	// Last valid end should not exceed total
	lastWindow := windows[len(windows)-1]
	if lastWindow.ValidEnd > 100 {
		t.Errorf("valid end (%d) should not exceed total (100)", lastWindow.ValidEnd)
	}

	// Each window should advance by stepSize
	for i := 1; i < len(windows); i++ {
		if windows[i].TrainStart != windows[i-1].TrainStart+10 {
			t.Errorf("expected windows to advance by stepSize")
		}
	}

	t.Logf("Generated %d windows for 100 candles (window=30, step=10)", len(windows))
}

// Helper: test SMA strategy that implements CandleStrategy
type testSMAStrategy struct {
	accountID string
	symbol    string
	quantity  string
	fastSMA   *indicator.SMA
	slowSMA   *indicator.SMA
	prevFast  float64
	prevSlow  float64
	warmup    int
}

func (s *testSMAStrategy) Name() string { return "testSMAStrategy" }

func (s *testSMAStrategy) OnTick(ctx context.Context) ([]strategy.Signal, error) {
	return nil, nil
}

func (s *testSMAStrategy) OnMarketCandle(ctx context.Context, candle marketdata.Candle) ([]strategy.Signal, error) {
	closePrice := utils.ParseFloat(candle.Close)
	s.warmup++

	fastVal := s.fastSMA.Update(closePrice)
	slowVal := s.slowSMA.Update(closePrice)

	var signals []strategy.Signal

	if s.warmup > 2 {
		if s.prevFast <= s.prevSlow && fastVal > slowVal {
			signals = append(signals, strategy.Signal{
				AccountID: s.accountID, Symbol: s.symbol,
				Action: "buy", Quantity: s.quantity, OrderType: "market",
			})
		}
		if s.prevFast >= s.prevSlow && fastVal < slowVal {
			signals = append(signals, strategy.Signal{
				AccountID: s.accountID, Symbol: s.symbol,
				Action: "sell", Quantity: s.quantity, OrderType: "market",
			})
		}
	}

	s.prevFast = fastVal
	s.prevSlow = slowVal
	return signals, nil
}

func generateSineCandles(count int, symbol, interval string) []marketdata.Candle {
	candles := make([]marketdata.Candle, count)
	baseTime := time.Now().Add(-time.Duration(count) * time.Minute)
	for i := 0; i < count; i++ {
		price := 100.0 + 20.0*sinFloat(float64(i)*0.1)
		p := fmt.Sprintf("%.2f", price)
		candles[i] = marketdata.Candle{
			Symbol:    symbol,
			Interval:  interval,
			Open:      p,
			High:      fmt.Sprintf("%.2f", price+1),
			Low:       fmt.Sprintf("%.2f", price-1),
			Close:     p,
			Volume:    "100",
			Timestamp: baseTime.Add(time.Duration(i) * time.Minute),
		}
	}
	return candles
}

func sinFloat(x float64) float64 {
	// Simple Taylor series approximation for sin
	result := x
	term := x
	for i := 1; i <= 10; i++ {
		term *= -x * x / float64(2*i) / float64(2*i+1)
		result += term
	}
	return result
}
