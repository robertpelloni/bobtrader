package orders

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type Record struct {
	Timestamp time.Time      `json:"timestamp"`
	AccountID string         `json:"account_id"`
	Exchange  string         `json:"exchange"`
	OrderID   string         `json:"order_id"`
	Symbol    string         `json:"symbol"`
	Side      string         `json:"side"`
	Type      string         `json:"type"`
	Status    string         `json:"status"`
	Quantity  string         `json:"quantity"`
	Price     string         `json:"price"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type Store struct {
	path string
	mu   sync.Mutex
}

func NewStore(path string) (*Store, error) {
	if path == "" {
		return nil, fmt.Errorf("orders path is empty")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create order store directory: %w", err)
	}
	return &Store{path: path}, nil
}

func (s *Store) Append(_ context.Context, record Record) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if record.Timestamp.IsZero() {
		record.Timestamp = time.Now().UTC()
	}
	if record.AccountID == "" {
		return fmt.Errorf("account id is required")
	}
	if record.OrderID == "" {
		return fmt.Errorf("order id is required")
	}

	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("open order store: %w", err)
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(record); err != nil {
		return fmt.Errorf("encode order record: %w", err)
	}
	return nil
}

func (s *Store) Path() string { return s.path }
