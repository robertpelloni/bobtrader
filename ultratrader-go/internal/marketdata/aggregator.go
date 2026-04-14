package marketdata

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/logging"
)

// AggregationStrategy defines how prices from multiple exchanges are aggregated.
type AggregationStrategy string

const (
	AveragePrice AggregationStrategy = "average"
	MedianPrice  AggregationStrategy = "median"
	BestBidAsk   AggregationStrategy = "best"
	Failover     AggregationStrategy = "failover"
)

// Aggregator merges multiple market data feeds into a single feed.
type Aggregator struct {
	feeds    map[string]Feed
	strategy AggregationStrategy
	logger   *logging.Logger
	mu       sync.RWMutex
}

// NewAggregator creates a new Aggregator.
func NewAggregator(strategy AggregationStrategy, logger *logging.Logger) *Aggregator {
	if logger == nil {
		logger, _ = logging.New(logging.Config{Stdout: true})
	}
	return &Aggregator{
		feeds:    make(map[string]Feed),
		strategy: strategy,
		logger:   logger,
	}
}

// AddFeed registers a feed with the given name (e.g., "binance", "kucoin").
func (a *Aggregator) AddFeed(name string, feed Feed) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.feeds[name] = feed
}

// LatestTick aggregates the latest tick from all registered feeds based on the aggregation strategy.
func (a *Aggregator) LatestTick(ctx context.Context, symbol string) (Tick, error) {
	a.mu.RLock()
	feeds := make([]Feed, 0, len(a.feeds))
	for _, f := range a.feeds {
		feeds = append(feeds, f)
	}
	a.mu.RUnlock()

	if len(feeds) == 0 {
		return Tick{}, fmt.Errorf("no feeds available")
	}

	var ticks []Tick
	for _, feed := range feeds {
		if t, err := feed.LatestTick(ctx, symbol); err == nil && t.Price != "" {
			ticks = append(ticks, t)
		}
	}

	if len(ticks) == 0 {
		return Tick{}, fmt.Errorf("no valid ticks found for symbol %q", symbol)
	}

	if a.strategy == Failover {
		// Just return the first one that succeeded
		return ticks[0], nil
	}

	var prices []float64
	for _, t := range ticks {
		if p, err := strconv.ParseFloat(t.Price, 64); err == nil {
			prices = append(prices, p)
		}
	}

	if len(prices) == 0 {
		return Tick{}, fmt.Errorf("could not parse any prices")
	}

	var aggregatedPrice float64

	switch a.strategy {
	case AveragePrice:
		var sum float64
		for _, p := range prices {
			sum += p
		}
		aggregatedPrice = sum / float64(len(prices))
	case MedianPrice:
		sort.Float64s(prices)
		mid := len(prices) / 2
		if len(prices)%2 == 0 {
			aggregatedPrice = (prices[mid-1] + prices[mid]) / 2.0
		} else {
			aggregatedPrice = prices[mid]
		}
	case BestBidAsk:
		// Fallback to average since Best Bid/Ask implies orderbook data not standard ticks
		var sum float64
		for _, p := range prices {
			sum += p
		}
		aggregatedPrice = sum / float64(len(prices))
	default:
		// Fallback to Failover (first available)
		return ticks[0], nil
	}

	return Tick{
		Symbol:    symbol,
		Price:     strconv.FormatFloat(aggregatedPrice, 'f', -1, 64),
		Timestamp: time.Now(),
	}, nil
}

// LatestCandle aggregates the latest candle from all registered feeds.
func (a *Aggregator) LatestCandle(ctx context.Context, symbol, interval string) (Candle, error) {
	a.mu.RLock()
	feeds := make([]Feed, 0, len(a.feeds))
	for _, f := range a.feeds {
		feeds = append(feeds, f)
	}
	a.mu.RUnlock()

	if len(feeds) == 0 {
		return Candle{}, fmt.Errorf("no feeds available")
	}

	var candles []Candle
	for _, feed := range feeds {
		if c, err := feed.LatestCandle(ctx, symbol, interval); err == nil && c.Close != "" {
			candles = append(candles, c)
		}
	}

	if len(candles) == 0 {
		return Candle{}, fmt.Errorf("no valid candles found for symbol %q", symbol)
	}

	// For simplicity, we implement Failover strategy for candles right now
	// Aggregating OHLC is complex because timestamps might not perfectly align
	return candles[0], nil
}
