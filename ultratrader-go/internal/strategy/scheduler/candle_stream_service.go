package scheduler

import (
	"context"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

type candleSubscriber interface {
	SubscribeCandles(ctx context.Context, symbol, interval string) marketdata.CandleSubscription
}

type candleRunner interface {
	RunCandle(ctx context.Context, candle marketdata.Candle) error
}

type CandleStreamService struct {
	runner   candleRunner
	feed     candleSubscriber
	symbols  []string
	interval string
}

func NewCandleStreamService(r candleRunner, feed candleSubscriber, symbols []string, interval string) *CandleStreamService {
	return &CandleStreamService{runner: r, feed: feed, symbols: symbols, interval: interval}
}

func (s *CandleStreamService) Start(ctx context.Context) {
	if s.runner == nil || s.feed == nil || len(s.symbols) == 0 {
		return
	}
	if s.interval == "" {
		s.interval = "1m"
	}
	for _, symbol := range s.symbols {
		sub := s.feed.SubscribeCandles(ctx, symbol, s.interval)
		go func(ch <-chan marketdata.Candle) {
			for {
				select {
				case <-ctx.Done():
					return
				case candle, ok := <-ch:
					if !ok {
						return
					}
					_ = s.runner.RunCandle(ctx, candle)
				}
			}
		}(sub.Chan())
	}
}
