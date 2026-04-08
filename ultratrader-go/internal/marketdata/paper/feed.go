package paper

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

type Feed struct {
	mu      sync.Mutex
	indices map[string]int
}

type tickSubscription struct{ ch <-chan marketdata.Tick }

func (s tickSubscription) Chan() <-chan marketdata.Tick { return s.ch }

type candleSubscription struct{ ch <-chan marketdata.Candle }

func (s candleSubscription) Chan() <-chan marketdata.Candle { return s.ch }

func New() *Feed { return &Feed{indices: map[string]int{}} }

func (f *Feed) LatestTick(_ context.Context, symbol string) (marketdata.Tick, error) {
	price, ok := defaultPrices()[symbol]
	if !ok {
		return marketdata.Tick{}, fmt.Errorf("unknown symbol %q", symbol)
	}
	return marketdata.Tick{Symbol: symbol, Price: price, Source: "paper", Timestamp: time.Now().UTC()}, nil
}

func (f *Feed) LatestCandle(_ context.Context, symbol, interval string) (marketdata.Candle, error) {
	price, ok := defaultPrices()[symbol]
	if !ok {
		return marketdata.Candle{}, fmt.Errorf("unknown symbol %q", symbol)
	}
	return marketdata.Candle{Symbol: symbol, Interval: interval, Open: price, High: price, Low: price, Close: price, Volume: "1000", Timestamp: time.Now().UTC()}, nil
}

func (f *Feed) SubscribeTicks(ctx context.Context, symbol string, interval time.Duration) marketdata.TickSubscription {
	ch := make(chan marketdata.Tick, 1)
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}
	go func() {
		defer close(ch)
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				tick, err := f.nextStreamTick(symbol)
				if err != nil {
					continue
				}
				select {
				case ch <- tick:
				default:
				}
			}
		}
	}()
	return tickSubscription{ch: ch}
}

func (f *Feed) nextStreamTick(symbol string) (marketdata.Tick, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	sequence, ok := streamPrices()[symbol]
	if !ok || len(sequence) == 0 {
		return marketdata.Tick{}, fmt.Errorf("unknown symbol %q", symbol)
	}
	idx := f.indices[symbol] % len(sequence)
	f.indices[symbol] = f.indices[symbol] + 1
	return marketdata.Tick{Symbol: symbol, Price: sequence[idx], Source: "paper-stream", Timestamp: time.Now().UTC()}, nil
}

func (f *Feed) SubscribeCandles(ctx context.Context, symbol, interval string) marketdata.CandleSubscription {
	ch := make(chan marketdata.Candle, 1)

	// Fake interval duration for paper trading (e.g. 5 seconds)
	dur := 5 * time.Second

	go func() {
		defer close(ch)
		ticker := time.NewTicker(dur)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				candle, err := f.nextStreamCandle(symbol, interval)
				if err != nil {
					continue
				}
				select {
				case ch <- candle:
				default:
				}
			}
		}
	}()
	return candleSubscription{ch: ch}
}

func (f *Feed) nextStreamCandle(symbol, interval string) (marketdata.Candle, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	sequence, ok := streamPrices()[symbol]
	if !ok || len(sequence) == 0 {
		return marketdata.Candle{}, fmt.Errorf("unknown symbol %q", symbol)
	}

	// Simply cycle through stream prices and mock a candle to keep it deterministic.
	idx := f.indices[symbol] % len(sequence)
	price := sequence[idx]
	// Normally we would increment idx, but let SubscribeTicks drive the index or just advance it.
	f.indices[symbol] = f.indices[symbol] + 1

	return marketdata.Candle{
		Symbol:    symbol,
		Interval:  interval,
		Open:      price,
		High:      price,
		Low:       price,
		Close:     price,
		Volume:    "100",
		Timestamp: time.Now().UTC(),
	}, nil
}

func defaultPrices() map[string]string {
	return map[string]string{"BTCUSDT": "65000.00", "ETHUSDT": "3200.00"}
}

func streamPrices() map[string][]string {
	return map[string][]string{
		"BTCUSDT": {"65000.00", "64950.00", "65050.00", "64975.00"},
		"ETHUSDT": {"3200.00", "3190.00", "3210.00", "3195.00"},
	}
}
