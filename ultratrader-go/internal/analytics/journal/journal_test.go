package journal

import (
	"math"
	"testing"
	"time"
)

func TestJournal_Record(t *testing.T) {
	j := New()
	entry := j.Record(Entry{
		Symbol:   "BTCUSDT",
		Side:     Buy,
		Price:    50000,
		Quantity: 1,
		GroupID:  "G1",
	})

	if entry.ID == "" {
		t.Error("expected auto-generated ID")
	}
	if entry.Timestamp.IsZero() {
		t.Error("expected auto-generated timestamp")
	}

	entries := j.Entries()
	if len(entries) != 1 {
		t.Fatalf("expected 1 entry, got %d", len(entries))
	}
	if entries[0].Symbol != "BTCUSDT" {
		t.Errorf("expected BTCUSDT, got %s", entries[0].Symbol)
	}
}

func TestJournal_TradeGroup(t *testing.T) {
	j := New()

	j.Record(Entry{
		Symbol:    "BTCUSDT",
		Side:      Buy,
		Price:     50000,
		Quantity:  1,
		GroupID:   "G1",
		Strategy:  "SMA",
		Timestamp: time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
	})

	j.Record(Entry{
		Symbol:    "BTCUSDT",
		Side:      Sell,
		Price:     55000,
		Quantity:  1,
		GroupID:   "G1",
		Strategy:  "SMA",
		Timestamp: time.Date(2026, 1, 1, 14, 0, 0, 0, time.UTC),
	})

	groups := j.TradeGroups()
	if len(groups) != 1 {
		t.Fatalf("expected 1 trade group, got %d", len(groups))
	}

	g := groups[0]
	if g.Symbol != "BTCUSDT" {
		t.Errorf("expected BTCUSDT, got %s", g.Symbol)
	}
	if g.EntryPrice != 50000 {
		t.Errorf("expected entry 50000, got %f", g.EntryPrice)
	}
	if g.ExitPrice != 55000 {
		t.Errorf("expected exit 55000, got %f", g.ExitPrice)
	}
	expectedPnL := (55000.0 - 50000.0) * 1.0 // +5000
	if math.Abs(g.PnL-expectedPnL) > 0.01 {
		t.Errorf("expected PnL %f, got %f", expectedPnL, g.PnL)
	}
	if g.Duration != 4*time.Hour {
		t.Errorf("expected 4h duration, got %s", g.Duration)
	}
}

func TestJournal_Stats(t *testing.T) {
	j := New()

	// Create 3 winning trades and 2 losing trades
	for i := 0; i < 5; i++ {
		buyPrice := 10000.0
		sellPrice := 10500.0
		if i >= 3 {
			sellPrice = 9800.0 // Loss
		}

		j.Record(Entry{
			Symbol:    "BTCUSDT",
			Side:      Buy,
			Price:     buyPrice,
			Quantity:  1,
			GroupID:   "G" + string(rune('1'+i)),
			Timestamp: time.Date(2026, 1, 1, i*2, 0, 0, 0, time.UTC),
		})
		j.Record(Entry{
			Symbol:    "BTCUSDT",
			Side:      Sell,
			Price:     sellPrice,
			Quantity:  1,
			GroupID:   "G" + string(rune('1'+i)),
			Timestamp: time.Date(2026, 1, 1, i*2+1, 0, 0, 0, time.UTC),
		})
	}

	stats := j.Stats()
	if stats.TotalTrades != 5 {
		t.Errorf("expected 5 total trades, got %d", stats.TotalTrades)
	}
	if stats.WinningTrades != 3 {
		t.Errorf("expected 3 winning, got %d", stats.WinningTrades)
	}
	if stats.LosingTrades != 2 {
		t.Errorf("expected 2 losing, got %d", stats.LosingTrades)
	}
	winRate := float64(3) / float64(5)
	if math.Abs(stats.WinRate-winRate) > 0.01 {
		t.Errorf("expected win rate %f, got %f", winRate, stats.WinRate)
	}
	if stats.ProfitFactor <= 0 {
		t.Errorf("expected positive profit factor, got %f", stats.ProfitFactor)
	}
}

