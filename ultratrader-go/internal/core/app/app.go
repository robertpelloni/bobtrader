package app

import (
	"context"
	"fmt"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/eventlog"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

type App struct {
	config         config.Config
	eventLog       *eventlog.Log
	accountService *account.Service
}

func New(cfg config.Config) (*App, error) {
	eventLog, err := eventlog.New(cfg.EventLog.Path)
	if err != nil {
		return nil, fmt.Errorf("create event log: %w", err)
	}

	accountService, err := account.NewService(cfg.Accounts)
	if err != nil {
		return nil, fmt.Errorf("create account service: %w", err)
	}

	return &App{
		config:         cfg,
		eventLog:       eventLog,
		accountService: accountService,
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	return a.eventLog.Append(ctx, eventlog.Entry{
		Type:   "app.started",
		Source: "ultratrader-go",
		Payload: map[string]any{
			"environment": a.config.Environment,
			"accounts":    len(a.accountService.List()),
		},
	})
}
