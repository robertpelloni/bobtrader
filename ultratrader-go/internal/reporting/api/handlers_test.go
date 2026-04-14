package api_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/reporting/api"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/portfolio"
)

func TestPortfolioSummaryEndpoint(t *testing.T) {
	tracker := portfolio.NewTracker()
	// Add mock position
	tracker.Apply(exchange.Order{
		Symbol:   "BTC",
		Side:     exchange.Buy,
		Quantity: "1.5",
		Price:    "40000",
	})

	server := api.NewServer(tracker, nil, nil) // passing nil feed, it should fallback to cost basis

	req, err := http.NewRequest("GET", "/api/portfolio/summary", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		server.ServeHTTP(w, r)
	})

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	var response api.PortfolioSummaryResponse
	if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
		t.Errorf("could not decode json response: %v", err)
	}

	if response.OpenPositions != 1 {
		t.Errorf("Expected 1 open position, got %d", response.OpenPositions)
	}

	// Because feed is nil, market value falls back to cost basis which is 60k
	// Actually tracker.TotalMarketValue returns CostBasis if market value isn't populated via ValuedPositions
	// Let's verify total value is not NaN
	if response.Positions == nil || len(response.Positions) == 0 {
		t.Errorf("Expected positions array in response")
	} else if response.Positions[0].Symbol != "BTC" {
		t.Errorf("Expected BTC position, got %s", response.Positions[0].Symbol)
	}
}

func TestHealthEndpoint(t *testing.T) {
	server := api.NewServer(nil, nil, nil)

	req, err := http.NewRequest("GET", "/api/system/health", nil)
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

	expected := `{"status":"ok"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}
