package ml_test

import (
	"context"
	"errors"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/analytics/ml"
)

type mockModel struct {
	name     string
	predH    float64
	predL    float64
	conf     float64
	err      error
	accuracy float64
}

func (m *mockModel) Name() string { return m.name }
func (m *mockModel) Predict(ctx context.Context, f ml.FeatureVector) (float64, float64, float64, error) {
	if m.err != nil {
		return 0, 0, 0, m.err
	}
	return m.predH, m.predL, m.conf, nil
}
func (m *mockModel) RollingAccuracy() float64    { return m.accuracy }
func (m *mockModel) UpdateAccuracy(h, l float64) {}

func TestEnsemblePredictor(t *testing.T) {
	ensemble := ml.NewEnsemblePredictor()

	m1 := &mockModel{name: "M1", predH: 0.10, predL: -0.05, conf: 0.8, accuracy: 0.9}
	m2 := &mockModel{name: "M2", predH: 0.05, predL: -0.02, conf: 0.6, accuracy: 0.1} // Low weight
	m3 := &mockModel{name: "M3", err: errors.New("failed")}                           // Should be skipped

	ensemble.AddModel(m1)
	ensemble.AddModel(m2)
	ensemble.AddModel(m3)

	res, err := ensemble.Predict(context.Background(), ml.FeatureVector{1, 2, 3})
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Calculate expected weighted average for high
	// (0.10 * 0.9) + (0.05 * 0.1) = 0.09 + 0.005 = 0.095
	// Total weight = 0.9 + 0.1 = 1.0
	// 0.095 / 1.0 = 0.095

	if res.PredictedHighPct < 0.094 || res.PredictedHighPct > 0.096 {
		t.Errorf("Expected ~0.095, got %f", res.PredictedHighPct)
	}

	if len(res.ModelVotes) != 2 {
		t.Errorf("Expected 2 votes recorded, got %d", len(res.ModelVotes))
	}
}

func TestRegistry_Evaluate(t *testing.T) {
	registry := ml.NewRegistry()

	champ := &mockModel{name: "Champion", accuracy: 0.50}
	chall := &mockModel{name: "Challenger", accuracy: 0.54} // 0.54 / 0.50 = 1.08 (8% improvement)

	registry.SetChampion(champ)
	registry.SetChallenger(chall)

	promoted := registry.Evaluate()

	if !promoted {
		t.Errorf("Expected challenger to be promoted (8%% > 5%%)")
	}

	if registry.Champion().Name() != "Challenger" {
		t.Errorf("Expected new champion to be Challenger, got %s", registry.Champion().Name())
	}

	if registry.Challenger() != nil {
		t.Errorf("Expected challenger slot to be cleared after promotion")
	}
}

func TestRegistry_Evaluate_NoPromotion(t *testing.T) {
	registry := ml.NewRegistry()

	champ := &mockModel{name: "Champion", accuracy: 0.50}
	chall := &mockModel{name: "Challenger", accuracy: 0.51} // Only 2% improvement

	registry.SetChampion(champ)
	registry.SetChallenger(chall)

	promoted := registry.Evaluate()

	if promoted {
		t.Errorf("Did not expect promotion (2%% < 5%%)")
	}

	if registry.Champion().Name() != "Champion" {
		t.Errorf("Champion should not have changed")
	}
}
