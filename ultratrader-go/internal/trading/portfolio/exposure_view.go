package portfolio

import (
	"context"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

type ExposureView struct {
	tracker *Tracker
	feed    marketdata.Feed
}

func NewExposureView(tracker *Tracker, feed marketdata.Feed) *ExposureView {
	return &ExposureView{tracker: tracker, feed: feed}
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
	return v.tracker.TotalMarketValue(context.Background(), v.feed)
}
