package runtime

import (
	"context"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/persistence/reports"
)

type stubStreamRunner struct{ calls atomic.Int32 }

func (r *stubStreamRunner) RunTick(ctx context.Context, tick marketdata.Tick) error {
	r.calls.Add(1)
	return nil
}
func (r *stubStreamRunner) RunCandle(ctx context.Context, candle marketdata.Candle) error {
	r.calls.Add(1)
	return nil
}

func TestReportingStreamRunnerWritesReports(t *testing.T) {
	store, err := reports.NewStore(filepath.Join(t.TempDir(), "reports.jsonl"))
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	runner := &stubStreamRunner{}
	reportingRunner := NewReportingStreamRunner(runner, runner, store, func(ctx context.Context) []reports.Report {
		return []reports.Report{{Type: "tick-summary", Timestamp: time.Now()}}
	})
	if err := reportingRunner.RunTick(context.Background(), marketdata.Tick{Symbol: "BTCUSDT"}); err != nil {
		t.Fatalf("RunTick returned error: %v", err)
	}
	if runner.calls.Load() != 1 {
		t.Fatalf("expected underlying runner to be called once for tick, got %d", runner.calls.Load())
	}

	if err := reportingRunner.RunCandle(context.Background(), marketdata.Candle{Symbol: "BTCUSDT"}); err != nil {
		t.Fatalf("RunCandle returned error: %v", err)
	}
	if runner.calls.Load() != 2 {
		t.Fatalf("expected underlying runner to be called twice, got %d", runner.calls.Load())
	}

	latest, err := store.LatestByType()
	if err != nil {
		t.Fatalf("LatestByType returned error: %v", err)
	}
	if _, ok := latest["tick-summary"]; !ok {
		t.Fatalf("expected tick-summary report, got %+v", latest)
	}
}
