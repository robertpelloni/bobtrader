package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
)

func TestControlledPaperRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping controlled paper run in short mode")
	}

	dir := t.TempDir()
	cfg := config.Default()
	cfg.Environment = "controlled-paper-run"
	cfg.Server.Enabled = false
	cfg.Scheduler.Enabled = true
	cfg.Scheduler.Mode = "stream"
	cfg.Scheduler.IntervalMS = 1000 // 1s polling
	cfg.Risk.AllowedSymbols = []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"}

	cfg.EventLog.Path = filepath.Join(dir, "events.jsonl")
	cfg.Snapshots.Path = filepath.Join(dir, "snapshots.jsonl")
	cfg.Orders.Path = filepath.Join(dir, "orders.jsonl")
	cfg.Reports.Path = filepath.Join(dir, "runtime.jsonl")
	cfg.Logging.Path = filepath.Join(dir, "app.jsonl")
	cfg.Logging.Stdout = false

	// Wire to use real binance feed
	cfg.Accounts = []config.AccountConfig{
		{
			ID: "controlled-test",
			Name: "Controlled Paper Run",
			Enabled: true,
			Exchange: "paper-market-aware",
		},
	}

	application, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to init: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
	defer cancel()

	t.Log("Starting 2-minute controlled paper run with real market data...")
	start := time.Now()
	if err := application.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Monitor for 120 seconds
	time.Sleep(120 * time.Second)
	duration := time.Since(start)

	t.Log("Shutting down controlled run...")
	if err := application.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Analyze results
	stats := application.signalLog.StatsByStrategy()
	totalSignals := 0
	for _, s := range stats {
		totalSignals += s.SignalsTotal
	}

	orders := application.executionRepo.List()
	pnl := application.portfolioTracker.TotalRealizedPnL()

	t.Logf("Controlled Run Results (Duration: %v):", duration)
	t.Logf("  Total Signals: %d", totalSignals)
	t.Logf("  Total Orders: %d", len(orders))
	t.Logf("  Realized PnL: %.4f", pnl)
	t.Logf("  Strategies Active: %d", len(stats))

	for name, s := range stats {
		t.Logf("    Strategy %s: Signals=%d SuccessRate=%.2f%%", name, s.SignalsTotal, s.SuccessRate*100)
	}

	// Verify persistence
	if data, _ := os.ReadFile(cfg.Reports.Path); len(data) > 0 {
		t.Logf("Persistence verified: %d bytes of report data.", len(data))
	}
}
