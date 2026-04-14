package sentiment_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/analytics/sentiment"
)

type mockProvider struct {
	name   string
	signal sentiment.Signal
	err    error
}

func (m *mockProvider) Name() string { return m.name }
func (m *mockProvider) FetchSentiment(ctx context.Context, symbol string) (sentiment.Signal, error) {
	if m.err != nil {
		return sentiment.Signal{}, m.err
	}
	return m.signal, nil
}

func TestEngine_AggregateSentiment(t *testing.T) {
	engine := sentiment.NewEngine(nil)

	p1 := &mockProvider{
		name: "Twitter",
		signal: sentiment.Signal{
			Source:    "Twitter",
			Symbol:    "BTC",
			Score:     0.8,
			Timestamp: time.Now(),
		},
	}
	p2 := &mockProvider{
		name: "News",
		signal: sentiment.Signal{
			Source:    "News",
			Symbol:    "BTC",
			Score:     0.4,
			Timestamp: time.Now(),
		},
	}
	p3 := &mockProvider{
		name: "Reddit",
		err:  errors.New("api rate limit"),
	}

	engine.RegisterProvider(p1)
	engine.RegisterProvider(p2)
	engine.RegisterProvider(p3)

	ctx := context.Background()
	score, signals, err := engine.AggregateSentiment(ctx, "BTC")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Should be (0.8 + 0.4) / 2 = 0.6
	if score < 0.599 || score > 0.601 {
		t.Errorf("Expected score 0.6, got %f", score)
	}

	if len(signals) != 2 {
		t.Errorf("Expected 2 signals, got %d", len(signals))
	}

	if _, ok := signals["Reddit"]; ok {
		t.Errorf("Did not expect Reddit signal to be present")
	}
}

func TestEngine_AggregateSentiment_Clamping(t *testing.T) {
	engine := sentiment.NewEngine(nil)

	p1 := &mockProvider{
		name: "Extreme",
		signal: sentiment.Signal{
			Source:    "Extreme",
			Symbol:    "BTC",
			Score:     5.0, // Should clamp to 1.0
			Timestamp: time.Now(),
		},
	}

	engine.RegisterProvider(p1)

	ctx := context.Background()
	score, _, err := engine.AggregateSentiment(ctx, "BTC")

	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if score != 1.0 {
		t.Errorf("Expected score 1.0 (clamped), got %f", score)
	}
}
