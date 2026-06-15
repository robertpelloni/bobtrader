package analytics

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

func TestRepository_Save(t *testing.T) {
	path := "test_performance.jsonl"
	defer os.Remove(path)

	repo, err := NewRepository(path)
	if err != nil {
		t.Fatalf("failed to create repo: %v", err)
	}

	snapshot := PerformanceSnapshot{
		Timestamp: time.Now(),
		StrategyStats: map[string]strategy.StrategyStats{
			"test-strat": {WinRate: 0.75, WinTrades: 3, LossTrades: 1},
		},
		TotalPnL: 100.50,
		Siphoned: 10.05,
	}

	if err := repo.Save(context.Background(), snapshot); err != nil {
		t.Fatalf("failed to save snapshot: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("expected file %s to exist", path)
	}
}
