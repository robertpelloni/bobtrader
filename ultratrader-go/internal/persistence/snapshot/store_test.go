package snapshot

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendSnapshot(t *testing.T) {
	path := filepath.Join(t.TempDir(), "snapshots.jsonl")
	store, err := NewStore(path)
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}

	err = store.Append(context.Background(), Snapshot{AccountID: "paper-main", AccountName: "Paper Main", Exchange: "paper"})
	if err != nil {
		t.Fatalf("Append returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile returned error: %v", err)
	}
	if !strings.Contains(string(data), "paper-main") {
		t.Fatalf("expected snapshot content, got %q", string(data))
	}
}
