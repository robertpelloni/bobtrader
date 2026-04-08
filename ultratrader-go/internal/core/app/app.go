package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/connectors/httpapi"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/eventlog"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/logging"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	exchangepaper "github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange/paper"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	marketdatapaper "github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata/paper"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/metrics"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/persistence/orders"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/persistence/reports"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/persistence/snapshot"
	reportinganalysis "github.com/robertpelloni/bobtrader/ultratrader-go/internal/reporting/analysis"
	reportingruntime "github.com/robertpelloni/bobtrader/ultratrader-go/internal/reporting/runtime"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/risk"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
	strategydemo "github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/demo"
	strategyscheduler "github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/scheduler"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/execution"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/portfolio"
)

type starter interface{ Start(context.Context) }
type cycleRunner interface{ RunOnce(context.Context) error }

type App struct {
	config           config.Config
	logger           *logging.Logger
	eventLog         *eventlog.Log
	snapshotStore    *snapshot.Store
	orderStore       *orders.Store
	reportStore      *reports.Store
	accountService   *account.Service
	exchangeRegistry *exchange.Registry
	marketDataFeed   marketdata.StreamFeed
	metricsTracker   *metrics.Tracker
	executionRepo    *execution.Repository
	portfolioTracker *portfolio.Tracker
	executionService *execution.Service
	strategyRuntime  *strategy.Runtime
	scheduler        *strategyscheduler.Scheduler
	schedulerService starter
	cycleRunner      cycleRunner
	pipeline         *risk.Pipeline
	httpHandler      http.Handler
	httpRuntime      *httpapi.Runtime
}

