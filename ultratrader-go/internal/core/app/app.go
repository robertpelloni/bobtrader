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
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange/binance"
	exchangepaper "github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange/paper"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata"
	marketdatabinance "github.com/robertpelloni/bobtrader/ultratrader-go/internal/marketdata/binance"
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
	config            config.Config
	logger            *logging.Logger
	eventLog          *eventlog.Log
	snapshotStore     *snapshot.Store
	orderStore        *orders.Store
	reportStore       *reports.Store
	accountService    *account.Service
	exchangeRegistry  *exchange.Registry
	marketDataFeed    marketdata.StreamFeed
	metricsTracker    *metrics.Tracker
	executionRepo     *execution.Repository
	portfolioTracker  *portfolio.Tracker
	executionService  *execution.Service
	executionManager  *execution.Manager
	strategyRuntime   *strategy.Runtime
	scheduler         *strategyscheduler.EnhancedScheduler
	schedulerService  starter
	cycleRunner       cycleRunner
	pipeline          *risk.Pipeline
	signalLog         *strategy.SignalLog
	marketAwarePaper  *exchangepaper.MarketAwareAdapter
	httpHandler       http.Handler
	httpRuntime       *httpapi.Runtime
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

	// ── Exchange Registry ──────────────────────────────────────
	registry := exchange.NewRegistry()
	if err := registry.Register("paper", func() exchange.Adapter {
		return exchangepaper.New()
	}); err != nil {
		return nil, fmt.Errorf("register paper exchange: %w", err)
	}
	if err := registry.Register("binance", func() exchange.Adapter {
		return binance.New(binance.Config{Testnet: true})
	}); err != nil {
		return nil, fmt.Errorf("register binance exchange: %w", err)
	}

	// ── Market Data Feed ───────────────────────────────────────
	marketDataFeed := buildMarketDataFeed(cfg, logger)

	// ── Market-Aware Paper Adapter ─────────────────────────────
	// This adapter fills simulated orders at real market prices
	initialBalance := 10000.0 // $10,000 starting USDT
	marketAwarePaper := exchangepaper.NewMarketAwareAdapter(marketDataFeed, initialBalance)
	if err := registry.Register("paper-market-aware", func() exchange.Adapter {
		return marketAwarePaper
	}); err != nil {
		return nil, fmt.Errorf("register market-aware paper exchange: %w", err)
	}

	// ── Core Services ──────────────────────────────────────────
	metricsTracker := metrics.NewTracker()
	executionRepo := execution.NewRepository()
	portfolioTracker := portfolio.NewTracker()
	exposureView := portfolio.NewExposureView(portfolioTracker, marketDataFeed)

	// ── Risk Pipeline ──────────────────────────────────────────
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

	// Determine the primary account ID for strategy signals
	primaryAccountID := "paper-main"
	for _, acct := range cfg.Accounts {
		if acct.Enabled {
			primaryAccountID = acct.ID
			break
		}
	}

	// ── Strategy Signal Log ────────────────────────────────────
	signalLog := strategy.NewSignalLog(10000)

	// ── Execution Service ──────────────────────────────────────
	executionService := execution.NewService(
		accountService, registry, pipeline, eventLog, orderStore,
		executionRepo, portfolioTracker, logger, metricsTracker,
	)

	// ── Execution Manager ──────────────────────────────────────
	executionManager := execution.NewManager()
	paperAdapter := exchangepaper.New()
	executionManager.Register(execution.NewMarketStrategy(paperAdapter))
	executionManager.Register(execution.NewWolfBotBollingerStrategy(paperAdapter, 3))

	// ── Strategy Runtime ───────────────────────────────────────
	strategyRuntime := buildAutonomousStrategyRuntime(
		cfg, primaryAccountID, marketDataFeed, portfolioTracker, marketAwarePaper,
	)

	// ── Enhanced Scheduler (position-aware, signal-logged) ─────
	scheduler := strategyscheduler.NewEnhanced(
		strategyRuntime, executionService, portfolioTracker, marketDataFeed, signalLog,
	)

	// ── Reporting ──────────────────────────────────────────────
	reportProvider := func(ctx context.Context) []reports.Report {
		return []reports.Report{
			{Type: "metrics-snapshot", Payload: map[string]any{"metrics": metricsTracker.Snapshot()}},
			{Type: "portfolio-valuation", Payload: map[string]any{
				"portfolio_value": portfolioTracker.TotalMarketValue(ctx, marketDataFeed),
				"realized_pnl":    portfolioTracker.TotalRealizedPnL(),
				"unrealized_pnl":  portfolioTracker.TotalUnrealizedPnL(ctx, marketDataFeed),
				"concentration":   portfolioTracker.Concentration(ctx, marketDataFeed),
				"usdt_balance":    marketAwarePaper.USDTBalance(),
			}},
			{Type: "execution-summary", Payload: map[string]any{"summary": executionRepo.Summary()}},
			{Type: "strategy-signals", Payload: map[string]any{
				"signal_count":   signalLog.Count(),
				"strategy_stats": signalLog.StatsByStrategy(),
			}},
		}
	}
	cycleRunner := reportingruntime.NewReportingRunner(scheduler, reportStore, reportProvider)

	// ── Scheduler Service ──────────────────────────────────────
	var schedulerService starter
	if cfg.Scheduler.Mode == "stream" {
		reportingStreamRunner := reportingruntime.NewReportingStreamRunner(scheduler, scheduler, reportStore, reportProvider)
		schedulerService = strategyscheduler.NewStreamService(
			reportingStreamRunner, marketDataFeed, cfg.Risk.AllowedSymbols,
			time.Duration(cfg.Scheduler.IntervalMS)*time.Millisecond,
		)
	} else if cfg.Scheduler.Mode == "candle-stream" {
		reportingStreamRunner := reportingruntime.NewReportingStreamRunner(scheduler, scheduler, reportStore, reportProvider)
		schedulerService = strategyscheduler.NewCandleStreamService(
			reportingStreamRunner, marketDataFeed, cfg.Risk.AllowedSymbols, "1m",
		)
	} else {
		schedulerService = strategyscheduler.NewService(
			cycleRunner, time.Duration(cfg.Scheduler.IntervalMS)*time.Millisecond,
		)
	}

	// ── HTTP API ───────────────────────────────────────────────
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
			return httpapi.PortfolioSnapshot{
				Positions:         portfolioTracker.ValuedPositions(context.Background(), marketDataFeed),
				Concentration:     currentConcentration(),
				TotalMarketValue:  portfolioTracker.TotalMarketValue(context.Background(), marketDataFeed),
				TotalRealizedPnL:  portfolioTracker.TotalRealizedPnL(),
				TotalUnrealizedPnL: portfolioTracker.TotalUnrealizedPnL(context.Background(), marketDataFeed),
			}
		},
		PortfolioSummaryProvider: func() httpapi.PortfolioSummary {
			return httpapi.PortfolioSummary{
				OpenPositions:     portfolioTracker.OpenPositionCount(),
				Concentration:     currentConcentration(),
				TotalMarketValue:  portfolioTracker.TotalMarketValue(context.Background(), marketDataFeed),
				TotalRealizedPnL:  portfolioTracker.TotalRealizedPnL(),
				TotalUnrealizedPnL: portfolioTracker.TotalUnrealizedPnL(context.Background(), marketDataFeed),
			}
		},
		OrdersProvider: func() []exchange.Order {
			return executionRepo.List()
		},
		ExecutionSummaryProvider: func() execution.Summary {
			return executionRepo.Summary()
		},
		ExecutionDiagnosticsProvider: func() httpapi.ExecutionDiagnostics {
			return httpapi.ExecutionDiagnostics{
				Summary: executionRepo.Summary(),
				Metrics: metricsTracker.Snapshot(),
			}
		},
		ExposureDiagnosticsProvider: func() httpapi.ExposureDiagnostics {
			topSymbol, topPct := topConcentration()
			return httpapi.ExposureDiagnostics{
				OpenPositions:       portfolioTracker.OpenPositionCount(),
				Concentration:       currentConcentration(),
				TopConcentration:    topSymbol,
				TopConcentrationPct: topPct,
				TotalMarketValue:    portfolioTracker.TotalMarketValue(context.Background(), marketDataFeed),
				TotalRealizedPnL:    portfolioTracker.TotalRealizedPnL(),
				TotalUnrealizedPnL:  portfolioTracker.TotalUnrealizedPnL(context.Background(), marketDataFeed),
			}
		},
		MetricsProvider: func() metrics.Snapshot {
			return metricsTracker.Snapshot()
		},
		GuardNamesProvider: func() []string {
			return pipeline.Names()
		},
		ConfigProvider: func() httpapi.RuntimeConfig {
			return httpapi.RuntimeConfig{
				Environment: cfg.Environment,
				Scheduler:   httpapi.SchedulerInfo{Mode: cfg.Scheduler.Mode, IntervalMS: cfg.Scheduler.IntervalMS, Enabled: cfg.Scheduler.Enabled},
				Risk:        httpapi.RiskInfo{MaxNotional: cfg.Risk.MaxNotional, MaxNotionalPerSymbol: cfg.Risk.MaxNotionalPerSymbol, AllowedSymbols: cfg.Risk.AllowedSymbols, CooldownMS: cfg.Risk.CooldownMS, MaxOpenPositions: cfg.Risk.MaxOpenPositions, MaxConcentrationPct: cfg.Risk.MaxConcentrationPct},
			}
		},
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
		ReportTrendsProvider: func() reportinganalysis.RuntimeTrends {
			return buildReportTrends()
		},
		SignalLogProvider: func() []strategy.LoggedSignal {
			return signalLog.Recent(200)
		},
		StrategyStatsProvider: func() map[string]strategy.StrategyStats {
			return signalLog.StatsByStrategy()
		},
	})

	var runtime *httpapi.Runtime
	if cfg.Server.Enabled {
		runtime = httpapi.NewRuntime(cfg.Server.Address, handler)
	}

	return &App{
		config:           cfg,
		logger:           logger,
		eventLog:         eventLog,
		snapshotStore:    snapshotStore,
		orderStore:       orderStore,
		reportStore:      reportStore,
		accountService:   accountService,
		exchangeRegistry: registry,
		marketDataFeed:   marketDataFeed,
		metricsTracker:   metricsTracker,
		executionRepo:    executionRepo,
		portfolioTracker: portfolioTracker,
		executionService: executionService,
		executionManager: executionManager,
		strategyRuntime:  strategyRuntime,
		scheduler:        scheduler,
		schedulerService: schedulerService,
		cycleRunner:      cycleRunner,
		pipeline:         pipeline,
		signalLog:        signalLog,
		marketAwarePaper: marketAwarePaper,
		httpHandler:      handler,
		httpRuntime:      runtime,
	}, nil
}

