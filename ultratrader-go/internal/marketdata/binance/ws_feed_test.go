package binance

import (
	"context"
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
