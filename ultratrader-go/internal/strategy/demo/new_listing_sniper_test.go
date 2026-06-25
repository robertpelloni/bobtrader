package demo_test

import (
	"context"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/demo"
)

func TestNewListingSniper(t *testing.T) {
	strat := demo.NewTestListingSniper("test-acct", "1.0", "USDT")

	strat.PreWarm([]string{"BTCUSDT"})

	ctx := context.Background()

	signals, _ := strat.OnMarketTick(ctx, marketdata.Tick{Symbol: "BTCUSDT", Price: "50000"})
	if len(signals) > 0 {
		t.Errorf("Expected no signals for pre-warmed token, got %v", signals)
	}

	signals, _ = strat.OnMarketTick(ctx, marketdata.Tick{Symbol: "DOGEBTC", Price: "0.0001"})
	if len(signals) > 0 {
		t.Errorf("Expected no signals for wrong base currency, got %v", signals)
	}

	signals, _ = strat.OnMarketTick(ctx, marketdata.Tick{Symbol: "PEPEUSDT", Price: "0.000001"})
	if len(signals) != 1 || signals[0].Action != "buy" {
		t.Fatalf("Expected 1 buy signal for new token, got %v", signals)
	}

	if signals[0].Symbol != "PEPEUSDT" {
		t.Errorf("Expected symbol PEPEUSDT, got %s", signals[0].Symbol)
	}

	signals, _ = strat.OnMarketTick(ctx, marketdata.Tick{Symbol: "PEPEUSDT", Price: "0.000002"})
	if len(signals) > 0 {
		t.Errorf("Expected no signals for second tick of new token, got %v", signals)
	}
}
