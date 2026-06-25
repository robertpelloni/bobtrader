package backtest

import (
	"context"
	"fmt"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	marketdatabinance "github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata/binance"
)

// LiveHistoryProvider fetches real historical data from Binance for backtesting.
type LiveHistoryProvider struct {
	adapter *marketdatabinance.Adapter
	candles []marketdata.Candle
}

func NewLiveHistoryProvider(adapter *marketdatabinance.Adapter) *LiveHistoryProvider {
	return &LiveHistoryProvider{adapter: adapter}
}

// FetchCandles fetches a specific number of historical candles for a symbol.
func (p *LiveHistoryProvider) FetchCandles(ctx context.Context, symbol, interval string, limit int) ([]marketdata.Candle, error) {
	candles, err := p.adapter.GetKlines(ctx, symbol, interval, limit)
	if err != nil {
		return nil, fmt.Errorf("LiveHistoryProvider: failed to fetch klines: %w", err)
	}
	p.candles = candles
	return candles, nil
}

// Candles implements CandleHistoryProvider.
func (p *LiveHistoryProvider) Candles() []marketdata.Candle {
	return p.candles
}
