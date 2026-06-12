package app

import (
	"context"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
	strategydemo "github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/demo"
)

func TestPerformanceStress(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping performance stress test in short mode")
	}

	dir := t.TempDir()
	cfg := config.Default()
	cfg.Environment = "performance-stress-test"
	cfg.Scheduler.Enabled = true
	cfg.Scheduler.Mode = "stream"
	cfg.Scheduler.IntervalMS = 10 // ultra tight
	cfg.Risk.AllowedSymbols = []string{"BTCUSDT", "ETHUSDT", "SOLUSDT"}
	cfg.Logging.Stdout = false

	cfg.EventLog.Path = dir + "/events.jsonl"
	cfg.Snapshots.Path = dir + "/snapshots.jsonl"
	cfg.Orders.Path = dir + "/orders.jsonl"
	cfg.Reports.Path = dir + "/reports.jsonl"

	cfg.Accounts = []config.AccountConfig{
		{
			ID: "stress-account",
			Name: "Stress Test Account",
			Enabled: true,
			Exchange: "paper-market-aware",
		},
	}

	application, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to init App: %v", err)
	}

	// Create high-frequency noise for multiple symbols
	// interval=1 means signal on every tick
	newRuntime := strategy.NewRuntime(
		strategydemo.NewNoisyStrategy("stress-account", "BTCUSDT", 1),
		strategydemo.NewNoisyStrategy("stress-account", "ETHUSDT", 1),
		strategydemo.NewNoisyStrategy("stress-account", "SOLUSDT", 1),
	)
	application.strategyRuntime = newRuntime
	application.scheduler.SetRuntime(newRuntime)

	ctx, cancel := context.WithTimeout(context.Background(), 70*time.Second)
	defer cancel()

	t.Log("Starting 60-second performance stress test...")
	start := time.Now()
	if err := application.Start(ctx); err != nil {
		t.Fatalf("App failed to start: %v", err)
	}

	// Stress for 60 seconds
	time.Sleep(60 * time.Second)
	duration := time.Since(start)

	cancel()                           // Stop background stream services
	time.Sleep(100 * time.Millisecond) // Wait for worker loops to exit

	t.Log("Shutting down stress test...")
	if err := application.Shutdown(context.Background()); err != nil {
		t.Errorf("App failed to shutdown cleanly: %v", err)
	}

	// Capture metrics
	signalCount := application.signalLog.Count()
	execCount := len(application.executionRepo.List())
	metrics := application.metricsTracker.Snapshot()

	t.Logf("Stress Test Results (Duration: %v):", duration)
	t.Logf("Total Signals: %d", signalCount)
	t.Logf("Total Orders Executed: %d", execCount)
	t.Logf("Execution Success Rate: %.2f%%", float64(metrics.ExecutionSuccess)/float64(metrics.ExecutionAttempts)*100)
	t.Logf("Throughput: %.2f orders/minute", float64(execCount)/(duration.Minutes()))

	if execCount < 5 {
		t.Errorf("Low execution count (%d), expected high throughput stress", execCount)
	}

	// Explicitly close application and logs to ensure file handles are released
	application.Close()
	time.Sleep(100 * time.Millisecond) // Give OS a moment to release handles
}
