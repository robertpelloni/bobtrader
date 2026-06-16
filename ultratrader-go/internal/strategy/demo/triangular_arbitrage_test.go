package demo

import (
	"context"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

func TestTriangularArbitrage(t *testing.T) {
	// BTCUSDT = 50000
	// ETHBTC = 0.05
	// ETHUSDT = 2501 (Spread! 50000 * 0.05 = 2500)

	strategy := NewTriangularArbitrage("test", "USDT", "BTC", "ETH", 0.01)

	// Feed prices
	ctx := context.Background()
	_, _ = strategy.OnMarketTick(ctx, marketdata.Tick{Symbol: "BTCUSDT", Price: "50000"})
	_, _ = strategy.OnMarketTick(ctx, marketdata.Tick{Symbol: "ETHBTC", Price: "0.05"})
	signals, _ := strategy.OnMarketTick(ctx, marketdata.Tick{Symbol: "ETHUSDT", Price: "2501"})

	// Profit 1: (1 / 50000 / 0.05 * 2501) - 1 = (1 / 2500 * 2501) - 1 = 0.0004 = 0.04%
	if len(signals) == 0 {
		t.Errorf("expected arbitrage signals, got 0")
	}

	if signals[0].Symbol != "BTCUSDT" {
		t.Errorf("expected signal for BTCUSDT, got %s", signals[0].Symbol)
	}
}
