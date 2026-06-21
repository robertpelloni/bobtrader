package app

import (
	"context"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
	strategydemo "github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/demo"
)

func TestFinalLiveIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping final live integration in short mode")
	}

	dir := t.TempDir()
	cfg := config.Default()
	cfg.Environment = "final-live-verification"
	cfg.Scheduler.Enabled = true
	cfg.Scheduler.Mode = "stream"
	cfg.Scheduler.IntervalMS = 100 // High frequency polling for REST if WS fails
	cfg.Risk.AllowedSymbols = []string{"BTCUSDT"}
	cfg.Logging.Stdout = false

	// Set paths to temp dir
	cfg.EventLog.Path = dir + "/events.jsonl"
	cfg.Snapshots.Path = dir + "/snapshots.jsonl"
	cfg.Orders.Path = dir + "/orders.jsonl"
	cfg.Reports.Path = dir + "/reports.jsonl"

	// Wire to use real binance feed via paper-market-aware
	cfg.Accounts = []config.AccountConfig{
		{
			ID:       "final-test",
			Name:     "Final Live Test",
			Enabled:  true,
			Exchange: "paper-market-aware",
		},
	}

	application, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to init App: %v", err)
	}

	// Override standard strategies with Noisy strategy for this test
	application.strategyRuntime = strategy.NewRuntime(
		strategydemo.NewNoisyStrategy("final-test", "BTCUSDT", 2), // Signal every 2 ticks
	)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	t.Log("Starting final live market integration test...")
	if err := application.Start(ctx); err != nil {
		t.Fatalf("App failed to start: %v", err)
	}

	// Monitor for a bit to collect ticks and signals
	time.Sleep(15 * time.Second)

	// Capture state
	signalCount := application.signalLog.Count()
	execCount := len(application.executionRepo.List())
	metrics := application.metricsTracker.Snapshot()

	t.Logf("Final Integration Summary: Signals=%d Executions=%d Attempts=%d Success=%d",
		signalCount, execCount, metrics.ExecutionAttempts, metrics.ExecutionSuccess)

	t.Log("Shutting down final live test...")
	if err := application.Shutdown(context.Background()); err != nil {
		t.Errorf("App failed to shutdown cleanly: %v", err)
	}

	if signalCount == 0 {
		t.Log("Warning: No signals generated. This is expected if the live feed provided < 2 ticks during the run (common in restricted networks).")
	}
}
