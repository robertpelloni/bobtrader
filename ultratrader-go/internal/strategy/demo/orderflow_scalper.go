package demo

import (
	"context"
	"fmt"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/analytics"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// OrderflowScalper uses Cumulative Volume Delta (CVD) and divergence detection
// to identify high-frequency entry points where volume precedes price action.
type OrderflowScalper struct {
	mu           sync.Mutex
	accountID    string
	symbol       string
	quantity     string
	analyzer     *analytics.OrderFlowAnalyzer
	prices       []float64
	buyVolumes   []float64
	sellVolumes  []float64
	lookback     int
	lastAction   string
}

func NewOrderflowScalper(accountID, symbol, quantity string, lookback int) *OrderflowScalper {
	return &OrderflowScalper{
		accountID:   accountID,
		symbol:      symbol,
		quantity:    quantity,
		analyzer:    analytics.NewOrderFlowAnalyzer(lookback),
		lookback:    lookback,
		prices:      make([]float64, 0, 100),
		buyVolumes:  make([]float64, 0, 100),
		sellVolumes: make([]float64, 0, 100),
	}
}

func (s *OrderflowScalper) Name() string { return fmt.Sprintf("orderflow-scalper-%s", s.symbol) }

func (s *OrderflowScalper) OnTick(_ context.Context) ([]strategy.Signal, error) { return nil, nil }

func (s *OrderflowScalper) OnMarketTick(ctx context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	if tick.Symbol != s.symbol {
		return nil, nil
	}

	price := utils.ParseFloat(tick.Price)
	qty := utils.ParseFloat(tick.Quantity)
	if price <= 0 {
		return nil, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// In a real orderflow system, we'd need to know if the tick was a buy or sell.
	// For this demo, we'll heuristics: price increase = buy volume, decrease = sell volume.
	bv, sv := 0.0, 0.0
	if len(s.prices) > 0 {
		if price >= s.prices[len(s.prices)-1] {
			bv = qty
		} else {
			sv = qty
		}
	} else {
		bv = qty // initial
	}

	s.prices = append(s.prices, price)
	s.buyVolumes = append(s.buyVolumes, bv)
	s.sellVolumes = append(s.sellVolumes, sv)

	if len(s.prices) > 100 {
		s.prices = s.prices[1:]
		s.buyVolumes = s.buyVolumes[1:]
		s.sellVolumes = s.sellVolumes[1:]
	}

	if len(s.prices) < s.lookback {
		return nil, nil
	}

	_, divergence, err := s.analyzer.Analyze(s.prices, s.buyVolumes, s.sellVolumes)
	if err != nil {
		return nil, nil
	}

	if divergence == analytics.BullishDivergence && s.lastAction != "buy" {
		s.lastAction = "buy"
		return []strategy.Signal{{
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "buy",
			Quantity:     s.quantity,
			Reason:       "Bullish Orderflow Divergence (CVD increasing while price dropping)",
			StrategyName: s.Name(),
		}}, nil
	}

	if divergence == analytics.BearishDivergence && s.lastAction == "buy" {
		s.lastAction = "sell"
		return []strategy.Signal{{
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "sell",
			Quantity:     s.quantity,
			Reason:       "Bearish Orderflow Divergence (CVD decreasing while price rising)",
			StrategyName: s.Name(),
		}}, nil
	}

	return nil, nil
}
