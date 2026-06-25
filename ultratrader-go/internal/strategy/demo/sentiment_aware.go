package demo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/analytics/sentiment"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// SentimentAwareStrategy combines multiple sentiment sources with technical analysis.
// It uses news, YouTube, Fear/Greed, market events, and stock market correlation
// to make trading decisions.
type SentimentAwareStrategy struct {
	accountID      string
	symbol         string
	quantity       string
	engine         *sentiment.Engine
	minConfidence  float64 // Minimum sentiment score to trigger (e.g., 0.3)
	lastSignalTime time.Time
	cooldown       time.Duration
	priceHistory   []float64
	maxHistory     int
	mu             sync.RWMutex
}

func NewSentimentAwareStrategy(
	accountID, symbol, quantity string,
	engine *sentiment.Engine,
	minConfidence float64,
) *SentimentAwareStrategy {
	if minConfidence <= 0 {
		minConfidence = 0.3
	}
	return &SentimentAwareStrategy{
		accountID:     accountID,
		symbol:        symbol,
		quantity:      quantity,
		engine:        engine,
		minConfidence: minConfidence,
		cooldown:      60 * time.Second, // 1 minute cooldown between signals
		priceHistory:  make([]float64, 0, 100),
		maxHistory:    100,
	}
}

func (s *SentimentAwareStrategy) Name() string {
	return fmt.Sprintf("sentiment-aware-%s", s.symbol)
}

func (s *SentimentAwareStrategy) OnTick(_ context.Context) ([]strategy.Signal, error) {
	return nil, nil
}

func (s *SentimentAwareStrategy) OnMarketTick(_ context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	if tick.Symbol != s.symbol {
		return nil, nil
	}

	price := utils.ParseFloat(tick.Price)
	if price <= 0 {
		return nil, nil
	}

	// Track price history
	s.mu.Lock()
	s.priceHistory = append(s.priceHistory, price)
	if len(s.priceHistory) > s.maxHistory {
		s.priceHistory = s.priceHistory[1:]
	}
	s.mu.Unlock()

	// Cooldown check
	if time.Since(s.lastSignalTime) < s.cooldown {
		return nil, nil
	}

	// Need price history for trend analysis
	s.mu.RLock()
	historyLen := len(s.priceHistory)
	s.mu.RUnlock()

	if historyLen < 20 {
		return nil, nil
	}

	// Fetch aggregated sentiment from all providers
	sentimentScore, sentimentSignals, err := s.engine.AggregateSentiment(context.Background(), s.symbol)
	if err != nil {
		return nil, nil
	}

	sourceCount := len(sentimentSignals)

	// Calculate technical trend (simple: price vs 20-period SMA)
	s.mu.RLock()
	sma := calculateSMA(s.priceHistory, 20)
	currentPrice := s.priceHistory[len(s.priceHistory)-1]
	s.mu.RUnlock()

	priceVsSMA := (currentPrice - sma) / sma * 100 // % above/below SMA

	// Combine sentiment with technicals
	// Sentiment range: -1.0 (bearish) to 1.0 (bullish)
	// Price vs SMA: negative = below trend, positive = above trend

	var signals []strategy.Signal

	// BUY SIGNAL: Strong bullish sentiment AND price below SMA (buy the dip)
	if sentimentScore >= s.minConfidence && priceVsSMA < -0.1 {
		signals = append(signals, strategy.Signal{
			StrategyName: s.Name(),
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "buy",
			Quantity:     s.quantity,
			Reason: fmt.Sprintf("SENTIMENT BUY: score=%.2f (bullish), price %.2f%% below SMA, sources=%d",
				sentimentScore, priceVsSMA, sourceCount),
		})
		s.lastSignalTime = time.Now()
	}

	// SELL SIGNAL: Strong bearish sentiment AND price above SMA (sell the rally)
	if sentimentScore <= -s.minConfidence && priceVsSMA > 0.1 {
		signals = append(signals, strategy.Signal{
			StrategyName: s.Name(),
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "sell",
			Quantity:     s.quantity,
			Reason: fmt.Sprintf("SENTIMENT SELL: score=%.2f (bearish), price %.2f%% above SMA, sources=%d",
				sentimentScore, priceVsSMA, sourceCount),
		})
		s.lastSignalTime = time.Now()
	}

	return signals, nil
}

// GetSentimentStatus returns current sentiment for diagnostics
func (s *SentimentAwareStrategy) GetSentimentStatus(ctx context.Context) map[string]interface{} {
	score, signals, err := s.engine.AggregateSentiment(ctx, s.symbol)
	if err != nil {
		return map[string]interface{}{"error": err.Error()}
	}

	s.mu.RLock()
	sma := 0.0
	currentPrice := 0.0
	if len(s.priceHistory) >= 20 {
		sma = calculateSMA(s.priceHistory, 20)
		currentPrice = s.priceHistory[len(s.priceHistory)-1]
	}
	s.mu.RUnlock()

	return map[string]interface{}{
		"symbol":          s.symbol,
		"sentiment_score": score,
		"source_count":    len(signals),
		"current_price":   currentPrice,
		"sma_20":          sma,
		"price_vs_sma":    (currentPrice - sma) / sma * 100,
	}
}

func calculateSMA(prices []float64, period int) float64 {
	if len(prices) < period {
		period = len(prices)
	}
	if period == 0 {
		return 0
	}
	sum := 0.0
	for _, p := range prices[len(prices)-period:] {
		sum += p
	}
	return sum / float64(period)
}
