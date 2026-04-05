package eventlog

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Entry struct {
	Timestamp time.Time      `json:"timestamp"`
	Type      string         `json:"type"`
	Source    string         `json:"source"`
	Payload   map[string]any `json:"payload,omitempty"`
}

type Log struct {
	path string
	mu   sync.Mutex
}

func New(path string) (*Log, error) {
	if path == "" {
		return nil, fmt.Errorf("event log path is empty")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create event log directory: %w", err)
	}

	return &Log{path: path}, nil
}

func (l *Log) Append(_ context.Context, entry Entry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if entry.Timestamp.IsZero() {
		entry.Timestamp = time.Now().UTC()
	}

	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open event log: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	if err := enc.Encode(entry); err != nil {
		return fmt.Errorf("encode event entry: %w", err)
	}

	return nil
}

func (l *Log) Path() string {
	return l.path
}
