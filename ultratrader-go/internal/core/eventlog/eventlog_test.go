package eventlog

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppendWritesJSONL(t *testing.T) {
	path := filepath.Join(t.TempDir(), "events.jsonl")
	log, err := New(path)
	if err != nil {
		t.Fatalf("New returned error: %v", err)
	}

	err = log.Append(context.Background(), Entry{
		Type:   "app.started",
		Source: "test",
		Payload: map[string]any{
			"ok": true,
		},
	})
	if err != nil {
		t.Fatalf("Append returned error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read event log: %v", err)
	}

	text := string(data)
	if !strings.Contains(text, "app.started") {
		t.Fatalf("expected event type in log, got: %s", text)
	}
	if !strings.Contains(text, "\n") {
		t.Fatalf("expected newline-delimited json, got: %q", text)
	}
}
