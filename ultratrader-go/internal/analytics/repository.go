package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// PerformanceSnapshot captures the state of a strategy's performance at a point in time.
type PerformanceSnapshot struct {
	Timestamp    time.Time                         `json:"timestamp"`
	StrategyStats map[string]strategy.StrategyStats `json:"strategy_stats"`
	TotalPnL     float64                           `json:"total_pnl"`
	Siphoned     float64                           `json:"siphoned"`
}

// Repository handles persistent storage of performance snapshots.
type Repository struct {
	mu   sync.Mutex
	path string
}

// NewRepository creates a new performance repository.
func NewRepository(path string) (*Repository, error) {
	if path == "" {
		path = "data/analytics/performance.jsonl"
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create analytics dir: %w", err)
	}

	return &Repository{path: path}, nil
}

// Save appends a performance snapshot to the persistent store.
func (r *Repository) Save(ctx context.Context, snapshot PerformanceSnapshot) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	f, err := os.OpenFile(r.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open performance file: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("marshal performance snapshot: %w", err)
	}

	if _, err := f.Write(append(data, '\n')); err != nil {
		return fmt.Errorf("write performance snapshot: %w", err)
	}

	return nil
}
