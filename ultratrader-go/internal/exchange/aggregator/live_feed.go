package aggregator

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/logging"
)

// ExchangeHealth tracks the health status of an exchange.
type ExchangeHealth struct {
	Name         string
	IsHealthy    bool
	LastUpdate   time.Time
	ErrorCount   int
	ResponseTime time.Duration
}

// LiveMultiExchangeFeed provides real-time price aggregation across multiple exchanges.
type LiveMultiExchangeFeed struct {
	aggregator   *PriceAggregator
	exchanges    map[string]PriceProvider
	health       map[string]*ExchangeHealth
	quotes       map[string]map[string]PriceQuote // symbol -> exchange -> quote
	minSpreadPct float64
	feePct       float64
	mu           sync.RWMutex
	logger       *logging.Logger
}

func NewLiveMultiExchangeFeed(logger *logging.Logger) *LiveMultiExchangeFeed {
	if logger == nil {
		logger, _ = logging.New(logging.Config{Stdout: true})
	}
	return &LiveMultiExchangeFeed{
		aggregator:   NewPriceAggregator(),
		exchanges:    make(map[string]PriceProvider),
		health:       make(map[string]*ExchangeHealth),
		quotes:       make(map[string]map[string]PriceQuote),
		minSpreadPct: 0.05, // 0.05% minimum spread
		feePct:       0.1,  // 0.1% estimated fee
		logger:       logger,
	}
}

// RegisterExchange adds an exchange provider.
func (f *LiveMultiExchangeFeed) RegisterExchange(provider PriceProvider) {
	f.mu.Lock()
	defer f.mu.Unlock()

	name := provider.Name()
	f.exchanges[name] = provider
	f.aggregator.Register(provider)
	f.health[name] = &ExchangeHealth{
		Name:       name,
		IsHealthy:  true,
		LastUpdate: time.Now(),
	}
}

// GetBestPrice returns the best price across all exchanges.
func (f *LiveMultiExchangeFeed) GetBestPrice(ctx context.Context, symbol string) (PriceQuote, error) {
	f.mu.RLock()
	quotes := f.getQuotesForSymbol(symbol)
	f.mu.RUnlock()

	if len(quotes) == 0 {
		// Fetch fresh quotes
		fetched := f.fetchAllQuotes(ctx, symbol)
		if len(fetched) == 0 {
			return PriceQuote{}, fmt.Errorf("no quotes available for %s", symbol)
		}
		quotes = fetched
	}

	// Find best price (lowest for buy, highest for sell — we return lowest)
	best := quotes[0]
	for _, q := range quotes[1:] {
		if q.Price < best.Price && q.Price > 0 {
			best = q
		}
	}

	return best, nil
}

// GetArbitrageOpportunities finds arbitrage opportunities across exchanges.
func (f *LiveMultiExchangeFeed) GetArbitrageOpportunities(ctx context.Context, symbol string) []ArbitrageOpportunity {
	f.mu.RLock()
	quotes := f.getQuotesForSymbol(symbol)
	f.mu.RUnlock()

	if len(quotes) < 2 {
		return nil
	}

	var opportunities []ArbitrageOpportunity

	for i, buyQ := range quotes {
		for j, sellQ := range quotes {
			if i == j || buyQ.Price <= 0 || sellQ.Price <= 0 {
				continue
			}

			spreadPct := (sellQ.Price - buyQ.Price) / buyQ.Price * 100

			// Account for fees (buy fee + sell fee)
			estProfitPct := spreadPct - (f.feePct * 2)

			if spreadPct >= f.minSpreadPct && estProfitPct > 0 {
				opportunities = append(opportunities, ArbitrageOpportunity{
					Symbol:       symbol,
					BuyExchange:  buyQ.Exchange,
					SellExchange: sellQ.Exchange,
					BuyPrice:     buyQ.Price,
					SellPrice:    sellQ.Price,
					Spread:       spreadPct,
				})
			}
		}
	}

	return opportunities
}

// GetExchangeHealth returns health status of all exchanges.
func (f *LiveMultiExchangeFeed) GetExchangeHealth() map[string]ExchangeHealth {
	f.mu.RLock()
	defer f.mu.RUnlock()

	result := make(map[string]ExchangeHealth, len(f.health))
	for name, h := range f.health {
		result[name] = *h
	}
	return result
}

// GetAggregatedPrice returns the aggregated price using the specified method.
func (f *LiveMultiExchangeFeed) GetAggregatedPrice(ctx context.Context, symbol string, method AggregationMethod) (float64, error) {
	return f.aggregator.GetPrice(ctx, symbol, method)
}

// fetchAllQuotes fetches prices from all registered exchanges.
func (f *LiveMultiExchangeFeed) fetchAllQuotes(ctx context.Context, symbol string) []PriceQuote {
	f.mu.RLock()
	exchanges := make(map[string]PriceProvider, len(f.exchanges))
	for k, v := range f.exchanges {
		exchanges[k] = v
	}
	f.mu.RUnlock()

	var mu sync.Mutex
	var quotes []PriceQuote
	var wg sync.WaitGroup

	for name, provider := range exchanges {
		wg.Add(1)
		go func(name string, p PriceProvider) {
			defer wg.Done()

			start := time.Now()
			priceStr, err := p.GetTickerPrice(ctx, symbol)
			elapsed := time.Since(start)

			f.mu.Lock()
			health := f.health[name]
			health.LastUpdate = time.Now()
			health.ResponseTime = elapsed
			if err != nil {
				health.ErrorCount++
				health.IsHealthy = health.ErrorCount < 5
				f.mu.Unlock()
				return
			}
			health.ErrorCount = 0
			health.IsHealthy = true
			f.mu.Unlock()

			price := parseFloat(priceStr)
			if price <= 0 {
				return
			}

			quote := PriceQuote{
				Exchange: name,
				Symbol:   symbol,
				Price:    price,
				Time:     time.Now(),
			}

			mu.Lock()
			quotes = append(quotes, quote)
			mu.Unlock()

			// Cache the quote
			f.mu.Lock()
			if f.quotes[symbol] == nil {
				f.quotes[symbol] = make(map[string]PriceQuote)
			}
			f.quotes[symbol][name] = quote
			f.mu.Unlock()
		}(name, provider)
	}

	wg.Wait()
	return quotes
}

// getQuotesForSymbol returns cached quotes for a symbol.
func (f *LiveMultiExchangeFeed) getQuotesForSymbol(symbol string) []PriceQuote {
	exchangeQuotes, ok := f.quotes[symbol]
	if !ok {
		return nil
	}

	var quotes []PriceQuote
	for _, q := range exchangeQuotes {
		// Skip stale quotes (older than 30 seconds)
		if time.Since(q.Time) > 30*time.Second {
			continue
		}
		quotes = append(quotes, q)
	}
	return quotes
}

func parseFloat(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
