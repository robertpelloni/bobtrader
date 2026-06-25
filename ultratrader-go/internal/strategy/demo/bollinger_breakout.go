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

// BollingerBreakoutStrategy looks for "Bollinger Squeezes" (low bandwidth)
// and enters when price breaks out of the bands with strong volume.
type BollingerBreakoutStrategy struct {
	mu           sync.Mutex
	accountID    string
	symbol       string
	quantity     string
	bb           *indicator.BollingerBands
	smaVol       *indicator.SMA
	bwThreshold  float64 // Maximum bandwidth to consider it a "squeeze"
	lastAction   string
	warmup       int
}

func NewBollingerBreakout(accountID, symbol, quantity string, period int, stdDev float64, bwThreshold float64) *BollingerBreakoutStrategy {
	return &BollingerBreakoutStrategy{
		accountID:   accountID,
		symbol:      symbol,
		quantity:    quantity,
		bb:          indicator.NewBollingerBands(period, stdDev),
		smaVol:      indicator.NewSMA(period),
		bwThreshold: bwThreshold,
	}
}

func (s *BollingerBreakoutStrategy) Name() string { return fmt.Sprintf("bb-breakout-%s", s.symbol) }

func (s *BollingerBreakoutStrategy) OnTick(_ context.Context) ([]strategy.Signal, error) { return nil, nil }

func (s *BollingerBreakoutStrategy) OnMarketCandle(ctx context.Context, candle marketdata.Candle) ([]strategy.Signal, error) {
	if candle.Symbol != s.symbol {
		return nil, nil
	}

	price := utils.ParseFloat(candle.Close)
	volume := utils.ParseFloat(candle.Volume)
	s.mu.Lock()
	defer s.mu.Unlock()

	res := s.bb.Update(price)
	avgVol := s.smaVol.Update(volume)
	s.warmup++

	if s.warmup < 20 { // Ensure indicators are stable
		return nil, nil
	}

	// Logic: If in a squeeze (low bandwidth) AND price breaks upper band AND volume is above average
	if res.Bandwidth < s.bwThreshold && price > res.Upper && volume > avgVol*1.5 && s.lastAction != "buy" {
		s.lastAction = "buy"
		return []strategy.Signal{{
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "buy",
			Quantity:     s.quantity,
			Reason:       fmt.Sprintf("Bollinger Breakout: BW %.3f < %.3f with vol surge", res.Bandwidth, s.bwThreshold),
			StrategyName: s.Name(),
		}}, nil
	}

	// Exit if price falls back below middle band (SMA)
	if price < res.Middle && s.lastAction == "buy" {
		s.lastAction = "sell"
		return []strategy.Signal{{
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "sell",
			Quantity:     s.quantity,
			Reason:       "Breakout Mean Reversion: price below middle band",
			StrategyName: s.Name(),
		}}, nil
	}

	return nil, nil
}
