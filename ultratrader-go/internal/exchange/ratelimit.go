package exchange

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// RateLimiter manages outgoing request rates to prevent API bans.
// It implements a basic token bucket algorithm.
type RateLimiter struct {
	mu         sync.Mutex
	tokens     float64
	capacity   float64
	ratePerSec float64
	lastUpdate time.Time
}

// NewRateLimiter creates a new rate limiter with a specific capacity and refill rate.
func NewRateLimiter(capacity, ratePerSec float64) *RateLimiter {
	return &RateLimiter{
		tokens:     capacity,
		capacity:   capacity,
		ratePerSec: ratePerSec,
		lastUpdate: time.Now(),
	}
}

// Wait blocks until a token is available or the context is cancelled.
// weight allows consuming multiple tokens for "expensive" endpoints.
func (rl *RateLimiter) Wait(ctx context.Context, weight float64) error {
	for {
		rl.mu.Lock()

		// Refill
		now := time.Now()
		elapsed := now.Sub(rl.lastUpdate).Seconds()
		rl.tokens += elapsed * rl.ratePerSec
		if rl.tokens > rl.capacity {
			rl.tokens = rl.capacity
		}
		rl.lastUpdate = now

		if rl.tokens >= weight {
			rl.tokens -= weight
			rl.mu.Unlock()
			return nil
		}

		rl.mu.Unlock()

		// Wait and try again
		select {
		case <-ctx.Done():
			return fmt.Errorf("rate limit wait cancelled: %w", ctx.Err())
		case <-time.After(50 * time.Millisecond):
			// continue looping
		}
	}
}

// Transport wraps an http.RoundTripper with rate limiting.
type Transport struct {
	Base        http.RoundTripper
	RateLimiter *RateLimiter
	Weight      float64 // Default weight to consume per request (usually 1.0)
}

// RoundTrip executes a single HTTP transaction, waiting for a rate limit token first.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Base == nil {
		t.Base = http.DefaultTransport
	}

	weight := t.Weight
	if weight <= 0 {
		weight = 1.0
	}

	if err := t.RateLimiter.Wait(req.Context(), weight); err != nil {
		return nil, err
	}

	return t.Base.RoundTrip(req)
}