func TestJournal_StatsByStrategy(t *testing.T) {
	j := New()

	// Strategy A: 2 wins
	for i := 0; i < 2; i++ {
		j.Record(Entry{Symbol: "BTCUSDT", Side: Buy, Price: 100, Quantity: 1, GroupID: "A" + string(rune('1'+i)), Strategy: "Alpha"})
		j.Record(Entry{Symbol: "BTCUSDT", Side: Sell, Price: 110, Quantity: 1, GroupID: "A" + string(rune('1'+i)), Strategy: "Alpha"})
	}

	// Strategy B: 1 loss
	j.Record(Entry{Symbol: "ETHUSDT", Side: Buy, Price: 200, Quantity: 1, GroupID: "B1", Strategy: "Beta"})
	j.Record(Entry{Symbol: "ETHUSDT", Side: Sell, Price: 190, Quantity: 1, GroupID: "B1", Strategy: "Beta"})

	statsA := j.StatsByStrategy("Alpha")
	if statsA.TotalTrades != 2 {
		t.Errorf("Alpha: expected 2 trades, got %d", statsA.TotalTrades)
	}
	if statsA.WinningTrades != 2 {
		t.Errorf("Alpha: expected 2 wins, got %d", statsA.WinningTrades)
	}

	statsB := j.StatsByStrategy("Beta")
	if statsB.TotalTrades != 1 {
		t.Errorf("Beta: expected 1 trade, got %d", statsB.TotalTrades)
	}
	if statsB.LosingTrades != 1 {
		t.Errorf("Beta: expected 1 loss, got %d", statsB.LosingTrades)
	}
}

func TestJournal_StatsSince(t *testing.T) {
	j := New()

	baseTime := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	for i := 0; i < 3; i++ {
		j.Record(Entry{Symbol: "BTCUSDT", Side: Buy, Price: 100, Quantity: 1, GroupID: "S" + string(rune('1'+i)), Timestamp: baseTime.Add(time.Duration(i*24) * time.Hour)})
		j.Record(Entry{Symbol: "BTCUSDT", Side: Sell, Price: 110, Quantity: 1, GroupID: "S" + string(rune('1'+i)), Timestamp: baseTime.Add(time.Duration(i*24+1) * time.Hour)})
	}

	stats := j.StatsSince(baseTime.Add(26 * time.Hour))
	if stats.TotalTrades != 1 {
		t.Errorf("expected 1 trade since cutoff, got %d", stats.TotalTrades)
	}
}

func TestJournal_EmptyStats(t *testing.T) {
	j := New()
	stats := j.Stats()
	if stats.TotalTrades != 0 {
		t.Errorf("expected 0 trades for empty journal")
	}
	if stats.WinRate != 0 {
		t.Errorf("expected 0 win rate for empty journal")
	}
}

func TestMaxDrawdown_NoDrawdown(t *testing.T) {
	dd := MaxDrawdown([]float64{100, 100, 100})
	if dd != 0 {
		t.Errorf("expected 0 drawdown for monotonically increasing, got %f", dd)
	}
}

func TestMaxDrawdown_WithDrawdown(t *testing.T) {
	// Cumulative: 100, 300, 600, 750, 850 -> peak=850, min after peak = 850, so no DD
	// Let's use losses to create drawdown
	// PnLs: +100, +200, -200, -100 -> cumulative: 100, 300, 100, 0 -> DD=300
	dd := MaxDrawdown([]float64{100, 200, -200, -100})
	if math.Abs(dd-300) > 0.01 {
		t.Errorf("expected 300 drawdown, got %f", dd)
	}
}

func TestMaxDrawdown_Recovery(t *testing.T) {
	// PnLs: +100, -50, +70 -> cumulative: 100, 50, 120 -> DD = 50
	dd := MaxDrawdown([]float64{100, -50, 70})
	if math.Abs(dd-50) > 0.01 {
		t.Errorf("expected 50 drawdown, got %f", dd)
	}
}

func TestMaxDrawdown_Empty(t *testing.T) {
	dd := MaxDrawdown(nil)
	if dd != 0 {
		t.Errorf("expected 0 for empty input, got %f", dd)
	}
}
