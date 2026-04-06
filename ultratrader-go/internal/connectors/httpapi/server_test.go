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
	reportinganalysis "github.com/robertpelloni/bobtrader/ultratrader-go/internal/reporting/analysis"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/execution"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/portfolio"
)

func TestNewHandlerHealthAndReady(t *testing.T) {
	h := NewHandler(Dependencies{
		StatusProvider:               func() Status { return Status{Name: "ultratrader-go", Ready: true, AccountCount: 1} },
		PortfolioProvider:            func() PortfolioSnapshot { return PortfolioSnapshot{} },
		PortfolioSummaryProvider:     func() PortfolioSummary { return PortfolioSummary{} },
		OrdersProvider:               func() []exchange.Order { return nil },
		ExecutionSummaryProvider:     func() execution.Summary { return execution.Summary{} },
		ExecutionDiagnosticsProvider: func() ExecutionDiagnostics { return ExecutionDiagnostics{} },
		ExposureDiagnosticsProvider:  func() ExposureDiagnostics { return ExposureDiagnostics{} },
		MetricsProvider:              func() metrics.Snapshot { return metrics.Snapshot{} },
		GuardNamesProvider:           func() []string { return nil },
		LatestReportsProvider:        func() map[string]reports.Report { return nil },
		ReportHistoryProvider:        func(reportType string, limit int) []reports.Report { return nil },
		ReportTrendsProvider:         func() reportinganalysis.RuntimeTrends { return reportinganalysis.RuntimeTrends{} },
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
		PortfolioSummaryProvider: func() PortfolioSummary {
			return PortfolioSummary{OpenPositions: 1, Concentration: map[string]float64{"BTCUSDT": 1}, TotalMarketValue: 32500, TotalUnrealizedPnL: 2500}
		},
		OrdersProvider:           func() []exchange.Order { return []exchange.Order{{ID: "ord-1", Symbol: "BTCUSDT"}} },
		ExecutionSummaryProvider: func() execution.Summary { return execution.Summary{TotalOrders: 1, LastOrderID: "ord-1"} },
		ExecutionDiagnosticsProvider: func() ExecutionDiagnostics {
			return ExecutionDiagnostics{Summary: execution.Summary{TotalOrders: 1, LastOrderID: "ord-1"}, Metrics: metrics.Snapshot{ExecutionAttempts: 2, ExecutionSuccess: 1, ExecutionBlocked: 1}}
		},
		ExposureDiagnosticsProvider: func() ExposureDiagnostics {
			return ExposureDiagnostics{OpenPositions: 1, Concentration: map[string]float64{"BTCUSDT": 1}, TopConcentration: "BTCUSDT", TopConcentrationPct: 1, TotalMarketValue: 32500}
		},
		MetricsProvider: func() metrics.Snapshot {
			return metrics.Snapshot{ExecutionAttempts: 2, ExecutionSuccess: 1, ExecutionBlocked: 1, BlockReasons: map[string]int{"cooldown": 1}}
		},
		GuardNamesProvider: func() []string { return []string{"symbol-whitelist", "max-notional"} },
		LatestReportsProvider: func() map[string]reports.Report {
			return map[string]reports.Report{"startup-summary": {Timestamp: time.Now(), Type: "startup-summary"}}
		},
		ReportHistoryProvider: func(reportType string, limit int) []reports.Report {
			return []reports.Report{{Timestamp: time.Now(), Type: "startup-summary"}, {Timestamp: time.Now(), Type: "metrics-snapshot"}}
		},
		ReportTrendsProvider: func() reportinganalysis.RuntimeTrends {
			return reportinganalysis.RuntimeTrends{MetricsSamples: 2, PortfolioValue: reportinganalysis.NumericTrend{Latest: 32500, Previous: 30000, Delta: 2500}}
		},
	})

	w := httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/portfolio", nil))
	if !strings.Contains(w.Body.String(), "BTCUSDT") || !strings.Contains(w.Body.String(), "2500") || !strings.Contains(w.Body.String(), "concentration") {
		t.Fatalf("unexpected portfolio response %q", w.Body.String())
	}

	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/portfolio-summary", nil))
	if !strings.Contains(w.Body.String(), "open_positions") || !strings.Contains(w.Body.String(), "32500") {
		t.Fatalf("expected portfolio summary response, got %q", w.Body.String())
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
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/execution-diagnostics", nil))
	if !strings.Contains(w.Body.String(), "execution_attempts") || !strings.Contains(w.Body.String(), "ord-1") {
		t.Fatalf("expected execution diagnostics response, got %q", w.Body.String())
	}

	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/exposure-diagnostics", nil))
	if !strings.Contains(w.Body.String(), "top_concentration") || !strings.Contains(w.Body.String(), "BTCUSDT") {
		t.Fatalf("expected exposure diagnostics response, got %q", w.Body.String())
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

	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/runtime-reports/history?limit=2&type=startup-summary", nil))
	if !strings.Contains(w.Body.String(), "startup-summary") {
		t.Fatalf("expected report history response, got %q", w.Body.String())
	}

	w = httptest.NewRecorder()
	h.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/api/runtime-reports/trends", nil))
	if !strings.Contains(w.Body.String(), "32500") {
		t.Fatalf("expected report trends response, got %q", w.Body.String())
	}
}
