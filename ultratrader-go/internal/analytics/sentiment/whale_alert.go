package sentiment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/logging"
)

// WhaleAlertProvider monitors large crypto transfers that signal whale activity.
// Large inflows to exchanges = potential sell pressure (bearish)
// Large outflows from exchanges = accumulation (bullish)
// Large wallet-to-wallet = neutral but worth monitoring
type WhaleAlertProvider struct {
	name     string
	apiKey   string // Whale Alert API key (free tier available)
	client   *http.Client
	cache    map[string]cachedWhaleData
	mu       sync.RWMutex
	logger   *logging.Logger
	cacheTTL time.Duration
	minUSD   float64 // Minimum transaction size to consider (e.g., $1M)
}

type cachedWhaleData struct {
	signal    Signal
	timestamp time.Time
	alerts    []WhaleAlert
}

// WhaleAlert represents a large transaction detected
type WhaleAlert struct {
	TxHash        string
	From          string
	To            string
	Amount        float64
	Symbol        string
	USDValue      float64
	Timestamp     time.Time
	AlertType     string // "exchange_inflow", "exchange_outflow", "wallet_transfer", "mint", "burn"
	IsExchangeIn  bool
	IsExchangeOut bool
}

func NewWhaleAlertProvider(apiKey string, minUSD float64, logger *logging.Logger) *WhaleAlertProvider {
	if logger == nil {
		logger, _ = logging.New(logging.Config{Stdout: true})
	}
	if minUSD <= 0 {
		minUSD = 500000 // $500K minimum by default
	}
	return &WhaleAlertProvider{
		name:     "whale-alert",
		apiKey:   apiKey,
		client:   &http.Client{Timeout: 15 * time.Second},
		cache:    make(map[string]cachedWhaleData),
		logger:   logger,
		cacheTTL: 2 * time.Minute,
		minUSD:   minUSD,
	}
}

func (p *WhaleAlertProvider) Name() string { return p.name }

