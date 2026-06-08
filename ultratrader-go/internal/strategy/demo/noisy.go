package demo

import (
	"context"
	"fmt"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// NoisyStrategy generates a signal on every Nth tick for verification purposes.
type NoisyStrategy struct {
	mu        sync.Mutex
	accountID string
	symbol    string
	interval  int
	counter   int
}

func NewNoisyStrategy(accountID, symbol string, interval int) *NoisyStrategy {
	return &NoisyStrategy{
		accountID: accountID,
		symbol:    symbol,
		interval:  interval,
	}
}

func (s *NoisyStrategy) Name() string {
	return "noisy-verification"
}

func (s *NoisyStrategy) OnTick(ctx context.Context) ([]strategy.Signal, error) {
	return nil, nil
}

func (s *NoisyStrategy) OnMarketTick(ctx context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if tick.Symbol != s.symbol {
		return nil, nil
	}

	s.counter++
	if s.counter >= s.interval {
		s.counter = 0
		action := "buy"
		if timeToSell() { // Just a random toggle or alternating
			action = "sell"
		}
		return []strategy.Signal{{
			AccountID: s.accountID,
			Symbol:    s.symbol,
			Action:    action,
			Quantity:  "0.001",
			Reason:    fmt.Sprintf("noisy signal every %d ticks", s.interval),
		}}, nil
	}

	return nil, nil
}

var lastActionWasBuy bool
func timeToSell() bool {
	lastActionWasBuy = !lastActionWasBuy
	return lastActionWasBuy
}
