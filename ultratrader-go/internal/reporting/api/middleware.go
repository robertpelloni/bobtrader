package api

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter is a simple map-based token bucket rate limiter per IP address.
type RateLimiter struct {
	mu         sync.Mutex
	visitors   map[string]*visitor
	rate       float64       // Tokens added per second
	capacity   float64       // Maximum tokens per IP
	cleanupInt time.Duration // Interval to clean up stale visitors
}

type visitor struct {
	tokens   float64
	lastSeen time.Time
}

// NewRateLimiter initializes a per-IP rate limiter middleware.
func NewRateLimiter(requestsPerSecond, burstCapacity float64) *RateLimiter {
	rl := &RateLimiter{
		visitors:   make(map[string]*visitor),
		rate:       requestsPerSecond,
		capacity:   burstCapacity,
		cleanupInt: 5 * time.Minute,
	}
	go rl.cleanupLoop()
	return rl
}

func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(rl.cleanupInt)
	defer ticker.Stop()

	for {
		<-ticker.C
		rl.mu.Lock()
		now := time.Now()
		for ip, v := range rl.visitors {
			// If not seen in 5 minutes, delete
			if now.Sub(v.lastSeen) > rl.cleanupInt {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow checks if the given IP is allowed to make a request.
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[ip]
	if !exists {
		rl.visitors[ip] = &visitor{
			tokens:   rl.capacity - 1,
			lastSeen: now,
		}
		return true
	}

	// Refill tokens based on time passed
	elapsed := now.Sub(v.lastSeen).Seconds()
	v.tokens += elapsed * rl.rate
	if v.tokens > rl.capacity {
		v.tokens = rl.capacity
	}

	v.lastSeen = now

	if v.tokens >= 1 {
		v.tokens--
		return true
	}

	return false
}

// Middleware wraps an http.Handler with rate limiting logic.
func (rl *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extremely simple IP extraction.
		// In production behind a load balancer, use X-Forwarded-For or X-Real-IP.
		ip := r.RemoteAddr

		if !rl.Allow(ip) {
			http.Error(w, "429 Too Many Requests", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
