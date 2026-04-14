package ml

import (
	"context"
	"fmt"
	"math"
	"sync"
)

// FeatureVector represents an extracted set of numerical features for a market snapshot.
type FeatureVector []float64

// PredictionResult contains the aggregated output from an ML model or ensemble.
type PredictionResult struct {
	PredictedHighPct float64            // Expected upward move percentage
	PredictedLowPct  float64            // Expected downward move percentage
	Confidence       float64            // 0.0 to 1.0 confidence score
	ModelVotes       map[string]float64 // Individual model contributions to the prediction
}

// Model represents a single predictive algorithm (e.g., a kNN or regression model).
type Model interface {
	Name() string
	Predict(ctx context.Context, features FeatureVector) (float64, float64, float64, error) // Returns HighPct, LowPct, Confidence, Error
	RollingAccuracy() float64                                                               // Weighting factor for ensemble
	UpdateAccuracy(actualHigh, actualLow float64)                                           // Provide feedback on past predictions
}

// EnsemblePredictor combines multiple Model instances using adaptive weighting.
// Models with higher rolling accuracy are given more influence on the final prediction.
type EnsemblePredictor struct {
	models []Model
	mu     sync.RWMutex
}

// NewEnsemblePredictor creates an empty ensemble.
func NewEnsemblePredictor() *EnsemblePredictor {
	return &EnsemblePredictor{}
}

// AddModel registers a new Model into the ensemble.
func (e *EnsemblePredictor) AddModel(m Model) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.models = append(e.models, m)
}

// Predict calculates the weighted-average prediction across all registered models.
func (e *EnsemblePredictor) Predict(ctx context.Context, features FeatureVector) (PredictionResult, error) {
	e.mu.RLock()
	models := make([]Model, len(e.models))
	copy(models, e.models)
	e.mu.RUnlock()

	if len(models) == 0 {
		return PredictionResult{}, fmt.Errorf("no models in ensemble")
	}

	var totalWeight float64
	var weightedHigh float64
	var weightedLow float64
	var weightedConf float64

	votes := make(map[string]float64)
	validModels := 0

	for _, m := range models {
		predH, predL, conf, err := m.Predict(ctx, features)
		if err != nil || math.IsNaN(predH) || math.IsNaN(predL) {
			continue
		}

		weight := m.RollingAccuracy()
		if weight <= 0 {
			weight = 0.01 // Minimum weight to prevent 0 division if all models are terrible
		}

		weightedHigh += predH * weight
		weightedLow += predL * weight
		weightedConf += conf * weight
		totalWeight += weight

		votes[m.Name()] = predH // Store the high prediction for debugging/transparency
		validModels++
	}

	if validModels == 0 || totalWeight == 0 {
		return PredictionResult{}, fmt.Errorf("all models failed to produce a valid prediction")
	}

	return PredictionResult{
		PredictedHighPct: weightedHigh / totalWeight,
		PredictedLowPct:  weightedLow / totalWeight,
		Confidence:       weightedConf / totalWeight,
		ModelVotes:       votes,
	}, nil
}

// Registry manages model versioning and champion/challenger selection.
type Registry struct {
	champion   Model
	challenger Model
	mu         sync.RWMutex
}

// NewRegistry initializes an empty model registry.
func NewRegistry() *Registry {
	return &Registry{}
}

// SetChampion sets the primary active model.
func (r *Registry) SetChampion(m Model) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.champion = m
}

// SetChallenger sets the experimental model running in parallel.
func (r *Registry) SetChallenger(m Model) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.challenger = m
}

// Evaluate A/B tests the challenger against the champion based on rolling accuracy.
// If the challenger significantly outperforms the champion, it replaces it.
// Returns true if a promotion occurred.
func (r *Registry) Evaluate() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.champion == nil || r.challenger == nil {
		return false
	}

	champScore := r.champion.RollingAccuracy()
	challScore := r.challenger.RollingAccuracy()

	// 5% improvement threshold to prevent noisy thrashing
	if challScore > champScore*1.05 {
		r.champion = r.challenger
		r.challenger = nil
		return true
	}

	return false
}

// Champion returns the currently active primary model.
func (r *Registry) Champion() Model {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.champion
}

// Challenger returns the currently active experimental model.
func (r *Registry) Challenger() Model {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.challenger
}
