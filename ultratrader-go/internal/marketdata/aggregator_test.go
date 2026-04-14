package marketdata_test

import (
	"context"
	"errors"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

type mockFeed struct {
	tick marketdata.Tick
	err  error
}

func (m *mockFeed) LatestTick(ctx context.Context, symbol string) (marketdata.Tick, error) {
	if m.err != nil {
		return marketdata.Tick{}, m.err
	}
	return m.tick, nil
}

func (m *mockFeed) LatestCandle(ctx context.Context, symbol, interval string) (marketdata.Candle, error) {
	return marketdata.Candle{}, nil
}

func TestAggregator_AveragePrice(t *testing.T) {
	agg := marketdata.NewAggregator(marketdata.AveragePrice, nil)

	feed1 := &mockFeed{tick: marketdata.Tick{Price: "100.0"}}
	feed2 := &mockFeed{tick: marketdata.Tick{Price: "102.0"}}

	agg.AddFeed("f1", feed1)
	agg.AddFeed("f2", feed2)

	ctx := context.Background()
	tick, err := agg.LatestTick(ctx, "BTC")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if tick.Price != "101" {
		t.Errorf("Expected 101, got %s", tick.Price)
	}
}

func TestAggregator_MedianPrice(t *testing.T) {
	agg := marketdata.NewAggregator(marketdata.MedianPrice, nil)

	feed1 := &mockFeed{tick: marketdata.Tick{Price: "100.0"}}
	feed2 := &mockFeed{tick: marketdata.Tick{Price: "150.0"}}
	feed3 := &mockFeed{tick: marketdata.Tick{Price: "102.0"}}

	agg.AddFeed("f1", feed1)
	agg.AddFeed("f2", feed2)
	agg.AddFeed("f3", feed3)

	ctx := context.Background()
	tick, err := agg.LatestTick(ctx, "BTC")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if tick.Price != "102" && tick.Price != "102.0" {
		t.Errorf("Expected 102, got %s", tick.Price)
	}
}

func TestAggregator_Failover(t *testing.T) {
	agg := marketdata.NewAggregator(marketdata.Failover, nil)

	feed1 := &mockFeed{err: errors.New("timeout")}
	feed2 := &mockFeed{tick: marketdata.Tick{Price: "102.0"}}

	// Because iteration order of maps is random, test might be flaky if both succeeded.
	// But since feed1 always fails, it should always return feed2's tick.
	agg.AddFeed("f1", feed1)
	agg.AddFeed("f2", feed2)

	ctx := context.Background()
	tick, err := agg.LatestTick(ctx, "BTC")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if tick.Price != "102" && tick.Price != "102.0" {
		t.Errorf("Expected 102, got %s", tick.Price)
	}
}
