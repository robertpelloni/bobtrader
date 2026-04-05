package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
)

func TestAppStartWritesEventSnapshotOrderAndLog(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Default()
	cfg.Server.Enabled = false
	cfg.Scheduler.Enabled = false
	cfg.EventLog.Path = filepath.Join(dir, "events.jsonl")
	cfg.Snapshots.Path = filepath.Join(dir, "snapshots.jsonl")
	cfg.Orders.Path = filepath.Join(dir, "orders.jsonl")
	cfg.Logging.Path = filepath.Join(dir, "app.jsonl")
	cfg.Logging.Stdout = false

	application, err := New(cfg)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	defer func() { _ = application.Close() }()
	if err := application.Start(context.Background()); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	events, err := os.ReadFile(cfg.EventLog.Path)
	if err != nil {
		t.Fatalf("read event log: %v", err)
	}
	if !strings.Contains(string(events), "app.started") || !strings.Contains(string(events), "execution.order_placed") {
		t.Fatalf("expected startup and execution events, got %q", string(events))
	}

	snapshots, err := os.ReadFile(cfg.Snapshots.Path)
	if err != nil {
		t.Fatalf("read snapshot log: %v", err)
	}
	if !strings.Contains(string(snapshots), "paper-main") {
		t.Fatalf("expected paper-main snapshot, got %q", string(snapshots))
	}

	orders, err := os.ReadFile(cfg.Orders.Path)
	if err != nil {
		t.Fatalf("read order log: %v", err)
	}
	if !strings.Contains(string(orders), "BTCUSDT") || !strings.Contains(string(orders), "correlation_id") {
		t.Fatalf("expected BTCUSDT order and correlation id, got %q", string(orders))
	}

	logs, err := os.ReadFile(cfg.Logging.Path)
	if err != nil {
		t.Fatalf("read app log: %v", err)
	}
	if !strings.Contains(string(logs), "app startup completed") || !strings.Contains(string(logs), "realized_pnl") || !strings.Contains(string(logs), "execution_attempts") {
		t.Fatalf("expected startup completion log with pnl and metrics fields, got %q", string(logs))
	}
}
