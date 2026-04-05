package strategy

import (
	"context"
	"testing"
)

type testStrategy struct{}

func (testStrategy) Name() string { return "test" }
func (testStrategy) OnTick(_ context.Context) ([]Signal, error) {
	return []Signal{{AccountID: "paper-main", Symbol: "BTCUSDT", Action: "buy", Reason: "test"}}, nil
}

func TestRuntimeTickAggregatesSignals(t *testing.T) {
	runtime := NewRuntime(testStrategy{})
	signals, err := runtime.Tick(context.Background())
	if err != nil {
		t.Fatalf("Tick returned error: %v", err)
	}
	if len(signals) != 1 {
		t.Fatalf("expected 1 signal, got %d", len(signals))
	}
}
