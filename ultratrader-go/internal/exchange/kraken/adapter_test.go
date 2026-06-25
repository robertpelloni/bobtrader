package kraken

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestKrakenAdapter_ListMarkets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"error":[],"result":{"XXBTZUSD":{"altname":"XBTUSD","base":"XXBT","quote":"ZUSD"}}}`))
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
	if markets[0].Symbol != "XXBTZUSD" {
		t.Errorf("expected XXBTZUSD, got %s", markets[0].Symbol)
	}
}

func TestKrakenAdapter_Balances(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"error":[],"result":{"ZUSD":"100.50","XXBT":"0.5"}}`))
	}))
	defer server.Close()

	// Use a base64 string for secret key to avoid decode errors
	dummySecret := "S0JDcmFrZW5TZWNyZXRLZXlGb3JUZXN0aW5nMTIzNDU2Nzg5MA=="
	adapter := New(Config{BaseURL: server.URL, APIKey: "key", SecretKey: dummySecret})
	balances, err := adapter.Balances(context.Background())
	if err != nil {
		t.Fatalf("Balances failed: %v", err)
	}

	if len(balances) != 2 {
		t.Errorf("expected 2 balances, got %d", len(balances))
	}
}
