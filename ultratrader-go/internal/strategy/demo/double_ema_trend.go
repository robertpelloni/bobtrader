package demo

import (
	"context"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// DoubleEMATrendStrategy implements a trend-following strategy with an EMA filter,
// inspired by freqtrade references.
type DoubleEMATrendStrategy struct {
	mu          sync.Mutex
	accountID   string
	symbol      string
	quantity    string
	fastPeriod  int
	slowPeriod  int
	trendPeriod int
	history     []marketdata.Candle
	lastAction  string
}

func NewDoubleEMATrendStrategy(accountID, symbol, quantity string, fast, slow, trend int) *DoubleEMATrendStrategy {
	return &DoubleEMATrendStrategy{
		accountID:   accountID,
		symbol:      symbol,
		quantity:    quantity,
		fastPeriod:  fast,
		slowPeriod:  slow,
		trendPeriod: trend,
		history:     make([]marketdata.Candle, 0),
		lastAction:  "none",
	}
}

func (s *DoubleEMATrendStrategy) Name() string {
	return "double-ema-trend"
}

func (s *DoubleEMATrendStrategy) OnTick(ctx context.Context) ([]strategy.Signal, error) {
	return nil, nil
}

func (s *DoubleEMATrendStrategy) OnMarketTick(ctx context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	return nil, nil // Strategy operates on candles
}

func (s *DoubleEMATrendStrategy) OnMarketCandle(ctx context.Context, candle marketdata.Candle) ([]strategy.Signal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if candle.Symbol != s.symbol {
		return nil, nil
	}

	s.history = append(s.history, candle)
	if len(s.history) < s.trendPeriod {
		return nil, nil // Need more data for the trend EMA
	}

	// Simple EMA calculations (placeholder for a real indicator lib)
	prices := make([]float64, len(s.history))
	for i, c := range s.history {
		prices[i] = utils.ParseFloat(c.Close)
	}

	fastEMA := calculateEMA(prices, s.fastPeriod)
	slowEMA := calculateEMA(prices, s.slowPeriod)
	trendEMA := calculateEMA(prices, s.trendPeriod)

	currentPrice := utils.ParseFloat(candle.Close)

	// Buy criteria: fast > slow AND price > trend
	if fastEMA > slowEMA && currentPrice > trendEMA && s.lastAction != "buy" {
		s.lastAction = "buy"
		return []strategy.Signal{{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "buy",
			Quantity:  s.quantity,
			Reason:    "fast > slow EMA and price > trend EMA",
		}}, nil
	}

	// Sell criteria: fast < slow OR price < trend
	if (fastEMA < slowEMA || currentPrice < trendEMA) && s.lastAction == "buy" {
		s.lastAction = "sell"
		return []strategy.Signal{{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    "sell",
			Quantity:  s.quantity,
			Reason:    "trend reversal or trend filter violation",
		}}, nil
	}

	return nil, nil
}

func calculateEMA(prices []float64, period int) float64 {
	if len(prices) == 0 {
		return 0
	}
	k := 2.0 / (float64(period) + 1.0)
	ema := prices[0]
	for i := 1; i < len(prices); i++ {
		ema = prices[i]*k + ema*(1.0-k)
	}
	return ema
}
