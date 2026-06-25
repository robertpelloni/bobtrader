package kucoin

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

func TestKuCoinAdapter_ListMarkets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":"200000","data":[{"symbol":"BTC-USDT","baseCurrency":"BTC","quoteCurrency":"USDT","enableTrading":true}]}`))
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
	if markets[0].Symbol != "BTC-USDT" {
		t.Errorf("expected BTC-USDT, got %s", markets[0].Symbol)
	}
}

func TestKuCoinAdapter_Balances(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":"200000","data":[{"currency":"BTC","type":"trade","balance":"1.5","available":"1.0","holds":"0.5"}]}`))
	}))
	defer server.Close()

	adapter := New(Config{BaseURL: server.URL, APIKey: "key", SecretKey: "secret", Passphrase: "pass"})
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
	if balances[0].Total != "1.5" {
		t.Errorf("expected 1.5 total, got %s", balances[0].Total)
	}
}

func TestKuCoinAdapter_PlaceOrder(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"code":"200000","data":{"orderId":"5bd6e9286443497e8a83a05b"}}`))
	}))
	defer server.Close()

	adapter := New(Config{BaseURL: server.URL, APIKey: "key", SecretKey: "secret", Passphrase: "pass"})
	req := exchange.OrderRequest{
		Symbol:   "BTC-USDT",
		Side:     exchange.Buy,
		Type:     exchange.MarketOrder,
		Quantity: "0.01",
	}
	order, err := adapter.PlaceOrder(context.Background(), req)
	if err != nil {
		t.Fatalf("PlaceOrder failed: %v", err)
	}

	if order.ID != "5bd6e9286443497e8a83a05b" {
		t.Errorf("expected orderId 5bd6e9286443497e8a83a05b, got %s", order.ID)
	}
}
