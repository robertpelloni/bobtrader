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
