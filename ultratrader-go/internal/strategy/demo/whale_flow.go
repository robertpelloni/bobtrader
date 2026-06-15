package demo

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/analytics/sentiment"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// WhaleFlowStrategy uses Whale Alert data to identify accumulation/distribution phases.
// High exchange outflows (bullish sentiment) trigger entries.
type WhaleFlowStrategy struct {
	mu             sync.Mutex
	accountID      string
	symbol         string
	quantity       string
	provider       *sentiment.WhaleAlertProvider
	threshold      float64 // Minimum bullish score (outflows) to trigger
	lastSignalTime time.Time
	cooldown       time.Duration
	lastAction     string
}

func NewWhaleFlow(accountID, symbol, quantity string, provider *sentiment.WhaleAlertProvider, threshold float64) *WhaleFlowStrategy {
	return &WhaleFlowStrategy{
		accountID: accountID,
		symbol:    symbol,
		quantity:  quantity,
		provider:  provider,
		threshold: threshold,
		cooldown:  4 * time.Hour, // Whale movements are macro; long cooldown
	}
}

func (s *WhaleFlowStrategy) Name() string { return fmt.Sprintf("whale-flow-%s", s.symbol) }

func (s *WhaleFlowStrategy) OnTick(ctx context.Context) ([]strategy.Signal, error) {
	if time.Since(s.lastSignalTime) < s.cooldown {
		return nil, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	sig, err := s.provider.FetchSentiment(ctx, s.symbol)
	if err != nil {
		return nil, nil
	}

	// WhaleAlert score: positive = more outflows (bullish), negative = more inflows (bearish)
	if sig.Score >= s.threshold && s.lastAction != "buy" {
		s.lastAction = "buy"
		s.lastSignalTime = time.Now()
		return []strategy.Signal{{
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "buy",
			Quantity:     s.quantity,
			Reason:       fmt.Sprintf("Whale Accumulation: Bullish outflow score %.2f", sig.Score),
			StrategyName: s.Name(),
		}}, nil
	}

	// If large inflows detected, consider exiting macro position
	if sig.Score <= -s.threshold && s.lastAction == "buy" {
		s.lastAction = "sell"
		s.lastSignalTime = time.Now()
		return []strategy.Signal{{
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "sell",
			Quantity:     s.quantity,
			Reason:       fmt.Sprintf("Whale Distribution: Bearish inflow score %.2f", sig.Score),
			StrategyName: s.Name(),
		}}, nil
	}

	return nil, nil
}
