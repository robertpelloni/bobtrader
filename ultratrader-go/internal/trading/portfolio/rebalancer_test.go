package portfolio_test

import (
	"context"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/portfolio"
)

// mockFeed implements marketdata.Feed for testing
type mockFeed struct {
	ticks map[string]marketdata.Tick
}

func (m *mockFeed) LatestTick(ctx context.Context, symbol string) (marketdata.Tick, error) {
	if t, ok := m.ticks[symbol]; ok {
		return t, nil
	}
	return marketdata.Tick{}, nil
}

func (m *mockFeed) Subscribe(ctx context.Context, symbol string) (<-chan marketdata.Tick, error) {
	return nil, nil
}
func (m *mockFeed) Unsubscribe(ctx context.Context, symbol string) error { return nil }
func (m *mockFeed) LatestCandle(ctx context.Context, symbol, interval string) (marketdata.Candle, error) {
	return marketdata.Candle{}, nil
}

func TestRebalancer_CheckDrift(t *testing.T) {
	tracker := portfolio.NewTracker()
	// Total Value:
	// BTC: 1 * 65000 = $65000
	// ETH: 10 * 3500 = $35000
	// Total: $100000
	tracker.Apply(exchange.Order{Symbol: "BTC", Side: exchange.Buy, Quantity: "1", Price: "65000"})
	tracker.Apply(exchange.Order{Symbol: "ETH", Side: exchange.Buy, Quantity: "10", Price: "3500"})

	feed := &mockFeed{
		ticks: map[string]marketdata.Tick{
			"BTC": {Symbol: "BTC", Price: "65000"},
			"ETH": {Symbol: "ETH", Price: "3500"},
			"SOL": {Symbol: "SOL", Price: "150"},
		},
	}

	config := portfolio.RebalanceConfig{
		Enabled: true,
		TargetAllocations: map[string]float64{
			"BTC": 50.0, // Should be $50,000 (currently $65,000) -> Sell $15,000
			"ETH": 30.0, // Should be $30,000 (currently $35,000) -> Sell $5,000
			"SOL": 20.0, // Should be $20,000 (currently $0)      -> Buy $20,000
		},
		DriftThresholdPct: 4.0, // ETH is 35% vs 30% -> drift 5% -> exceeds threshold
	}

	rebalancer := portfolio.NewRebalancer(config, nil)
	ctx := context.Background()

	adjustments := rebalancer.CheckDrift(ctx, tracker, feed)

	if len(adjustments) != 3 {
		t.Fatalf("Expected 3 adjustments, got %d", len(adjustments))
	}

	for _, adj := range adjustments {
		if adj.Symbol == "BTC" {
			if adj.Action != exchange.Sell {
				t.Errorf("Expected BTC to be SELL, got %s", adj.Action)
			}
			if adj.DiffUSD != 15000 {
				t.Errorf("Expected BTC diff to be 15000, got %f", adj.DiffUSD)
			}
		} else if adj.Symbol == "ETH" {
			if adj.Action != exchange.Sell {
				t.Errorf("Expected ETH to be SELL, got %s", adj.Action)
			}
		} else if adj.Symbol == "SOL" {
			if adj.Action != exchange.Buy {
				t.Errorf("Expected SOL to be BUY, got %s", adj.Action)
			}
			if adj.DiffUSD != 20000 {
				t.Errorf("Expected SOL diff to be 20000, got %f", adj.DiffUSD)
			}
		}
	}
}

func TestRebalancer_WashSale(t *testing.T) {
	tracker := portfolio.NewTracker()
	// BTC Avg Cost: $70000
	tracker.Apply(exchange.Order{Symbol: "BTC", Side: exchange.Buy, Quantity: "1", Price: "70000"})

	feed := &mockFeed{
		ticks: map[string]marketdata.Tick{
			"BTC": {Symbol: "BTC", Price: "60000"}, // Current price < Avg Cost
		},
	}

	config := portfolio.RebalanceConfig{
		Enabled: true,
		TargetAllocations: map[string]float64{
			"BTC": 50.0, // Will want to sell because it's the only asset
		},
		DriftThresholdPct: 1.0,
		AvoidWashSales:    true,
	}

	recentOrders := []portfolio.OrderHistoryItem{
		{
			Symbol:    "BTC",
			Side:      exchange.Buy,
			Timestamp: time.Now().Add(-10 * 24 * time.Hour), // Bought 10 days ago
		},
	}

	rebalancer := portfolio.NewRebalancer(config, nil)
	ctx := context.Background()

	orders := rebalancer.GenerateOrders(ctx, tracker, feed, recentOrders)

	if len(orders) != 0 {
		t.Errorf("Expected wash sale to be prevented, but got %d orders", len(orders))
	}

	// Test without wash sale prevention
	config.AvoidWashSales = false
	rebalancer2 := portfolio.NewRebalancer(config, nil)
	orders2 := rebalancer2.GenerateOrders(ctx, tracker, feed, recentOrders)

	if len(orders2) == 0 {
		t.Errorf("Expected order to be generated when wash sale prevention is off")
	}
}
