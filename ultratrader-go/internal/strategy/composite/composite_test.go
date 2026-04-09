package composite

import (
	"context"
	"errors"
	"testing"
)

// mockEvaluator for testing
type mockEvaluator struct {
	name    string
	signal  Signal
	conf    Confidence
	err     error
}

func (m *mockEvaluator) Name() string { return m.name }
func (m *mockEvaluator) Evaluate(_ context.Context) (SignalResult, error) {
	if m.err != nil {
		return SignalResult{}, m.err
	}
	return SignalResult{Signal: m.signal, Confidence: m.conf, Source: m.name}, nil
}

func TestSignal_String(t *testing.T) {
	if SignalBuy.String() != "BUY" {
		t.Errorf("expected BUY")
	}
	if SignalSell.String() != "SELL" {
		t.Errorf("expected SELL")
	}
	if SignalNone.String() != "NONE" {
		t.Errorf("expected NONE")
	}
}

func TestComposite_Unanimous_AllBuy(t *testing.T) {
	c := NewCompositeStrategy("test", Unanimous)
	c.AddStrategy(&mockEvaluator{name: "s1", signal: SignalBuy, conf: ConfidenceHigh}, 1)
	c.AddStrategy(&mockEvaluator{name: "s2", signal: SignalBuy, conf: ConfidenceMedium}, 1)
	c.AddStrategy(&mockEvaluator{name: "s3", signal: SignalBuy, conf: ConfidenceAbsolute}, 1)

	result, err := c.Evaluate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Signal != SignalBuy {
		t.Errorf("expected BUY, got %s", result.Signal)
	}
}

func TestComposite_Unanimous_Disagree(t *testing.T) {
	c := NewCompositeStrategy("test", Unanimous)
	c.AddStrategy(&mockEvaluator{name: "s1", signal: SignalBuy, conf: ConfidenceHigh}, 1)
	c.AddStrategy(&mockEvaluator{name: "s2", signal: SignalSell, conf: ConfidenceHigh}, 1)

	result, _ := c.Evaluate(context.Background())
	if result.Signal != SignalNone {
		t.Errorf("expected NONE for disagreement, got %s", result.Signal)
	}
}

func TestComposite_Majority(t *testing.T) {
	c := NewCompositeStrategy("test", Majority)
	c.AddStrategy(&mockEvaluator{name: "s1", signal: SignalBuy, conf: ConfidenceHigh}, 1)
	c.AddStrategy(&mockEvaluator{name: "s2", signal: SignalBuy, conf: ConfidenceMedium}, 1)
	c.AddStrategy(&mockEvaluator{name: "s3", signal: SignalSell, conf: ConfidenceHigh}, 1)

	result, _ := c.Evaluate(context.Background())
	if result.Signal != SignalBuy {
		t.Errorf("expected BUY (majority), got %s", result.Signal)
	}
}

func TestComposite_Majority_NoMajority(t *testing.T) {
	c := NewCompositeStrategy("test", Majority)
	c.AddStrategy(&mockEvaluator{name: "s1", signal: SignalBuy, conf: ConfidenceHigh}, 1)
	c.AddStrategy(&mockEvaluator{name: "s2", signal: SignalSell, conf: ConfidenceHigh}, 1)

	result, _ := c.Evaluate(context.Background())
	if result.Signal != SignalNone {
		t.Errorf("expected NONE (tied), got %s", result.Signal)
	}
}

func TestComposite_Any(t *testing.T) {
	c := NewCompositeStrategy("test", Any)
	c.AddStrategy(&mockEvaluator{name: "s1", signal: SignalNone, conf: ConfidenceLow}, 1)
	c.AddStrategy(&mockEvaluator{name: "s2", signal: SignalBuy, conf: ConfidenceHigh}, 1)
	c.AddStrategy(&mockEvaluator{name: "s3", signal: SignalNone, conf: ConfidenceLow}, 1)

	result, _ := c.Evaluate(context.Background())
	if result.Signal != SignalBuy {
		t.Errorf("expected BUY (any mode), got %s", result.Signal)
	}
}

