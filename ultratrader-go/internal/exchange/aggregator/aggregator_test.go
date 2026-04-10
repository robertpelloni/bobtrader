package aggregator

import (
	"context"
	"testing"
)

// mockProvider implements PriceProvider for testing.
type mockProvider struct {
	name  string
	price string
	err   error
}

func (m *mockProvider) Name() string { return m.name }
func (m *mockProvider) GetTickerPrice(_ context.Context, _ string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.price, nil
}

func TestPriceAggregator_Median(t *testing.T) {
	pa := NewPriceAggregator()
	pa.Register(&mockProvider{name: "ex1", price: "100"})
	pa.Register(&mockProvider{name: "ex2", price: "102"})
	pa.Register(&mockProvider{name: "ex3", price: "104"})

	price, err := pa.GetPrice(context.Background(), "BTCUSDT", MethodMedian)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if price != 102 { // Median of 100, 102, 104
		t.Errorf("expected median 102, got %f", price)
	}
}

func TestPriceAggregator_Mean(t *testing.T) {
	pa := NewPriceAggregator()
	pa.Register(&mockProvider{name: "ex1", price: "100"})
	pa.Register(&mockProvider{name: "ex2", price: "200"})

	price, err := pa.GetPrice(context.Background(), "BTCUSDT", MethodMean)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if price != 150 {
		t.Errorf("expected mean 150, got %f", price)
	}
}

func TestPriceAggregator_NoExchanges(t *testing.T) {
	pa := NewPriceAggregator()
	price, err := pa.GetPrice(context.Background(), "BTCUSDT", MethodMedian)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if price != 0 {
		t.Errorf("expected 0 for no exchanges, got %f", price)
	}
}

func TestPriceAggregator_HealthTracking(t *testing.T) {
	pa := NewPriceAggregator()
	pa.Register(&mockProvider{name: "good", price: "100"})
	pa.Register(&mockProvider{name: "bad", err: context.Canceled})

	pa.GetPrice(context.Background(), "BTCUSDT", MethodMedian)

	health := pa.HealthStatus()
	if !health["good"] {
		t.Error("expected good exchange to be healthy")
	}
	if health["bad"] {
		t.Error("expected bad exchange to be unhealthy")
	}
}

func TestPriceAggregator_ArbitrageDetection(t *testing.T) {
	pa := NewPriceAggregator()
	pa.Register(&mockProvider{name: "ex1", price: "100"})
	pa.Register(&mockProvider{name: "ex2", price: "105"})

	opportunities := pa.DetectArbitrage(context.Background(), "BTCUSDT", 0.01)
	if len(opportunities) != 1 {
		t.Fatalf("expected 1 opportunity, got %d", len(opportunities))
	}
	if opportunities[0].Spread < 0.04 {
		t.Errorf("expected spread > 4%%, got %f%%", opportunities[0].Spread*100)
	}
	if opportunities[0].BuyExchange != "ex1" {
		t.Errorf("expected buy on ex1 (cheaper), got %s", opportunities[0].BuyExchange)
	}
}

func TestPriceAggregator_NoArbitrage(t *testing.T) {
	pa := NewPriceAggregator()
	pa.Register(&mockProvider{name: "ex1", price: "100"})
	pa.Register(&mockProvider{name: "ex2", price: "100.5"})

	opportunities := pa.DetectArbitrage(context.Background(), "BTCUSDT", 0.10) // 10% minimum
	if len(opportunities) != 0 {
		t.Errorf("expected no opportunities with high threshold, got %d", len(opportunities))
	}
}

func TestPriceAggregator_GetAllQuotes(t *testing.T) {
	pa := NewPriceAggregator()
	pa.Register(&mockProvider{name: "ex1", price: "100"})
	pa.Register(&mockProvider{name: "ex2", price: "200"})

	quotes := pa.GetAllQuotes(context.Background(), "BTCUSDT")
	if len(quotes) != 2 {
		t.Errorf("expected 2 quotes, got %d", len(quotes))
	}
}

func TestPriceAggregator_SingleExchange(t *testing.T) {
	pa := NewPriceAggregator()
	pa.Register(&mockProvider{name: "only", price: "50000"})

	price, err := pa.GetPrice(context.Background(), "BTCUSDT", MethodMedian)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if price != 50000 {
		t.Errorf("expected 50000, got %f", price)
	}
}
