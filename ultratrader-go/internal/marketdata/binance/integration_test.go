package binance

import (
	"context"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange/binance"
)

func TestMarketDataFeedAccuracy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	adapter := binance.New(binance.Config{Testnet: false})
	feed := NewFeed(adapter)
	ctx := context.Background()

	symbol := "BTCUSDT"

	// 1. Verify REST Accuracy
	tick, err := feed.LatestTick(ctx, symbol)
	if err != nil {
		t.Fatalf("Failed to fetch latest tick via REST: %v", err)
	}
	if tick.Price == "" || tick.Price == "0" {
		t.Errorf("Received invalid price from REST: %s", tick.Price)
	}
	t.Logf("REST Tick verified: %s", tick.Price)
}
