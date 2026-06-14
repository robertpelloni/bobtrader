package sentiment

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/logging"
)

// YouTubeSentimentProvider analyzes crypto YouTube channels for market sentiment.
// Monitors channels like Arcane Bear, Benjamin Cowen, Coin Bureau, etc.
type YouTubeSentimentProvider struct {
	name       string
	apiKey     string // YouTube Data API key
	client     *http.Client
	cache      map[string]cachedSentiment
	mu         sync.RWMutex
	logger     *logging.Logger
	cacheTTL   time.Duration
	channels   []YouTubeChannel
	keywords   SentimentKeywords
}

type YouTubeChannel struct {
	ChannelID   string
	Name        string
	Weight      float64 // How influential (0.0-1.0)
	Focus       []string // e.g., ["BTC", "ETH", "macro"]
}

type SentimentKeywords struct {
	Bullish []string
	Bearish []string
}

func NewYouTubeSentimentProvider(apiKey string, logger *logging.Logger) *YouTubeSentimentProvider {
	if logger == nil {
		logger, _ = logging.New(logging.Config{Stdout: true})
	}

	return &YouTubeSentimentProvider{
		name:     "youtube-sentiment",
		apiKey:   apiKey,
		client:   &http.Client{Timeout: 15 * time.Second},
		cache:    make(map[string]cachedSentiment),
		logger:   logger,
		cacheTTL: 30 * time.Minute, // YouTube videos change slowly
		channels: []YouTubeChannel{
			// Top crypto YouTube channels with their focus areas
			{ChannelID: "UCAL3JXZSzSm8AlZyD3pQqiA", Name: "Arcane Bear", Weight: 0.8, Focus: []string{"BTC", "macro", "trading"}},
			{ChannelID: "UCRvqjQPSeaWn-uEx-w0XOIg", Name: "Benjamin Cowen", Weight: 0.9, Focus: []string{"BTC", "ETH", "analysis"}},
			{ChannelID: "UCqK_GSMbpiV8spgD3ZGloSw", Name: "Coin Bureau", Weight: 0.85, Focus: []string{"altcoins", "macro", "regulation"}},
			{ChannelID: "UCbNQ8DPGm_bq2jxqDlfWJ8w", Name: "The Moon", Weight: 0.7, Focus: []string{"BTC", "trading"}},
			{ChannelID: "UCbiWJYRgKluBdUERjKCkVRA", Name: "BitBoy Crypto", Weight: 0.6, Focus: []string{"altcoins", "news"}},
			{ChannelID: "UCjemQfjaXg3L1Broz2uXHxA", Name: "Crypto Banter", Weight: 0.75, Focus: []string{"trading", "macro", "BTC"}},
			{ChannelID: "UCMiJtkE4q7R5bGNr6dAbpNg", Name: "InvestAnswers", Weight: 0.8, Focus: []string{"BTC", "stocks", "macro"}},
			{ChannelID: "UCkN2MFbYESqLo6R2x3pXsYQ", Name: "Crypto Zombie", Weight: 0.65, Focus: []string{"BTC", "trading"}},
		},
		keywords: SentimentKeywords{
			Bullish: []string{
				"bullish", "moon", "pump", "rally", "breakout", "accumulate", "buy the dip",
				"long", "bull run", "new highs", "adoption", "institutional", "etf approval",
				"halving", "supply shock", "diamond hands", "hodl", "undervalued", "bottom is in",
				"massive gains", "100k", "200k", "500k", "million", "explosive", "parabolic",
			},
			Bearish: []string{
				"bearish", "crash", "dump", "sell", "short", "bear market", "correction",
				"overvalued", "bubble", "scam", "fraud", "regulation", "ban", "hack",
				"death cross", "capitulation", "rekt", "liquidation", "panic", "fear",
				"bottom not in", "more pain", "lower lows", "resistance", "reject",
			},
		},
	}
}

func (p *YouTubeSentimentProvider) Name() string { return p.name }

