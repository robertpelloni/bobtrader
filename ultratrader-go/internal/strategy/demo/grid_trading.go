package demo

import (
	"context"
	"fmt"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// GridTradingStrategy implements a neutral grid trading strategy.
// It places buy orders below the current price and sell orders above.
type GridTradingStrategy struct {
	mu           sync.Mutex
	accountID    string
	symbol       string
	quantity     string
	gridLevels   int
	gridSpacing  float64 // percentage spacing between levels
	centerPrice  float64
	initialized  bool
}

func NewGridTrading(accountID, symbol, quantity string, levels int, spacing float64) *GridTradingStrategy {
	return &GridTradingStrategy{
		accountID:   accountID,
		symbol:      symbol,
		quantity:    quantity,
		gridLevels:  levels,
		gridSpacing: spacing,
	}
}

func (s *GridTradingStrategy) Name() string { return fmt.Sprintf("grid-%s", s.symbol) }

func (s *GridTradingStrategy) OnTick(_ context.Context) ([]strategy.Signal, error) { return nil, nil }

func (s *GridTradingStrategy) OnMarketTick(ctx context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	if tick.Symbol != s.symbol {
		return nil, nil
	}

	price := utils.ParseFloat(tick.Price)
	if price <= 0 {
		return nil, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.initialized {
		s.centerPrice = price
		s.initialized = true
		return nil, nil
	}

	// Simple logic: if price moved more than spacing, emit signal to capture volatility
	// and potentially re-center (simplified for demo)
	diff := (price - s.centerPrice) / s.centerPrice
	if diff >= s.gridSpacing {
		s.centerPrice = price // Re-center
		return []strategy.Signal{{
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "sell",
			Quantity:     s.quantity,
			Reason:       "grid level reached: profit take",
			StrategyName: s.Name(),
		}}, nil
	} else if diff <= -s.gridSpacing {
		s.centerPrice = price // Re-center
		return []strategy.Signal{{
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "buy",
			Quantity:     s.quantity,
			Reason:       "grid level reached: buy dip",
			StrategyName: s.Name(),
		}}, nil
	}

	return nil, nil
}
