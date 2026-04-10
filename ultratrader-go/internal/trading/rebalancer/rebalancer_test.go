package rebalancer

import (
	"math"
	"testing"
)

func TestRebalancer_Balanced(t *testing.T) {
	r := New([]Allocation{
		{Symbol: "BTC", Weight: 0.5},
		{Symbol: "ETH", Weight: 0.5},
	}, 0.05)

	result := r.Compute([]Holding{
		{Symbol: "BTC", Quantity: 1, Value: 50000},
		{Symbol: "ETH", Quantity: 10, Value: 50000},
	})

	if result.NeedsRebalance {
		t.Error("should not need rebalance when perfectly balanced")
	}
	if len(result.Orders) != 0 {
		t.Errorf("expected 0 orders, got %d", len(result.Orders))
	}
}

func TestRebalancer_Overweight(t *testing.T) {
	r := New([]Allocation{
		{Symbol: "BTC", Weight: 0.5},
		{Symbol: "ETH", Weight: 0.5},
	}, 0.05)

	result := r.Compute([]Holding{
		{Symbol: "BTC", Quantity: 1, Value: 70000},
		{Symbol: "ETH", Quantity: 10, Value: 30000},
	})

	if !result.NeedsRebalance {
		t.Error("should need rebalance when overweight")
	}
	if len(result.Orders) == 0 {
		t.Fatal("expected orders")
	}

	// BTC should be sold (overweight)
	// ETH should be bought (underweight)
	sells := 0
	buys := 0
	for _, o := range result.Orders {
		if o.Side == "sell" {
			sells++
		} else {
			buys++
		}
	}
	if sells != 1 || buys != 1 {
		t.Errorf("expected 1 sell + 1 buy, got %d sells + %d buys", sells, buys)
	}
}

func TestRebalancer_ThreeAssets(t *testing.T) {
	r := New([]Allocation{
		{Symbol: "BTC", Weight: 0.4},
		{Symbol: "ETH", Weight: 0.4},
		{Symbol: "DOGE", Weight: 0.2},
	}, 0.05)

	result := r.Compute([]Holding{
		{Symbol: "BTC", Quantity: 1, Value: 50000},
		{Symbol: "ETH", Quantity: 10, Value: 40000},
		{Symbol: "DOGE", Quantity: 10000, Value: 10000},
	})

	// BTC: 50% vs 40% target -> overweight 10%
	// ETH: 40% vs 40% -> balanced
	// DOGE: 10% vs 20% -> underweight 10%
	if !result.NeedsRebalance {
		t.Error("should need rebalance")
	}

	if result.MaxDrift < 0.09 {
		t.Errorf("expected max drift ~10%%, got %.1f%%", result.MaxDrift*100)
	}
}

func TestRebalancer_EmptyPortfolio(t *testing.T) {
	r := New([]Allocation{
		{Symbol: "BTC", Weight: 1.0},
	}, 0.05)

	result := r.Compute(nil)
	if result.NeedsRebalance {
		t.Error("empty portfolio should not trigger rebalance")
	}
}

func TestRebalancer_SingleAsset(t *testing.T) {
	r := New([]Allocation{
		{Symbol: "BTC", Weight: 1.0},
	}, 0.05)

	result := r.Compute([]Holding{
		{Symbol: "BTC", Quantity: 1, Value: 50000},
	})

	if result.NeedsRebalance {
		t.Error("single asset at 100% should be balanced")
	}
}

func TestRebalancer_Drift(t *testing.T) {
	r := New([]Allocation{
		{Symbol: "BTC", Weight: 0.5},
		{Symbol: "ETH", Weight: 0.5},
	}, 0.05)

	drifts := r.Drift([]Holding{
		{Symbol: "BTC", Quantity: 1, Value: 60000},
		{Symbol: "ETH", Quantity: 10, Value: 40000},
	})

	if math.Abs(drifts["BTC"]-0.1) > 0.001 {
		t.Errorf("BTC drift: expected 0.1, got %f", drifts["BTC"])
	}
	if math.Abs(drifts["ETH"]-(-0.1)) > 0.001 {
		t.Errorf("ETH drift: expected -0.1, got %f", drifts["ETH"])
	}
}

func TestRebalancer_DriftMissingAsset(t *testing.T) {
	r := New([]Allocation{
		{Symbol: "BTC", Weight: 0.5},
		{Symbol: "ETH", Weight: 0.5},
	}, 0.05)

	drifts := r.Drift([]Holding{
		{Symbol: "BTC", Quantity: 1, Value: 100000},
	})

	if math.Abs(drifts["ETH"]-(-0.5)) > 0.001 {
		t.Errorf("ETH (missing) drift: expected -0.5, got %f", drifts["ETH"])
	}
}

func TestRebalancer_Summary(t *testing.T) {
	r := New([]Allocation{
		{Symbol: "BTC", Weight: 0.5},
		{Symbol: "ETH", Weight: 0.5},
	}, 0.05)

	// Balanced portfolio
	result := r.Compute([]Holding{
		{Symbol: "BTC", Quantity: 1, Value: 50000},
		{Symbol: "ETH", Quantity: 10, Value: 50000},
	})
	if result.Summary() == "" {
		t.Error("expected non-empty summary")
	}

	// Unbalanced portfolio
	result = r.Compute([]Holding{
		{Symbol: "BTC", Quantity: 1, Value: 70000},
		{Symbol: "ETH", Quantity: 10, Value: 30000},
	})
	s := result.Summary()
	if s == "" {
		t.Error("expected non-empty summary")
	}
	if !result.NeedsRebalance {
		t.Error("unbalanced portfolio should need rebalance")
	}
}

func TestRebalancer_DefaultThreshold(t *testing.T) {
	r := New([]Allocation{
		{Symbol: "BTC", Weight: 1.0},
	}, 0) // Zero threshold should default to 5%

	if r.threshold != 0.05 {
		t.Errorf("expected default threshold 0.05, got %f", r.threshold)
	}
}
