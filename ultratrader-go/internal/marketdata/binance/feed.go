package binance

import (
	"context"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange/binance"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

// Feed implements marketdata.StreamFeed using the Binance REST adapter.
// It polls for latest ticks and candles at configurable intervals.
type Feed struct {
	adapter *binance.Adapter
	mu      sync.Mutex
}

func NewFeed(adapter *binance.Adapter) *Feed {
	return &Feed{adapter: adapter}
}

func (f *Feed) LatestTick(ctx context.Context, symbol string) (marketdata.Tick, error) {
	price, err := f.adapter.GetTickerPrice(ctx, symbol)
	if err != nil {
		return marketdata.Tick{}, err
	}
	return marketdata.Tick{
		Symbol:    symbol,
		Price:     price,
		Source:    "binance",
		Timestamp: time.Now().UTC(),
	}, nil
}

func (f *Feed) LatestCandle(ctx context.Context, symbol, interval string) (marketdata.Candle, error) {
	klines, err := f.adapter.GetKlines(ctx, symbol, interval, 1)
	if err != nil {
		return marketdata.Candle{}, err
	}
	if len(klines) == 0 {
		return marketdata.Candle{}, nil
	}
	k := klines[0]
	return marketdata.Candle{
		Symbol:    symbol,
		Interval:  interval,
		Open:      k.Open,
		High:      k.High,
		Low:       k.Low,
		Close:     k.Close,
		Volume:    k.Volume,
		Timestamp: time.UnixMilli(k.OpenTime).UTC(),
	}, nil
}

func (f *Feed) SubscribeTicks(ctx context.Context, symbol string, interval time.Duration) marketdata.TickSubscription {
	ch := make(chan marketdata.Tick, 1)
	if interval <= 0 {
		interval = 1 * time.Second
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
				tick, err := f.LatestTick(ctx, symbol)
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
	return tickSub{ch: ch}
}

func (f *Feed) SubscribeCandles(ctx context.Context, symbol, interval string) marketdata.CandleSubscription {
	ch := make(chan marketdata.Candle, 1)
	// Parse interval to duration for polling
	dur := candleIntervalToDuration(interval)
	go func() {
		defer close(ch)
		ticker := time.NewTicker(dur)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				candle, err := f.LatestCandle(ctx, symbol, interval)
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
	return candleSub{ch: ch}
}

func candleIntervalToDuration(interval string) time.Duration {
	// Parse Binance interval strings like "1m", "5m", "1h", "1d"
	if len(interval) < 2 {
		return 1 * time.Minute
	}
	num := utils.ParseFloat(interval[:len(interval)-1])
	unit := interval[len(interval)-1]
	switch unit {
	case 's':
		return time.Duration(num * float64(time.Second))
	case 'm':
		return time.Duration(num * float64(time.Minute))
	case 'h':
		return time.Duration(num * float64(time.Hour))
	case 'd':
		return time.Duration(num * float64(24*time.Hour))
	case 'w':
		return time.Duration(num * float64(7*24*time.Hour))
	default:
		return 1 * time.Minute
	}
}

type tickSub struct{ ch <-chan marketdata.Tick }

func (s tickSub) Chan() <-chan marketdata.Tick { return s.ch }

type candleSub struct{ ch <-chan marketdata.Candle }

func (s candleSub) Chan() <-chan marketdata.Candle { return s.ch }
