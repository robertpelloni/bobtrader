package api_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/reporting/api"
)

func TestRateLimiter_Allow(t *testing.T) {
	// Create limiter that allows 2 requests immediately (burst=2), then refills 10 per sec (0.1s each)
	limiter := api.NewRateLimiter(10, 2)
	ip := "192.168.1.1:1234"

	// Request 1: Should pass
	if !limiter.Allow(ip) {
		t.Errorf("First request should be allowed")
	}

	// Request 2: Should pass
	if !limiter.Allow(ip) {
		t.Errorf("Second request should be allowed (burst of 2)")
	}

	// Request 3: Should fail (bucket empty, no time elapsed)
	if limiter.Allow(ip) {
		t.Errorf("Third request should fail due to rate limit")
	}

	// Wait 0.15 seconds to refill at least 1 token
	time.Sleep(150 * time.Millisecond)

	// Request 4: Should pass after waiting
	if !limiter.Allow(ip) {
		t.Errorf("Fourth request should be allowed after waiting for refill")
	}
}

func TestRateLimiter_Middleware(t *testing.T) {
	limiter := api.NewRateLimiter(10, 1)

	// A dummy handler
	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap handler
	handler := limiter.Middleware(nextHandler)

	req, _ := http.NewRequest("GET", "/", nil)
	req.RemoteAddr = "10.0.0.1:5678" // mock IP

	// Request 1: Burst=1, should pass
	rr1 := httptest.NewRecorder()
	handler.ServeHTTP(rr1, req)
	if rr1.Code != http.StatusOK {
		t.Errorf("Expected 200 OK, got %d", rr1.Code)
	}

	// Request 2: Immediate, should fail with 429
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("Expected 429 Too Many Requests, got %d", rr2.Code)
	}
}
