package paper

import (
	"context"
	"testing"
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
