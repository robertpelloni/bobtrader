package demo

import (
	"context"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

func buildCompositeWarmupTicks(n int, basePrice float64) []marketdata.Tick {
	ticks := make([]marketdata.Tick, n)
	for i := 0; i < n; i++ {
		// Small oscillation within bands, no indicator extremes
		offset := float64(i%3-1) * basePrice * 0.001
		ticks[i] = marketdata.Tick{
			Symbol: "BTCUSDT",
			Price:  formatFloat(basePrice + offset),
			Source: "binance",
		}
	}
	return ticks
}

func TestRSIBollingerComposite_Warmup(t *testing.T) {
	s := NewRSIBollingerComposite("paper-main", "BTCUSDT", "0.01", 14, 30, 70, 20, 2.0, nil)

	// Feed fewer than rsiPeriod + bbPeriod ticks
	for _, tick := range buildCompositeWarmupTicks(15, 100) {
		signals, err := s.OnMarketTick(context.Background(), tick)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(signals) > 0 {
			t.Errorf("expected no signals during warmup, got %d", len(signals))
		}
	}
}

func TestRSIBollingerComposite_BuySignal(t *testing.T) {
	s := NewRSIBollingerComposite("paper-main", "BTCUSDT", "0.01", 14, 30, 70, 20, 2.0, nil)

	// Warm up the indicators
	for _, tick := range buildCompositeWarmupTicks(30, 100) {
		s.OnMarketTick(context.Background(), tick)
	}

	// Trigger a buy: drop price sharply so that RSI becomes oversold AND price is below lower BB
	// Feed a series of falling ticks to push RSI down and touch the lower band
	prices := []float64{90, 80, 70, 60, 50, 40, 30}
	var lastSignals []marketdata.Tick
	for _, p := range prices {
		tick := marketdata.Tick{Symbol: "BTCUSDT", Price: formatFloat(p), Source: "binance"}
		lastSignals = append(lastSignals, tick)
	}

	buySignalsCount := 0
	for _, tick := range lastSignals {
		signals, err := s.OnMarketTick(context.Background(), tick)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, sig := range signals {
			if sig.Action == "buy" {
				buySignalsCount++
			}
		}
	}

	if buySignalsCount != 1 {
		t.Errorf("expected exactly 1 buy signal during drop sequence, got %d", buySignalsCount)
	}
}

func TestRSIBollingerComposite_SellSignal(t *testing.T) {
	s := NewRSIBollingerComposite("paper-main", "BTCUSDT", "0.01", 14, 30, 70, 20, 2.0, nil)

	// Warm up
	for _, tick := range buildCompositeWarmupTicks(30, 100) {
		s.OnMarketTick(context.Background(), tick)
	}

	// Trigger a sell: rise price sharply so RSI is overbought AND price is above upper BB
	prices := []float64{110, 120, 130, 140, 150, 160, 170}
	var lastSignals []marketdata.Tick
	for _, p := range prices {
		tick := marketdata.Tick{Symbol: "BTCUSDT", Price: formatFloat(p), Source: "binance"}
		lastSignals = append(lastSignals, tick)
	}

	sellSignalsCount := 0
	for _, tick := range lastSignals {
		signals, err := s.OnMarketTick(context.Background(), tick)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		for _, sig := range signals {
			if sig.Action == "sell" {
				sellSignalsCount++
			}
		}
	}

	if sellSignalsCount != 1 {
		t.Errorf("expected exactly 1 sell signal during rise sequence, got %d", sellSignalsCount)
	}
}

func TestRSIBollingerComposite_Deduplication(t *testing.T) {
	s := NewRSIBollingerComposite("paper-main", "BTCUSDT", "0.01", 14, 30, 70, 20, 2.0, nil)

	// Warm up
	for _, tick := range buildCompositeWarmupTicks(30, 100) {
		s.OnMarketTick(context.Background(), tick)
	}

	// Sharp drop to trigger buy
	prices := []float64{90, 80, 70, 60, 50, 40, 30}
	for _, p := range prices {
		tick := marketdata.Tick{Symbol: "BTCUSDT", Price: formatFloat(p), Source: "binance"}
		s.OnMarketTick(context.Background(), tick)
	}

	// Next tick at same low price should NOT trigger buy again (deduplication)
	tick := marketdata.Tick{Symbol: "BTCUSDT", Price: formatFloat(30), Source: "binance"}
	signals, _ := s.OnMarketTick(context.Background(), tick)
	for _, sig := range signals {
		if sig.Action == "buy" {
			t.Errorf("expected no repeat buy signal (deduplicated)")
		}
	}
}

func TestRSIBollingerComposite_NeutralZoneReset(t *testing.T) {
	s := NewRSIBollingerComposite("paper-main", "BTCUSDT", "0.01", 14, 30, 70, 20, 2.0, nil)

	// Warm up
	for _, tick := range buildCompositeWarmupTicks(30, 100) {
		s.OnMarketTick(context.Background(), tick)
	}

	// Sharp drop to trigger buy
	prices := []float64{90, 80, 70, 60, 50, 40, 30}
	firstDropBuyCount := 0
	for _, p := range prices {
		tick := marketdata.Tick{Symbol: "BTCUSDT", Price: formatFloat(p), Source: "binance"}
		signals, _ := s.OnMarketTick(context.Background(), tick)
		for _, sig := range signals {
			if sig.Action == "buy" {
				firstDropBuyCount++
			}
		}
	}
	if firstDropBuyCount != 1 {
		t.Fatalf("expected exactly 1 buy signal on first drop, got %d", firstDropBuyCount)
	}

	// Move price back to neutral zone (middle of bands, RSI around 50)
	// Warm up again at base price 100
	for _, tick := range buildCompositeWarmupTicks(25, 100) {
		s.OnMarketTick(context.Background(), tick)
	}

	// Now a new sharp drop should trigger a new buy signal
	secondDropBuyCount := 0
	for _, p := range prices {
		tick := marketdata.Tick{Symbol: "BTCUSDT", Price: formatFloat(p), Source: "binance"}
		signals, _ := s.OnMarketTick(context.Background(), tick)
		for _, sig := range signals {
			if sig.Action == "buy" {
				secondDropBuyCount++
			}
		}
	}
	if secondDropBuyCount != 1 {
		t.Errorf("expected exactly 1 buy signal on second drop after neutral zone reset, got %d", secondDropBuyCount)
	}
}
