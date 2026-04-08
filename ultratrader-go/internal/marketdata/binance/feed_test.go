package binance

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	exchangebinance "github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange/binance"
)

func TestFeed_LatestTick(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{"symbol": "BTCUSDT", "price": "65500.00"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	adapter := exchangebinance.New(exchangebinance.Config{})
	adapter.SetBaseURL(server.URL)

	feed := NewFeed(adapter)
	tick, err := feed.LatestTick(context.Background(), "BTCUSDT")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tick.Symbol != "BTCUSDT" {
		t.Errorf("expected BTCUSDT, got %s", tick.Symbol)
	}
	if tick.Price != "65500.00" {
		t.Errorf("expected 65500.00, got %s", tick.Price)
	}
	if tick.Source != "binance" {
		t.Errorf("expected binance source, got %s", tick.Source)
	}
}

func TestFeed_LatestCandle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := [][]interface{}{
			{1609459200000, "65000.00", "65100.00", "64900.00", "65050.00", "100.5", 1609459259999, "6528250.00", 150, "50.25", "3264125.00", "0"},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	adapter := exchangebinance.New(exchangebinance.Config{})
	adapter.SetBaseURL(server.URL)

	feed := NewFeed(adapter)
	candle, err := feed.LatestCandle(context.Background(), "BTCUSDT", "1m")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if candle.Symbol != "BTCUSDT" {
		t.Errorf("expected BTCUSDT, got %s", candle.Symbol)
	}
	if candle.Close != "65050.00" {
		t.Errorf("expected close 65050.00, got %s", candle.Close)
	}
}

func TestFeed_SubscribeTicks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]string{"symbol": "BTCUSDT", "price": "65000.00"}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	adapter := exchangebinance.New(exchangebinance.Config{})
	adapter.SetBaseURL(server.URL)

	feed := NewFeed(adapter)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sub := feed.SubscribeTicks(ctx, "BTCUSDT", 50*time.Millisecond)

	select {
	case tick := <-sub.Chan():
		if tick.Symbol != "BTCUSDT" {
			t.Errorf("expected BTCUSDT, got %s", tick.Symbol)
		}
	case <-time.After(200 * time.Millisecond):
		t.Errorf("expected to receive tick within timeout")
	}
}

func TestCandleIntervalToDuration(t *testing.T) {
	tests := []struct {
		interval string
		expected time.Duration
	}{
		{"1m", 1 * time.Minute},
		{"5m", 5 * time.Minute},
		{"1h", 1 * time.Hour},
		{"1d", 24 * time.Hour},
		{"4h", 4 * time.Hour},
	}
	for _, tt := range tests {
		got := candleIntervalToDuration(tt.interval)
		if got != tt.expected {
			t.Errorf("interval %s: expected %v, got %v", tt.interval, tt.expected, got)
		}
	}
}
