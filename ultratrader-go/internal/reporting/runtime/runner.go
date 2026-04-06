package runtime

import (
	"context"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/persistence/reports"
)

type runner interface {
	RunOnce(ctx context.Context) error
}

type ReportProvider func(context.Context) []reports.Report

type ReportingRunner struct {
	runner   runner
	store    *reports.Store
	provider ReportProvider
}

func NewReportingRunner(r runner, store *reports.Store, provider ReportProvider) *ReportingRunner {
	return &ReportingRunner{runner: r, store: store, provider: provider}
}

func (r *ReportingRunner) RunOnce(ctx context.Context) error {
	if err := r.runner.RunOnce(ctx); err != nil {
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
