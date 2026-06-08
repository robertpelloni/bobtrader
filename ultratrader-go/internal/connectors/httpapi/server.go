package httpapi

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/metrics"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/persistence/reports"
	reportinganalysis "github.com/robertpelloni/bobtrader/ultratrader-go/internal/reporting/analysis"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
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
	Concentration      map[string]float64   `json:"concentration,omitempty"`
	TotalMarketValue   float64              `json:"total_market_value"`
	TotalRealizedPnL   float64              `json:"total_realized_pnl"`
	TotalUnrealizedPnL float64              `json:"total_unrealized_pnl"`
}

type PortfolioSummary struct {
	OpenPositions      int                `json:"open_positions"`
	Concentration      map[string]float64 `json:"concentration,omitempty"`
	TotalMarketValue   float64            `json:"total_market_value"`
	TotalRealizedPnL   float64            `json:"total_realized_pnl"`
	TotalUnrealizedPnL float64            `json:"total_unrealized_pnl"`
}

type ExecutionDiagnostics struct {
	Summary execution.Summary `json:"summary"`
	Metrics metrics.Snapshot  `json:"metrics"`
}

type GuardDiagnostics struct {
	ActiveGuards []string        `json:"active_guards"`
	Metrics      metrics.Snapshot `json:"metrics"`
}

type ExposureDiagnostics struct {
	OpenPositions       int                `json:"open_positions"`
	Concentration       map[string]float64 `json:"concentration,omitempty"`
	TopConcentration    string             `json:"top_concentration,omitempty"`
	TopConcentrationPct float64            `json:"top_concentration_pct,omitempty"`
	TotalMarketValue    float64            `json:"total_market_value"`
	TotalRealizedPnL    float64            `json:"total_realized_pnl"`
	TotalUnrealizedPnL  float64            `json:"total_unrealized_pnl"`
}

type RuntimeConfig struct {
	Environment string        `json:"environment"`
	Scheduler   SchedulerInfo `json:"scheduler"`
	Risk        RiskInfo      `json:"risk"`
	Strategy    StrategyInfo  `json:"strategy"`
	MarketData  MarketDataInfo `json:"market_data"`
}

type SchedulerInfo struct {
	Mode       string `json:"mode"`
	IntervalMS int    `json:"interval_ms"`
	Enabled    bool   `json:"enabled"`
}

type RiskInfo struct {
	MaxNotional          float64  `json:"max_notional"`
	MaxNotionalPerSymbol float64  `json:"max_notional_per_symbol"`
	AllowedSymbols      []string `json:"allowed_symbols"`
	CooldownMS          int      `json:"cooldown_ms"`
	MaxOpenPositions    int      `json:"max_open_positions"`
	MaxConcentrationPct float64  `json:"max_concentration_pct"`
}

type StrategyInfo struct {
	RiskPct                float64 `json:"risk_pct"`
	MaxNotional            float64 `json:"max_notional"`
	TrailingActivatePct    float64 `json:"trailing_activate_pct"`
	TrailingGapPct         float64 `json:"trailing_gap_pct"`
	TrailingStopLossPct    float64 `json:"trailing_stop_loss_pct"`
	TrailingMaxHoldMinutes int     `json:"trailing_max_hold_minutes"`
	BollingerPeriod        int     `json:"bollinger_period"`
	BollingerStdDev        float64 `json:"bollinger_std_dev"`
	RSIPeriod              int     `json:"rsi_period"`
	RSIOversold            float64 `json:"rsi_oversold"`
	RSIOverbought          float64 `json:"rsi_overbought"`
	EMAFast                int     `json:"ema_fast"`
	EMASlow                int     `json:"ema_slow"`
}

type MarketDataInfo struct {
	Source         string  `json:"source"`
	InitialBalance float64 `json:"initial_balance"`
}


type Dependencies struct {
	StatusProvider               func() Status
	PortfolioProvider            func() PortfolioSnapshot
	PortfolioSummaryProvider     func() PortfolioSummary
	OrdersProvider               func() []exchange.Order
	ExecutionSummaryProvider     func() execution.Summary
	ExecutionDiagnosticsProvider func() ExecutionDiagnostics
	ExposureDiagnosticsProvider  func() ExposureDiagnostics
	MetricsProvider              func() metrics.Snapshot
	GuardNamesProvider           func() []string
	ConfigProvider               func() RuntimeConfig
	LatestReportsProvider        func() map[string]reports.Report
	ReportHistoryProvider        func(reportType string, limit int) []reports.Report
	ReportTrendsProvider         func() reportinganalysis.RuntimeTrends
	SignalLogProvider            func() []strategy.LoggedSignal
	StrategyStatsProvider        func() map[string]strategy.StrategyStats
}

func NewHandler(deps Dependencies) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(dashboardHTML))
	})

	mux.HandleFunc("/dashboard", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(dashboardHTML))
	})

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

	mux.HandleFunc("/api/portfolio-summary", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(deps.PortfolioSummaryProvider())
	})

	mux.HandleFunc("/api/orders", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(deps.OrdersProvider())
	})

	mux.HandleFunc("/api/execution-summary", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(deps.ExecutionSummaryProvider())
	})

	mux.HandleFunc("/api/execution-diagnostics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(deps.ExecutionDiagnosticsProvider())
	})

	mux.HandleFunc("/api/exposure-diagnostics", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(deps.ExposureDiagnosticsProvider())
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
		_ = json.NewEncoder(w).Encode(GuardDiagnostics{
			ActiveGuards: deps.GuardNamesProvider(),
			Metrics:      deps.MetricsProvider(),
		})
	})

	mux.HandleFunc("/api/config", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(deps.ConfigProvider())
	})

	mux.HandleFunc("/api/runtime-reports/latest", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(deps.LatestReportsProvider())
	})

	mux.HandleFunc("/api/runtime-reports/history", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		reportType := r.URL.Query().Get("type")
		_ = json.NewEncoder(w).Encode(deps.ReportHistoryProvider(reportType, limit))
	})

	mux.HandleFunc("/api/runtime-reports/trends", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(deps.ReportTrendsProvider())
	})

	mux.HandleFunc("/api/signals", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(deps.SignalLogProvider())
	})

	mux.HandleFunc("/api/strategy-stats", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(deps.StrategyStatsProvider())
	})

	return mux
}
