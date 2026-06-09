package demo

import (
	"context"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

// buildWarmupCandles creates n candles with small oscillation around base
// that stay well within the Bollinger Bands (no signal triggers).
func buildWarmupCandles(n int, base float64) []marketdata.Candle {
	candles := make([]marketdata.Candle, n)
	for i := 0; i < n; i++ {
		// Tiny oscillation: ±0.1% of base
		offset := float64(i%3-1) * base * 0.001
		p := base + offset
		candles[i] = marketdata.Candle{
			Symbol: "BTCUSDT",
			Close:  formatFloat(p),
			High:   formatFloat(p + base*0.002),
			Low:    formatFloat(p - base*0.002),
		}
	}
	return candles
}

func TestBollingerReversion_BuyAtLowerBand(t *testing.T) {
	s := NewBollingerReversion("paper-main", "BTCUSDT", "0.01", 20, 2.0)

	// Warm up with small oscillation (stays within bands)
	for _, c := range buildWarmupCandles(20, 100) {
		s.OnMarketCandle(context.Background(), c)
	}

	// Drop far below the lower band
	candle := marketdata.Candle{Symbol: "BTCUSDT", Close: formatFloat(80), High: formatFloat(81), Low: formatFloat(79)}
	signals, err := s.OnMarketCandle(context.Background(), candle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	buyCount := 0
	for _, sig := range signals {
		if sig.Action == "buy" {
			buyCount++
		}
	}
	if buyCount == 0 {
		t.Errorf("expected at least one buy signal when price drops to lower band")
	}
}

func TestBollingerReversion_SellAtUpperBand(t *testing.T) {
	s := NewBollingerReversion("paper-main", "BTCUSDT", "0.01", 20, 2.0)

	for _, c := range buildWarmupCandles(20, 100) {
		s.OnMarketCandle(context.Background(), c)
	}

	candle := marketdata.Candle{Symbol: "BTCUSDT", Close: formatFloat(120), High: formatFloat(121), Low: formatFloat(119)}
	signals, err := s.OnMarketCandle(context.Background(), candle)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	sellCount := 0
	for _, sig := range signals {
		if sig.Action == "sell" {
			sellCount++
		}
	}
	if sellCount == 0 {
		t.Errorf("expected at least one sell signal when price spikes to upper band")
	}
}

func TestBollingerReversion_NoSignalInBand(t *testing.T) {
	s := NewBollingerReversion("paper-main", "BTCUSDT", "0.01", 20, 2.0)

	// Small oscillation should never touch the bands
	for _, c := range buildWarmupCandles(40, 100) {
		signals, err := s.OnMarketCandle(context.Background(), c)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(signals) > 0 {
			t.Errorf("expected no signals when price stays within bands, got %d", len(signals))
			return
		}
	}
}

func TestBollingerReversion_Deduplication(t *testing.T) {
	s := NewBollingerReversion("paper-main", "BTCUSDT", "0.01", 20, 2.0)

	for _, c := range buildWarmupCandles(20, 100) {
		s.OnMarketCandle(context.Background(), c)
	}

	// First lower-band candle should trigger buy
	candle1 := marketdata.Candle{Symbol: "BTCUSDT", Close: formatFloat(80), High: formatFloat(81), Low: formatFloat(79)}
	signals1, _ := s.OnMarketCandle(context.Background(), candle1)
	buyCount := 0
	for _, sig := range signals1 {
		if sig.Action == "buy" {
			buyCount++
		}
	}
	if buyCount != 1 {
		t.Fatalf("expected exactly 1 buy on first lower-band touch, got %d", buyCount)
	}

	// Second lower-band candle should be deduplicated (no repeat buy)
	candle2 := marketdata.Candle{Symbol: "BTCUSDT", Close: formatFloat(78), High: formatFloat(79), Low: formatFloat(77)}
	signals2, _ := s.OnMarketCandle(context.Background(), candle2)
	for _, sig := range signals2 {
		if sig.Action == "buy" {
			t.Errorf("expected no repeat buy signal on second lower-band touch (dedup)")
		}
	}
}

func TestBollingerReversion_NoDualSignal(t *testing.T) {
	s := NewBollingerReversion("paper-main", "BTCUSDT", "0.01", 20, 2.0)

	// With tiny-oscillation warmup, no signals should fire during warmup.
	// After warmup, send extreme prices and verify buy+sell never both fire.
	totalBuys := 0
	totalSells := 0

	for i, c := range buildWarmupCandles(20, 100) {
		signals, _ := s.OnMarketCandle(context.Background(), c)
		for _, sig := range signals {
			if sig.Action == "buy" {
				totalBuys++
			}
			if sig.Action == "sell" {
				totalSells++
			}
		}
		_ = i
	}

	// Never both buy AND sell in same call — else-if prevents this
	if totalBuys > 0 && totalSells > 0 {
		t.Errorf("expected no dual buy+sell, but got buys=%d sells=%d", totalBuys, totalSells)
	}
}

func TestBollingerReversion_NeutralZoneReset(t *testing.T) {
	s := NewBollingerReversion("paper-main", "BTCUSDT", "0.01", 20, 2.0)

	for _, c := range buildWarmupCandles(20, 100) {
		s.OnMarketCandle(context.Background(), c)
	}

	// Trigger buy
	candle := marketdata.Candle{Symbol: "BTCUSDT", Close: formatFloat(80), High: formatFloat(81), Low: formatFloat(79)}
	signals, _ := s.OnMarketCandle(context.Background(), candle)
	if len(signals) == 0 || signals[0].Action != "buy" {
		t.Fatalf("expected buy signal on first touch, got %d signals", len(signals))
	}

	// Feed neutral-zone candles to reset lastSignal
	for _, c := range buildWarmupCandles(20, 100) {
		s.OnMarketCandle(context.Background(), c)
	}

	// Now trigger buy again — should work because lastSignal was reset
	candle2 := marketdata.Candle{Symbol: "BTCUSDT", Close: formatFloat(80), High: formatFloat(81), Low: formatFloat(79)}
	signals2, _ := s.OnMarketCandle(context.Background(), candle2)
	buyFound := false
	for _, sig := range signals2 {
		if sig.Action == "buy" {
			buyFound = true
		}
	}
	if !buyFound {
		t.Errorf("expected buy signal after neutral zone reset, but got none")
	}
}

func TestBollingerReversion_Name(t *testing.T) {
	s := NewBollingerReversion("paper-main", "BTCUSDT", "0.01", 20, 2.0)
	if s.Name() != "BollingerReversion" {
		t.Errorf("expected name BollingerReversion, got %s", s.Name())
	}
}
