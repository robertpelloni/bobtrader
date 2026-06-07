package scheduler

import (
	"context"
	"fmt"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/execution"
)

// EnhancedScheduler is like Scheduler but adds position awareness, real notional
// calculation, and signal logging for every generated signal.
type EnhancedScheduler struct {
	runtime   *strategy.Runtime
	execution *execution.Service
	portfolio PositionChecker
	feed      marketdata.Feed
	signalLog *strategy.SignalLog
}

// NewEnhanced creates a scheduler that logs signals and respects positions.
func NewEnhanced(
	runtime *strategy.Runtime,
	execService *execution.Service,
	portfolio PositionChecker,
	feed marketdata.Feed,
	signalLog *strategy.SignalLog,
) *EnhancedScheduler {
	return &EnhancedScheduler{
		runtime:   runtime,
		execution: execService,
		portfolio: portfolio,
		feed:      feed,
		signalLog: signalLog,
	}
}

func (s *EnhancedScheduler) RunOnce(ctx context.Context) error {
	signals, err := s.runtime.Tick(ctx)
	if err != nil {
		return fmt.Errorf("runtime tick: %w", err)
	}
	ExecuteSignals(ctx, signals, s.execution, s.portfolio, s.feed, s.signalLog)
	return nil
}

func (s *EnhancedScheduler) RunTick(ctx context.Context, tick marketdata.Tick) error {
	signals, err := s.runtime.TickEvent(ctx, tick)
	if err != nil {
		return fmt.Errorf("runtime tick event: %w", err)
	}
	ExecuteSignals(ctx, signals, s.execution, s.portfolio, s.feed, s.signalLog)
	return nil
}

func (s *EnhancedScheduler) RunCandle(ctx context.Context, candle marketdata.Candle) error {
	signals, err := s.runtime.CandleEvent(ctx, candle)
	if err != nil {
		return fmt.Errorf("runtime candle event: %w", err)
	}
	ExecuteSignals(ctx, signals, s.execution, s.portfolio, s.feed, s.signalLog)
	return nil
}
