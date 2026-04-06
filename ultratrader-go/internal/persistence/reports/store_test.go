package reports

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendReport(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reports.jsonl")
	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	if err := store.Append(context.Background(), Report{Type: "startup", Payload: map[string]any{"ok": true}}); err != nil {
		t.Fatalf("Append returned error: %v", err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if !strings.Contains(string(data), "startup") {
		t.Fatalf("expected report content, got %q", string(data))
	}
}

func TestLatestAndLatestByType(t *testing.T) {
	path := filepath.Join(t.TempDir(), "reports.jsonl")
	store, _ := NewStore(path)
	_ = store.Append(context.Background(), Report{Type: "metrics", Payload: map[string]any{"a": 1}})
	_ = store.Append(context.Background(), Report{Type: "valuation", Payload: map[string]any{"v": 2}})
	_ = store.Append(context.Background(), Report{Type: "metrics", Payload: map[string]any{"a": 3}})

	latest, err := store.Latest(2)
	if err != nil {
		t.Fatalf("Latest returned error: %v", err)
	}
	if len(latest) != 2 {
		t.Fatalf("expected 2 latest reports, got %d", len(latest))
	}
	byType, err := store.LatestByType()
	if err != nil {
		t.Fatalf("LatestByType returned error: %v", err)
	}
	if byType["metrics"].Payload["a"].(float64) != 3 {
		t.Fatalf("expected latest metrics payload 3, got %+v", byType["metrics"].Payload)
	}
	filtered, err := store.ListByType("metrics", 10)
	if err != nil {
		t.Fatalf("ListByType returned error: %v", err)
	}
	if len(filtered) != 2 {
		t.Fatalf("expected 2 filtered reports, got %d", len(filtered))
	}
}
