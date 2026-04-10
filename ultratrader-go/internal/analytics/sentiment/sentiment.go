package sentiment

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"time"
)

// Config holds settings for the Sentiment Analyzer.
type Config struct {
	Enabled                bool
	FearGreedEnabled       bool
	SocialSentimentEnabled bool
	MCPURL                 string
	ImpactMultiplier       float64
}

// Analyzer handles fetching and normalizing sentiment data from external MCP servers.
type Analyzer struct {
	config Config
	client *http.Client
}

// NewAnalyzer creates a new SentimentAnalyzer.
func NewAnalyzer(config Config) *Analyzer {
	return &Analyzer{
		config: config,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// makeMCPRequest is a generic helper to connect to an MCP server exposed via HTTP.
func (a *Analyzer) makeMCPRequest(endpoint string, payload interface{}) (map[string]interface{}, error) {
	if a.config.MCPURL == "" {
		return nil, fmt.Errorf("MCP URL is empty")
	}

	url := fmt.Sprintf("%s/%s", a.config.MCPURL, endpoint)

	var req *http.Request
	var err error

	if payload != nil {
		body, _ := json.Marshal(payload)
		req, err = http.NewRequest("POST", url, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest("GET", url, nil)
	}

	if err != nil {
		return nil, err
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("MCP server returned status code %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(respBody, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// GetGlobalFearGreedIndex fetches the global fear and greed index.
// Returns a normalized score: -1.0 (Extreme Fear) to 1.0 (Extreme Greed).
func (a *Analyzer) GetGlobalFearGreedIndex() float64 {
	if !a.config.FearGreedEnabled {
		return 0.0
	}

	data, err := a.makeMCPRequest("api/fear-greed", nil)
	if err != nil {
		log.Printf("Failed to get fear/greed index: %v", err)
		return 0.0
	}

	if score, ok := data["score"].(float64); ok {
		// Assuming MCP returns 0-100
		// 0 = Extreme Fear, 100 = Extreme Greed
		normalized := (score - 50.0) / 50.0
		return normalized
	}

	return 0.0
}

// GetCoinSentiment fetches the social/news sentiment specifically for a given coin.
// Returns a normalized score from -1.0 to 1.0.
func (a *Analyzer) GetCoinSentiment(symbol string) float64 {
	if !a.config.SocialSentimentEnabled {
		return 0.0
	}

	payload := map[string]string{"symbol": symbol}
	data, err := a.makeMCPRequest("api/coin-sentiment", payload)
	if err != nil {
		log.Printf("Failed to get coin sentiment for %s: %v", symbol, err)
		return 0.0
	}

	if score, ok := data["sentiment_score"].(float64); ok {
		return score
	}

	return 0.0
}

// CalculateCombinedSentiment calculates a blended sentiment score.
func (a *Analyzer) CalculateCombinedSentiment(symbol string) float64 {
	if !a.config.Enabled {
		return 0.0
	}

	globalFG := a.GetGlobalFearGreedIndex()
	coinScore := a.GetCoinSentiment(symbol)

	// Blend logic: 70% Coin-specific sentiment, 30% Global Market trend
	blended := (coinScore * 0.70) + (globalFG * 0.30)

	// Apply the user's impact multiplier
	finalScore := blended * a.config.ImpactMultiplier

	// Bound strictly between -1.0 and 1.0
	return math.Max(-1.0, math.Min(1.0, finalScore))
}
