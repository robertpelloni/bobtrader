package sentiment

import (
	"context"
	"encoding/xml"
	"net/http"
	"strings"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/logging"
)

type RSS struct {
	Items []Item `xml:"channel>item"`
}

type Item struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

// NewsAggregator fetches and classifies news from RSS feeds.
type NewsAggregator struct {
	logger *logging.Logger
	feeds  []string
}

func NewNewsAggregator(logger *logging.Logger) *NewsAggregator {
	return &NewsAggregator{
		logger: logger,
		feeds: []string{
			"https://cointelegraph.com/rss",
			"https://www.coindesk.com/arc/outboundfeeds/rss/",
		},
	}
}

// GetSentiment calculates a keyword-based sentiment score (-1.0 to 1.0).
func (a *NewsAggregator) GetSentiment(ctx context.Context, query string) (float64, error) {
	var allItems []Item
	client := &http.Client{Timeout: 5 * time.Second}

	for _, feed := range a.feeds {
		req, _ := http.NewRequestWithContext(ctx, "GET", feed, nil)
		resp, err := client.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		var rss RSS
		if err := xml.NewDecoder(resp.Body).Decode(&rss); err != nil {
			continue
		}
		allItems = append(allItems, rss.Items...)
	}

	if len(allItems) == 0 {
		return 0, nil
	}

	bullishKeys := []string{"moon", "surge", "breakout", "rally", "gain", "etf", "adoption", "high", "buy"}
	bearishKeys := []string{"crash", "dump", "plummet", "sec", "lawsuit", "scam", "hack", "low", "sell"}

	score := 0.0
	count := 0

	for _, item := range allItems {
		text := strings.ToLower(item.Title + " " + item.Description)
		if query != "" && !strings.Contains(text, strings.ToLower(query)) {
			continue
		}

		itemScore := 0.0
		for _, k := range bullishKeys {
			if strings.Contains(text, k) {
				itemScore += 0.2
			}
		}
		for _, k := range bearishKeys {
			if strings.Contains(text, k) {
				itemScore -= 0.2
			}
		}

		// Clamp item score
		if itemScore > 1.0 { itemScore = 1.0 }
		if itemScore < -1.0 { itemScore = -1.0 }

		score += itemScore
		count++
	}

	if count == 0 {
		return 0, nil
	}

	return score / float64(count), nil
}

// AggregatorProvider wraps NewsAggregator to implement sentiment.Provider.
type AggregatorProvider struct {
	agg *NewsAggregator
}

func NewAggregatorProvider(logger *logging.Logger) *AggregatorProvider {
	return &AggregatorProvider{agg: NewNewsAggregator(logger)}
}

func (p *AggregatorProvider) Name() string { return "rss-aggregator" }

func (p *AggregatorProvider) FetchSentiment(ctx context.Context, symbol string) (Signal, error) {
	score, err := p.agg.GetSentiment(ctx, symbol)
	if err != nil {
		return Signal{}, err
	}
	return Signal{
		Source:    p.Name(),
		Symbol:    symbol,
		Score:     score,
		Timestamp: time.Now(),
	}, nil
}
