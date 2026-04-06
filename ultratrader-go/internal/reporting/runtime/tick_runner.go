package runtime

import (
	"context"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/persistence/reports"
)

type tickRunner interface {
	RunTick(ctx context.Context, tick marketdata.Tick) error
}

type ReportingTickRunner struct {
	runner   tickRunner
	store    *reports.Store
	provider ReportProvider
}

func NewReportingTickRunner(r tickRunner, store *reports.Store, provider ReportProvider) *ReportingTickRunner {
	return &ReportingTickRunner{runner: r, store: store, provider: provider}
}

func (r *ReportingTickRunner) RunTick(ctx context.Context, tick marketdata.Tick) error {
	if err := r.runner.RunTick(ctx, tick); err != nil {
		return err
	}
	if r.store == nil || r.provider == nil {
		return nil
	}
	for _, report := range r.provider(ctx) {
		if err := r.store.Append(ctx, report); err != nil {
			return err
		}
	}
	return nil
}
