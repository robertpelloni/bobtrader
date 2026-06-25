package demo

import (
	"context"
	"fmt"
	"math"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// MicroScalper is a high-frequency tick-based strategy.
// It looks for sudden price movements (volatility spikes) and volume surges.
type MicroScalper struct {
	mu            sync.Mutex
	accountID     string
	symbol        string
	quantity      string
	lookbackTicks int
	prices        []float64
	volumes       []float64
	thresholdPct  float64
	lastAction    string
}

func NewMicroScalper(accountID, symbol, quantity string, lookback int, threshold float64) *MicroScalper {
	return &MicroScalper{
		accountID:     accountID,
		symbol:        symbol,
		quantity:      quantity,
		lookbackTicks: lookback,
		thresholdPct:  threshold,
	}
}

func (s *MicroScalper) Name() string {
	return fmt.Sprintf("micro-scalper-%s", s.symbol)
}

func (s *MicroScalper) OnTick(ctx context.Context) ([]strategy.Signal, error) {
	return nil, nil
}

func (s *MicroScalper) OnMarketTick(ctx context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	if tick.Symbol != s.symbol {
		return nil, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	price := utils.ParseFloat(tick.Price)
	volume := utils.ParseFloat(tick.Quantity)
	if price <= 0 {
		return nil, nil
	}

	s.prices = append(s.prices, price)
	s.volumes = append(s.volumes, volume)

	if len(s.prices) > s.lookbackTicks {
		s.prices = s.prices[1:]
		s.volumes = s.volumes[1:]
	}

	if len(s.prices) < s.lookbackTicks {
		return nil, nil
	}

	// Calculate movement
	first := s.prices[0]
	last := s.prices[len(s.prices)-1]
	changePct := ((last - first) / first) * 100

	// Calculate volume surge
	var avgVol float64
	for _, v := range s.volumes[:len(s.volumes)-1] {
		avgVol += v
	}
	avgVol /= float64(len(s.volumes) - 1)

	// Signal logic: volatility spike + volume > 2x average
	if math.Abs(changePct) >= s.thresholdPct && volume > avgVol*2 {
		action := "none"
		reason := ""

		if changePct > 0 && s.lastAction != "buy" {
			action = "buy"
			reason = fmt.Sprintf("micro bullish spike: %.3f%% with vol surge", changePct)
		} else if changePct < 0 && s.lastAction != "sell" {
			action = "sell"
			reason = fmt.Sprintf("micro bearish spike: %.3f%% with vol surge", changePct)
		}

		if action != "none" {
			s.lastAction = action
			return []strategy.Signal{{
				AccountID:    s.accountID,
				Symbol:       s.symbol,
				Action:       action,
				Quantity:     s.quantity,
				Reason:       reason,
				StrategyName: s.Name(),
			}}, nil
		}
	}

	return nil, nil
}
