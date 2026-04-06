package demo

import (
	"context"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

func TestTickMomentumBurstBuySignal(t *testing.T) {
	strategy := NewTickMomentumBurst("paper-main", "BTCUSDT", "0.01", 3, 0.05, 0.05)
	_, _ = strategy.OnMarketTick(context.Background(), marketdata.Tick{Symbol: "BTCUSDT", Price: "100.00", Timestamp: time.Now()})
	_, _ = strategy.OnMarketTick(context.Background(), marketdata.Tick{Symbol: "BTCUSDT", Price: "100.02", Timestamp: time.Now()})
	signals, err := strategy.OnMarketTick(context.Background(), marketdata.Tick{Symbol: "BTCUSDT", Price: "100.10", Timestamp: time.Now()})
	if err != nil {
		t.Fatalf("OnMarketTick returned error: %v", err)
	}
	if len(signals) != 1 || signals[0].Action != "buy" {
		t.Fatalf("expected buy signal, got %+v", signals)
	}
}

func TestTickMomentumBurstSellSignal(t *testing.T) {
	strategy := NewTickMomentumBurst("paper-main", "BTCUSDT", "0.01", 3, 0.05, 0.05)
	_, _ = strategy.OnMarketTick(context.Background(), marketdata.Tick{Symbol: "BTCUSDT", Price: "100.00", Timestamp: time.Now()})
	_, _ = strategy.OnMarketTick(context.Background(), marketdata.Tick{Symbol: "BTCUSDT", Price: "99.98", Timestamp: time.Now()})
	signals, err := strategy.OnMarketTick(context.Background(), marketdata.Tick{Symbol: "BTCUSDT", Price: "99.80", Timestamp: time.Now()})
	if err != nil {
		t.Fatalf("OnMarketTick returned error: %v", err)
	}
	if len(signals) != 1 || signals[0].Action != "sell" {
		t.Fatalf("expected sell signal, got %+v", signals)
	}
}
