package execution

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/eventlog"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/logging"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/utils"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/metrics"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/notifications"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/persistence/orders"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/risk"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/account"
	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/trading/portfolio"
)

type Service struct {
	accounts   *account.Service
	registry   *exchange.Registry
	pipeline   *risk.Pipeline
	events     *eventlog.Log
	orders     *orders.Store
	repository *Repository
	portfolio  *portfolio.Tracker
	siphoning  *SiphoningManager
	logger        *logging.Logger
	metrics       *metrics.Tracker
	notifications *notifications.Manager
}

func NewService(accounts *account.Service, registry *exchange.Registry, pipeline *risk.Pipeline, events *eventlog.Log, orderStore *orders.Store, repository *Repository, portfolioTracker *portfolio.Tracker, logger *logging.Logger, metricsTracker *metrics.Tracker) *Service {
	if logger == nil {
		logger, _ = logging.New(logging.Config{Stdout: true})
	}
	if metricsTracker == nil {
		metricsTracker = metrics.NewTracker()
	}
	return &Service{accounts: accounts, registry: registry, pipeline: pipeline, events: events, orders: orderStore, repository: repository, portfolio: portfolioTracker, logger: logger, metrics: metricsTracker}
}

func (s *Service) SetSiphoningManager(m *SiphoningManager) {
	s.siphoning = m
}

func (s *Service) SetNotificationManager(m *notifications.Manager) {
	s.notifications = m
}

func (s *Service) Execute(ctx context.Context, accountID string, request exchange.OrderRequest, intent risk.OrderIntent) (exchange.Order, error) {
	ctx, correlationID := logging.NewCorrelationContext(ctx, "exec")
	log := s.logger.WithContext(ctx)
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

	if s.metrics != nil {
		s.metrics.RecordAttempt()
	}
	log.Info("execution requested", map[string]any{"account_id": accountID, "symbol": request.Symbol, "side": request.Side, "type": request.Type})
	if err := s.pipeline.Run(ctx, acct, intent); err != nil {
		var guardErr risk.GuardError
		reason := "unknown"
		if errors.As(err, &guardErr) {
			reason = guardErr.GuardName
		}
		if s.metrics != nil {
			s.metrics.RecordBlocked(reason)
		}
		log.Error("execution blocked by guard", map[string]any{"account_id": accountID, "symbol": request.Symbol, "error": err.Error(), "guard": reason})
		return exchange.Order{}, err
	}

	adapter, err := s.registry.CreateForAccount(acct.ExchangeName, acct.APIKey, acct.SecretKey, acct.Testnet)
	if err != nil {
		return exchange.Order{}, err
	}
	order, err := adapter.PlaceOrder(ctx, request)
	if err != nil {
		return exchange.Order{}, fmt.Errorf("place order: %w", err)
	}

	if s.repository != nil {
		s.repository.Save(order)
	}
	if s.portfolio != nil {
		// Calculate realized PnL before applying sell order if we have a siphoning manager
		var realizedPnL float64
		if order.Side == exchange.Sell && s.siphoning != nil {
			avgEntry := s.portfolio.AverageEntryPrice(order.Symbol)
			qty := utils.ParseFloat(order.Quantity)
			price := utils.ParseFloat(order.Price)
			if avgEntry > 0 && price > 0 {
				realizedPnL = (price - avgEntry) * qty
			}
		}

		s.portfolio.Apply(order)

		// Trigger siphoning if we just realized a profit
		if realizedPnL > 0 && s.siphoning != nil {
			go func() {
				// Use a fresh context or background context for siphoning to avoid
				// cancelling the macro trade if the micro trade context is short-lived.
				s.siphoning.OnTradeExit(context.Background(), order.Symbol, realizedPnL)
			}()
		}
	}
	if s.orders != nil {
		if err := s.orders.Append(ctx, orders.Record{AccountID: acct.ID, Exchange: acct.ExchangeName, OrderID: order.ID, Symbol: order.Symbol, Side: string(order.Side), Type: string(order.Type), Status: string(order.Status), Quantity: order.Quantity, Price: order.Price, Metadata: map[string]any{"account_name": acct.Name, "correlation_id": correlationID}}); err != nil {
			return exchange.Order{}, fmt.Errorf("append order record: %w", err)
		}
	}

	if s.notifications != nil {
		s.notifications.Notify(ctx, notifications.Notification{
			Level:   notifications.Trade,
			Source:  "ExecutionService",
			Message: fmt.Sprintf("Order executed: %s %s %s @ %s", order.Side, order.Quantity, order.Symbol, order.Price),
		})
	}
	if s.events != nil {
		_ = s.events.Append(ctx, eventlog.Entry{Type: "execution.order_placed", Source: "execution-service", Payload: map[string]any{"account_id": accountID, "exchange": acct.ExchangeName, "symbol": order.Symbol, "side": order.Side, "type": order.Type, "status": string(order.Status), "order_id": order.ID, "correlation_id": correlationID}})
	}
	if s.metrics != nil {
		s.metrics.RecordSuccess()
	}
	log.Info("execution completed", map[string]any{"account_id": accountID, "order_id": order.ID, "symbol": order.Symbol, "status": string(order.Status)})
	return order, nil
}

func (s *Service) MetricsSnapshot() metrics.Snapshot {
	if s.metrics == nil {
		return metrics.Snapshot{}
	}
	return s.metrics.Snapshot()
}
