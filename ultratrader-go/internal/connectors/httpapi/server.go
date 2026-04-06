package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/metrics"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/persistence/reports"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/execution"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/portfolio"
)

type Status struct {
	Name         string `json:"name"`
	Ready        bool   `json:"ready"`
	AccountCount int    `json:"account_count"`
}

type PortfolioSnapshot struct {
	Positions          []portfolio.Position `json:"positions"`
	TotalMarketValue   float64              `json:"total_market_value"`
	TotalRealizedPnL   float64              `json:"total_realized_pnl"`
	TotalUnrealizedPnL float64              `json:"total_unrealized_pnl"`
}

type GuardDiagnostics struct {
	ActiveGuards []string         `json:"active_guards"`
	Metrics      metrics.Snapshot `json:"metrics"`
}

type Dependencies struct {
	StatusProvider           func() Status
	PortfolioProvider        func() PortfolioSnapshot
	OrdersProvider           func() []exchange.Order
	ExecutionSummaryProvider func() execution.Summary
	MetricsProvider          func() metrics.Snapshot
	GuardNamesProvider       func() []string
	LatestReportsProvider    func() map[string]reports.Report
}

func NewHandler(deps Dependencies) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		status := deps.StatusProvider()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"ok": true, "name": status.Name})
	})
	mux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		status := deps.StatusProvider()
		w.Header().Set("Content-Type", "application/json")
		code := http.StatusOK
		if !status.Ready {
			code = http.StatusServiceUnavailable
		}
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(status)
	})
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(deps.StatusProvider())
	})
	mux.HandleFunc("/api/portfolio", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(deps.PortfolioProvider())
	})
	mux.HandleFunc("/api/orders", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(deps.OrdersProvider())
	})
	mux.HandleFunc("/api/execution-summary", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(deps.ExecutionSummaryProvider())
	})
	mux.HandleFunc("/api/metrics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(deps.MetricsProvider())
	})
	mux.HandleFunc("/api/guards", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"guards": deps.GuardNamesProvider()})
	})
	mux.HandleFunc("/api/guard-diagnostics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(GuardDiagnostics{ActiveGuards: deps.GuardNamesProvider(), Metrics: deps.MetricsProvider()})
	})
	mux.HandleFunc("/api/runtime-reports/latest", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(deps.LatestReportsProvider())
	})
	return mux
}
