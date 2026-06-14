package app

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
)

func TestSystemSimulation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping system simulation in short mode")
	}

	dir := t.TempDir()
	cfg := config.Default()
	cfg.Environment = "system-test"
	cfg.Server.Enabled = false
	cfg.Scheduler.Enabled = true
	cfg.Scheduler.Mode = "stream"
	cfg.Scheduler.IntervalMS = 100
	cfg.Risk.AllowedSymbols = []string{"BTCUSDT", "ETHUSDT"}
	cfg.EventLog.Path = filepath.Join(dir, "events.jsonl")
	cfg.Snapshots.Path = filepath.Join(dir, "snapshots.jsonl")
	cfg.Orders.Path = filepath.Join(dir, "orders.jsonl")
	cfg.Reports.Path = filepath.Join(dir, "runtime.jsonl")
	cfg.Logging.Path = filepath.Join(dir, "app.jsonl")
	cfg.Logging.Stdout = false

	application, err := New(cfg)
	if err != nil {
		t.Fatalf("Failed to create App: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	t.Log("Starting system simulation...")
	if err := application.Start(ctx); err != nil {
		t.Fatalf("App failed to start: %v", err)
	}

	// Verify API connectivity if address is available (not available in this test as server is disabled)
	// Instead, we verify internal providers
	if application.httpHandler == nil {
		t.Error("HTTP handler not initialized")
	}

	// Wait for several scheduler cycles
	time.Sleep(1500 * time.Millisecond)

	cancel()                           // Stop background stream services
	time.Sleep(100 * time.Millisecond) // Wait for worker loops to exit

	t.Log("Stopping system simulation...")
	if err := application.Shutdown(context.Background()); err != nil {
		t.Errorf("App failed to shutdown cleanly: %v", err)
	}

	// Verify persistence artifacts
	verifyFileExists(t, cfg.EventLog.Path, "Event Log")
	verifyFileExists(t, cfg.Snapshots.Path, "Snapshot Store")
	verifyFileExists(t, cfg.Orders.Path, "Order Store")
	verifyFileExists(t, cfg.Reports.Path, "Report Store")
	verifyFileExists(t, cfg.Logging.Path, "App Log")

	// Check if orders were actually generated during simulation
	ordersData, _ := os.ReadFile(cfg.Orders.Path)
	if len(ordersData) > 0 {
		t.Logf("Simulation successful: %d bytes of order data recorded", len(ordersData))
	} else {
		t.Log("Simulation completed with no orders placed (typical for short runs without signal triggers)")
	}

	// Verify signal log
	if application.signalLog.Count() > 0 {
		t.Logf("Recorded %d strategy signals during simulation", application.signalLog.Count())
	}

	// Explicitly close application and logs to ensure file handles are released
	application.Close()
	time.Sleep(100 * time.Millisecond) // Give OS a moment to release handles
}

func verifyFileExists(t *testing.T, path string, name string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("%s was not created at %s", name, path)
	}
}
