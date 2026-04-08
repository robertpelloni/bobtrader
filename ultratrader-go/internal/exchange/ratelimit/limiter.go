package ratelimit

import (
	"context"
	"sync"
	"time"
)

// Limiter implements a token bucket rate limiter for API call compliance.
type Limiter struct {
	mu         sync.Mutex
	maxTokens  int
	tokens     int
	refillRate time.Duration // How often to add a token
	lastRefill time.Time
}

// New creates a new rate limiter with the specified capacity and refill interval.
// Example: New(1200, time.Minute) allows 1200 requests per minute.
func New(maxTokens int, refillInterval time.Duration) *Limiter {
	return &Limiter{
		maxTokens:  maxTokens,
		tokens:     maxTokens,
		refillRate: refillInterval,
		lastRefill: time.Now(),
	}
}

// Wait blocks until a token is available, then consumes it.
// Returns ctx.Err() if the context is cancelled while waiting.
func (l *Limiter) Wait(ctx context.Context) error {
	for {
		if l.tryAcquire() {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(10 * time.Millisecond):
			continue
		}
	}
}

// TryAcquire attempts to acquire a token without blocking.
// Returns true if a token was acquired, false otherwise.
func (l *Limiter) TryAcquire() bool {
	return l.tryAcquire()
}

// Remaining returns the number of available tokens.
func (l *Limiter) Remaining() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refill()
	return l.tokens
}

func (l *Limiter) tryAcquire() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.refill()
	if l.tokens > 0 {
		l.tokens--
		return true
	}
	return false
}

func (l *Limiter) refill() {
	now := time.Now()
	elapsed := now.Sub(l.lastRefill)
	if elapsed < l.refillRate {
		return
	}
	// Add tokens based on elapsed time
	tokensToAdd := int(elapsed / l.refillRate)
	if tokensToAdd > 0 {
		l.tokens += tokensToAdd
		if l.tokens > l.maxTokens {
			l.tokens = l.maxTokens
		}
		l.lastRefill = l.lastRefill.Add(time.Duration(tokensToAdd) * l.refillRate)
	}
}

// BinanceSpotLimiter returns a rate limiter configured for Binance spot API limits.
// Binance spot allows 1200 request weight per minute.
// We use a conservative 1000 tokens with per-second refill to spread load evenly.
func BinanceSpotLimiter() *Limiter {
	// 1000 requests per 60 seconds = ~16.7 per second
	// We add 1 token every 60ms for smooth distribution
	return &Limiter{
		maxTokens:  1000,
		tokens:     1000,
		refillRate: 60 * time.Millisecond,
		lastRefill: time.Now(),
	}
}

// BinanceOrderLimiter returns a rate limiter for Binance order endpoints.
// Binance allows 50 orders per 10 seconds, 160000 per 24 hours.
func BinanceOrderLimiter() *Limiter {
	return &Limiter{
		maxTokens:  50,
		tokens:     50,
		refillRate: 200 * time.Millisecond, // 50 per 10 seconds
		lastRefill: time.Now(),
	}
}
