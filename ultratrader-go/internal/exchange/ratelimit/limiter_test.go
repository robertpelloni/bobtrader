package ratelimit

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestLimiter_AllowsUpToMax(t *testing.T) {
	limiter := New(5, 10*time.Millisecond)

	allowed := 0
	for i := 0; i < 10; i++ {
		if limiter.TryAcquire() {
			allowed++
		}
	}
	if allowed != 5 {
		t.Errorf("expected 5 acquisitions, got %d", allowed)
	}
}

func TestLimiter_Refills(t *testing.T) {
	limiter := New(3, 10*time.Millisecond)

	// Consume all tokens
	for i := 0; i < 3; i++ {
		if !limiter.TryAcquire() {
			t.Fatalf("expected token %d to be available", i)
		}
	}

	// Should be empty now
	if limiter.TryAcquire() {
		t.Errorf("expected no tokens available")
	}

	// Wait for refill
	time.Sleep(50 * time.Millisecond)

	// Should have refilled
	if limiter.Remaining() == 0 {
		t.Errorf("expected tokens to refill after waiting")
	}
}

func TestLimiter_WaitBlocks(t *testing.T) {
	limiter := New(2, 20*time.Millisecond)

	// Consume both tokens
	limiter.TryAcquire()
	limiter.TryAcquire()

	// Wait should block until refill
	start := time.Now()
	err := limiter.Wait(context.Background())
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed < 10*time.Millisecond {
		t.Errorf("expected Wait to block, but it returned in %v", elapsed)
	}
}

func TestLimiter_WaitContextCancellation(t *testing.T) {
	limiter := New(1, 10*time.Second) // Very slow refill
	limiter.TryAcquire()             // Consume the only token

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := limiter.Wait(ctx)
	if err == nil {
		t.Errorf("expected context cancellation error")
	}
}

func TestLimiter_Remaining(t *testing.T) {
	limiter := New(10, 100*time.Millisecond)

	if limiter.Remaining() != 10 {
		t.Errorf("expected 10 remaining, got %d", limiter.Remaining())
	}

	limiter.TryAcquire()
	if limiter.Remaining() != 9 {
		t.Errorf("expected 9 remaining, got %d", limiter.Remaining())
	}
}

func TestLimiter_ConcurrentSafety(t *testing.T) {
	limiter := New(100, time.Millisecond)

	var acquired atomic.Int32
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Launch many goroutines trying to acquire
	for i := 0; i < 200; i++ {
		go func() {
			select {
			case <-ctx.Done():
				return
			default:
				if limiter.TryAcquire() {
					acquired.Add(1)
				}
			}
		}()
	}

	time.Sleep(50 * time.Millisecond)

	count := acquired.Load()
	if count > 102 { // Allow 2 for refill race during concurrent burst
		t.Errorf("expected at most ~100 acquisitions, got %d", count)
	}
}

func TestBinanceSpotLimiter(t *testing.T) {
	limiter := BinanceSpotLimiter()
	if limiter.Remaining() != 1000 {
		t.Errorf("expected 1000 initial tokens, got %d", limiter.Remaining())
	}
}

func TestBinanceOrderLimiter(t *testing.T) {
	limiter := BinanceOrderLimiter()
	if limiter.Remaining() != 50 {
		t.Errorf("expected 50 initial tokens, got %d", limiter.Remaining())
	}
}
