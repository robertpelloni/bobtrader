package portfolio

import (
	"context"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

// USDTBalanceReader reads the available USDT cash balance.
type USDTBalanceReader interface {
	USDTBalance() float64
}

type ExposureView struct {
	tracker    *Tracker
	feed       marketdata.Feed
	balanceReader USDTBalanceReader
}

func NewExposureView(tracker *Tracker, feed marketdata.Feed) *ExposureView {
	return &ExposureView{tracker: tracker, feed: feed}
}

// NewExposureViewWithBalance creates an exposure view that includes USDT balance
// in total portfolio value calculations.
func NewExposureViewWithBalance(tracker *Tracker, feed marketdata.Feed, balanceReader USDTBalanceReader) *ExposureView {
	return &ExposureView{tracker: tracker, feed: feed, balanceReader: balanceReader}
}

func (v *ExposureView) CurrentValue(symbol string) float64 {
	for _, position := range v.tracker.ValuedPositions(context.Background(), v.feed) {
		if position.Symbol == symbol {
			if position.MarketValue > 0 {
				return position.MarketValue
			}
			return position.CostBasis
		}
	}
	return 0
}

func (v *ExposureView) TotalValue() float64 {
	total := v.tracker.TotalMarketValue(context.Background(), v.feed)
	// Include USDT cash balance if available
	if v.balanceReader != nil {
		total += v.balanceReader.USDTBalance()
	}
	return total
}
