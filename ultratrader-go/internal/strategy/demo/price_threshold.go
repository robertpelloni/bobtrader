package demo

import (
	"context"
	"strings"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

type PriceThreshold struct {
	accountID   string
	symbol      string
	quantity    string
	maxBuyPrice string
	feed        marketdata.Feed
	emitted     bool
}

func NewPriceThreshold(accountID, symbol, quantity, maxBuyPrice string, feed marketdata.Feed) *PriceThreshold {
	return &PriceThreshold{accountID: accountID, symbol: symbol, quantity: quantity, maxBuyPrice: maxBuyPrice, feed: feed}
}

func (s *PriceThreshold) Name() string { return "demo-price-threshold" }

func (s *PriceThreshold) OnTick(ctx context.Context) ([]strategy.Signal, error) {
	if s.emitted {
		return nil, nil
	}
	tick, err := s.feed.LatestTick(ctx, s.symbol)
	if err != nil {
		return nil, err
	}
	if compareDecimalStrings(tick.Price, s.maxBuyPrice) <= 0 {
		s.emitted = true
		return []strategy.Signal{{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "buy",
			Reason:    "price at or below threshold",
			Quantity:  s.quantity,
			OrderType: "market",
		}}, nil
	}
	return nil, nil
}

func compareDecimalStrings(left, right string) int {
	left = normalizeDecimal(left)
	right = normalizeDecimal(right)
	if left == right {
		return 0
	}
	if left < right {
		return -1
	}
	return 1
}

func normalizeDecimal(value string) string {
	parts := strings.SplitN(strings.TrimSpace(value), ".", 2)
	whole := parts[0]
	frac := ""
	if len(parts) == 2 {
		frac = parts[1]
	}
	whole = strings.TrimLeft(whole, "0")
	if whole == "" {
		whole = "0"
	}
	for len(whole) < 20 {
		whole = "0" + whole
	}
	for len(frac) < 10 {
		frac += "0"
	}
	if len(frac) > 10 {
		frac = frac[:10]
	}
	return whole + "." + frac
}
