package app

import (
	"context"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange/binance"
	marketdatabinance "github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata/binance"
)

func TestLiveMarketMonitor(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live market monitor in short mode")
	}

	// Use public Binance feed (requires no API keys for public streams)
	adapter := binance.New(binance.Config{Testnet: false})
	feed := marketdatabinance.NewStreamFeed(adapter)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	symbol := "BTCUSDT"
	t.Logf("Connecting to live WebSocket feed for %s...", symbol)

	sub := feed.SubscribeTicks(ctx, symbol, 0)

	tickCount := 0
	for {
		select {
		case <-ctx.Done():
			if tickCount == 0 {
				t.Log("Warning: No ticks received from live feed during timeout. This may happen if the network is restricted or Binance is down.")
			} else {
				t.Logf("Successfully monitored %d ticks from live market.", tickCount)
			}
			return
		case tick := <-sub.Chan():
			tickCount++
			if tickCount == 1 {
				t.Logf("First tick received: Price=%s Source=%s", tick.Price, tick.Source)
			}
			if tickCount >= 5 {
				t.Logf("Received %d ticks, live feed confirmed.", tickCount)
				cancel()
			}
		}
	}
}
