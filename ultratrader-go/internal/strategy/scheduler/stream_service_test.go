package scheduler

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

type fakeSub struct{ ch <-chan marketdata.Tick }

func (s fakeSub) Chan() <-chan marketdata.Tick { return s.ch }

type fakeFeed struct{}

func (f fakeFeed) SubscribeTicks(ctx context.Context, symbol string, interval time.Duration) marketdata.TickSubscription {
	ch := make(chan marketdata.Tick, 1)
	go func() {
		defer close(ch)
		select {
		case <-ctx.Done():
			return
		case <-time.After(10 * time.Millisecond):
			ch <- marketdata.Tick{Symbol: symbol}
		}
	}()
	return fakeSub{ch: ch}
}

type countRunner struct{ count atomic.Int32 }

func (r *countRunner) RunTick(ctx context.Context, tick marketdata.Tick) error {
	r.count.Add(1)
	return nil
}

func TestStreamServiceRunsOnTick(t *testing.T) {
	runner := &countRunner{}
	service := NewStreamService(runner, fakeFeed{}, []string{"BTCUSDT"}, 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	service.Start(ctx)
	time.Sleep(40 * time.Millisecond)
	if runner.count.Load() < 1 {
		t.Fatalf("expected runner to be triggered by stream, got %d", runner.count.Load())
	}
}
