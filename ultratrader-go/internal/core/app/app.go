package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/connectors/httpapi"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/eventlog"
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
)

type App struct {
	config           config.Config
	eventLog         *eventlog.Log
	snapshotStore    *snapshot.Store
	orderStore       *orders.Store
	accountService   *account.Service
	exchangeRegistry *exchange.Registry
	marketDataFeed   marketdata.Feed
	executionService *execution.Service
	strategyRuntime  *strategy.Runtime
	scheduler        *strategyscheduler.Scheduler
	httpHandler      http.Handler
	httpRuntime      *httpapi.Runtime
}

func New(cfg config.Config) (*App, error) {
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

	pipeline := risk.NewPipeline()
	executionService := execution.NewService(accountService, registry, pipeline, eventLog, orderStore)
	marketDataFeed := marketdatapaper.New()
	_ = marketDataFeed

	strategyRuntime := strategy.NewRuntime(
		strategydemo.NewBuyOnce("paper-main", "BTCUSDT", "0.01"),
	)
	scheduler := strategyscheduler.New(strategyRuntime, executionService)

	handler := httpapi.NewHandler(httpapi.Status{
		Name:         "ultratrader-go",
		Ready:        true,
		AccountCount: len(accountService.List()),
	})

	var runtime *httpapi.Runtime
	if cfg.Server.Enabled {
		runtime = httpapi.NewRuntime(cfg.Server.Address, handler)
	}

	return &App{
		config:           cfg,
		eventLog:         eventLog,
		snapshotStore:    snapshotStore,
		orderStore:       orderStore,
		accountService:   accountService,
		exchangeRegistry: registry,
		marketDataFeed:   marketDataFeed,
		executionService: executionService,
		strategyRuntime:  strategyRuntime,
		scheduler:        scheduler,
		httpHandler:      handler,
		httpRuntime:      runtime,
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	if err := a.eventLog.Append(ctx, eventlog.Entry{
		Type:   "app.started",
		Source: "ultratrader-go",
		Payload: map[string]any{
			"environment": a.config.Environment,
			"accounts":    len(a.accountService.List()),
		},
	}); err != nil {
		return err
	}

	for _, acct := range a.accountService.List() {
		if err := a.snapshotStore.Append(ctx, snapshot.Snapshot{
			AccountID:   acct.ID,
			AccountName: acct.Name,
			Exchange:    acct.ExchangeName,
			Metadata: map[string]any{
				"enabled": acct.Enabled,
			},
		}); err != nil {
			return fmt.Errorf("append bootstrap snapshot for %s: %w", acct.ID, err)
		}
	}

	if a.httpRuntime != nil {
		if err := a.httpRuntime.Start(ctx); err != nil {
			return fmt.Errorf("start http runtime: %w", err)
		}
	}

	if err := a.scheduler.RunOnce(ctx); err != nil {
		return fmt.Errorf("run strategy scheduler: %w", err)
	}
	return nil
}

func (a *App) Handler() http.Handler {
	return a.httpHandler
}
