package app

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
)

func TestAppStartWritesEventSnapshotOrderLogAndReport(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Default()
	cfg.Server.Enabled = false
	cfg.Scheduler.Enabled = false
	cfg.EventLog.Path = filepath.Join(dir, "events.jsonl")
	cfg.Snapshots.Path = filepath.Join(dir, "snapshots.jsonl")
	cfg.Orders.Path = filepath.Join(dir, "orders.jsonl")
	cfg.Reports.Path = filepath.Join(dir, "runtime.jsonl")
	cfg.Logging.Path = filepath.Join(dir, "app.jsonl")
	cfg.Logging.Stdout = false

	application, err := New(cfg)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	defer func() { _ = application.Shutdown(context.Background()) }()
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

	reports, err := os.ReadFile(cfg.Reports.Path)
	if err != nil {
		t.Fatalf("read report log: %v", err)
	}
	if !strings.Contains(string(reports), "startup-summary") || !strings.Contains(string(reports), "portfolio-valuation") || !strings.Contains(string(reports), "metrics-snapshot") {
		t.Fatalf("expected runtime reports, got %q", string(reports))
	}

	logs, err := os.ReadFile(cfg.Logging.Path)
	if err != nil {
		t.Fatalf("read app log: %v", err)
	}
	if !strings.Contains(string(logs), "app startup completed") || !strings.Contains(string(logs), "realized_pnl") || !strings.Contains(string(logs), "execution_attempts") || !strings.Contains(string(logs), "symbol-whitelist") {
		t.Fatalf("expected startup completion log with pnl, metrics, and guards fields, got %q", string(logs))
	}
}

func TestAppShutdownWithServer(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Default()
	cfg.Server.Enabled = true
	cfg.Scheduler.Enabled = false
	cfg.Logging.Stdout = false
	cfg.Logging.Path = filepath.Join(dir, "app.jsonl")
	cfg.Reports.Path = filepath.Join(dir, "runtime.jsonl")
	cfg.EventLog.Path = filepath.Join(dir, "events.jsonl")
	cfg.Snapshots.Path = filepath.Join(dir, "snapshots.jsonl")
	cfg.Orders.Path = filepath.Join(dir, "orders.jsonl")
	application, err := New(cfg)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := application.Start(ctx); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	time.Sleep(20 * time.Millisecond)
	resp, err := http.Get("http://" + application.Address() + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz failed: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK || len(body) == 0 {
		t.Fatalf("unexpected health response: status=%d body=%q", resp.StatusCode, string(body))
	}
	cancel()
	if err := application.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown returned error: %v", err)
	}
}
