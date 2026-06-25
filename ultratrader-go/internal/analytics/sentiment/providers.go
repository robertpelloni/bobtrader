package sentiment

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/logging"
)

// CryptoNewsProvider fetches sentiment from crypto news APIs.
// Integrates with CryptoPanic, CoinGecko trending, and Fear/Greed Index.
type CryptoNewsProvider struct {
	name     string
	apiKey   string
	baseURL  string
	client   *http.Client
	cache    map[string]cachedSentiment
	mu       sync.RWMutex
	logger   *logging.Logger
	cacheTTL time.Duration
}

type cachedSentiment struct {
	signal    Signal
	timestamp time.Time
}

func NewCryptoNewsProvider(apiKey string, logger *logging.Logger) *CryptoNewsProvider {
	if logger == nil {
		logger, _ = logging.New(logging.Config{Stdout: true})
	}
	return &CryptoNewsProvider{
		name:     "crypto-news",
		apiKey:   apiKey,
		baseURL:  "https://cryptopanic.com/api/v1",
		client:   &http.Client{Timeout: 10 * time.Second},
		cache:    make(map[string]cachedSentiment),
		logger:   logger,
		cacheTTL: 5 * time.Minute,
	}
}

func (p *CryptoNewsProvider) Name() string { return p.name }

func (p *CryptoNewsProvider) FetchSentiment(ctx context.Context, symbol string) (Signal, error) {
	// Check cache
	p.mu.RLock()
	if cached, ok := p.cache[symbol]; ok {
		if time.Since(cached.timestamp) < p.cacheTTL {
			p.mu.RUnlock()
			return cached.signal, nil
		}
	}
	p.mu.RUnlock()

	// If no API key, return neutral sentiment immediately without making request
	if p.apiKey == "" {
		return Signal{
			Source:    p.name,
			Symbol:    symbol,
			Score:     0.0,
			Timestamp: time.Now(),
		}, nil
	}

	// Fetch from CryptoPanic API
	url := fmt.Sprintf("%s/posts/?auth_token=%s&currencies=%s&kind=news&filter=hot",
		p.baseURL, p.apiKey, strings.TrimSuffix(symbol, "USDT"))

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return Signal{
			Source:    p.name,
			Symbol:    symbol,
			Score:     0.0,
			Timestamp: time.Now(),
		}, nil
	}

	resp, err := p.client.Do(req)
	if err != nil {
		// Fallback: return neutral sentiment on error
		return Signal{
			Source:    p.name,
			Symbol:    symbol,
			Score:     0.0,
			Timestamp: time.Now(),
		}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.logger.Info("CryptoPanic API returned non-OK status", map[string]any{
			"status": resp.Status,
			"symbol": symbol,
		})
		return Signal{
			Source:    p.name,
			Symbol:    symbol,
			Score:     0.0,
			Timestamp: time.Now(),
		}, nil
	}

	var result struct {
		Results []struct {
			Title string `json:"title"`
			Votes struct {
				Positive int `json:"positive"`
				Negative int `json:"negative"`
			} `json:"votes"`
		} `json:"results"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		p.logger.Info("failed to decode CryptoPanic response", map[string]any{
			"error":  err.Error(),
			"symbol": symbol,
		})
		return Signal{
			Source:    p.name,
			Symbol:    symbol,
			Score:     0.0,
			Timestamp: time.Now(),
		}, nil
	}

	// Calculate sentiment from vote ratios
	score := p.calculateScore(result.Results)

	signal := Signal{
		Source:    p.name,
		Symbol:    symbol,
		Score:     score,
		Timestamp: time.Now(),
	}

	// Cache the result
	p.mu.Lock()
	p.cache[symbol] = cachedSentiment{signal: signal, timestamp: time.Now()}
	p.mu.Unlock()

	return signal, nil
}

func (p *CryptoNewsProvider) calculateScore(articles []struct {
	Title string `json:"title"`
	Votes struct {
		Positive int `json:"positive"`
		Negative int `json:"negative"`
	} `json:"votes"`
}) float64 {
	if len(articles) == 0 {
		return 0.0
	}

	totalPositive := 0
	totalNegative := 0

	for _, article := range articles {
		totalPositive += article.Votes.Positive
		totalNegative += article.Votes.Negative
	}

	total := totalPositive + totalNegative
	if total == 0 {
		return 0.0
	}

	// Score from -1.0 (all negative) to 1.0 (all positive)
	return float64(totalPositive-totalNegative) / float64(total)
}

// FearGreedProvider fetches the Crypto Fear & Greed Index.
type FearGreedProvider struct {
	name   string
	client *http.Client
	cache  *cachedSentiment
	mu     sync.RWMutex
	logger *logging.Logger
}

func NewFearGreedProvider(logger *logging.Logger) *FearGreedProvider {
	if logger == nil {
		logger, _ = logging.New(logging.Config{Stdout: true})
	}
	return &FearGreedProvider{
		name:   "fear-greed-index",
		client: &http.Client{Timeout: 10 * time.Second},
		logger: logger,
	}
}

func (p *FearGreedProvider) Name() string { return p.name }

func (p *FearGreedProvider) FetchSentiment(ctx context.Context, symbol string) (Signal, error) {
	// Fear/Greed is market-wide, not per-symbol
	p.mu.RLock()
	if p.cache != nil && time.Since(p.cache.timestamp) < 15*time.Minute {
		p.mu.RUnlock()
		return p.cache.signal, nil
	}
	p.mu.RUnlock()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.alternative.me/fng/?limit=1", nil)
	if err != nil {
		return Signal{Source: p.name, Symbol: symbol, Score: 0, Timestamp: time.Now()}, nil
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return Signal{Source: p.name, Symbol: symbol, Score: 0, Timestamp: time.Now()}, nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		p.logger.Info("FearGreed API returned non-OK status", map[string]any{
			"status": resp.Status,
		})
		return Signal{Source: p.name, Symbol: symbol, Score: 0, Timestamp: time.Now()}, nil
	}

	var result struct {
		Data []struct {
			Value string `json:"value"`
			Class string `json:"value_classification"`
		} `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		p.logger.Info("failed to decode FearGreed response", map[string]any{
			"error": err.Error(),
		})
		return Signal{Source: p.name, Symbol: symbol, Score: 0, Timestamp: time.Now()}, nil
	}

	if len(result.Data) == 0 {
		return Signal{Source: p.name, Symbol: symbol, Score: 0, Timestamp: time.Now()}, nil
	}

	// Convert 0-100 scale to -1.0 to 1.0
	// 0 = Extreme Fear, 50 = Neutral, 100 = Extreme Greed
	var value int
	fmt.Sscanf(result.Data[0].Value, "%d", &value)
	score := (float64(value) - 50) / 50 // -1.0 to 1.0

	signal := Signal{
		Source:    p.name,
		Symbol:    symbol,
		Score:     score,
		Timestamp: time.Now(),
	}

	p.mu.Lock()
	p.cache = &cachedSentiment{signal: signal, timestamp: time.Now()}
	p.mu.Unlock()

	return signal, nil
}

