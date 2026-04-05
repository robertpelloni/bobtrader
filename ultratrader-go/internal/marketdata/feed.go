package marketdata

import "context"

type Feed interface {
	LatestTick(ctx context.Context, symbol string) (Tick, error)
	LatestCandle(ctx context.Context, symbol, interval string) (Candle, error)
}
