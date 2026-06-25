package demo

import (
	"context"
	"fmt"
	"sync"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/analytics/features"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/analytics/ml"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
)

// MLAwareStrategy uses an ML ensemble to filter or generate signals.
type MLAwareStrategy struct {
	mu            sync.Mutex
	accountID     string
	symbol        string
	quantity      string
	ensemble      *ml.EnsemblePredictor
	extractor     *features.Extractor
	minConfidence float64
	thresholdPct  float64 // Minimum predicted move to trigger a signal
	lastAction    string
}

func NewMLAwareStrategy(accountID, symbol, quantity string, ensemble *ml.EnsemblePredictor, minConf float64, threshold float64) *MLAwareStrategy {
	return &MLAwareStrategy{
		accountID:     accountID,
		symbol:        symbol,
		quantity:      quantity,
		ensemble:      ensemble,
		extractor:     features.NewExtractor(20, 20),
		minConfidence: minConf,
		thresholdPct:  threshold,
	}
}

func (s *MLAwareStrategy) Name() string { return fmt.Sprintf("ml-aware-%s", s.symbol) }

func (s *MLAwareStrategy) OnTick(_ context.Context) ([]strategy.Signal, error) { return nil, nil }

func (s *MLAwareStrategy) OnMarketCandle(ctx context.Context, candle marketdata.Candle) ([]strategy.Signal, error) {
	if candle.Symbol != s.symbol {
		return nil, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Update features
	fMap := s.extractor.Update(features.CandleData{
		Open:   utils.ParseFloat(candle.Open),
		High:   utils.ParseFloat(candle.High),
		Low:    utils.ParseFloat(candle.Low),
		Close:  utils.ParseFloat(candle.Close),
		Volume: utils.ParseFloat(candle.Volume),
	})

	// Convert FeatureMap to FeatureVector
	names := s.extractor.Names()
	vector := make(ml.FeatureVector, len(names))
	for i, name := range names {
		vector[i] = fMap[name]
	}

	// Predict
	prediction, err := s.ensemble.Predict(ctx, vector)
	if err != nil {
		return nil, nil // Silently skip if ML fails
	}

	if prediction.Confidence < s.minConfidence {
		return nil, nil
	}

	var signals []strategy.Signal

	// Logic: If predicted high move > threshold and confidence is high, buy.
	if prediction.PredictedHighPct >= s.thresholdPct && s.lastAction != "buy" {
		s.lastAction = "buy"
		signals = append(signals, strategy.Signal{
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "buy",
			Quantity:     s.quantity,
			Reason:       fmt.Sprintf("ML Bullish: +%.2f%% expected (conf %.2f)", prediction.PredictedHighPct, prediction.Confidence),
			StrategyName: s.Name(),
		})
	} else if prediction.PredictedLowPct >= s.thresholdPct && s.lastAction == "buy" {
		s.lastAction = "sell"
		signals = append(signals, strategy.Signal{
			AccountID:    s.accountID,
			Symbol:       s.symbol,
			Action:       "sell",
			Quantity:     s.quantity,
			Reason:       fmt.Sprintf("ML Bearish: -%.2f%% expected (conf %.2f)", prediction.PredictedLowPct, prediction.Confidence),
			StrategyName: s.Name(),
		})
	}

	return signals, nil
}
