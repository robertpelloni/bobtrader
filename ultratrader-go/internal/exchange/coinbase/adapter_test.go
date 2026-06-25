package coinbase

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCoinbaseAdapter_ListMarkets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"products":[{"product_id":"BTC-USD","base_currency_id":"BTC","quote_currency_id":"USD","is_disabled":false}]}`))
	}))
	defer server.Close()

	adapter := New(Config{BaseURL: server.URL})
	markets, err := adapter.ListMarkets(context.Background())
	if err != nil {
		t.Fatalf("ListMarkets failed: %v", err)
	}

	if len(markets) != 1 {
		t.Errorf("expected 1 market, got %d", len(markets))
	}
	if markets[0].Symbol != "BTC-USD" {
		t.Errorf("expected BTC-USD, got %s", markets[0].Symbol)
	}
}

func TestCoinbaseAdapter_Balances(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"accounts":[{"currency":"BTC","available_balance":{"value":"1.5"},"hold":{"value":"0.5"}}]}`))
	}))
	defer server.Close()

	adapter := New(Config{BaseURL: server.URL, APIKey: "key", SecretKey: "secret"})
	balances, err := adapter.Balances(context.Background())
	if err != nil {
		t.Fatalf("Balances failed: %v", err)
	}

	if len(balances) != 1 {
		t.Errorf("expected 1 balance, got %d", len(balances))
	}
	if balances[0].Asset != "BTC" {
		t.Errorf("expected BTC, got %s", balances[0].Asset)
	}
}