func New(cfg config.Config) (*App, error) {
	logger, err := logging.New(logging.Config{Path: cfg.Logging.Path, Stdout: cfg.Logging.Stdout})
	if err != nil {
		return nil, fmt.Errorf("create logger: %w", err)
	}
	eventLog, err := eventlog.New(cfg.EventLog.Path)
	if err != nil {
		return nil, fmt.Errorf("create event log: %w", err)
	}
	snapshotStore, err := snapshot.NewStore(cfg.Snapshots.Path)
	if err != nil {
		return nil, fmt.Errorf("create snapshot store: %w", err)
	}
	orderStore, err := orders.NewStore(cfg.Orders.Path)
	if err != nil {
		return nil, fmt.Errorf("create order store: %w", err)
	}
	reportStore, err := reports.NewStore(cfg.Reports.Path)
	if err != nil {
		return nil, fmt.Errorf("create report store: %w", err)
	}
	buildReportTrends := func() reportinganalysis.RuntimeTrends {
		metricHistory, _ := reportStore.ListByType("metrics-snapshot", 100)
		valuationHistory, _ := reportStore.ListByType("portfolio-valuation", 100)
		executionHistory, _ := reportStore.ListByType("execution-summary", 100)
		return reportinganalysis.BuildRuntimeTrends(metricHistory, valuationHistory, executionHistory)
	}
	accountService, err := account.NewService(cfg.Accounts)
	if err != nil {
		return nil, fmt.Errorf("create account service: %w", err)
	}

	registry := exchange.NewRegistry()
	if err := registry.Register("paper", func() exchange.Adapter { return exchangepaper.New() }); err != nil {
		return nil, fmt.Errorf("register paper exchange: %w", err)
	}

	marketDataFeed := marketdatapaper.New()
	metricsTracker := metrics.NewTracker()
	executionRepo := execution.NewRepository()
	portfolioTracker := portfolio.NewTracker()
	exposureView := portfolio.NewExposureView(portfolioTracker, marketDataFeed)
	pipeline := risk.NewPipeline(
		risk.NewSymbolWhitelistGuard(cfg.Risk.AllowedSymbols),
		risk.NewMaxNotionalGuard(cfg.Risk.MaxNotional),
		risk.NewMaxNotionalPerSymbolGuard(cfg.Risk.MaxNotionalPerSymbol, exposureView),
		risk.NewCooldownGuard(time.Duration(cfg.Risk.CooldownMS)*time.Millisecond),
		risk.NewDuplicateSymbolGuard(executionRepo, time.Duration(cfg.Risk.DuplicateWindowMS)*time.Millisecond),
		risk.NewDuplicateSideGuard(executionRepo, exchange.Buy, time.Duration(cfg.Risk.DuplicateSideWindowMS)*time.Millisecond),
		risk.NewMaxOpenPositionsGuard(cfg.Risk.MaxOpenPositions, portfolioTracker),
		risk.NewMaxConcentrationGuard(cfg.Risk.MaxConcentrationPct, exposureView),
	)
	executionService := execution.NewService(accountService, registry, pipeline, eventLog, orderStore, executionRepo, portfolioTracker, logger, metricsTracker)
	var strategyRuntime *strategy.Runtime
	if cfg.Scheduler.Mode == "stream" {
		strategyRuntime = strategy.NewRuntime(
			strategydemo.NewTickPriceThreshold("paper-main", "BTCUSDT", "0.01", "70000.00"),
			strategydemo.NewTickMomentumBurst("paper-main", "BTCUSDT", "0.01", 3, 0.05, 0.05),
			strategydemo.NewTickMeanReversion("paper-main", "BTCUSDT", "0.01", 3, 0.1, 0.1),
		)
	} else {
		strategyRuntime = strategy.NewRuntime(
			strategydemo.NewPriceThreshold("paper-main", "BTCUSDT", "0.01", "70000.00", marketDataFeed),
			strategydemo.NewEMACrossover("paper-main", "ETHUSDT", "0.1", 5, 10, marketDataFeed),
		)
	}
	scheduler := strategyscheduler.New(strategyRuntime, executionService)
	reportProvider := func(ctx context.Context) []reports.Report {
		return []reports.Report{
			{Type: "metrics-snapshot", Payload: map[string]any{"metrics": metricsTracker.Snapshot()}},
			{Type: "portfolio-valuation", Payload: map[string]any{"portfolio_value": portfolioTracker.TotalMarketValue(ctx, marketDataFeed), "realized_pnl": portfolioTracker.TotalRealizedPnL(), "unrealized_pnl": portfolioTracker.TotalUnrealizedPnL(ctx, marketDataFeed), "concentration": portfolioTracker.Concentration(ctx, marketDataFeed)}},
			{Type: "execution-summary", Payload: map[string]any{"summary": executionRepo.Summary()}},
		}
	}
	cycleRunner := reportingruntime.NewReportingRunner(scheduler, reportStore, reportProvider)
	var schedulerService interface{ Start(context.Context) }
	if cfg.Scheduler.Mode == "stream" {
		reportingTickRunner := reportingruntime.NewReportingTickRunner(scheduler, reportStore, reportProvider)
		schedulerService = strategyscheduler.NewStreamService(reportingTickRunner, marketDataFeed, cfg.Risk.AllowedSymbols, time.Duration(cfg.Scheduler.IntervalMS)*time.Millisecond)
	} else {
		schedulerService = strategyscheduler.NewService(cycleRunner, time.Duration(cfg.Scheduler.IntervalMS)*time.Millisecond)
	}

	currentConcentration := func() map[string]float64 {
		return portfolioTracker.Concentration(context.Background(), marketDataFeed)
	}
	topConcentration := func() (string, float64) {
		concentration := currentConcentration()
		var topSymbol string
		var topPct float64
		for symbol, pct := range concentration {
			if pct > topPct {
				topSymbol = symbol
				topPct = pct
			}
		}
		return topSymbol, topPct
	}

	handler := httpapi.NewHandler(httpapi.Dependencies{
		StatusProvider: func() httpapi.Status {
			return httpapi.Status{Name: "ultratrader-go", Ready: true, AccountCount: len(accountService.List())}
		},
		PortfolioProvider: func() httpapi.PortfolioSnapshot {
			return httpapi.PortfolioSnapshot{Positions: portfolioTracker.ValuedPositions(context.Background(), marketDataFeed), Concentration: currentConcentration(), TotalMarketValue: portfolioTracker.TotalMarketValue(context.Background(), marketDataFeed), TotalRealizedPnL: portfolioTracker.TotalRealizedPnL(), TotalUnrealizedPnL: portfolioTracker.TotalUnrealizedPnL(context.Background(), marketDataFeed)}
		},
		PortfolioSummaryProvider: func() httpapi.PortfolioSummary {
			return httpapi.PortfolioSummary{OpenPositions: portfolioTracker.OpenPositionCount(), Concentration: currentConcentration(), TotalMarketValue: portfolioTracker.TotalMarketValue(context.Background(), marketDataFeed), TotalRealizedPnL: portfolioTracker.TotalRealizedPnL(), TotalUnrealizedPnL: portfolioTracker.TotalUnrealizedPnL(context.Background(), marketDataFeed)}
		},
		OrdersProvider:           func() []exchange.Order { return executionRepo.List() },
		ExecutionSummaryProvider: func() execution.Summary { return executionRepo.Summary() },
		ExecutionDiagnosticsProvider: func() httpapi.ExecutionDiagnostics {
			return httpapi.ExecutionDiagnostics{Summary: executionRepo.Summary(), Metrics: metricsTracker.Snapshot()}
		},
		ExposureDiagnosticsProvider: func() httpapi.ExposureDiagnostics {
			topSymbol, topPct := topConcentration()
			return httpapi.ExposureDiagnostics{OpenPositions: portfolioTracker.OpenPositionCount(), Concentration: currentConcentration(), TopConcentration: topSymbol, TopConcentrationPct: topPct, TotalMarketValue: portfolioTracker.TotalMarketValue(context.Background(), marketDataFeed), TotalRealizedPnL: portfolioTracker.TotalRealizedPnL(), TotalUnrealizedPnL: portfolioTracker.TotalUnrealizedPnL(context.Background(), marketDataFeed)}
		},
		MetricsProvider:    func() metrics.Snapshot { return metricsTracker.Snapshot() },
		GuardNamesProvider: func() []string { return pipeline.Names() },
		LatestReportsProvider: func() map[string]reports.Report {
			latest, err := reportStore.LatestByType()
			if err != nil {
				return map[string]reports.Report{}
			}
			return latest
		},
		ReportHistoryProvider: func(reportType string, limit int) []reports.Report {
			history, err := reportStore.ListByType(reportType, limit)
			if err != nil {
				return nil
			}
			return history
		},
		ReportTrendsProvider: func() reportinganalysis.RuntimeTrends { return buildReportTrends() },
	})

	var runtime *httpapi.Runtime
	if cfg.Server.Enabled {
		runtime = httpapi.NewRuntime(cfg.Server.Address, handler)
	}
	return &App{config: cfg, logger: logger, eventLog: eventLog, snapshotStore: snapshotStore, orderStore: orderStore, reportStore: reportStore, accountService: accountService, exchangeRegistry: registry, marketDataFeed: marketDataFeed, metricsTracker: metricsTracker, executionRepo: executionRepo, portfolioTracker: portfolioTracker, executionService: executionService, strategyRuntime: strategyRuntime, scheduler: scheduler, schedulerService: schedulerService, cycleRunner: cycleRunner, pipeline: pipeline, httpHandler: handler, httpRuntime: runtime}, nil
}

