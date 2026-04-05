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
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/persistence/orders"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/persistence/snapshot"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/risk"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
	strategydemo "github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/demo"
	strategyscheduler "github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy/scheduler"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/execution"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/portfolio"
)

type App struct {
	config           config.Config
	logger           *logging.Logger
	eventLog         *eventlog.Log
	snapshotStore    *snapshot.Store
	orderStore       *orders.Store
	accountService   *account.Service
	exchangeRegistry *exchange.Registry
	marketDataFeed   marketdata.Feed
	executionRepo    *execution.Repository
	portfolioTracker *portfolio.Tracker
	executionService *execution.Service
	strategyRuntime  *strategy.Runtime
	scheduler        *strategyscheduler.Scheduler
	schedulerService *strategyscheduler.Service
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
	accountService, err := account.NewService(cfg.Accounts)
	if err != nil {
		return nil, fmt.Errorf("create account service: %w", err)
	}

	registry := exchange.NewRegistry()
	if err := registry.Register("paper", func() exchange.Adapter { return exchangepaper.New() }); err != nil {
		return nil, fmt.Errorf("register paper exchange: %w", err)
	}

	marketDataFeed := marketdatapaper.New()
	executionRepo := execution.NewRepository()
	portfolioTracker := portfolio.NewTracker()
	pipeline := risk.NewPipeline(
		risk.NewSymbolWhitelistGuard(cfg.Risk.AllowedSymbols),
		risk.NewMaxNotionalGuard(cfg.Risk.MaxNotional),
		risk.NewCooldownGuard(time.Duration(cfg.Risk.CooldownMS)*time.Millisecond),
		risk.NewDuplicateSymbolGuard(executionRepo, time.Duration(cfg.Risk.DuplicateWindowMS)*time.Millisecond),
	)
	executionService := execution.NewService(accountService, registry, pipeline, eventLog, orderStore, executionRepo, portfolioTracker, logger)
	strategyRuntime := strategy.NewRuntime(strategydemo.NewPriceThreshold("paper-main", "BTCUSDT", "0.01", "70000.00", marketDataFeed))
	scheduler := strategyscheduler.New(strategyRuntime, executionService)
	schedulerService := strategyscheduler.NewService(scheduler, time.Duration(cfg.Scheduler.IntervalMS)*time.Millisecond)

	handler := httpapi.NewHandler(httpapi.Dependencies{
		StatusProvider: func() httpapi.Status {
			return httpapi.Status{Name: "ultratrader-go", Ready: true, AccountCount: len(accountService.List())}
		},
		PortfolioProvider: func() httpapi.PortfolioSnapshot {
			return httpapi.PortfolioSnapshot{
				Positions:          portfolioTracker.ValuedPositions(context.Background(), marketDataFeed),
				TotalMarketValue:   portfolioTracker.TotalMarketValue(context.Background(), marketDataFeed),
				TotalRealizedPnL:   portfolioTracker.TotalRealizedPnL(),
				TotalUnrealizedPnL: portfolioTracker.TotalUnrealizedPnL(context.Background(), marketDataFeed),
			}
		},
		OrdersProvider:           func() []exchange.Order { return executionRepo.List() },
		ExecutionSummaryProvider: func() execution.Summary { return executionRepo.Summary() },
	})

	var runtime *httpapi.Runtime
	if cfg.Server.Enabled {
		runtime = httpapi.NewRuntime(cfg.Server.Address, handler)
	}
	return &App{config: cfg, logger: logger, eventLog: eventLog, snapshotStore: snapshotStore, orderStore: orderStore, accountService: accountService, exchangeRegistry: registry, marketDataFeed: marketDataFeed, executionRepo: executionRepo, portfolioTracker: portfolioTracker, executionService: executionService, strategyRuntime: strategyRuntime, scheduler: scheduler, schedulerService: schedulerService, httpHandler: handler, httpRuntime: runtime}, nil
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
		a.logger.Info("http runtime started", map[string]any{"address": a.config.Server.Address})
	}
	if a.config.Scheduler.Enabled {
		a.schedulerService.Start(ctx)
		a.logger.Info("scheduler service started", map[string]any{"interval_ms": a.config.Scheduler.IntervalMS})
	}
	if err := a.scheduler.RunOnce(ctx); err != nil {
		return fmt.Errorf("run strategy scheduler: %w", err)
	}
	a.logger.Info("app startup completed", map[string]any{"orders": len(a.executionRepo.List()), "portfolio_value": a.portfolioTracker.TotalMarketValue(ctx, a.marketDataFeed), "realized_pnl": a.portfolioTracker.TotalRealizedPnL(), "unrealized_pnl": a.portfolioTracker.TotalUnrealizedPnL(ctx, a.marketDataFeed)})
	return nil
}

func (a *App) Handler() http.Handler { return a.httpHandler }
