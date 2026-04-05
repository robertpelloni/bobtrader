package execution

import (
	"context"
	"fmt"
	"strings"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/eventlog"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/persistence/orders"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/risk"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
)

type Service struct {
	accounts *account.Service
	registry *exchange.Registry
	pipeline *risk.Pipeline
	events   *eventlog.Log
	orders   *orders.Store
}

func NewService(accounts *account.Service, registry *exchange.Registry, pipeline *risk.Pipeline, events *eventlog.Log, orderStore *orders.Store) *Service {
	return &Service{accounts: accounts, registry: registry, pipeline: pipeline, events: events, orders: orderStore}
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

	if s.orders != nil {
		if err := s.orders.Append(ctx, orders.Record{
			AccountID: acct.ID,
			Exchange:  acct.ExchangeName,
			OrderID:   order.ID,
			Symbol:    order.Symbol,
			Side:      string(order.Side),
			Type:      string(order.Type),
			Status:    order.Status,
			Quantity:  order.Quantity,
			Price:     order.Price,
			Metadata: map[string]any{
				"account_name": acct.Name,
			},
		}); err != nil {
			return exchange.Order{}, fmt.Errorf("append order record: %w", err)
		}
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
