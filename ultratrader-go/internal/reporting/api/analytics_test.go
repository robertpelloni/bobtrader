package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/analytics"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/reporting/api"
)

func TestAnalyticsServer_PatternRecognition(t *testing.T) {
	server := api.NewAnalyticsServer(nil, nil)

	req, err := http.NewRequest("GET", "/api/analytics/patterns?symbol=BTC", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.ServeHTTP(w, r)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response api.PatternResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("could not decode json response: %v", err)
	}

	if response.Symbol != "BTC" {
		t.Errorf("Expected BTC, got %s", response.Symbol)
	}

	// We mocked prices in the handler. If length > 0 it scanned correctly.
	if response.Patterns == nil {
		t.Errorf("Patterns slice should be initialized")
	}
}

func TestAnalyticsServer_Arbitrage(t *testing.T) {
	server := api.NewAnalyticsServer(nil, nil)

	req, err := http.NewRequest("GET", "/api/analytics/arbitrage?symbol_a=BTC&symbol_b=ETH", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.ServeHTTP(w, r)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response api.ArbitrageResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("could not decode json response: %v", err)
	}

	if response.SymbolA != "BTC" || response.SymbolB != "ETH" {
		t.Errorf("Expected BTC/ETH pair")
	}

	if response.Stats.Correlation == 0 {
		t.Errorf("Expected non-zero correlation calculation in JSON")
	}

	// We hardcoded a divergence so Action should be a trade intent
	if response.Action == "HOLD" {
		t.Errorf("Expected actionable z-score divergence, got HOLD")
	}
}

func TestAnalyticsServer_OrderFlow(t *testing.T) {
	server := api.NewAnalyticsServer(nil, nil)

	req, err := http.NewRequest("GET", "/api/analytics/orderflow?symbol=BTC", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.ServeHTTP(w, r)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var response api.OrderFlowResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("could not decode json response: %v", err)
	}

	if response.Symbol != "BTC" {
		t.Errorf("Expected BTC")
	}

	if len(response.Data) != 5 {
		t.Errorf("Expected 5 order flow data points, got %d", len(response.Data))
	}

	if response.Divergence != analytics.BearishDivergence {
		t.Errorf("Expected BearishDivergence mock to trigger, got %s", response.Divergence)
	}
}
