package composite

import (
	"context"
	"fmt"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// RegimeFilter wraps a strategy and suppresses its signals if they conflict
// with the current macro market regime provided by a SignalEvaluator.
type RegimeFilter struct {
	mu           sync.RWMutex
	inner        strategy.Strategy
	macro        SignalEvaluator
	lastRegime   SignalResult
	suppressNone bool // If true, suppress all signals if macro is SignalNone
}

func NewRegimeFilter(inner strategy.Strategy, macro SignalEvaluator, suppressNone bool) *RegimeFilter {
	return &RegimeFilter{
		inner:        inner,
		macro:        macro,
		suppressNone: suppressNone,
	}
}

func (f *RegimeFilter) Name() string {
	return fmt.Sprintf("regime-filter(%s)", f.inner.Name())
}

func (f *RegimeFilter) OnTick(ctx context.Context) ([]strategy.Signal, error) {
	// Update regime
	regime, err := f.macro.Evaluate(ctx)
	if err == nil {
		f.mu.Lock()
		f.lastRegime = regime
		f.mu.Unlock()
	}

	signals, err := f.inner.OnTick(ctx)
	if err != nil {
		return nil, err
	}

	return f.filter(signals), nil
}

func (f *RegimeFilter) OnMarketTick(ctx context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	// If inner is not a TickStrategy, skip
	ts, ok := f.inner.(strategy.TickStrategy)
	if !ok {
		return nil, nil
	}

	signals, err := ts.OnMarketTick(ctx, tick)
	if err != nil {
		return nil, err
	}

	return f.filter(signals), nil
}

func (f *RegimeFilter) OnMarketCandle(ctx context.Context, candle marketdata.Candle) ([]strategy.Signal, error) {
	// Update regime first if macro is also a CandleStrategy
	if cs, ok := f.macro.(strategy.CandleStrategy); ok {
		_, _ = cs.OnMarketCandle(ctx, candle)
	}

	// Update cached regime
	regime, err := f.macro.Evaluate(ctx)
	if err == nil {
		f.mu.Lock()
		f.lastRegime = regime
		f.mu.Unlock()
	}

	// If inner is not a CandleStrategy, skip
	cs, ok := f.inner.(strategy.CandleStrategy)
	if !ok {
		return nil, nil
	}

	signals, err := cs.OnMarketCandle(ctx, candle)
	if err != nil {
		return nil, err
	}

	return f.filter(signals), nil
}

func (f *RegimeFilter) filter(signals []strategy.Signal) []strategy.Signal {
	if len(signals) == 0 {
		return nil
	}

	f.mu.RLock()
	regime := f.lastRegime
	f.mu.RUnlock()

	var filtered []strategy.Signal
	for _, s := range signals {
		allowed := false

		switch regime.Signal {
		case SignalBuy:
			if s.Action == "buy" {
				allowed = true
			}
		case SignalSell:
			if s.Action == "sell" {
				allowed = true
			}
		case SignalNone:
			if !f.suppressNone {
				allowed = true
			}
		}

		// ALWAYS allow sells if they are reasons of "STOP LOSS" or "TAKE PROFIT"
		if s.Action == "sell" {
			allowed = true
		}

		if allowed {
			s.Reason = fmt.Sprintf("%s [Regime: %s]", s.Reason, regime.Signal)
			filtered = append(filtered, s)
		}
	}

	return filtered
}

// Ensure interfaces are met
var _ strategy.Strategy = (*RegimeFilter)(nil)
var _ strategy.TickStrategy = (*RegimeFilter)(nil)
var _ strategy.CandleStrategy = (*RegimeFilter)(nil)