func TestComposite_Any_AllNone(t *testing.T) {
	c := NewCompositeStrategy("test", Any)
	c.AddStrategy(&mockEvaluator{name: "s1", signal: SignalNone, conf: ConfidenceLow}, 1)
	c.AddStrategy(&mockEvaluator{name: "s2", signal: SignalNone, conf: ConfidenceLow}, 1)

	result, _ := c.Evaluate(context.Background())
	if result.Signal != SignalNone {
		t.Errorf("expected NONE, got %s", result.Signal)
	}
}

func TestComposite_Weighted(t *testing.T) {
	c := NewCompositeStrategy("test", Weighted)
	c.AddStrategy(&mockEvaluator{name: "s1", signal: SignalBuy, conf: ConfidenceHigh}, 0.5)
	c.AddStrategy(&mockEvaluator{name: "s2", signal: SignalSell, conf: ConfidenceAbsolute}, 1.0)

	// Sell has weight 1.0 vs Buy weight 0.5
	result, _ := c.Evaluate(context.Background())
	if result.Signal != SignalSell {
		t.Errorf("expected SELL (higher weight), got %s", result.Signal)
	}
}

func TestComposite_AllErrors(t *testing.T) {
	c := NewCompositeStrategy("test", Majority)
	c.AddStrategy(&mockEvaluator{name: "s1", err: errors.New("fail")}, 1)
	c.AddStrategy(&mockEvaluator{name: "s2", err: errors.New("fail")}, 1)

	_, err := c.Evaluate(context.Background())
	if err == nil {
		t.Error("expected error when all evaluators fail")
	}
}

func TestComposite_PartialErrors(t *testing.T) {
	c := NewCompositeStrategy("test", Majority)
	c.AddStrategy(&mockEvaluator{name: "s1", signal: SignalBuy, conf: ConfidenceHigh}, 1)
	c.AddStrategy(&mockEvaluator{name: "s2", err: errors.New("fail")}, 1)

	result, err := c.Evaluate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Signal != SignalBuy {
		t.Errorf("expected BUY from surviving evaluator, got %s", result.Signal)
	}
}

func TestComposite_Empty(t *testing.T) {
	c := NewCompositeStrategy("test", Majority)
	result, err := c.Evaluate(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Signal != SignalNone {
		t.Errorf("expected NONE for empty composite, got %s", result.Signal)
	}
}

func TestComposite_StrategyCount(t *testing.T) {
	c := NewCompositeStrategy("test", Majority)
	c.AddStrategy(&mockEvaluator{name: "s1"}, 1)
	c.AddStrategy(&mockEvaluator{name: "s2"}, 1)
	c.AddStrategy(&mockEvaluator{name: "s3"}, 1)

	if c.StrategyCount() != 3 {
		t.Errorf("expected 3 strategies, got %d", c.StrategyCount())
	}
}

func TestComposite_Name(t *testing.T) {
	c := NewCompositeStrategy("my-composite", Majority)
	if c.Name() != "my-composite" {
		t.Errorf("expected my-composite, got %s", c.Name())
	}
}

func TestComposite_Any_HighestConfidence(t *testing.T) {
	c := NewCompositeStrategy("test", Any)
	c.AddStrategy(&mockEvaluator{name: "s1", signal: SignalSell, conf: ConfidenceLow}, 1)
	c.AddStrategy(&mockEvaluator{name: "s2", signal: SignalBuy, conf: ConfidenceAbsolute}, 1)

	result, _ := c.Evaluate(context.Background())
	if result.Signal != SignalBuy {
		t.Errorf("expected BUY (higher confidence), got %s", result.Signal)
	}
	if result.Confidence != ConfidenceAbsolute {
		t.Errorf("expected absolute confidence, got %f", result.Confidence)
	}
}
