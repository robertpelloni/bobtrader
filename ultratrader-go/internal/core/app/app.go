package app

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
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
	sentimentsentiment "github.com/robertpelloni/bobtrader/ultratrader-go/internal/analytics/sentiment"
	strategydemo "github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/demo"
	strategyscheduler "github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/scheduler"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/execution"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/portfolio"
)

type starter interface{ Start(context.Context) }
type cycleRunner interface{ RunOnce(context.Context) error }

type App struct {
	config                  config.Config
	logger                  *logging.Logger
	eventLog                *eventlog.Log
	snapshotStore           *snapshot.Store
	orderStore              *orders.Store
	reportStore             *reports.Store
	accountService          *account.Service
	exchangeRegistry        *exchange.Registry
	marketDataFeed          marketdata.StreamFeed
	metricsTracker          *metrics.Tracker
	executionRepo           *execution.Repository
	portfolioTracker        *portfolio.Tracker
	executionService        *execution.Service
	executionManager        *execution.Manager
	strategyRuntime         *strategy.Runtime
	scheduler               *strategyscheduler.EnhancedScheduler
	schedulerService        starter
	cycleRunner             cycleRunner
	pipeline                *risk.Pipeline
	signalLog               *strategy.SignalLog
	signalLogStop           func()
	marketAwarePaper        *exchangepaper.MarketAwareAdapter
	httpHandler             http.Handler
	httpRuntime             *httpapi.Runtime
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
	if err := registry.Register("paper", func() exchange.Adapter { return exchangepaper.New() }); err != nil {
		return nil, fmt.Errorf("register paper exchange: %w", err)
	}
	if err := registry.Register("binance", func() exchange.Adapter { return binance.New(binance.Config{Testnet: true}) }); err != nil {
		return nil, fmt.Errorf("register binance exchange: %w", err)
	}
	if err := registry.RegisterAccountFactory("binance", func(apiKey, secretKey string, testnet bool) exchange.Adapter {
		return binance.New(binance.Config{
			APIKey:    apiKey,
			SecretKey: secretKey,
			Testnet:   testnet,
		})
	}); err != nil {
		return nil, fmt.Errorf("register binance account factory: %w", err)
	}

	// ── Market Data Feed ───────────────────────────────────────
	marketDataFeed := buildMarketDataFeed(cfg, logger)

	// ── Market-Aware Paper Adapter ─────────────────────────────
	initialBalance := cfg.MarketData.InitialBalance
	if initialBalance <= 0 {
		initialBalance = 10000
	}
	marketAwarePaper := exchangepaper.NewMarketAwareAdapter(marketDataFeed, initialBalance)
	if err := registry.Register("paper-market-aware", func() exchange.Adapter { return marketAwarePaper }); err != nil {
		return nil, fmt.Errorf("register market-aware paper exchange: %w", err)
	}

	// ── Core Services ──────────────────────────────────────────
	metricsTracker := metrics.NewTracker()
	executionRepo := execution.NewRepository()
	portfolioTracker := portfolio.NewTracker()
	exposureView := portfolio.NewExposureViewWithBalance(portfolioTracker, marketDataFeed, marketAwarePaper)

	// ── Risk Pipeline ──────────────────────────────────────────
	pipeline := risk.NewPipeline(
		risk.NewSymbolWhitelistGuard(cfg.Risk.AllowedSymbols),
		risk.NewMaxNotionalGuard(cfg.Risk.MaxNotional),
		risk.NewMaxNotionalPerSymbolGuard(cfg.Risk.MaxNotionalPerSymbol, exposureView),
		risk.NewCooldownGuard(time.Duration(cfg.Risk.CooldownMS)*time.Millisecond),
		risk.NewDuplicateSymbolGuard(executionRepo, time.Duration(cfg.Risk.DuplicateWindowMS)*time.Millisecond),
		risk.NewDuplicateSideGuard(executionRepo, time.Duration(cfg.Risk.DuplicateSideWindowMS)*time.Millisecond),
		risk.NewMaxOpenPositionsGuard(cfg.Risk.MaxOpenPositions, portfolioTracker),
		risk.NewMaxConcentrationGuard(cfg.Risk.MaxConcentrationPct, exposureView),
	)

	// Determine the primary account ID for strategy signals.
	// Prefer paper accounts when available (safe default for paper-trading mode).
	// If no paper account exists, use the first enabled account.
	primaryAccountID := ""
	for _, acct := range cfg.Accounts {
		if acct.Enabled && (acct.Exchange == "paper" || acct.Exchange == "paper-market-aware") {
			primaryAccountID = acct.ID
			break
		}
	}
	if primaryAccountID == "" {
		for _, acct := range cfg.Accounts {
			if acct.Enabled {
				primaryAccountID = acct.ID
				break
			}
		}
	}
	if primaryAccountID == "" {
		primaryAccountID = "paper-main" // ultimate fallback
	}

	// ── Strategy Signal Log ────────────────────────────────────
	signalLog := strategy.NewSignalLog(10000)
	if err := signalLog.EnablePersistence(filepath.Join("data", "signals", "signals.jsonl")); err != nil {
		logger.Info("signal log persistence disabled", map[string]any{"error": err.Error()})
	}
	signalLogStop := signalLog.StartAutoFlush(30 * time.Second)

	// ── Execution Service ──────────────────────────────────────
	executionService := execution.NewService(
		accountService, registry, pipeline, eventLog,
		orderStore, executionRepo, portfolioTracker, logger, metricsTracker,
	)

	// ── Execution Manager ──────────────────────────────────────
	executionManager := execution.NewManager()
	paperAdapter := exchangepaper.New()
	executionManager.Register(execution.NewMarketStrategy(paperAdapter))
	executionManager.Register(execution.NewWolfBotBollingerStrategy(paperAdapter, 3))

	// ── Balance Reader for Strategy Sizing ──────────────────────────────────────
	// Use paper balance (simulated) when trading on paper account.
	// Use real Binance balance only when the primary account is a real Binance account.
	var balanceReader strategydemo.BalanceReader = marketAwarePaper
	if primaryAccountID != "" {
		for _, acct := range cfg.Accounts {
			if acct.Enabled && acct.ID == primaryAccountID && acct.Exchange == "binance" {
				binanceAdapter := binance.New(binance.Config{
					APIKey:   acct.APIKey,
					SecretKey: acct.SecretKey,
					Testnet:  acct.Testnet,
				})
				balanceReader = strategydemo.NewBinanceBalanceReader(binanceAdapter, 30*time.Second)
				break
			}
		}
	}

	// ── Strategy Runtime ───────────────────────────────────────
	strategyRuntime := buildAutonomousStrategyRuntime(
		cfg, primaryAccountID, marketDataFeed, portfolioTracker, balanceReader,
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
		OrdersProvider: func() []exchange.Order { return executionRepo.List() },
		ExecutionSummaryProvider: func() execution.Summary { return executionRepo.Summary() },
		ExecutionDiagnosticsProvider: func() httpapi.ExecutionDiagnostics {
			return httpapi.ExecutionDiagnostics{
				Summary: executionRepo.Summary(),
				Metrics: metricsTracker.Snapshot(),
			}
		},
		ExposureDiagnosticsProvider: func() httpapi.ExposureDiagnostics {
			topSymbol, topPct := topConcentration()
			return httpapi.ExposureDiagnostics{
				OpenPositions:        portfolioTracker.OpenPositionCount(),
				Concentration:        currentConcentration(),
				TopConcentration:     topSymbol,
				TopConcentrationPct:  topPct,
				TotalMarketValue:     portfolioTracker.TotalMarketValue(context.Background(), marketDataFeed),
				TotalRealizedPnL:     portfolioTracker.TotalRealizedPnL(),
				TotalUnrealizedPnL:   portfolioTracker.TotalUnrealizedPnL(context.Background(), marketDataFeed),
			}
		},
		MetricsProvider:    func() metrics.Snapshot { return metricsTracker.Snapshot() },
		GuardNamesProvider: func() []string { return pipeline.Names() },
		ConfigProvider: func() httpapi.RuntimeConfig {
			return httpapi.RuntimeConfig{
				Environment: cfg.Environment,
				Scheduler:   httpapi.SchedulerInfo{Mode: cfg.Scheduler.Mode, IntervalMS: cfg.Scheduler.IntervalMS, Enabled: cfg.Scheduler.Enabled},
				Risk:        httpapi.RiskInfo{MaxNotional: cfg.Risk.MaxNotional, MaxNotionalPerSymbol: cfg.Risk.MaxNotionalPerSymbol, AllowedSymbols: cfg.Risk.AllowedSymbols, CooldownMS: cfg.Risk.CooldownMS, MaxOpenPositions: cfg.Risk.MaxOpenPositions, MaxConcentrationPct: cfg.Risk.MaxConcentrationPct},
				Strategy:    httpapi.StrategyInfo{RiskPct: cfg.Strategy.RiskPct, MaxNotional: cfg.Strategy.MaxNotional, TrailingActivatePct: cfg.Strategy.TrailingActivatePct, TrailingGapPct: cfg.Strategy.TrailingGapPct, TrailingStopLossPct: cfg.Strategy.TrailingStopLossPct, TrailingMaxHoldMinutes: cfg.Strategy.TrailingMaxHoldMinutes, BollingerPeriod: cfg.Strategy.BollingerPeriod, BollingerStdDev: cfg.Strategy.BollingerStdDev, RSIPeriod: cfg.Strategy.RSIPeriod, RSIOversold: cfg.Strategy.RSIOversold, RSIOverbought: cfg.Strategy.RSIOverbought, EMAFast: cfg.Strategy.EMAFast, EMASlow: cfg.Strategy.EMASlow},
				MarketData:  httpapi.MarketDataInfo{Source: cfg.MarketData.Source, InitialBalance: cfg.MarketData.InitialBalance},
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
		ReportTrendsProvider: func() reportinganalysis.RuntimeTrends { return buildReportTrends() },
		SignalLogProvider:    func() []strategy.LoggedSignal { return signalLog.Recent(200) },
		StrategyStatsProvider: func() map[string]strategy.StrategyStats { return signalLog.StatsByStrategy() },
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
		signalLogStop:   signalLogStop,
		marketAwarePaper: marketAwarePaper,
		httpHandler:      handler,
		httpRuntime:      runtime,
	}, nil
}

// buildMarketDataFeed creates the appropriate market data feed based on config.
// Supports "rest" (default) and "websocket" sources for real Binance data.
// Falls back to the deterministic paper feed if no Binance accounts are configured.
func buildMarketDataFeed(cfg config.Config, logger *logging.Logger) marketdata.StreamFeed {
	for _, acct := range cfg.Accounts {
		if acct.Exchange == "binance" || acct.Exchange == "paper-market-aware" {
			adapter := binance.New(binance.Config{APIKey: acct.APIKey, SecretKey: acct.SecretKey, Testnet: acct.Testnet})

			switch cfg.MarketData.Source {
			case "websocket", "ws":
				wsFeed := marketdatabinance.NewStreamFeed(adapter)
				logger.Info("using Binance WebSocket market data feed", map[string]any{"testnet": acct.Testnet})
				return wsFeed
			default:
				restFeed := marketdatabinance.NewFeed(adapter)
				logger.Info("using Binance REST market data feed", map[string]any{"testnet": acct.Testnet})
				return restFeed
			}
		}
	}
	logger.Info("using paper market data feed", nil)
	return marketdatapaper.New()
}

// buildAutonomousStrategyRuntime creates a multi-strategy runtime for
// autonomous trading. Each symbol gets:
//   - Entry strategies: EMA crossover, Bollinger band reversion, RSI reversion
//   - Exit strategy: Trailing take-profit (configurable activation, trail, stop-loss, max hold)
//   - Position sizing: Portfolio-aware sizing based on balance
//
// All strategy parameters are driven from cfg.Strategy (config file or defaults).
func buildAutonomousStrategyRuntime(
	cfg config.Config,
	accountID string,
	feed marketdata.Feed,
	portfolioTracker *portfolio.Tracker,
	balanceReader strategydemo.BalanceReader,
) *strategy.Runtime {
	symbols := cfg.Risk.AllowedSymbols
	if len(symbols) == 0 {
		symbols = []string{"BTCUSDT", "ETHUSDT"}
	}

	sc := cfg.Strategy
	maxNotional := sc.MaxNotional
	if cfg.Risk.MaxNotional > 0 && cfg.Risk.MaxNotional < maxNotional {
		maxNotional = cfg.Risk.MaxNotional
	}

	var strategies []strategy.Strategy

	switch cfg.Scheduler.Mode {
	case "stream", "":
		// Default to stream for autonomous trading
		for _, symbol := range symbols {
			// ── Entry Strategy 1: EMA Crossover ──────────
			emaBase := strategydemo.NewEMATickCrossover(accountID, symbol, "0.001", sc.EMAFast, sc.EMASlow)
			emaSized := strategydemo.NewPortfolioSizer(emaBase, symbol, balanceReader, feed, sc.RiskPct, maxNotional)
			strategies = append(strategies, emaSized)

			// ── Entry Strategy 2: Bollinger Band Reversion ──
			bbBase := strategydemo.NewBollingerTickReversion(accountID, symbol, "0.001", sc.BollingerPeriod, sc.BollingerStdDev)
			bbSized := strategydemo.NewPortfolioSizer(bbBase, symbol, balanceReader, feed, sc.RiskPct, maxNotional)
			strategies = append(strategies, bbSized)

			// ── Entry Strategy 3: RSI Reversion ──────────
			rsiBase := strategydemo.NewRSIReversion(accountID, symbol, "0.001", sc.RSIPeriod, sc.RSIOversold, sc.RSIOverbought)
			rsiSized := strategydemo.NewPortfolioSizer(rsiBase, symbol, balanceReader, feed, sc.RiskPct, maxNotional)
			strategies = append(strategies, rsiSized)

			// ── Exit Strategy: Trailing Take Profit ──────
			trailingTP := strategydemo.NewTrailingTakeProfit(
				accountID, symbol, "0.001",
				sc.TrailingActivatePct,
				sc.TrailingGapPct,
				strategydemo.WithStopLossPct(sc.TrailingStopLossPct),
				strategydemo.WithMaxHoldMinutes(sc.TrailingMaxHoldMinutes),
				strategydemo.WithPortfolioEntry(portfolioTracker),
				strategydemo.WithFeed(feed),
			)
			strategies = append(strategies, trailingTP)

			// ── Entry Strategy 4: Tick Momentum Burst ────
			momentumBase := strategydemo.NewTickMomentumBurst(accountID, symbol, "0.001", 10, 0.15, 0.15)
			momentumSized := strategydemo.NewPortfolioSizer(momentumBase, symbol, balanceReader, feed, sc.RiskPct, maxNotional)
			strategies = append(strategies, momentumSized)

			// ── Entry Strategy 5: Tick Mean Reversion ────
			meanRevBase := strategydemo.NewTickMeanReversion(accountID, symbol, "0.001", 20, 0.10, 0.10)
			meanRevSized := strategydemo.NewPortfolioSizer(meanRevBase, symbol, balanceReader, feed, sc.RiskPct, maxNotional)
			strategies = append(strategies, meanRevSized)

			// ── Entry Strategy 6: Double EMA Trend ─────
			doubleEMABase := strategydemo.NewDoubleEMATrendStrategy(accountID, symbol, "0.001", 5, 13, 50)
			doubleEMASized := strategydemo.NewPortfolioSizer(doubleEMABase, symbol, balanceReader, feed, sc.RiskPct, maxNotional)
			strategies = append(strategies, doubleEMASized)

			// ── Entry Strategy 7: Tick Price Threshold ──
			// Buy when price drops below a dynamic threshold
			priceThresholdBase := strategydemo.NewTickPriceThreshold(accountID, symbol, "0.001", "60000.00")
			strategies = append(strategies, priceThresholdBase)
		}

		// ── USDT Stablecoin Scalp Strategy ──────────
		// Trades USDT fluctuations around $1.00 peg
		// Buy at 0.9992, sell at 0.9999, stop loss at 0.98
		usdtScalp := strategydemo.NewUSDTStablecoinScalp(
			accountID, "USDTUSD", "100",
			0.9992, 0.9999, 0.9800, 500.0,
		)
		strategies = append(strategies, usdtScalp)

		// ── USDC Stablecoin Scalp Strategy ──────────
		// USDC is more volatile than USDT — wider thresholds
		// Buy at 0.9985, sell at 0.9998, stop loss at 0.97
		usdcScalp := strategydemo.NewUSDTStablecoinScalp(
			accountID, "USDCUSD", "100",
			0.9985, 0.9998, 0.9700, 500.0,
		)
		strategies = append(strategies, usdcScalp)

		// ── Sentiment Engine Setup ───────────────────
		// Aggregates sentiment from multiple sources:
		// - CryptoPanic news API
		// - Fear & Greed Index
		// - Market events (halving, FOMC, ETF decisions)
		// - Stock market correlation (SPY)
		sentLogger, _ := logging.New(logging.Config{Stdout: false})
		sentimentEngine := sentimentsentiment.NewEngine(sentLogger)
		sentimentEngine.RegisterProvider(sentimentsentiment.NewFearGreedProvider(sentLogger))
		sentimentEngine.RegisterProvider(sentimentsentiment.NewMarketEventsProvider(sentLogger))
		// CryptoNews and YouTube providers need API keys — register with empty key for now
		sentimentEngine.RegisterProvider(sentimentsentiment.NewCryptoNewsProvider("", sentLogger))
		sentimentEngine.RegisterProvider(sentimentsentiment.NewStockMarketCorrelation("", sentLogger))
		sentimentEngine.RegisterProvider(sentimentsentiment.NewWhaleAlertProvider("", 500000, sentLogger))

		// ── Sentiment-Aware Strategy ─────────────────
		// Combines all sentiment sources with technical analysis
		for _, symbol := range cfg.Risk.AllowedSymbols {
			sentimentBase := strategydemo.NewSentimentAwareStrategy(accountID, symbol, "0.001", sentimentEngine, 0.2)
			sentimentSized := strategydemo.NewPortfolioSizer(sentimentBase, symbol, balanceReader, feed, sc.RiskPct, maxNotional)
			strategies = append(strategies, sentimentSized)
		}

		// ── Time-Based Cycle Strategies ──────────────
		for _, symbol := range cfg.Risk.AllowedSymbols {
			// Weekly Cycle: Buy Monday dip, sell Sunday peak
			weeklyCycleBase := strategydemo.NewWeeklyCycleStrategy(accountID, symbol, "0.001")
			weeklyCycleSized := strategydemo.NewPortfolioSizer(weeklyCycleBase, symbol, balanceReader, feed, sc.RiskPct, maxNotional)
			strategies = append(strategies, weeklyCycleSized)

			// China Session: Buy pre-Asia quiet, sell Asia volatility spike
			chinaSessionBase := strategydemo.NewChinaSessionStrategy(accountID, symbol, "0.001")
			chinaSessionSized := strategydemo.NewPortfolioSizer(chinaSessionBase, symbol, balanceReader, feed, sc.RiskPct, maxNotional)
			strategies = append(strategies, chinaSessionSized)

			// Whale Alert: Trade based on large whale movements
			whaleAlertBase := strategydemo.NewWhaleAlertStrategy(accountID, symbol, "0.001")
			whaleAlertSized := strategydemo.NewPortfolioSizer(whaleAlertBase, symbol, balanceReader, feed, sc.RiskPct, maxNotional)
			strategies = append(strategies, whaleAlertSized)
		}

	case "candle-stream":
		for _, symbol := range symbols {
			strategies = append(strategies, strategydemo.NewMACDCrossover(accountID, symbol, "0.001", 12, 26, 9))
			strategies = append(strategies, strategydemo.NewBollingerReversion(accountID, symbol, "0.001", sc.BollingerPeriod, sc.BollingerStdDev))
			strategies = append(strategies, strategydemo.NewCandleSMACross(accountID, symbol, "0.001", 5, 20))
			strategies = append(strategies, strategydemo.NewATRSizing(accountID, symbol, "0.001", 0.01, 7, 25, 14))

			trailingTP := strategydemo.NewTrailingTakeProfit(
				accountID, symbol, "0.001",
				sc.TrailingActivatePct,
				sc.TrailingGapPct,
				strategydemo.WithStopLossPct(sc.TrailingStopLossPct),
				strategydemo.WithMaxHoldMinutes(sc.TrailingMaxHoldMinutes),
				strategydemo.WithPortfolioEntry(portfolioTracker),
				strategydemo.WithFeed(feed),
			)
			strategies = append(strategies, trailingTP)
		}

	default: // timer mode
		for _, symbol := range symbols {
			strategies = append(strategies, strategydemo.NewEMACrossover(accountID, symbol, "0.001", sc.EMAFast, sc.EMASlow, feed))
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
		Type:   "app.started",
		Source: "ultratrader-go",
		Payload: map[string]any{
			"environment":  a.config.Environment,
			"accounts":     len(a.accountService.List()),
			"market_data":  a.config.MarketData.Source,
			"initial_usdt": a.config.MarketData.InitialBalance,
		},
	}); err != nil {
		return err
	}

	for _, acct := range a.accountService.List() {
		if err := a.snapshotStore.Append(ctx, snapshot.Snapshot{
			AccountID:   acct.ID,
			AccountName: acct.Name,
			Exchange:    acct.ExchangeName,
			Metadata:    map[string]any{"enabled": acct.Enabled},
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
		"market_data":    a.config.MarketData.Source,
		"strategy":       a.config.Strategy,
	}
	if err := a.reportStore.Append(ctx, reports.Report{Type: "startup-summary", Payload: reportPayload}); err != nil {
		return fmt.Errorf("append runtime report: %w", err)
	}

	a.logger.Info("app startup completed", reportPayload)
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	// Flush signal log before shutdown
	if a.signalLogStop != nil {
		a.signalLogStop()
	}
	a.logger.Info("app shutdown initiated", map[string]any{
		"signal_count":    a.signalLog.Count(),
		"strategy_stats":  a.signalLog.StatsByStrategy(),
		"portfolio_value": a.portfolioTracker.TotalMarketValue(ctx, a.marketDataFeed),
		"realized_pnl":    a.portfolioTracker.TotalRealizedPnL(),
		"unrealized_pnl":  a.portfolioTracker.TotalUnrealizedPnL(ctx, a.marketDataFeed),
		"usdt_balance":    a.marketAwarePaper.USDTBalance(),
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
