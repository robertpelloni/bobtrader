package strategy

import (
	"context"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

type testStrategy struct{}

func (testStrategy) Name() string { return "test" }
func (testStrategy) OnTick(_ context.Context) ([]Signal, error) {
	return []Signal{{AccountID: "paper-main", Symbol: "BTCUSDT", Action: "buy", Reason: "test"}}, nil
}

type tickOnlyStrategy struct{}

func (tickOnlyStrategy) Name() string                               { return "tick" }
func (tickOnlyStrategy) OnTick(_ context.Context) ([]Signal, error) { return nil, nil }
func (tickOnlyStrategy) OnMarketTick(_ context.Context, tick marketdata.Tick) ([]Signal, error) {
	return []Signal{{AccountID: "paper-main", Symbol: tick.Symbol, Action: "buy", Reason: "tick"}}, nil
}

type candleOnlyStrategy struct{}

func (candleOnlyStrategy) Name() string                               { return "candle" }
func (candleOnlyStrategy) OnTick(_ context.Context) ([]Signal, error) { return nil, nil }
func (candleOnlyStrategy) OnMarketCandle(_ context.Context, candle marketdata.Candle) ([]Signal, error) {
	return []Signal{{AccountID: "paper-main", Symbol: candle.Symbol, Action: "buy", Reason: "candle"}}, nil
}

func TestRuntimeTickAggregatesSignals(t *testing.T) {
	runtime := NewRuntime(testStrategy{})
	signals, err := runtime.Tick(context.Background())
	if err != nil {
		t.Fatalf("Tick returned error: %v", err)
	}
	if len(signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(signals))
	}
}

func TestRuntimeTickEventAggregatesTickStrategies(t *testing.T) {
	runtime := NewRuntime(testStrategy{}, tickOnlyStrategy{})
	signals, err := runtime.TickEvent(context.Background(), marketdata.Tick{Symbol: "BTCUSDT", Timestamp: time.Now()})
	if err != nil {
		t.Fatalf("TickEvent returned error: %v", err)
	}
	if len(signals) != 1 || signals[0].Reason != "tick" {
		t.Fatalf("unexpected tick signals: %+v", signals)
	}
}

func TestRuntimeCandleEventAggregatesCandleStrategies(t *testing.T) {
	runtime := NewRuntime(testStrategy{}, candleOnlyStrategy{})
	signals, err := runtime.CandleEvent(context.Background(), marketdata.Candle{Symbol: "BTCUSDT", Timestamp: time.Now()})
	if err != nil {
		t.Fatalf("CandleEvent returned error: %v", err)
	}
	if len(signals) != 1 || signals[0].Reason != "candle" {
		t.Fatalf("unexpected candle signals: %+v", signals)
	}
}
