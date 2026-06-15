package demo

import (
	"context"
	"fmt"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/indicator"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/composite"
)

// MacroRegimeStrategy provides high-level trend context (Bullish/Bearish/Neutral).
// It implements composite.SignalEvaluator to be used as a filter or voting member.
type MacroRegimeStrategy struct {
	mu           sync.RWMutex
	symbol       string
	ema          *indicator.EMA
	adx          *indicator.ADX
	warmup       int
	warmupPeriod int
	lastRegime   composite.Signal
}

func NewMacroRegimeStrategy(symbol string, emaPeriod, adxPeriod int) *MacroRegimeStrategy {
	return &MacroRegimeStrategy{
		symbol:       symbol,
		ema:          indicator.NewEMA(emaPeriod),
		adx:          indicator.NewADX(adxPeriod),
		warmupPeriod: emaPeriod,
	}
}

func (s *MacroRegimeStrategy) Name() string {
	return fmt.Sprintf("macro-regime-%s", s.symbol)
}

func (s *MacroRegimeStrategy) Evaluate(ctx context.Context) (composite.SignalResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.warmup < s.warmupPeriod {
		return composite.SignalResult{
			Signal:     composite.SignalNone,
			Confidence: composite.ConfidenceLow,
			Source:     s.Name(),
			Reason:     "warming up",
		}, nil
	}

	return composite.SignalResult{
		Signal:     s.lastRegime,
		Confidence: composite.ConfidenceHigh,
		Source:     s.Name(),
		Reason:     fmt.Sprintf("macro regime: %s", s.lastRegime),
	}, nil
}

// OnMarketCandle updates the internal indicators.
func (s *MacroRegimeStrategy) OnMarketCandle(ctx context.Context, candle marketdata.Candle) ([]strategy.Signal, error) {
	if candle.Symbol != s.symbol {
		return nil, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	price := utils.ParseFloat(candle.Close)
	high := utils.ParseFloat(candle.High)
	low := utils.ParseFloat(candle.Low)

	emaVal := s.ema.Update(price)
	adxVal := s.adx.Update(high, low, price)

	s.warmup++

	// Determine regime
	// Trend is strong if ADX > 25
	if adxVal > 25 {
		if price > emaVal {
			s.lastRegime = composite.SignalBuy // Macro Bullish
		} else {
			s.lastRegime = composite.SignalSell // Macro Bearish
		}
	} else {
		s.lastRegime = composite.SignalNone // Ranging/Neutral
	}

	return nil, nil
}

// Ensure it implements the expected interfaces
var _ composite.SignalEvaluator = (*MacroRegimeStrategy)(nil)
var _ strategy.CandleStrategy = (*MacroRegimeStrategy)(nil)

func (s *MacroRegimeStrategy) OnTick(ctx context.Context) ([]strategy.Signal, error) {
	return nil, nil
}

func (s *MacroRegimeStrategy) OnMarketTick(ctx context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	return nil, nil
}
