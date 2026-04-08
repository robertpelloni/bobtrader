package runtime

import (
	"context"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/persistence/reports"
)

type tickRunner interface {
	RunTick(ctx context.Context, tick marketdata.Tick) error
}

type candleRunner interface {
	RunCandle(ctx context.Context, candle marketdata.Candle) error
}

type ReportingStreamRunner struct {
	tickRunner   tickRunner
	candleRunner candleRunner
	store        *reports.Store
	provider     ReportProvider
}

func NewReportingStreamRunner(tr tickRunner, cr candleRunner, store *reports.Store, provider ReportProvider) *ReportingStreamRunner {
	return &ReportingStreamRunner{tickRunner: tr, candleRunner: cr, store: store, provider: provider}
}

func (r *ReportingStreamRunner) RunTick(ctx context.Context, tick marketdata.Tick) error {
	if r.tickRunner == nil {
		return nil
	}
	if err := r.tickRunner.RunTick(ctx, tick); err != nil {
		return err
	}
	return r.record(ctx)
}

func (r *ReportingStreamRunner) RunCandle(ctx context.Context, candle marketdata.Candle) error {
	if r.candleRunner == nil {
		return nil
	}
	if err := r.candleRunner.RunCandle(ctx, candle); err != nil {
		return err
	}
	return r.record(ctx)
}

func (r *ReportingStreamRunner) record(ctx context.Context) error {
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
