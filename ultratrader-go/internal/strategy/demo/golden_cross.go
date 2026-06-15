package demo

import (
	"context"
	"fmt"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/indicator"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// GoldenCrossStrategy implements the classic 50/200 SMA crossover.
type GoldenCrossStrategy struct {
	mu           sync.Mutex
	accountID    string
	symbol       string
	quantity     string
	fastSMA      *indicator.SMA
	slowSMA      *indicator.SMA
	lastAction   string
	warmup       int
	slowPeriod   int
}

func NewGoldenCross(accountID, symbol, quantity string, fast, slow int) *GoldenCrossStrategy {
	return &GoldenCrossStrategy{
		accountID:  accountID,
		symbol:     symbol,
		quantity:   quantity,
		fastSMA:    indicator.NewSMA(fast),
		slowSMA:    indicator.NewSMA(slow),
		slowPeriod: slow,
	}
}

func (s *GoldenCrossStrategy) Name() string { return fmt.Sprintf("golden-cross-%s", s.symbol) }

func (s *GoldenCrossStrategy) OnTick(_ context.Context) ([]strategy.Signal, error) { return nil, nil }

func (s *GoldenCrossStrategy) OnMarketCandle(ctx context.Context, candle marketdata.Candle) ([]strategy.Signal, error) {
	if candle.Symbol != s.symbol {
		return nil, nil
	}

	price := utils.ParseFloat(candle.Close)
	s.mu.Lock()
	defer s.mu.Unlock()

	fast := s.fastSMA.Update(price)
	slow := s.slowSMA.Update(price)
	s.warmup++

	if s.warmup < s.slowPeriod {
		return nil, nil
	}

	if fast > slow && s.lastAction != "buy" {
		s.lastAction = "buy"
		return []strategy.Signal{{
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "buy",
			Quantity:     s.quantity,
			Reason:       "Golden Cross: 50 SMA > 200 SMA",
			StrategyName: s.Name(),
		}}, nil
	}

	if fast < slow && s.lastAction == "buy" {
		s.lastAction = "sell"
		return []strategy.Signal{{
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "sell",
			Quantity:     s.quantity,
			Reason:       "Death Cross: 50 SMA < 200 SMA",
			StrategyName: s.Name(),
		}}, nil
	}

	return nil, nil
}