func (a *App) Close() error {
	if a.logger != nil {
		return a.logger.Close()
	}
	return nil
}

func (a *App) Start(ctx context.Context) error {
	a.logger.Info("app startup initiated", map[string]any{"environment": a.config.Environment})
	if err := a.eventLog.Append(ctx, eventlog.Entry{Type: "app.started", Source: "ultratrader-go", Payload: map[string]any{"environment": a.config.Environment, "accounts": len(a.accountService.List())}}); err != nil {
		return err
	}
	for _, acct := range a.accountService.List() {
		if err := a.snapshotStore.Append(ctx, snapshot.Snapshot{AccountID: acct.ID, AccountName: acct.Name, Exchange: acct.ExchangeName, Metadata: map[string]any{"enabled": acct.Enabled}}); err != nil {
			return fmt.Errorf("append bootstrap snapshot for %s: %w", acct.ID, err)
		}
	}
	if a.httpRuntime != nil {
		if err := a.httpRuntime.Start(ctx); err != nil {
			return fmt.Errorf("start http runtime: %w", err)
		}
		a.logger.Info("http runtime started", map[string]any{"address": a.httpRuntime.Address()})
	}
	if a.config.Scheduler.Enabled {
		a.schedulerService.Start(ctx)
		a.logger.Info("scheduler service started", map[string]any{"interval_ms": a.config.Scheduler.IntervalMS, "mode": a.config.Scheduler.Mode})
	}
	if err := a.cycleRunner.RunOnce(ctx); err != nil {
		return fmt.Errorf("run strategy scheduler: %w", err)
	}
	reportPayload := map[string]any{"orders": len(a.executionRepo.List()), "portfolio_value": a.portfolioTracker.TotalMarketValue(ctx, a.marketDataFeed), "realized_pnl": a.portfolioTracker.TotalRealizedPnL(), "unrealized_pnl": a.portfolioTracker.TotalUnrealizedPnL(ctx, a.marketDataFeed), "metrics": a.metricsTracker.Snapshot(), "guards": a.pipeline.Names()}
	if err := a.reportStore.Append(ctx, reports.Report{Type: "startup-summary", Payload: reportPayload}); err != nil {
		return fmt.Errorf("append runtime report: %w", err)
	}
	a.logger.Info("app startup completed", reportPayload)
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	if a.httpRuntime != nil {
		if err := a.httpRuntime.Shutdown(ctx); err != nil {
			return err
		}
	}
	return a.Close()
}

func (a *App) Handler() http.Handler { return a.httpHandler }
func (a *App) Address() string {
	if a.httpRuntime != nil {
		return a.httpRuntime.Address()
	}
	return ""
}
