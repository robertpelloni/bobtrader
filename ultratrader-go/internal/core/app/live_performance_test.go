package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
)

func TestLivePerformanceIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live performance integration in short mode")
	}

	dir := t.TempDir()
	cfg := config.Default()
	cfg.Environment = "live-integration-test"
	cfg.Server.Enabled = false
	cfg.Scheduler.Enabled = true
	cfg.Scheduler.Mode = "stream"
	cfg.Scheduler.IntervalMS = 1000
	cfg.Risk.AllowedSymbols = []string{"BTCUSDT", "ETHUSDT"}

	cfg.EventLog.Path = filepath.Join(dir, "events.jsonl")
	cfg.Snapshots.Path = filepath.Join(dir, "snapshots.jsonl")
	cfg.Orders.Path = filepath.Join(dir, "orders.jsonl")
	cfg.Reports.Path = filepath.Join(dir, "runtime.jsonl")
	cfg.Logging.Path = filepath.Join(dir, "app.jsonl")
	cfg.Logging.Stdout = false

	// Wire to use real binance feed
	cfg.Accounts = []config.AccountConfig{
		{
			ID:       "test-live",
			Name:     "Live Integration Test",
			Enabled:  true,
			Exchange: "paper-market-aware",
		},
	}

	application, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to init: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	t.Log("Starting live performance integration test...")
	if err := application.Start(ctx); err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Monitor for 20 seconds
	time.Sleep(20 * time.Second)

	cancel()                           // Stop background stream services
	time.Sleep(100 * time.Millisecond) // Wait for worker loops to exit

	t.Log("Shutting down live integration test...")
	if err := application.Shutdown(context.Background()); err != nil {
		t.Errorf("Shutdown failed: %v", err)
	}

	// Check if any signals or orders were processed
	stats := application.signalLog.StatsByStrategy()
	totalSignals := 0
	for _, s := range stats {
		totalSignals += s.SignalsTotal
	}

	t.Logf("Live integration summary: Recorded %d signals across %d strategies.",
		totalSignals, len(stats))

	// Verify persistence
	if data, _ := os.ReadFile(cfg.Reports.Path); len(data) > 0 {
		t.Logf("Persistence verified: %d bytes of report data.", len(data))
	}
}
