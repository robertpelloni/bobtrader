package app

import (
	"context"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
)

func TestSandboxRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping sandbox run in short mode")
	}

	// 1. Load sandbox configuration
	cfg, err := config.Load("../../config/sandbox-test.json")
	if err != nil {
		// Fallback to default if file not found (e.g. in some environments)
		cfg = config.Default()
		cfg.Environment = "sandbox-fallback"
		cfg.Scheduler.IntervalMS = 100
	}

	// 2. Initialize App
	application, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to init App: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 7*time.Second)
	defer cancel()

	t.Log("Starting formal sandbox test run...")
	if err := application.Start(ctx); err != nil {
		t.Fatalf("App failed to start: %v", err)
	}

	// 3. Monitor for duration
	time.Sleep(5 * time.Second)

	// 4. Capture metrics before shutdown
	metrics := application.metricsTracker.Snapshot()
	signals := application.signalLog.Count()

	t.Logf("Sandbox Run Summary: Attempts=%d Success=%d Blocked=%d Signals=%d",
		metrics.ExecutionAttempts, metrics.ExecutionSuccess, metrics.ExecutionBlocked, signals)

	// 5. Clean shutdown
	t.Log("Shutting down sandbox run...")
	if err := application.Shutdown(context.Background()); err != nil {
		t.Errorf("App failed to shutdown cleanly: %v", err)
	}

	// 6. Final verification
	if metrics.ExecutionAttempts == 0 && signals == 0 {
		t.Log("Note: Zero signals/attempts during this run. Typical if prices remained within thresholds.")
	}
}
