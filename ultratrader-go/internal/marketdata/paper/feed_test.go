package paper

import (
	"context"
	"testing"
	"time"
)

func TestLatestTick(t *testing.T) {
	feed := New()
	tick, err := feed.LatestTick(context.Background(), "BTCUSDT")
	if err != nil {
		t.Fatalf("LatestTick returned error: %v", err)
	}
	if tick.Price == "" {
		t.Fatal("expected price")
	}
}

func TestSubscribeTicks(t *testing.T) {
	feed := New()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sub := feed.SubscribeTicks(ctx, "BTCUSDT", 5*time.Millisecond)
	select {
	case tick, ok := <-sub.Chan():
		if !ok || tick.Symbol != "BTCUSDT" {
			t.Fatalf("unexpected tick: %+v open=%v", tick, ok)
		}
	case <-time.After(50 * time.Millisecond):
		t.Fatal("timed out waiting for tick")
	}
}
