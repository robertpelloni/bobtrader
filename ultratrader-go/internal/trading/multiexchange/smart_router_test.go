package multiexchange

import (
	"fmt"
	"testing"
)

type mockManager struct {
	prices map[string]map[string]Ticker
}

func (m *mockManager) GetTicker(coin, exchange string) (Ticker, error) {
	if exMap, ok := m.prices[coin]; ok {
		if ticker, ok := exMap[exchange]; ok {
			return ticker, nil
		}
	}
	return Ticker{}, fmt.Errorf("not found")
}

func (m *mockManager) GetExchanges() []string {
	return []string{"binance", "kucoin", "coinbase", "kraken"}
}

func TestSmartRouter_BestExecution(t *testing.T) {
	prices := map[string]map[string]Ticker{
		"BTC": {
			"binance":  {Bid: 49990, Ask: 50010},
			"kucoin":   {Bid: 49980, Ask: 50005},
			"coinbase": {Bid: 50000, Ask: 50020},
			"kraken":   {Bid: 49985, Ask: 50015},
		},
	}

	fees := map[string]float64{
		"binance":  0.1,  // 50010 + 50.01 = 50060.01
		"kucoin":   0.2,  // 50005 + 100.01 = 50105.01
		"coinbase": 0.05, // 50020 + 25.01 = 50045.01 (Winner for Buy)
		"kraken":   0.15, // 50015 + 75.02 = 50090.02
	}

	manager := &mockManager{prices: prices}
	router := NewSmartRouter(manager, fees)

	// Test Buy
	routes, err := router.CompareRoutes("BTC", "buy", 1.0)
	if err != nil {
		t.Fatalf("CompareRoutes failed: %v", err)
	}

	if routes[0].Exchange != "coinbase" {
		t.Errorf("expected coinbase to be best buy route, got %s (Price: %f, Total: %f)",
			routes[0].Exchange, routes[0].Price, routes[0].TotalCost)
	}

	// Test Sell
	// binance: 49990 - 49.99 = 49940.01
	// kucoin: 49980 - 99.96 = 49880.04
	// coinbase: 50000 - 25.00 = 49975.00 (Winner for Sell)
	// kraken: 49985 - 74.98 = 49910.02
	routes, _ = router.CompareRoutes("BTC", "sell", 1.0)
	if routes[0].Exchange != "coinbase" {
		t.Errorf("expected coinbase to be best sell route, got %s", routes[0].Exchange)
	}
}