// buildMarketDataFeed creates the appropriate market data feed based on config.
// If any account uses Binance or paper-market-aware, we use Binance real data.
// Otherwise we fall back to the deterministic paper feed.
func buildMarketDataFeed(cfg config.Config, logger *logging.Logger) marketdata.StreamFeed {
	for _, acct := range cfg.Accounts {
		if acct.Exchange == "binance" || acct.Exchange == "paper-market-aware" {
			adapter := binance.New(binance.Config{Testnet: acct.Testnet})
			feed := marketdatabinance.NewFeed(adapter)
			logger.Info("using Binance real market data feed", map[string]any{"testnet": acct.Testnet})
			return feed
		}
	}
	logger.Info("using paper market data feed", nil)
	return marketdatapaper.New()
}

// buildAutonomousStrategyRuntime creates a multi-strategy runtime for
// autonomous trading. Each symbol gets:
//   - Entry strategies: EMA crossover, Bollinger band reversion, RSI reversion
//   - Exit strategy: Trailing take-profit
//   - Position sizing: Portfolio-aware sizing based on balance
//
// Entry strategies are wrapped with PortfolioSizer for dynamic sizing.
// The trailing take-profit handles all exits.
func buildAutonomousStrategyRuntime(
	cfg config.Config,
	accountID string,
	feed marketdata.Feed,
	portfolioTracker *portfolio.Tracker,
	marketAwarePaper *exchangepaper.MarketAwareAdapter,
) *strategy.Runtime {
	symbols := cfg.Risk.AllowedSymbols
	if len(symbols) == 0 {
		symbols = []string{"BTCUSDT", "ETHUSDT"}
	}

	// Risk parameters from config
	riskPct := 2.0          // 2% of balance per trade
	maxNotional := 1000.0   // max $1000 per trade
	if cfg.Risk.MaxNotional > 0 && cfg.Risk.MaxNotional < maxNotional {
		maxNotional = cfg.Risk.MaxNotional
	}

	var strategies []strategy.Strategy

	switch cfg.Scheduler.Mode {
	case "stream", "": // Default to stream for autonomous trading
		for _, symbol := range symbols {
			// ── Entry Strategy 1: EMA Crossover (9/21) ──────────
			// Trend-following: buys on golden cross, sells on death cross
			emaBase := strategydemo.NewEMATickCrossover(accountID, symbol, "0.001", 9, 21)
			emaSized := strategydemo.NewPortfolioSizer(emaBase, symbol, marketAwarePaper, feed, riskPct, maxNotional)
			strategies = append(strategies, emaSized)

			// ── Entry Strategy 2: Bollinger Band Reversion (20, 2.0) ─
			// Mean-reversion: buys at lower band, sells at upper band
			bbBase := strategydemo.NewBollingerTickReversion(accountID, symbol, "0.001", 20, 2.0)
			bbSized := strategydemo.NewPortfolioSizer(bbBase, symbol, marketAwarePaper, feed, riskPct, maxNotional)
			strategies = append(strategies, bbSized)

			// ── Entry Strategy 3: RSI Reversion (14, 30/70) ─────
			// Mean-reversion: buys oversold, sells overbought
			rsiBase := strategydemo.NewRSIReversion(accountID, symbol, "0.001", 14, 30, 70)
			rsiSized := strategydemo.NewPortfolioSizer(rsiBase, symbol, marketAwarePaper, feed, riskPct, maxNotional)
			strategies = append(strategies, rsiSized)

			// ── Exit Strategy: Trailing Take Profit ─────────────
			// Activates at 2% profit, trails with 0.5% gap
			// Sells entire position when price drops below trail
			trailingTP := strategydemo.NewTrailingTakeProfit(
				accountID, symbol, "0.001",
				2.0,  // activate at 2% profit
				0.5,  // trail 0.5% below high
				portfolioTracker, feed,
			)
			strategies = append(strategies, trailingTP)
		}

	case "candle-stream":
		for _, symbol := range symbols {
			// Candle-based entry strategies
			strategies = append(strategies, strategydemo.NewMACDCrossover(accountID, symbol, "0.001", 12, 26, 9))
			strategies = append(strategies, strategydemo.NewBollingerReversion(accountID, symbol, "0.001", 20, 2.0))
			strategies = append(strategies, strategydemo.NewCandleSMACross(accountID, symbol, "0.001", 5, 20))
			strategies = append(strategies, strategydemo.NewATRSizing(accountID, symbol, "0.001", 0.01, 7, 25, 14))
			// Trailing exit for candle mode too
			trailingTP := strategydemo.NewTrailingTakeProfit(
				accountID, symbol, "0.001",
				2.0, 0.5, portfolioTracker, feed,
			)
			strategies = append(strategies, trailingTP)
		}

	default: // timer mode
		for _, symbol := range symbols {
			strategies = append(strategies, strategydemo.NewEMACrossover(accountID, symbol, "0.001", 9, 21, feed))
			strategies = append(strategies, strategydemo.NewPriceThreshold(accountID, symbol, "0.001", "70000.00", feed))
		}
	}

	return strategy.NewRuntime(strategies...)
}