// MarketEventsProvider tracks major crypto market events.
type MarketEventsProvider struct {
	name   string
	events []MarketEvent
	mu     sync.RWMutex
	logger *logging.Logger
}

type MarketEvent struct {
	Name      string
	Date      time.Time
	Impact    float64 // -1.0 (bearish) to 1.0 (bullish)
	Symbol    string  // Empty = market-wide
	Recurring bool
}

func NewMarketEventsProvider(logger *logging.Logger) *MarketEventsProvider {
	if logger == nil {
		logger, _ = logging.New(logging.Config{Stdout: true})
	}

	p := &MarketEventsProvider{
		name:   "market-events",
		logger: logger,
	}

	// Pre-populate known events
	p.events = []MarketEvent{
		// Bitcoin Halvings
		{Name: "BTC Halving 2024", Date: time.Date(2024, 4, 20, 0, 0, 0, 0, time.UTC), Impact: 0.8, Symbol: "BTCUSDT"},
		{Name: "BTC Halving 2028", Date: time.Date(2028, 4, 0, 0, 0, 0, 0, time.UTC), Impact: 0.8, Symbol: "BTCUSDT"},

		// FOMC Meetings (approximate - 8 per year)
		// These affect risk assets including crypto
		{Name: "FOMC Rate Decision", Date: time.Date(2026, 1, 28, 19, 0, 0, 0, time.UTC), Impact: -0.3, Symbol: ""},
		{Name: "FOMC Rate Decision", Date: time.Date(2026, 3, 18, 19, 0, 0, 0, time.UTC), Impact: -0.3, Symbol: ""},
		{Name: "FOMC Rate Decision", Date: time.Date(2026, 5, 6, 19, 0, 0, 0, time.UTC), Impact: -0.3, Symbol: ""},
		{Name: "FOMC Rate Decision", Date: time.Date(2026, 6, 17, 19, 0, 0, 0, time.UTC), Impact: -0.3, Symbol: ""},
		{Name: "FOMC Rate Decision", Date: time.Date(2026, 7, 29, 19, 0, 0, 0, time.UTC), Impact: -0.3, Symbol: ""},
		{Name: "FOMC Rate Decision", Date: time.Date(2026, 9, 16, 19, 0, 0, 0, time.UTC), Impact: -0.3, Symbol: ""},
		{Name: "FOMC Rate Decision", Date: time.Date(2026, 11, 4, 19, 0, 0, 0, time.UTC), Impact: -0.3, Symbol: ""},
		{Name: "FOMC Rate Decision", Date: time.Date(2026, 12, 16, 19, 0, 0, 0, time.UTC), Impact: -0.3, Symbol: ""},

		// Bitcoin ETF decisions
		{Name: "BTC ETF Anniversary", Date: time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC), Impact: 0.2, Symbol: "BTCUSDT"},

		// Tax season (typically bearish for crypto)
		{Name: "US Tax Season", Date: time.Date(2026, 4, 15, 0, 0, 0, 0, time.UTC), Impact: -0.2, Symbol: ""},

		// Ethereum upgrades
		{Name: "ETH Pectra Upgrade", Date: time.Date(2026, 3, 0, 0, 0, 0, 0, time.UTC), Impact: 0.3, Symbol: "ETHUSDT"},
	}

	return p
}

