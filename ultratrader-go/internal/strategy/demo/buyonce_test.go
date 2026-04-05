package demo

import (
	"context"
	"testing"
)

func TestBuyOnceEmitsOnlyOnce(t *testing.T) {
	strategy := NewBuyOnce("paper-main", "BTCUSDT", "0.01")

	first, err := strategy.OnTick(context.Background())
	if err != nil {
		t.Fatalf("OnTick returned error: %v", err)
	}
	if len(first) != 1 {
		t.Fatalf("expected 1 signal on first tick, got %d", len(first))
	}

	second, err := strategy.OnTick(context.Background())
	if err != nil {
		t.Fatalf("OnTick returned error: %v", err)
	}
	if len(second) != 0 {
		t.Fatalf("expected 0 signals on second tick, got %d", len(second))
	}
}