func (p *WhaleAlertProvider) FetchSentiment(ctx context.Context, symbol string) (Signal, error) {
	// Check cache
	p.mu.RLock()
	if cached, ok := p.cache[symbol]; ok {
		if time.Since(cached.timestamp) < p.cacheTTL {
			p.mu.RUnlock()
			return cached.signal, nil
		}
	}
	p.mu.RUnlock()

	// If no API key, use mock data based on known patterns
	if p.apiKey == "" {
		return p.generatePatternBasedSentiment(symbol), nil
	}

	// Fetch from Whale Alert API
	// Free tier: 10 requests/min, last 3600 seconds of transactions
	baseCoin := trimSuffix(symbol, "USDT")
	url := fmt.Sprintf(
		"https://api.whale-alert.io/v1/transactions?api_key=%s&min_value=%d&currency=%s&limit=20",
		p.apiKey, int64(p.minUSD), baseCoin,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return p.generatePatternBasedSentiment(symbol), nil
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return p.generatePatternBasedSentiment(symbol), nil
	}
	defer resp.Body.Close()

	var result struct {
		Count        int `json:"count"`
		Transactions []struct {
			Hash      string `json:"hash"`
			Timestamp int64  `json:"timestamp"`
			From      struct {
				Address   string `json:"address"`
				Owner     string `json:"owner"`
				OwnerType string `json:"owner_type"` // "exchange", "unknown", "wallet"
			} `json:"from"`
			To struct {
				Address   string `json:"address"`
				Owner     string `json:"owner"`
				OwnerType string `json:"owner_type"`
			} `json:"to"`
			Amount           float64 `json:"amount"`
			Symbol           string  `json:"symbol"`
			TransactionCount int     `json:"transaction_count"`
		} `json:"transactions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return p.generatePatternBasedSentiment(symbol), nil
	}

	// Analyze whale movements
	alerts := p.analyzeTransactions(result.Transactions)
	score := p.calculateWhaleScore(alerts)

	signal := Signal{
		Source:    p.name,
		Symbol:    symbol,
		Score:     score,
		Timestamp: time.Now(),
	}

	// Cache
	p.mu.Lock()
	p.cache[symbol] = cachedWhaleData{
		signal:    signal,
		timestamp: time.Now(),
		alerts:    alerts,
	}
	p.mu.Unlock()

	return signal, nil
}

func (p *WhaleAlertProvider) analyzeTransactions(txs []struct {
	Hash      string `json:"hash"`
	Timestamp int64  `json:"timestamp"`
	From      struct {
		Address   string `json:"address"`
		Owner     string `json:"owner"`
		OwnerType string `json:"owner_type"`
	} `json:"from"`
	To struct {
		Address   string `json:"address"`
		Owner     string `json:"owner"`
		OwnerType string `json:"owner_type"`
	} `json:"to"`
	Amount           float64 `json:"amount"`
	Symbol           string  `json:"symbol"`
	TransactionCount int     `json:"transaction_count"`
}) []WhaleAlert {

	var alerts []WhaleAlert
	for _, tx := range txs {
		alert := WhaleAlert{
			TxHash:    tx.Hash,
			From:      tx.From.Owner,
			To:        tx.To.Owner,
			Amount:    tx.Amount,
			Symbol:    tx.Symbol,
			Timestamp: time.Unix(tx.Timestamp, 0),
		}

		// Classify the transaction
		if tx.From.OwnerType == "exchange" && tx.To.OwnerType != "exchange" {
			alert.AlertType = "exchange_outflow"
			alert.IsExchangeOut = true
		} else if tx.From.OwnerType != "exchange" && tx.To.OwnerType == "exchange" {
			alert.AlertType = "exchange_inflow"
			alert.IsExchangeIn = true
		} else if tx.From.OwnerType == "exchange" && tx.To.OwnerType == "exchange" {
			alert.AlertType = "exchange_transfer"
		} else {
			alert.AlertType = "wallet_transfer"
		}

		alerts = append(alerts, alert)
	}
	return alerts
}

func (p *WhaleAlertProvider) calculateWhaleScore(alerts []WhaleAlert) float64 {
	if len(alerts) == 0 {
		return 0
	}

	inflowVolume := 0.0  // Into exchanges (bearish)
	outflowVolume := 0.0 // Out of exchanges (bullish)

	for _, alert := range alerts {
		if alert.IsExchangeIn {
			inflowVolume += alert.USDValue
		} else if alert.IsExchangeOut {
			outflowVolume += alert.USDValue
		}
	}

	totalVolume := inflowVolume + outflowVolume
	if totalVolume == 0 {
		return 0
	}

	// Net flow: positive = more outflows (bullish), negative = more inflows (bearish)
	netFlow := outflowVolume - inflowVolume

	// Normalize to [-1, 1] range
	// Cap at $50M for normalization
	normalized := netFlow / 50000000
	if normalized > 1.0 {
		normalized = 1.0
	}
	if normalized < -1.0 {
		normalized = -1.0
	}

	return normalized
}

// generatePatternBasedSentiment provides sentiment based on known whale patterns
// when API is unavailable
func (p *WhaleAlertProvider) generatePatternBasedSentiment(symbol string) Signal {
	// Known whale behavior patterns:
	// - Whales tend to accumulate during dips
	// - Large outflows from exchanges often precede pumps
	// - Sunday/Monday often sees accumulation
	// - End of month often sees profit-taking

	now := time.Now().UTC()
	hour := now.Hour()
	weekday := now.Weekday()
	dayOfMonth := now.Day()

	score := 0.0

	// Pattern 1: Sunday accumulation (whales buy weekend dips)
	if weekday == time.Sunday {
		score += 0.15
	}

	// Pattern 2: Early morning Asia (whale activity window)
	if hour >= 0 && hour <= 4 {
		score += 0.1
	}

	// Pattern 3: End of month profit-taking
	if dayOfMonth >= 28 {
		score -= 0.1
	}

	// Pattern 4: Monday morning dip buying
	if weekday == time.Monday && hour >= 6 && hour <= 10 {
		score += 0.2
	}

	// Clamp to [-1, 1]
	if score > 1.0 {
		score = 1.0
	}
	if score < -1.0 {
		score = -1.0
	}

	return Signal{
		Source:    p.name,
		Symbol:    symbol,
		Score:     score,
		Timestamp: now,
	}
}

// GetRecentAlerts returns recent whale alerts for diagnostics
func (p *WhaleAlertProvider) GetRecentAlerts(symbol string) []WhaleAlert {
	p.mu.RLock()
	defer p.mu.RUnlock()
	if cached, ok := p.cache[symbol]; ok {
		return cached.alerts
	}
	return nil
}

func trimSuffix(s, suffix string) string {
	if len(s) >= len(suffix) && s[len(s)-len(suffix):] == suffix {
		return s[:len(s)-len(suffix)]
	}
	return s
}