func (p *MarketEventsProvider) Name() string { return p.name }

func (p *MarketEventsProvider) FetchSentiment(ctx context.Context, symbol string) (Signal, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	now := time.Now()
	totalImpact := 0.0
	eventCount := 0

	for _, event := range p.events {
		// Check if event is within 7 days (before or after)
		daysUntil := event.Date.Sub(now).Hours() / 24

		// Event is relevant if within 7 days before or after
		if daysUntil >= -7 && daysUntil <= 7 {
			// Weight by proximity (closer = stronger impact)
			proximityWeight := 1.0 - (math.Abs(daysUntil) / 7.0)

			// Check if event applies to this symbol
			if event.Symbol == "" || event.Symbol == symbol {
				totalImpact += event.Impact * proximityWeight
				eventCount++
			}
		}
	}

	score := 0.0
	if eventCount > 0 {
		score = totalImpact / float64(eventCount)
		// Clamp to [-1, 1]
		if score > 1.0 {
			score = 1.0
		}
		if score < -1.0 {
			score = -1.0
		}
	}

	return Signal{
		Source:    p.name,
		Symbol:    symbol,
		Score:     score,
		Timestamp: now,
	}, nil
}

// AddEvent allows dynamic addition of market events
func (p *MarketEventsProvider) AddEvent(event MarketEvent) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = append(p.events, event)
}

// StockMarketCorrelation provides sentiment based on stock market movements.
// Crypto often correlates with risk-on assets like QQQ/SPY.
type StockMarketCorrelation struct {
	name   string
	client *http.Client
	cache  *cachedSentiment
	mu     sync.RWMutex
	logger *logging.Logger
	apiKey string // Alpha Vantage or similar
}

func NewStockMarketCorrelation(apiKey string, logger *logging.Logger) *StockMarketCorrelation {
	if logger == nil {
		logger, _ = logging.New(logging.Config{Stdout: true})
	}
	return &StockMarketCorrelation{
		name:   "stock-market-correlation",
		client: &http.Client{Timeout: 10 * time.Second},
		logger: logger,
		apiKey: apiKey,
	}
}

func (p *StockMarketCorrelation) Name() string { return p.name }

func (p *StockMarketCorrelation) FetchSentiment(ctx context.Context, symbol string) (Signal, error) {
	p.mu.RLock()
	if p.cache != nil && time.Since(p.cache.timestamp) < 30*time.Minute {
		p.mu.RUnlock()
		return p.cache.signal, nil
	}
	p.mu.RUnlock()

	// If no API key, return neutral
	if p.apiKey == "" {
		return Signal{
			Source:    p.name,
			Symbol:    symbol,
			Score:     0.0,
			Timestamp: time.Now(),
		}, nil
	}

	// Fetch SPY (S&P 500 ETF) performance as proxy for risk sentiment
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol=SPY&apikey=%s", p.apiKey)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return Signal{Source: p.name, Symbol: symbol, Score: 0, Timestamp: time.Now()}, nil
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return Signal{Source: p.name, Symbol: symbol, Score: 0, Timestamp: time.Now()}, nil
	}
	defer resp.Body.Close()

	var result struct {
		GlobalQuote struct {
			ChangePercent string `json:"10. change percent"`
		} `json:"Global Quote"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Signal{Source: p.name, Symbol: symbol, Score: 0, Timestamp: time.Now()}, nil
	}

	// Parse change percent (e.g., "1.25%")
	var changePct float64
	fmt.Sscanf(strings.TrimRight(result.GlobalQuote.ChangePercent, "%"), "%f", &changePct)

	// Convert to sentiment: SPY up = risk-on = bullish for crypto
	// Clamp to [-1, 1] range (5% move = max sentiment)
	score := changePct / 5.0
	if score > 1.0 {
		score = 1.0
	}
	if score < -1.0 {
		score = -1.0
	}

	signal := Signal{
		Source:    p.name,
		Symbol:    symbol,
		Score:     score,
		Timestamp: time.Now(),
	}

	p.mu.Lock()
	p.cache = &cachedSentiment{signal: signal, timestamp: time.Now()}
	p.mu.Unlock()

	return signal, nil
}
