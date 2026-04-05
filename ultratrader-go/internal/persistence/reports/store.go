package reports

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Report struct {
	Timestamp time.Time      `json:"timestamp"`
	Type      string         `json:"type"`
	Payload   map[string]any `json:"payload,omitempty"`
}

type Store struct {
	path string
	mu   sync.Mutex
}

func NewStore(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("report path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create report store directory: %w", err)
	}
	return &Store{path: path}, nil
}

func (s *Store) Append(_ context.Context, report Report) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if report.Timestamp.IsZero() {
		report.Timestamp = time.Now().UTC()
	}
	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open report store: %w", err)
	}
	defer f.Close()
	if err := json.NewEncoder(f).Encode(report); err != nil {
		return fmt.Errorf("encode report: %w", err)
	}
	return nil
}

func (s *Store) Path() string { return s.path }
