package binance

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange/binance"
)

func TestMarketDataAccuracy(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping market data accuracy test in short mode")
	}

	// Use US endpoint to avoid restricted location error in sandbox
	adapter := binance.New(binance.Config{Testnet: false})
	feed := NewFeed(adapter)

	symbols := []string{"BTCUSDT", "ETHUSDT"}

	for _, symbol := range symbols {
		t.Run(symbol, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()

			tick, err := feed.LatestTick(ctx, symbol)
			if err != nil {
				t.Fatalf("Failed to fetch tick for %s: %v", symbol, err)
			}

			if tick.Symbol != symbol {
				t.Errorf("Expected symbol %s, got %s", symbol, tick.Symbol)
			}

			price, _ := strconv.ParseFloat(tick.Price, 64)
			if price <= 0 {
				t.Errorf("Invalid price for %s: %f", symbol, price)
			}

			// Reasonable bounds check for 2024-2026 era
			if symbol == "BTCUSDT" && (price < 10000 || price > 250000) {
				t.Errorf("BTC price %f out of expected sanity range (10k-250k)", price)
			}
			if symbol == "ETHUSDT" && (price < 500 || price > 20000) {
				t.Errorf("ETH price %f out of expected sanity range (500-20k)", price)
			}

			t.Logf("%s Price: %f", symbol, price)

			candle, err := feed.LatestCandle(ctx, symbol, "1m")
			if err != nil {
				t.Fatalf("Failed to fetch candle for %s: %v", symbol, err)
			}

			if candle.Symbol != symbol {
				t.Errorf("Expected symbol %s, got %s", symbol, candle.Symbol)
			}

			cp, _ := strconv.ParseFloat(candle.Close, 64)
			if cp <= 0 {
				t.Errorf("Invalid candle close for %s: %f", symbol, cp)
			}

			vol, _ := strconv.ParseFloat(candle.Volume, 64)
			if vol < 0 {
				t.Errorf("Invalid candle volume for %s: %f", symbol, vol)
			}
			if vol == 0 {
				t.Logf("Warning: %s candle volume is 0", symbol)
			}

			t.Logf("%s 1m Candle Close: %f Volume: %f", symbol, cp, vol)
		})
	}
}
