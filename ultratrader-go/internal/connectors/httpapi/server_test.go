package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/execution"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/portfolio"
)

func TestNewHandlerHealthAndReady(t *testing.T) {
	h := NewHandler(Dependencies{
		StatusProvider:           func() Status { return Status{Name: "ultratrader-go", Ready: true, AccountCount: 1} },
		PortfolioProvider:        func() PortfolioSnapshot { return PortfolioSnapshot{} },
		OrdersProvider:           func() []exchange.Order { return nil },
		ExecutionSummaryProvider: func() execution.Summary { return execution.Summary{} },
	})

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("healthz expected 200, got %d", w.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/readyz", nil)
	w = httptest.NewRecorder()
	h.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("readyz expected 200, got %d", w.Code)
	}
}

func TestPortfolioOrdersAndSummaryEndpoints(t *testing.T) {
	h := NewHandler(Dependencies{
		StatusProvider: func() Status { return Status{Name: "ultratrader-go", Ready: true, AccountCount: 1} },
		PortfolioProvider: func() PortfolioSnapshot {
			return PortfolioSnapshot{Positions: []portfolio.Position{{Symbol: "BTCUSDT", Quantity: 0.5}}, TotalMarketValue: 32500, TotalUnrealizedPnL: 2500}
		},
		OrdersProvider:           func() []exchange.Order { return []exchange.Order{{ID: "ord-1", Symbol: "BTCUSDT"}} },
		ExecutionSummaryProvider: func() execution.Summary { return execution.Summary{TotalOrders: 1, LastOrderID: "ord-1"} },
	})

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/portfolio", nil))
	if !strings.Contains(w.Body.String(), "BTCUSDT") || !strings.Contains(w.Body.String(), "2500") {
		t.Fatalf("unexpected portfolio response %q", w.Body.String())
	}

	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/orders", nil))
	if !strings.Contains(w.Body.String(), "ord-1") {
		t.Fatalf("expected ord-1 in orders response, got %q", w.Body.String())
	}

	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/execution-summary", nil))
	if !strings.Contains(w.Body.String(), "ord-1") {
		t.Fatalf("expected execution summary in response, got %q", w.Body.String())
	}
}
