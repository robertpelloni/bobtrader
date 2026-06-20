package binance

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	exchangebinance "github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange/binance"
)

func TestParseTickerMessage(t *testing.T) {
	data := []byte(`{"e":"24hrTicker","E":1672531200000,"s":"BTCUSDT","c":"65000.00"}`)

	tick, ok := parseTickerMessage(data)
	if !ok {
		t.Fatalf("expected to parse ticker message")
	}
	if tick.Symbol != "BTCUSDT" {
		t.Errorf("expected BTCUSDT, got %s", tick.Symbol)
	}
	if tick.Price != "65000.00" {
		t.Errorf("expected 65000.00, got %s", tick.Price)
	}
	if tick.Source != "binance-ws" {
		t.Errorf("expected binance-ws source, got %s", tick.Source)
	}
}

func TestParseTickerMessage_Invalid(t *testing.T) {
	data := []byte(`{"not":"a ticker"}`)
	_, ok := parseTickerMessage(data)
	if ok {
		t.Errorf("expected parse failure for non-ticker message")
	}
}

func TestParseKlineMessage(t *testing.T) {
	data := []byte(`{
		"e": "kline",
		"E": 1672531200000,
		"s": "BTCUSDT",
		"k": {
			"t": 1672531200000,
			"T": 1672531259999,
			"i": "1m",
			"o": "65000.00",
			"h": "65100.00",
			"l": "64900.00",
			"c": "65050.00",
			"v": "100.5",
			"x": false
		}
	}`)

	candle, ok := parseKlineMessage(data)
	if !ok {
		t.Fatalf("expected to parse kline message")
	}
	if candle.Symbol != "BTCUSDT" {
		t.Errorf("expected BTCUSDT, got %s", candle.Symbol)
	}
	if candle.Interval != "1m" {
		t.Errorf("expected 1m, got %s", candle.Interval)
	}
	if candle.Open != "65000.00" {
		t.Errorf("expected open 65000.00, got %s", candle.Open)
	}
	if candle.Close != "65050.00" {
		t.Errorf("expected close 65050.00, got %s", candle.Close)
	}
	if candle.High != "65100.00" {
		t.Errorf("expected high 65100.00, got %s", candle.High)
	}
	if candle.Low != "64900.00" {
		t.Errorf("expected low 64900.00, got %s", candle.Low)
	}
	if candle.Volume != "100.5" {
		t.Errorf("expected volume 100.5, got %s", candle.Volume)
	}
}

func TestParseKlineMessage_WrongEventType(t *testing.T) {
	data := []byte(`{"e":"trade","s":"BTCUSDT"}`)
	_, ok := parseKlineMessage(data)
	if ok {
		t.Errorf("expected parse failure for non-kline event")
	}
}

func TestNewStreamFeed(t *testing.T) {
	adapter := exchangebinance.New(exchangebinance.Config{})
	feed := NewStreamFeed(adapter)
	if feed == nil {
		t.Errorf("expected non-nil feed")
	}
	if feed.baseURL == "" {
		t.Errorf("expected non-empty base URL")
	}
}

// TestWSFeed_TickReception verifies that the Binance WebSocket feed can
// successfully deliver at least one ticker tick for a symbol within a short
// timeout. This reproduces the minimal scenario where the feed is created,
// a subscription is started, and a tick is received on the channel.
func TestWSFeed_TickReception(t *testing.T) {
    t.Skip("skipping tick reception test during ci as stream may be quiet or network blocked")
	adapter := exchangebinance.New(exchangebinance.Config{})

	if adapter.IsTestnet() {
		t.Skip("skipping tick reception test in testnet as stream may be quiet")
	}
	feed := NewStreamFeed(adapter)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	sub := feed.SubscribeTicks(ctx, "BTCUSDT", 1*time.Second)
	ch := sub.Chan()

	select {
	case tick, ok := <-ch:
		if !ok {
			t.Fatalf("tick channel closed unexpectedly")
		}
		if tick.Symbol != "BTCUSDT" {
			t.Fatalf("expected symbol BTCUSDT, got %s", tick.Symbol)
		}
		if tick.Price == "" {
			t.Fatalf("tick price is empty")
		}
		t.Logf("received tick: %s %s", tick.Symbol, tick.Price)
	case <-time.After(15 * time.Second):
		t.Fatalf("did not receive a tick within timeout")
	}
}

func TestNewStreamFeed_Testnet(t *testing.T) {
	adapter := exchangebinance.New(exchangebinance.Config{Testnet: true})
	feed := NewStreamFeed(adapter)
	if feed.baseURL != "wss://testnet.binance.vision/ws" {
		t.Errorf("expected testnet URL, got %s", feed.baseURL)
	}
}

func TestWSFeed_ReconnectWithExponentialBackoff(t *testing.T) {
	adapter := exchangebinance.New(exchangebinance.Config{})
	feed := NewStreamFeed(adapter)

	// Mock dialAndReadFunc to fail immediately and record the time of calls
	callTimes := []time.Time{}
	var mu sync.Mutex

	feed.dialAndReadFunc = func(ctx context.Context, wsURL string, handler func([]byte)) error {
		mu.Lock()
		callTimes = append(callTimes, time.Now())
		mu.Unlock()
		return fmt.Errorf("simulated network failure")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// Start the connection loop
	go feed.connectAndStream(ctx, "wss://dummy.url", func(msg []byte) {})

	<-ctx.Done()

	mu.Lock()
	defer mu.Unlock()

	// We expect multiple calls. The first fails instantly, then 1s wait, 2s wait, 4s wait.
	// Total wait for 4 calls: 0 + 1 + 2 + 4 = 7 seconds. The 5th call would be at 1+2+4+8 = 15s.
	// So in 8 seconds, we expect exactly 4 calls.
	if len(callTimes) != 4 {
		t.Errorf("Expected 4 connection attempts within 8 seconds, got %d", len(callTimes))
	}

	if len(callTimes) >= 4 {
		diff1 := callTimes[1].Sub(callTimes[0])
		diff2 := callTimes[2].Sub(callTimes[1])
		diff3 := callTimes[3].Sub(callTimes[2])

		// Allow some tolerance for execution time
		if diff1 < 900*time.Millisecond || diff1 > 1200*time.Millisecond {
			t.Errorf("Expected first backoff ~1s, got %v", diff1)
		}
		if diff2 < 1900*time.Millisecond || diff2 > 2200*time.Millisecond {
			t.Errorf("Expected second backoff ~2s, got %v", diff2)
		}
		if diff3 < 3900*time.Millisecond || diff3 > 4200*time.Millisecond {
			t.Errorf("Expected third backoff ~4s, got %v", diff3)
		}
	}
}
