package demo

import (
	"context"
	"fmt"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/composite"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/regime"
)

// MacroRegimeStrategy provides high-level market state context (Trending/Ranging/Volatile).
// It implements composite.SignalEvaluator to be used as a filter or voting member.
type MacroRegimeStrategy struct {
	mu           sync.RWMutex
	symbol       string
	detector     regime.Detector
	history      []regime.CandleData
	maxHistory   int
	lastRegime   regime.Regime
}

func NewMacroRegimeStrategy(symbol string) *MacroRegimeStrategy {
	// Use a composite detector for robust classification
	detector := regime.NewCompositeDetector(
		regime.NewTrendDetector(14, 25, 50),
		regime.NewVolatilityDetector(0.05, 0.01, 14),
		regime.NewBollingerBandwidthDetector(20, 2.0, 0.05, 0.15),
	)

	return &MacroRegimeStrategy{
		symbol:     symbol,
		detector:   detector,
		maxHistory: 100,
		history:    make([]regime.CandleData, 0, 100),
	}
}

func (s *MacroRegimeStrategy) Name() string {
	return fmt.Sprintf("macro-regime-%s", s.symbol)
}

func (s *MacroRegimeStrategy) Evaluate(ctx context.Context) (composite.SignalResult, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.history) < 20 {
		return composite.SignalResult{
			Signal:     composite.SignalNone,
			Confidence: composite.ConfidenceLow,
			Source:     s.Name(),
			Reason:     "warming up",
		}, nil
	}

	// Map regime.Regime to composite.Signal for filtering
	// (Simplified mapping: Trending = directional signal based on price action)
	sig := composite.SignalNone
	if s.lastRegime == regime.RegimeTrending {
		// Detect trend direction
		first := s.history[0].Close
		last := s.history[len(s.history)-1].Close
		if last > first {
			sig = composite.SignalBuy
		} else {
			sig = composite.SignalSell
		}
	}

	return composite.SignalResult{
		Signal:     sig,
		Confidence: composite.ConfidenceHigh,
		Source:     s.Name(),
		Reason:     fmt.Sprintf("market regime: %s", s.lastRegime),
	}, nil
}

// OnMarketCandle updates the internal indicators.
func (s *MacroRegimeStrategy) OnMarketCandle(ctx context.Context, candle marketdata.Candle) ([]strategy.Signal, error) {
	if candle.Symbol != s.symbol {
		return nil, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	data := regime.CandleData{
		Open:   utils.ParseFloat(candle.Open),
		High:   utils.ParseFloat(candle.High),
		Low:    utils.ParseFloat(candle.Low),
		Close:  utils.ParseFloat(candle.Close),
		Volume: utils.ParseFloat(candle.Volume),
	}

	s.history = append(s.history, data)
	if len(s.history) > s.maxHistory {
		s.history = s.history[1:]
	}

	if len(s.history) >= 20 {
		s.lastRegime = s.detector.Detect(s.history)
	}

	return nil, nil
}

func (s *MacroRegimeStrategy) CurrentRegime() regime.Regime {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.lastRegime
}

// Ensure it implements the expected interfaces
var _ composite.SignalEvaluator = (*MacroRegimeStrategy)(nil)
var _ strategy.CandleStrategy = (*MacroRegimeStrategy)(nil)

func (s *MacroRegimeStrategy) OnTick(ctx context.Context) ([]strategy.Signal, error) { return nil, nil }
func (s *MacroRegimeStrategy) OnMarketTick(ctx context.Context, tick marketdata.Tick) ([]strategy.Signal, error) {
	return nil, nil
}
