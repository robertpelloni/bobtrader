package paper

import (
	"context"
	"fmt"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

type Feed struct{}

func New() *Feed {
	return &Feed{}
}

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

func defaultPrices() map[string]string {
	return map[string]string{
		"BTCUSDT": "65000.00",
		"ETHUSDT": "3200.00",
	}
}
