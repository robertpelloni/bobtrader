package execution

import (
	"context"
	"fmt"
	"strings"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/eventlog"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/risk"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

type Service struct {
	accounts *account.Service
	registry *exchange.Registry
	pipeline *risk.Pipeline
	events   *eventlog.Log
}

func NewService(accounts *account.Service, registry *exchange.Registry, pipeline *risk.Pipeline, events *eventlog.Log) *Service {
	return &Service{accounts: accounts, registry: registry, pipeline: pipeline, events: events}
}

func (s *Service) Execute(ctx context.Context, accountID string, request exchange.OrderRequest, intent risk.OrderIntent) (exchange.Order, error) {
	acct, ok := s.accounts.Get(accountID)
	if !ok {
		return exchange.Order{}, fmt.Errorf("account %q not found", accountID)
	}
	if !acct.Enabled {
		return exchange.Order{}, fmt.Errorf("account %q is disabled", accountID)
	}
	if strings.TrimSpace(request.Symbol) == "" {
		return exchange.Order{}, fmt.Errorf("symbol is required")
	}

	if err := s.pipeline.Run(ctx, acct, intent); err != nil {
		return exchange.Order{}, err
	}

	adapter, err := s.registry.Create(acct.ExchangeName)
	if err != nil {
		return exchange.Order{}, err
	}

	order, err := adapter.PlaceOrder(ctx, request)
	if err != nil {
		return exchange.Order{}, fmt.Errorf("place order: %w", err)
	}

	if s.events != nil {
		_ = s.events.Append(ctx, eventlog.Entry{
			Type:   "execution.order_placed",
			Source: "execution-service",
			Payload: map[string]any{
				"account_id": accountID,
				"exchange":   acct.ExchangeName,
				"symbol":     order.Symbol,
				"side":       order.Side,
				"type":       order.Type,
				"status":     order.Status,
				"order_id":   order.ID,
			},
		})
	}

	return order, nil
}
