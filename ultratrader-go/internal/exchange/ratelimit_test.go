package exchange_test

import (
	"context"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

func TestRateLimiter_Wait(t *testing.T) {
	// Create limiter: Capacity 2, refilling 10 per second
	rl := exchange.NewRateLimiter(2.0, 10.0)

	ctx := context.Background()

	start := time.Now()

	// Should consume immediately (tokens = 1)
	err := rl.Wait(ctx, 1.0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should consume immediately (tokens = 0)
	err = rl.Wait(ctx, 1.0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should block until refill (needs 1 token, refilling at 10/s means 100ms)
	err = rl.Wait(ctx, 1.0)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	elapsed := time.Since(start)

	// Since we had a capacity of 2 and asked for 3, the third request had to wait.
	// 1 token / 10 per sec = 0.1 seconds (100ms).
	if elapsed < 90*time.Millisecond {
		t.Errorf("Expected wait of ~100ms, but completed in %v", elapsed)
	}
}

func TestRateLimiter_ContextCancel(t *testing.T) {
	// Limiter refilling incredibly slowly
	rl := exchange.NewRateLimiter(0.0, 0.0001)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := rl.Wait(ctx, 1.0)

	if err == nil {
		t.Fatalf("Expected wait to be cancelled by context")
	}
}
