package sentiment_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/analytics/sentiment"
)

func TestSentimentAnalyzer_CombinedSentiment(t *testing.T) {
	// Setup mock MCP server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/api/fear-greed" {
			// Returns 75 (Greed) -> normalized to (75-50)/50 = 0.5
			json.NewEncoder(w).Encode(map[string]interface{}{"score": 75.0})
			return
		}

		if r.URL.Path == "/api/coin-sentiment" {
			// Returns 0.8 (Very Positive)
			json.NewEncoder(w).Encode(map[string]interface{}{"sentiment_score": 0.8})
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	config := sentiment.Config{
		Enabled:                true,
		FearGreedEnabled:       true,
		SocialSentimentEnabled: true,
		MCPURL:                 server.URL,
		ImpactMultiplier:       1.0,
	}

	analyzer := sentiment.NewAnalyzer(config)

	score := analyzer.CalculateCombinedSentiment("BTC")

	// Expected calculation:
	// globalFG = 0.5
	// coinScore = 0.8
	// blended = (0.8 * 0.7) + (0.5 * 0.3) = 0.56 + 0.15 = 0.71

	assert.InDelta(t, 0.71, score, 0.001)

	// Test bounds with high multiplier
	config.ImpactMultiplier = 2.0
	analyzer2 := sentiment.NewAnalyzer(config)
	score2 := analyzer2.CalculateCombinedSentiment("BTC")
	// 0.71 * 2 = 1.42 -> bounded to 1.0
	assert.Equal(t, 1.0, score2)
}

func TestSentimentAnalyzer_DisabledFeatures(t *testing.T) {
	config := sentiment.Config{
		Enabled:                true,
		FearGreedEnabled:       false,
		SocialSentimentEnabled: false,
		ImpactMultiplier:       1.0,
	}

	analyzer := sentiment.NewAnalyzer(config)
	score := analyzer.CalculateCombinedSentiment("BTC")
	assert.Equal(t, 0.0, score)
}
