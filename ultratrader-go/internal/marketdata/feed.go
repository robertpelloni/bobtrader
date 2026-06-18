package marketdata

import (
	"context"
	"time"
)

type Feed interface {
	LatestTick(ctx context.Context, symbol string) (Tick, error)
	LatestCandle(ctx context.Context, symbol, interval string) (Candle, error)
	CandleHistory(ctx context.Context, symbol, interval string, limit int) ([]Candle, error)
}

type TickSubscription interface {
	Chan() <-chan Tick
}

type CandleSubscription interface {
	Chan() <-chan Candle
}

type DepthSubscription interface {
	Chan() <-chan DepthUpdate
}

type DepthUpdate struct {
	Symbol    string
	Bids      [][2]string // [price, quantity]
	Asks      [][2]string
	Timestamp time.Time
}

type StreamFeed interface {
	Feed
	SubscribeTicks(ctx context.Context, symbol string, interval time.Duration) TickSubscription
	SubscribeCandles(ctx context.Context, symbol, interval string) CandleSubscription
	SubscribeDepth(ctx context.Context, symbol string) DepthSubscription
}
