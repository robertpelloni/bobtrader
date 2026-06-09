package execution

import (
	"context"
	"fmt"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

// ProfitBank implements a strategy to sell when a specific profit target is hit.
type ProfitBank struct {
	mu           sync.Mutex
	adapter      exchange.Adapter
	targetProfit float64 // percentage
}

func NewProfitBank(adapter exchange.Adapter, targetPct float64) *ProfitBank {
	return &ProfitBank{
		adapter:      adapter,
		targetProfit: targetPct,
	}
}

func (s *ProfitBank) Name() string { return "profit-bank" }

func (s *ProfitBank) Update(ctx context.Context, currentPrice, entryPrice float64) bool {
	margin := (currentPrice - entryPrice) / entryPrice * 100
	return margin >= s.targetProfit
}

// PreventLoss implements a strategy to sell before profit turns into a loss.
type PreventLoss struct {
	mu            sync.Mutex
	adapter       exchange.Adapter
	triggerProfit float64 // profit level to activate prevent loss
	minProfit     float64 // level to sell at if price drops back
	activated     bool
}

func NewPreventLoss(adapter exchange.Adapter, triggerPct, minPct float64) *PreventLoss {
	return &PreventLoss{
		adapter:       adapter,
		triggerProfit: triggerPct,
		minProfit:     minPct,
	}
}

func (s *PreventLoss) Name() string { return "prevent-loss" }

func (s *PreventLoss) Update(ctx context.Context, currentPrice, entryPrice float64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	margin := (currentPrice - entryPrice) / entryPrice * 100

	if !s.activated && margin >= s.triggerProfit {
		s.activated = true
		fmt.Println("PreventLoss: Activated")
	}

	if s.activated && margin <= s.minProfit {
		return true
	}

	return false
}
