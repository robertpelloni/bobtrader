package httpapi

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/metrics"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/persistence/reports"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/execution"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/portfolio"
)

func TestNewHandlerHealthAndReady(t *testing.T) {
	h := NewHandler(Dependencies{
		StatusProvider:           func() Status { return Status{Name: "ultratrader-go", Ready: true, AccountCount: 1} },
		PortfolioProvider:        func() PortfolioSnapshot { return PortfolioSnapshot{} },
		OrdersProvider:           func() []exchange.Order { return nil },
		ExecutionSummaryProvider: func() execution.Summary { return execution.Summary{} },
		MetricsProvider:          func() metrics.Snapshot { return metrics.Snapshot{} },
		GuardNamesProvider:       func() []string { return nil },
		LatestReportsProvider:    func() map[string]reports.Report { return nil },
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

func TestDiagnosticsEndpoints(t *testing.T) {
	h := NewHandler(Dependencies{
		StatusProvider: func() Status { return Status{Name: "ultratrader-go", Ready: true, AccountCount: 1} },
		PortfolioProvider: func() PortfolioSnapshot {
			return PortfolioSnapshot{Positions: []portfolio.Position{{Symbol: "BTCUSDT", Quantity: 0.5}}, Concentration: map[string]float64{"BTCUSDT": 1}, TotalMarketValue: 32500, TotalUnrealizedPnL: 2500}
		},
		OrdersProvider:           func() []exchange.Order { return []exchange.Order{{ID: "ord-1", Symbol: "BTCUSDT"}} },
		ExecutionSummaryProvider: func() execution.Summary { return execution.Summary{TotalOrders: 1, LastOrderID: "ord-1"} },
		MetricsProvider: func() metrics.Snapshot {
			return metrics.Snapshot{ExecutionAttempts: 2, ExecutionSuccess: 1, ExecutionBlocked: 1, BlockReasons: map[string]int{"cooldown": 1}}
		},
		GuardNamesProvider: func() []string { return []string{"symbol-whitelist", "max-notional"} },
		LatestReportsProvider: func() map[string]reports.Report {
			return map[string]reports.Report{"startup-summary": {Timestamp: time.Now(), Type: "startup-summary"}}
		},
	})

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/portfolio", nil))
	if !strings.Contains(w.Body.String(), "BTCUSDT") || !strings.Contains(w.Body.String(), "2500") || !strings.Contains(w.Body.String(), "concentration") {
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

	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/metrics", nil))
	if !strings.Contains(w.Body.String(), "execution_attempts") {
		t.Fatalf("expected metrics response, got %q", w.Body.String())
	}

	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/guards", nil))
	if !strings.Contains(w.Body.String(), "max-notional") {
		t.Fatalf("expected guard response, got %q", w.Body.String())
	}

	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/guard-diagnostics", nil))
	if !strings.Contains(w.Body.String(), "cooldown") || !strings.Contains(w.Body.String(), "symbol-whitelist") {
		t.Fatalf("expected guard diagnostics response, got %q", w.Body.String())
	}

	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/runtime-reports/latest", nil))
	if !strings.Contains(w.Body.String(), "startup-summary") {
		t.Fatalf("expected latest reports response, got %q", w.Body.String())
	}
}
