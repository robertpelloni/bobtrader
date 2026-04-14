package ml_test

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/analytics/ml"
)

func TestKNNModel_Predict(t *testing.T) {
	model := ml.NewKNNModel("test_knn", 2)

	dataset := ml.Dataset{
		Samples: []ml.TrainingSample{
			{Features: ml.FeatureVector{1.0, 1.0}, OutcomeHigh: 0.05, OutcomeLow: -0.01},
			{Features: ml.FeatureVector{5.0, 5.0}, OutcomeHigh: 0.15, OutcomeLow: -0.10},
			{Features: ml.FeatureVector{10.0, 10.0}, OutcomeHigh: 0.30, OutcomeLow: -0.20},
		},
	}

	ctx := context.Background()

	// Before training should error
	_, _, _, err := model.Predict(ctx, ml.FeatureVector{2.0, 2.0})
	if err == nil {
		t.Errorf("Expected untrained model error")
	}

	trainer := ml.NewTrainer(os.TempDir(), nil)
	err = trainer.Train(ctx, model, dataset)
	if err != nil {
		t.Fatalf("Failed to train model: %v", err)
	}

	// Predict close to the first sample
	high, low, conf, err := model.Predict(ctx, ml.FeatureVector{1.5, 1.5})
	if err != nil {
		t.Fatalf("Prediction failed: %v", err)
	}

	if high < 0.05 || high > 0.15 {
		t.Errorf("Expected high pred between 0.05 and 0.15 (neighbors), got %f", high)
	}

	if low > -0.01 || low < -0.10 {
		t.Errorf("Expected low pred between -0.01 and -0.10, got %f", low)
	}

	if conf <= 0 {
		t.Errorf("Expected positive confidence")
	}
}

func TestTrainer_SaveModel(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "ml_models_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	trainer := ml.NewTrainer(tmpDir, nil)
	model := ml.NewKNNModel("save_test", 5)

	dataset := ml.Dataset{
		Samples: []ml.TrainingSample{
			{Features: ml.FeatureVector{1.0}, OutcomeHigh: 0.05},
		},
	}

	err = trainer.Train(context.Background(), model, dataset)
	if err != nil {
		t.Fatalf("Failed to train and save: %v", err)
	}

	files, err := ioutil.ReadDir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to read model dir: %v", err)
	}

	if len(files) != 1 {
		t.Errorf("Expected 1 model file saved, got %d", len(files))
	}
}
