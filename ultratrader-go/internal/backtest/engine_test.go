package backtest

import (
	"context"
	"math"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// mockStrategy buys on every tick below 100, sells above 100.
type mockStrategy struct{}

func (mockStrategy) Name() string                                        { return "mock" }
func (mockStrategy) OnTick(_ context.Context) ([]strategy.Signal, error) { return nil, nil }
func (mockStrategy) OnMarketTick(_ context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	if tick.Price == "90.00" {
		return []strategy.Signal{{AccountID: "test", Symbol: "BTCUSDT", Action: "buy", Quantity: "1.0"}}, nil
	} else if tick.Price == "110.00" {
		return []strategy.Signal{{AccountID: "test", Symbol: "BTCUSDT", Action: "sell", Quantity: "1.0"}}, nil
	}
	return nil, nil
}

func TestEngineRun(t *testing.T) {
	now := time.Now()
	history := NewMemoryHistory([]marketdata.Tick{
		{Symbol: "BTCUSDT", Price: "90.00", Timestamp: now},
		{Symbol: "BTCUSDT", Price: "110.00", Timestamp: now.Add(time.Second)},
	})

	strat := mockStrategy{}
	engine := NewEngine(strat, history, 1000.0)

	result, err := engine.Run(context.Background())
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
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
