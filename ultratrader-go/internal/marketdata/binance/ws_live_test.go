//go:build wslocal

package binance

import (
	"context"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange/binance"
)

func TestWSLiveFeed(t *testing.T) {
	adapter := binance.New(binance.Config{Testnet: false})
	feed := NewStreamFeed(adapter)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	sub := feed.SubscribeTicks(ctx, "BTCUSDT", 1*time.Second)
	ch := sub.Chan()

	select {
	case tick, ok := <-ch:
		if !ok {
			t.Fatal("Channel closed")
		}
		t.Logf("Tick received: %s %s", tick.Symbol, tick.Price)
		if tick.Symbol != "BTCUSDT" {
			t.Fatalf("Wrong symbol: %s", tick.Symbol)
		}
		if tick.Price == "" {
			t.Fatal("Empty price")
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Timeout - no ticks received from WebSocket")
	}
}
