package marketdata

import (
	"context"
	"time"
)

type Feed interface {
	LatestTick(ctx context.Context, symbol string) (Tick, error)
	LatestCandle(ctx context.Context, symbol, interval string) (Candle, error)
}

type TickSubscription interface {
	Chan() <-chan Tick
}

type StreamFeed interface {
	Feed
	SubscribeTicks(ctx context.Context, symbol string, interval time.Duration) TickSubscription
}
