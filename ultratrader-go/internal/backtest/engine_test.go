package backtest

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

type mockTickStrategy struct{}

func (mockTickStrategy) Name() string                                        { return "mock" }
func (mockTickStrategy) OnTick(_ context.Context) ([]strategy.Signal, error) { return nil, nil }
func (mockTickStrategy) OnMarketTick(_ context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	if tick.Price == "90.00" {
		return []strategy.Signal{{AccountID: "test", Symbol: "BTCUSDT", Action: "buy", Quantity: "1.0"}}, nil
	} else if tick.Price == "110.00" {
		return []strategy.Signal{{AccountID: "test", Symbol: "BTCUSDT", Action: "sell", Quantity: "1.0"}}, nil
	}
	return nil, nil
}

type mockCandleStrategy struct{}

func (mockCandleStrategy) Name() string                                        { return "mock-candle" }
func (mockCandleStrategy) OnTick(_ context.Context) ([]strategy.Signal, error) { return nil, nil }
func (mockCandleStrategy) OnMarketCandle(_ context.Context, candle marketdata.Candle) ([]strategy.Signal, error) {
	if candle.Close == "90.00" {
		return []strategy.Signal{{AccountID: "test", Symbol: "BTCUSDT", Action: "buy", Quantity: "1.0"}}, nil
	} else if candle.Close == "110.00" {
		return []strategy.Signal{{AccountID: "test", Symbol: "BTCUSDT", Action: "sell", Quantity: "1.0"}}, nil
	}
	return nil, nil
}

func TestEngineRunTicks(t *testing.T) {
	now := time.Now()
	history := NewMemoryHistory([]marketdata.Tick{
		{Symbol: "BTCUSDT", Price: "90.00", Timestamp: now},
		{Symbol: "BTCUSDT", Price: "110.00", Timestamp: now.Add(time.Second)},
	})

	strat := mockTickStrategy{}
	// Use zero friction to test perfect PnL logic
	opts := EmulatorOptions{MakerFeeRate: 0, TakerFeeRate: 0, SlippageRate: 0}
	engine := NewEngineWithOptions(strat, 1000.0, opts)

	result, err := engine.RunTicks(context.Background(), history)
	if err != nil {
		t.Fatalf("RunTicks returned error: %v", err)
	}

	if result.TotalTrades != 2 {
		t.Fatalf("Expected 2 trades, got %d", result.TotalTrades)
	}

	if math.Abs(result.RealizedPnL-20.0) > 0.001 {
		t.Fatalf("Expected $20 realized PnL, got %f", result.RealizedPnL)
	}

	if len(result.Orders) != 2 {
		t.Fatalf("Expected 2 orders, got %d", len(result.Orders))
	}
	if result.Orders[0].Side != "buy" || result.Orders[1].Side != "sell" {
		t.Fatalf("Unexpected order sides: %v, %v", result.Orders[0].Side, result.Orders[1].Side)
	}
}

func TestEngineRunCandles(t *testing.T) {
	now := time.Now()
	history := NewMemoryCandleHistory([]marketdata.Candle{
		{Symbol: "BTCUSDT", Close: "90.00", Timestamp: now},
		{Symbol: "BTCUSDT", Close: "110.00", Timestamp: now.Add(time.Hour)},
	})

	strat := mockCandleStrategy{}
	// Use zero friction
	opts := EmulatorOptions{MakerFeeRate: 0, TakerFeeRate: 0, SlippageRate: 0}
	engine := NewEngineWithOptions(strat, 1000.0, opts)

	result, err := engine.RunCandles(context.Background(), history)
	if err != nil {
		t.Fatalf("RunCandles returned error: %v", err)
	}

	if result.TotalTrades != 2 {
		t.Fatalf("Expected 2 trades, got %d", result.TotalTrades)
	}

	if math.Abs(result.RealizedPnL-20.0) > 0.001 {
		t.Fatalf("Expected $20 realized PnL, got %f", result.RealizedPnL)
	}
}

type frictionMock struct{}

func TestEngineRunFriction(t *testing.T) {
	now := time.Now()
	history := NewMemoryHistory([]marketdata.Tick{
		{Symbol: "BTCUSDT", Price: "100.00", Timestamp: now},                  // Force buy at 100
		{Symbol: "BTCUSDT", Price: "200.00", Timestamp: now.Add(time.Second)}, // Force sell at 200
	})

	strat := frictionMock{}

	opts := EmulatorOptions{
		MakerFeeRate: 0.00,
		TakerFeeRate: 0.01, // 1% fee
		SlippageRate: 0.05, // 5% slippage
	}
	engine := NewEngineWithOptions(&strat, 1000.0, opts)

	result, err := engine.RunTicks(context.Background(), history)
	if err != nil {
		t.Fatalf("RunTicks returned error: %v", err)
	}

	if result.TotalTrades != 2 {
		t.Fatalf("Expected 2 trades, got %d", result.TotalTrades)
	}

	// Execution trace:
	// BUY at 100. With 5% slippage -> 105. With 1% fee -> 105 * 1.01 = 106.05
	// SELL at 200. With 5% slippage -> 190. With 1% fee -> 190 * 0.99 = 188.10
	// Realized PnL = 188.10 - 106.05 = 82.05

	if math.Abs(result.RealizedPnL-82.05) > 0.001 {
		t.Fatalf("Expected $82.05 realized PnL, got %f", result.RealizedPnL)
	}
}

func (frictionMock) Name() string                                        { return "friction-mock" }
func (frictionMock) OnTick(_ context.Context) ([]strategy.Signal, error) { return nil, nil }
func (frictionMock) OnMarketTick(_ context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	if tick.Price == "100.00" {
		return []strategy.Signal{{AccountID: "test", Symbol: "BTCUSDT", Action: "buy", Quantity: "1.0"}}, nil
	} else if tick.Price == "200.00" {
		return []strategy.Signal{{AccountID: "test", Symbol: "BTCUSDT", Action: "sell", Quantity: "1.0"}}, nil
	}
	return nil, nil
}
