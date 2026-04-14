package sentiment

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/logging"
)

// Signal represents an individual sentiment score.
// Score is typically -1.0 (extreme fear/bearish) to 1.0 (extreme greed/bullish).
type Signal struct {
	Source    string
	Symbol    string
	Score     float64
	Timestamp time.Time
}

// Provider defines the interface for an external sentiment data source.
type Provider interface {
	Name() string
	FetchSentiment(ctx context.Context, symbol string) (Signal, error)
}

// Engine aggregates sentiment from multiple providers.
type Engine struct {
	providers []Provider
	logger    *logging.Logger
	mu        sync.RWMutex
}

// NewEngine creates a new sentiment aggregation engine.
func NewEngine(logger *logging.Logger) *Engine {
	if logger == nil {
		logger, _ = logging.New(logging.Config{Stdout: true})
	}
	return &Engine{
		logger: logger,
	}
}

// RegisterProvider adds a new sentiment provider to the engine.
func (e *Engine) RegisterProvider(p Provider) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.providers = append(e.providers, p)
}

// AggregateSentiment fetches and averages sentiment from all registered providers for a given symbol.
// Returns a combined score [-1.0, 1.0] and a map of individual signals.
func (e *Engine) AggregateSentiment(ctx context.Context, symbol string) (float64, map[string]Signal, error) {
	e.mu.RLock()
	providers := make([]Provider, len(e.providers))
	copy(providers, e.providers)
	e.mu.RUnlock()

	if len(providers) == 0 {
		return 0, nil, fmt.Errorf("no sentiment providers registered")
	}

	signals := make(map[string]Signal)
	var totalScore float64
	var validCount int

	// Optional: Could be done concurrently using errgroup if we have slow external APIs.
	for _, p := range providers {
		sig, err := p.FetchSentiment(ctx, symbol)
		if err != nil {
			e.logger.WithContext(ctx).Info("Failed to fetch sentiment", map[string]any{
				"provider": p.Name(),
				"symbol":   symbol,
				"error":    err.Error(),
			})
			continue
		}

		// Clamp score to [-1.0, 1.0] for safety
		if sig.Score < -1.0 {
			sig.Score = -1.0
		} else if sig.Score > 1.0 {
			sig.Score = 1.0
		}

		signals[p.Name()] = sig
		totalScore += sig.Score
		validCount++
	}

	if validCount == 0 {
		return 0, signals, fmt.Errorf("could not fetch sentiment from any provider")
	}

	averageScore := totalScore / float64(validCount)
	return averageScore, signals, nil
}
