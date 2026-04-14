package backtest

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
)

// SyncCandle represents a point-in-time state across multiple symbols.
type SyncCandle struct {
	Timestamp time.Time
	Candles   map[string]marketdata.Candle
}

// MultiSymbolFeed handles synchronized historical data playback for multiple symbols.
type MultiSymbolFeed struct {
	symbols []string
	data    map[string][]marketdata.Candle
	mu      sync.Mutex
}

// NewMultiSymbolFeed creates a new engine for managing synchronized backtesting data.
func NewMultiSymbolFeed(symbols []string) *MultiSymbolFeed {
	return &MultiSymbolFeed{
		symbols: symbols,
		data:    make(map[string][]marketdata.Candle),
	}
}

// LoadData accepts historical candles for a specific symbol.
func (m *MultiSymbolFeed) LoadData(symbol string, candles []marketdata.Candle) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure they are sorted chronologically
	sort.Slice(candles, func(i, j int) bool {
		return candles[i].Timestamp.Before(candles[j].Timestamp)
	})

	m.data[symbol] = candles
}

// FetchData concurrently loads data from a given feed for all registered symbols.
func (m *MultiSymbolFeed) FetchData(ctx context.Context, feed marketdata.Feed, interval string, limit int) error {
	var wg sync.WaitGroup
	errs := make(chan error, len(m.symbols))

	type result struct {
		symbol  string
		candles []marketdata.Candle
	}
	results := make(chan result, len(m.symbols))

	// Some feeds might not support historical data retrieval in this exact signature,
	// so we assume we have a HistoricalFeed interface here. If the feed doesn't implement it, we fail.
	type HistoricalFeed interface {
		HistoricalCandles(ctx context.Context, symbol string, interval string, limit int) ([]marketdata.Candle, error)
	}

	histFeed, ok := feed.(HistoricalFeed)
	if !ok {
		return fmt.Errorf("provided feed does not support HistoricalCandles")
	}

	for _, sym := range m.symbols {
		wg.Add(1)
		go func(symbol string) {
			defer wg.Done()
			candles, err := histFeed.HistoricalCandles(ctx, symbol, interval, limit)
			if err != nil {
				errs <- fmt.Errorf("failed to fetch %s: %w", symbol, err)
				return
			}
			results <- result{symbol: symbol, candles: candles}
		}(sym)
	}

	wg.Wait()
	close(errs)
	close(results)

	if len(errs) > 0 {
		return <-errs // Return first error
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	for r := range results {
		m.data[r.symbol] = r.candles
	}

	return nil
}

// Synchronize aligns all loaded candles by timestamp and returns an ordered slice of SyncCandles.
// Missing data points for a symbol at a specific timestamp will result in an empty candle or the previous candle state
// depending on strategy (currently omitted from the map for that timestamp).
func (m *MultiSymbolFeed) Synchronize() []SyncCandle {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Collect all unique timestamps
	timestamps := make(map[time.Time]struct{})
	for _, candles := range m.data {
		for _, c := range candles {
			timestamps[c.Timestamp] = struct{}{}
		}
	}

	var orderedTimes []time.Time
	for ts := range timestamps {
		orderedTimes = append(orderedTimes, ts)
	}

	sort.Slice(orderedTimes, func(i, j int) bool {
		return orderedTimes[i].Before(orderedTimes[j])
	})

	var syncTimeline []SyncCandle

	// Track pointers into each symbol's slice to avoid O(N^2) search
	ptrs := make(map[string]int)
	for _, sym := range m.symbols {
		ptrs[sym] = 0
	}

	for _, ts := range orderedTimes {
		sc := SyncCandle{
			Timestamp: ts,
			Candles:   make(map[string]marketdata.Candle),
		}

		for _, sym := range m.symbols {
			candles := m.data[sym]
			idx := ptrs[sym]

			// Fast-forward pointer to current timestamp
			for idx < len(candles) && candles[idx].Timestamp.Before(ts) {
				idx++
			}
			ptrs[sym] = idx

			if idx < len(candles) && candles[idx].Timestamp.Equal(ts) {
				sc.Candles[sym] = candles[idx]
			}
		}

		syncTimeline = append(syncTimeline, sc)
	}

	return syncTimeline
}
