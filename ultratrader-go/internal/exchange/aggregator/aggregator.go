package aggregator

import (
	"context"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
)

// PriceQuote represents a price quote from a single exchange.
type PriceQuote struct {
	Exchange string
	Symbol   string
	Price    float64
	Bid      float64
	Ask      float64
	Volume   float64
	Time     time.Time
}

// AggregationMethod determines how prices from multiple exchanges are combined.
type AggregationMethod int

const (
	MethodMedian AggregationMethod = iota
	MethodVWAP
	MethodBestBidAsk
	MethodMean
)

// PriceProvider is the interface adapters implement for price fetching.
type PriceProvider interface {
	Name() string
	GetTickerPrice(ctx context.Context, symbol string) (string, error)
}

// PriceAggregator combines prices from multiple exchange adapters.
type PriceAggregator struct {
	mu        sync.RWMutex
	exchanges map[string]PriceProvider
	health    map[string]bool
}

// NewPriceAggregator creates a new multi-exchange price aggregator.
func NewPriceAggregator() *PriceAggregator {
	return &PriceAggregator{
		exchanges: make(map[string]exchange.Adapter),
		health:    make(map[string]bool),
	}
}

// Register adds a price provider.
func (pa *PriceAggregator) Register(provider PriceProvider) {
	pa.mu.Lock()
	defer pa.mu.Unlock()
	pa.exchanges[provider.Name()] = provider
	pa.health[provider.Name()] = true
}

// GetPrice fetches prices from all registered exchanges and aggregates.
func (pa *PriceAggregator) GetPrice(ctx context.Context, symbol string, method AggregationMethod) (float64, error) {
	pa.mu.RLock()
	defer pa.mu.RUnlock()

	quotes := pa.fetchQuotes(ctx, symbol)
	if len(quotes) == 0 {
		return 0, nil
	}

	switch method {
	case MethodMedian:
		return median(quotes), nil
	case MethodMean:
		return mean(quotes), nil
	case MethodVWAP:
		return vwap(quotes), nil
	case MethodBestBidAsk:
		return bestMidpoint(quotes), nil
	default:
		return median(quotes), nil
	}
}

// GetAllQuotes fetches quotes from all exchanges.
func (pa *PriceAggregator) GetAllQuotes(ctx context.Context, symbol string) []PriceQuote {
	pa.mu.RLock()
	defer pa.mu.RUnlock()
	return pa.fetchQuotes(ctx, symbol)
}

// DetectArbitrage checks for price differences between exchanges.
func (pa *PriceAggregator) DetectArbitrage(ctx context.Context, symbol string, minSpread float64) []ArbitrageOpportunity {
	quotes := pa.fetchQuotes(ctx, symbol)
	if len(quotes) < 2 {
		return nil
	}

	var opportunities []ArbitrageOpportunity
	for i := 0; i < len(quotes); i++ {
		for j := i + 1; j < len(quotes); j++ {
			q1 := quotes[i]
			q2 := quotes[j]

			lowQuote, highQuote := q1, q2
			if q1.Price > q2.Price {
				lowQuote, highQuote = q2, q1
			}

			spread := (highQuote.Price - lowQuote.Price) / lowQuote.Price
			if spread >= minSpread {
				opportunities = append(opportunities, ArbitrageOpportunity{
					Symbol:      symbol,
					BuyExchange: lowQuote.Exchange,
					SellExchange: highQuote.Exchange,
					BuyPrice:    lowQuote.Price,
					SellPrice:   highQuote.Price,
					Spread:      spread,
				})
			}
		}
	}

	sort.Slice(opportunities, func(i, j int) bool {
		return opportunities[i].Spread > opportunities[j].Spread
	})

	return opportunities
}

// HealthStatus returns the health of each exchange.
func (pa *PriceAggregator) HealthStatus() map[string]bool {
	pa.mu.RLock()
	defer pa.mu.RUnlock()
	result := make(map[string]bool)
	for k, v := range pa.health {
		result[k] = v
	}
	return result
}

// ArbitrageOpportunity represents a potential arbitrage trade.
type ArbitrageOpportunity struct {
	Symbol       string  `json:"symbol"`
	BuyExchange  string  `json:"buy_exchange"`
	SellExchange string  `json:"sell_exchange"`
	BuyPrice     float64 `json:"buy_price"`
	SellPrice    float64 `json:"sell_price"`
	Spread       float64 `json:"spread"`
}

func (pa *PriceAggregator) fetchQuotes(ctx context.Context, symbol string) []PriceQuote {
	type result struct {
		quote PriceQuote
		err   error
	}

	ch := make(chan result, len(pa.exchanges))
	for name, provider := range pa.exchanges {
		go func(name string, provider PriceProvider) {
			priceStr, err := provider.GetTickerPrice(ctx, symbol)
			if err != nil {
				pa.mu.Lock()
				pa.health[name] = false
				pa.mu.Unlock()
				ch <- result{err: err}
				return
			}
			price := utils.ParseFloat(priceStr)
			pa.mu.Lock()
			pa.health[name] = true
			pa.mu.Unlock()
			ch <- result{quote: PriceQuote{
				Exchange: name,
				Symbol:   symbol,
				Price:    price,
				Time:     time.Now().UTC(),
			}}
		}(name, adapter)
	}

	var quotes []PriceQuote
	for i := 0; i < len(pa.exchanges); i++ {
		r := <-ch
		if r.err == nil {
			quotes = append(quotes, r.quote)
		}
	}
	return quotes
}

func median(quotes []PriceQuote) float64 {
	prices := make([]float64, len(quotes))
	for i, q := range quotes {
		prices[i] = q.Price
	}
	sort.Float64s(prices)

	n := len(prices)
	if n%2 == 1 {
		return prices[n/2]
	}
	return (prices[n/2-1] + prices[n/2]) / 2
}

func mean(quotes []PriceQuote) float64 {
	var sum float64
	for _, q := range quotes {
		sum += q.Price
	}
	return sum / float64(len(quotes))
}

func vwap(quotes []PriceQuote) float64 {
	var totalVolPrice, totalVol float64
	for _, q := range quotes {
		totalVolPrice += q.Price * q.Volume
		totalVol += q.Volume
	}
	if totalVol == 0 {
		return mean(quotes)
	}
	return totalVolPrice / totalVol
}

func bestMidpoint(quotes []PriceQuote) float64 {
	var midpoints []float64
	for _, q := range quotes {
		if q.Bid > 0 && q.Ask > 0 {
			midpoints = append(midpoints, (q.Bid+q.Ask)/2)
		}
	}
	if len(midpoints) == 0 {
		return median(quotes)
	}
	var sum float64
	for _, m := range midpoints {
		sum += m
	}
	return sum / float64(len(midpoints))
}

func sqrt(x float64) float64 {
	return math.Sqrt(x)
}

var _ = sqrt
