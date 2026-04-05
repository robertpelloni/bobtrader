package snapshot

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Snapshot struct {
	Timestamp   time.Time      `json:"timestamp"`
	AccountID   string         `json:"account_id"`
	AccountName string         `json:"account_name"`
	Exchange    string         `json:"exchange"`
	Balances    map[string]any `json:"balances,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

type Store struct {
	path string
	mu   sync.Mutex
}

func NewStore(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("snapshot path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create snapshot directory: %w", err)
	}
	return &Store{path: path}, nil
}

func (s *Store) Append(_ context.Context, snapshot Snapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if snapshot.Timestamp.IsZero() {
		snapshot.Timestamp = time.Now().UTC()
	}
	if snapshot.AccountID == "" {
		return fmt.Errorf("account id is required")
	}

	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open snapshot store: %w", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(snapshot); err != nil {
		return fmt.Errorf("encode snapshot: %w", err)
	}
	return nil
}

func (s *Store) Path() string { return s.path }