func (a *App) Close() error {
	if a.logger != nil {
		return a.logger.Close()
	}
	return nil
}

func (a *App) Start(ctx context.Context) error {
	a.logger.Info("app startup initiated", map[string]any{"environment": a.config.Environment})

	if err := a.eventLog.Append(ctx, eventlog.Entry{
		Type: "app.started", Source: "ultratrader-go",
		Payload: map[string]any{"environment": a.config.Environment, "accounts": len(a.accountService.List())},
	}); err != nil {
		return err
	}

	for _, acct := range a.accountService.List() {
		if err := a.snapshotStore.Append(ctx, snapshot.Snapshot{
			AccountID: acct.ID, AccountName: acct.Name, Exchange: acct.ExchangeName,
			Metadata: map[string]any{"enabled": acct.Enabled},
		}); err != nil {
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
		a.logger.Info("scheduler service started", map[string]any{
			"interval_ms": a.config.Scheduler.IntervalMS,
			"mode":        a.config.Scheduler.Mode,
		})
	}

	if err := a.cycleRunner.RunOnce(ctx); err != nil {
		return fmt.Errorf("run strategy scheduler: %w", err)
	}

	reportPayload := map[string]any{
		"orders":         len(a.executionRepo.List()),
		"portfolio_value": a.portfolioTracker.TotalMarketValue(ctx, a.marketDataFeed),
		"realized_pnl":   a.portfolioTracker.TotalRealizedPnL(),
		"unrealized_pnl": a.portfolioTracker.TotalUnrealizedPnL(ctx, a.marketDataFeed),
		"metrics":        a.metricsTracker.Snapshot(),
		"guards":         a.pipeline.Names(),
		"signal_count":   a.signalLog.Count(),
		"usdt_balance":   a.marketAwarePaper.USDTBalance(),
	}
	if err := a.reportStore.Append(ctx, reports.Report{Type: "startup-summary", Payload: reportPayload}); err != nil {
		return fmt.Errorf("append runtime report: %w", err)
	}

	a.logger.Info("app startup completed", reportPayload)
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	a.logger.Info("app shutdown initiated", map[string]any{
		"signal_count":   a.signalLog.Count(),
		"strategy_stats": a.signalLog.StatsByStrategy(),
		"portfolio_value": a.portfolioTracker.TotalMarketValue(ctx, a.marketDataFeed),
		"realized_pnl":   a.portfolioTracker.TotalRealizedPnL(),
		"unrealized_pnl": a.portfolioTracker.TotalUnrealizedPnL(ctx, a.marketDataFeed),
		"usdt_balance":   a.marketAwarePaper.USDTBalance(),
	})
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
