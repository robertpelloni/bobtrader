package runtime

import (
	"context"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/persistence/reports"
)

type stubRunner struct{ calls atomic.Int32 }

func (r *stubRunner) RunOnce(ctx context.Context) error {
	r.calls.Add(1)
	return nil
}

func TestReportingRunnerWritesReports(t *testing.T) {
	store, err := reports.NewStore(filepath.Join(t.TempDir(), "reports.jsonl"))
	if err != nil {
		t.Fatalf("NewStore returned error: %v", err)
	}
	runner := &stubRunner{}
	reportingRunner := NewReportingRunner(runner, store, func(ctx context.Context) []reports.Report {
		return []reports.Report{{Type: "cycle-summary"}}
	})
	if err := reportingRunner.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce returned error: %v", err)
	}
	if runner.calls.Load() != 1 {
		t.Fatalf("expected underlying runner to be called once, got %d", runner.calls.Load())
	}
	latest, err := store.LatestByType()
	if err != nil {
		t.Fatalf("LatestByType returned error: %v", err)
	}
	if _, ok := latest["cycle-summary"]; !ok {
		t.Fatalf("expected cycle-summary report, got %+v", latest)
	}
}
