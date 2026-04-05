package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/connectors/httpapi"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/eventlog"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange/paper"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/persistence/snapshot"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/risk"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/strategy"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/execution"
)

type App struct {
	config           config.Config
	eventLog         *eventlog.Log
	snapshotStore    *snapshot.Store
	accountService   *account.Service
	exchangeRegistry *exchange.Registry
	executionService *execution.Service
	strategyRuntime  *strategy.Runtime
	httpHandler      http.Handler
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

	accountService, err := account.NewService(cfg.Accounts)
	if err != nil {
		return nil, fmt.Errorf("create account service: %w", err)
	}

	registry := exchange.NewRegistry()
	if err := registry.Register("paper", func() exchange.Adapter { return paper.New() }); err != nil {
		return nil, fmt.Errorf("register paper exchange: %w", err)
	}

	pipeline := risk.NewPipeline()
	executionService := execution.NewService(accountService, registry, pipeline, eventLog)
	strategyRuntime := strategy.NewRuntime()

	handler := httpapi.NewHandler(httpapi.Status{
		Name:         "ultratrader-go",
		Ready:        true,
		AccountCount: len(accountService.List()),
	})

	return &App{
		config:           cfg,
		eventLog:         eventLog,
		snapshotStore:    snapshotStore,
		accountService:   accountService,
		exchangeRegistry: registry,
		executionService: executionService,
		strategyRuntime:  strategyRuntime,
		httpHandler:      handler,
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

	_, _ = a.strategyRuntime.Tick(ctx)
	return nil
}

func (a *App) Handler() http.Handler {
	return a.httpHandler
}