func (p *YouTubeSentimentProvider) FetchSentiment(ctx context.Context, symbol string) (Signal, error) {
	// Check cache
	p.mu.RLock()
	if cached, ok := p.cache[symbol]; ok {
		if time.Since(cached.timestamp) < p.cacheTTL {
			p.mu.RUnlock()
			return cached.signal, nil
		}
	}
	p.mu.RUnlock()

	// If no API key, use keyword-based fallback
	if p.apiKey == "" {
		return p.fallbackSentiment(symbol), nil
	}

	// Fetch recent videos from each channel
	totalScore := 0.0
	totalWeight := 0.0

	for _, channel := range p.channels {
		score, weight, err := p.analyzeChannel(ctx, channel, symbol)
		if err != nil {
			p.logger.Info("youtube channel analysis failed", map[string]any{
				"channel": channel.Name,
				"error":   err.Error(),
			})
			continue
		}
		totalScore += score * weight
		totalWeight += weight
	}

	// Normalize
	normalizedScore := 0.0
	if totalWeight > 0 {
		normalizedScore = totalScore / totalWeight
	}

	signal := Signal{
		Source:    p.name,
		Symbol:    symbol,
		Score:     normalizedScore,
		Timestamp: time.Now(),
	}

	// Cache
	p.mu.Lock()
	p.cache[symbol] = cachedSentiment{signal: signal, timestamp: time.Now()}
	p.mu.Unlock()

	return signal, nil
}

func (p *YouTubeSentimentProvider) analyzeChannel(ctx context.Context, channel YouTubeChannel, symbol string) (float64, float64, error) {
	// Fetch recent videos from channel
	url := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/search?key=%s&channelId=%s&part=snippet&order=date&maxResults=5&type=video",
		p.apiKey, channel.ChannelID,
	)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return 0, 0, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	var result struct {
		Items []struct {
			Snippet struct {
				Title       string `json:"title"`
				Description string `json:"description"`
			} `json:"snippet"`
		} `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, 0, err
	}

	// Analyze sentiment from titles and descriptions
	score := p.analyzeTextSentiment(result.Items, symbol)

	return score, channel.Weight, nil
}

func (p *YouTubeSentimentProvider) analyzeTextSentiment(videos []struct {
	Snippet struct {
		Title       string `json:"title"`
		Description string `json:"description"`
	} `json:"snippet"`
}, symbol string) float64 {

	baseCoin := strings.TrimSuffix(symbol, "USDT")
	bullCount := 0
	bearCount := 0

	for _, video := range videos {
		text := strings.ToLower(video.Snippet.Title + " " + video.Snippet.Description)

		// Check if video is relevant to this symbol
		if !strings.Contains(text, strings.ToLower(baseCoin)) &&
			!strings.Contains(text, "crypto") &&
			!strings.Contains(text, "bitcoin") {
			continue
		}

		// Count bullish/bearish keywords
		for _, kw := range p.keywords.Bullish {
			if strings.Contains(text, kw) {
				bullCount++
			}
		}
		for _, kw := range p.keywords.Bearish {
			if strings.Contains(text, kw) {
				bearCount++
			}
		}
	}

	total := bullCount + bearCount
	if total == 0 {
		return 0.0
	}

	return float64(bullCount-bearCount) / float64(total)
}

// fallbackSentiment returns neutral sentiment when API is unavailable
func (p *YouTubeSentimentProvider) fallbackSentiment(symbol string) Signal {
	return Signal{
		Source:    p.name,
		Symbol:    symbol,
		Score:     0.0, // Neutral when we can't fetch
		Timestamp: time.Now(),
	}
}

// AddChannel allows adding new YouTube channels dynamically
func (p *YouTubeSentimentProvider) AddChannel(channel YouTubeChannel) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.channels = append(p.channels, channel)
}

// GetMonitoredChannels returns the list of channels being monitored
func (p *YouTubeSentimentProvider) GetMonitoredChannels() []YouTubeChannel {
	p.mu.RLock()
	defer p.mu.RUnlock()
	channels := make([]YouTubeChannel, len(p.channels))
	copy(channels, p.channels)
	return channels
}
