package scheduler

import (
	"context"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

type tickSubscriber interface {
	SubscribeTicks(ctx context.Context, symbol string, interval time.Duration) marketdata.TickSubscription
}

type tickRunner interface {
	RunTick(ctx context.Context, tick marketdata.Tick) error
}

type StreamService struct {
	runner   tickRunner
	feed     tickSubscriber
	symbols  []string
	interval time.Duration
}

func NewStreamService(r tickRunner, feed tickSubscriber, symbols []string, interval time.Duration) *StreamService {
	return &StreamService{runner: r, feed: feed, symbols: symbols, interval: interval}
}

func (s *StreamService) Start(ctx context.Context) {
	if s.runner == nil || s.feed == nil || len(s.symbols) == 0 {
		return
	}
	if s.interval <= 0 {
		s.interval = 100 * time.Millisecond
	}
	for _, symbol := range s.symbols {
		sub := s.feed.SubscribeTicks(ctx, symbol, s.interval)
		go func(ch <-chan marketdata.Tick) {
			for {
				select {
				case <-ctx.Done():
					return
				case tick, ok := <-ch:
					if !ok {
						return
					}
					_ = s.runner.RunTick(ctx, tick)
				}
			}
		}(sub.Chan())
	}
}
