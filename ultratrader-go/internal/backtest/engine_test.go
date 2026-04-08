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
	engine := NewEngine(strat, 1000.0)

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
	engine := NewEngine(strat, 1000.0)

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
