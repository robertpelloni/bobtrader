package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
)

func TestAppStartWritesEventAndSnapshot(t *testing.T) {
	dir := t.TempDir()
	cfg := config.Default()
	cfg.EventLog.Path = filepath.Join(dir, "events.jsonl")
	cfg.Snapshots.Path = filepath.Join(dir, "snapshots.jsonl")

	application, err := New(cfg)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	if err := application.Start(context.Background()); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}

	events, err := os.ReadFile(cfg.EventLog.Path)
	if err != nil {
		t.Fatalf("read event log: %v", err)
	}
	if !strings.Contains(string(events), "app.started") {
		t.Fatalf("expected app.started event, got %q", string(events))
	}

	snapshots, err := os.ReadFile(cfg.Snapshots.Path)
	if err != nil {
		t.Fatalf("read snapshot log: %v", err)
	}
	if !strings.Contains(string(snapshots), "paper-main") {
		t.Fatalf("expected paper-main snapshot, got %q", string(snapshots))
	}
}
