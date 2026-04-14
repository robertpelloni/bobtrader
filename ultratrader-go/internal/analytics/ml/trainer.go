package ml

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/logging"
)

// TrainingSample represents a single historical observation.
type TrainingSample struct {
	Features    FeatureVector `json:"features"`
	OutcomeHigh float64       `json:"outcome_high"` // The max % move upward within a lookahead window
	OutcomeLow  float64       `json:"outcome_low"`  // The max % move downward within a lookahead window
}

// Dataset represents a collection of historical training samples.
type Dataset struct {
	Samples []TrainingSample `json:"samples"`
}

// KNNModel implements the Model interface using a simple k-Nearest-Neighbors algorithm.
type KNNModel struct {
	name      string
	dataset   Dataset
	k         int
	accuracy  float64
	totalPred int
	mu        sync.RWMutex
}

// NewKNNModel creates a new kNN model.
func NewKNNModel(name string, k int) *KNNModel {
	if k <= 0 {
		k = 5
	}
	return &KNNModel{
		name:     name,
		k:        k,
		accuracy: 0.5, // start neutral
	}
}

// Name returns the model name.
func (m *KNNModel) Name() string { return m.name }

// RollingAccuracy returns the accuracy weight.
func (m *KNNModel) RollingAccuracy() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.accuracy
}

// UpdateAccuracy gently updates the running accuracy of the model.
func (m *KNNModel) UpdateAccuracy(actualHigh, actualLow float64) {
	// A real implementation would compare the actuals to its last prediction.
	// For now, we stub this out as a placeholder.
	m.mu.Lock()
	m.totalPred++
	m.mu.Unlock()
}

// Predict performs a kNN search across the historical dataset to find similar patterns.
func (m *KNNModel) Predict(ctx context.Context, features FeatureVector) (float64, float64, float64, error) {
	m.mu.RLock()
	data := m.dataset.Samples
	m.mu.RUnlock()

	if len(data) == 0 {
		return 0, 0, 0, fmt.Errorf("model is untrained")
	}

	type Distance struct {
		Dist float64
		High float64
		Low  float64
	}

	var distances []Distance
	for _, sample := range data {
		if len(sample.Features) != len(features) {
			continue // skip mismatched feature dimensions
		}

		var sumSq float64
		for i, f := range features {
			diff := f - sample.Features[i]
			sumSq += diff * diff
		}

		dist := math.Sqrt(sumSq)
		distances = append(distances, Distance{
			Dist: dist,
			High: sample.OutcomeHigh,
			Low:  sample.OutcomeLow,
		})
	}

	if len(distances) == 0 {
		return 0, 0, 0, fmt.Errorf("no comparable samples found")
	}

	// In a production Go ML system, we would sort the slice to find the k-nearest.
	// A naive O(N^2) or sorting approach:
	for i := 0; i < len(distances); i++ {
		for j := i + 1; j < len(distances); j++ {
			if distances[i].Dist > distances[j].Dist {
				distances[i], distances[j] = distances[j], distances[i]
			}
		}
	}

	neighbors := m.k
	if len(distances) < m.k {
		neighbors = len(distances)
	}

	var sumH, sumL, weightSum float64
	for i := 0; i < neighbors; i++ {
		d := distances[i]
		// inverse distance weighting
		w := 1.0 / (d.Dist + 1e-6)
		sumH += d.High * w
		sumL += d.Low * w
		weightSum += w
	}

	predH := sumH / weightSum
	predL := sumL / weightSum
	confidence := 1.0 - (distances[0].Dist / (distances[0].Dist + 1.0)) // heuristic confidence

	return predH, predL, confidence, nil
}

// Trainer manages the offline training, testing, and persistence of ML models.
type Trainer struct {
	logger  *logging.Logger
	dataDir string
}

// NewTrainer creates a new model trainer instance.
func NewTrainer(dataDir string, logger *logging.Logger) *Trainer {
	if logger == nil {
		logger, _ = logging.New(logging.Config{Stdout: true})
	}
	return &Trainer{
		logger:  logger,
		dataDir: dataDir,
	}
}

// Train loads a dataset into a model and persists the artifact to disk.
func (t *Trainer) Train(ctx context.Context, model *KNNModel, data Dataset) error {
	t.logger.WithContext(ctx).Info("Training model", map[string]any{
		"model_name":   model.Name(),
		"samples_size": len(data.Samples),
	})

	model.mu.Lock()
	model.dataset = data
	// Evaluate naive accuracy
	model.accuracy = 0.85 // mock training score
	model.mu.Unlock()

	return t.SaveModel(model)
}

// SaveModel serializes the model dataset and metadata to JSON.
func (t *Trainer) SaveModel(model *KNNModel) error {
	if err := os.MkdirAll(t.dataDir, 0755); err != nil {
		return fmt.Errorf("failed to create model dir: %w", err)
	}

	filename := fmt.Sprintf("%s_%d.json", model.Name(), time.Now().Unix())
	path := filepath.Join(t.dataDir, filename)

	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create model file: %w", err)
	}
	defer file.Close()

	model.mu.RLock()
	defer model.mu.RUnlock()

	enc := json.NewEncoder(file)
	if err := enc.Encode(model.dataset); err != nil {
		return fmt.Errorf("failed to serialize model data: %w", err)
	}

	t.logger.Info("Model saved", map[string]any{"path": path, "name": model.Name()})
	return nil
}
