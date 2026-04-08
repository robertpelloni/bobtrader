package scheduler_test

import (
	"context"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/scheduler"
)

type mockCandleFeed struct {
	ch chan marketdata.Candle
}

func (m *mockCandleFeed) SubscribeCandles(ctx context.Context, symbol, interval string) marketdata.CandleSubscription {
	return mockCandleSub{ch: m.ch}
}

type mockCandleSub struct {
	ch chan marketdata.Candle
}

func (m mockCandleSub) Chan() <-chan marketdata.Candle { return m.ch }

type mockCandleRunner struct {
	called bool
}

func (m *mockCandleRunner) RunCandle(ctx context.Context, candle marketdata.Candle) error {
	m.called = true
	return nil
}

func TestCandleStreamService_StartAndReceive(t *testing.T) {
	ch := make(chan marketdata.Candle, 1)
	feed := &mockCandleFeed{ch: ch}
	runner := &mockCandleRunner{}

	service := scheduler.NewCandleStreamService(runner, feed, []string{"BTCUSDT"}, "1m")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	service.Start(ctx)

	// Send a candle
	ch <- marketdata.Candle{Symbol: "BTCUSDT", Close: "100.0"}

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	if !runner.called {
		t.Errorf("expected runner to be called with candle")
	}
}
