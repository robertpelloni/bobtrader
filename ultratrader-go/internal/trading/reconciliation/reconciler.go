package reconciliation

import (
	"context"
	"fmt"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

// OrderStatus represents the current state of an order on the exchange.
type OrderStatus struct {
	ID              string
	Symbol          string
	Side            exchange.OrderSide
	Type            exchange.OrderType
	Status          string // NEW, PARTIALLY_FILLED, FILLED, CANCELED, REJECTED, EXPIRED
	Quantity        string
	ExecutedQty     string
	Price           string
	TransactionTime time.Time
}

// Reconciler verifies and updates internal order state against the exchange.
type Reconciler struct {
	adapter exchange.Adapter
}

// NewReconciler creates a new order reconciler for the given exchange adapter.
func NewReconciler(adapter exchange.Adapter) *Reconciler {
	return &Reconciler{adapter: adapter}
}

// ReconcileResult holds the outcome of a reconciliation pass.
type ReconcileResult struct {
	TotalChecked    int
	Matched         int
	Filled          int
	PartiallyFilled int
	Canceled        int
	Rejected        int
	Expired         int
	Unknown         int
	Discrepancies   []Discrepancy
}

// Discrepancy represents a mismatch between internal and exchange state.
type Discrepancy struct {
	OrderID        string
	InternalStatus string
	ExchangeStatus string
	Description    string
}

// OrderQuerier is an optional interface that adapters can implement to query
// individual order status. If not implemented, the reconciler falls back to
// a simple consistency check.
type OrderQuerier interface {
	QueryOrder(ctx context.Context, symbol, orderID string) (OrderStatus, error)
}

// ReconcileOrders checks a batch of local orders against the exchange.
func (r *Reconciler) ReconcileOrders(ctx context.Context, localOrders []exchange.Order) (*ReconcileResult, error) {
	result := &ReconcileResult{
		TotalChecked: len(localOrders),
	}

	querier, hasQuerier := r.adapter.(OrderQuerier)
	if !hasQuerier {
		// Without a query interface, just count local states
		for _, o := range localOrders {
			switch o.Status {
			case "filled":
				result.Filled++
			case "canceled":
				result.Canceled++
			default:
				result.Unknown++
			}
		}
		return result, nil
	}

	for _, local := range localOrders {
		select {
		case <-ctx.Done():
			return result, ctx.Err()
		default:
		}

		remote, err := querier.QueryOrder(ctx, local.Symbol, local.ID)
		if err != nil {
			result.Unknown++
			result.Discrepancies = append(result.Discrepancies, Discrepancy{
				OrderID:        local.ID,
				InternalStatus: local.Status,
				ExchangeStatus: "unknown",
				Description:    fmt.Sprintf("query error: %v", err),
			})
			continue
		}

		// Compare statuses
		if normalizeStatus(local.Status) == normalizeStatus(remote.Status) {
			result.Matched++
		} else {
			result.Discrepancies = append(result.Discrepancies, Discrepancy{
				OrderID:        local.ID,
				InternalStatus: local.Status,
				ExchangeStatus: remote.Status,
				Description:    fmt.Sprintf("status mismatch: local=%s exchange=%s", local.Status, remote.Status),
			})
		}

		switch normalizeStatus(remote.Status) {
		case "filled":
			result.Filled++
		case "partially_filled":
			result.PartiallyFilled++
		case "canceled":
			result.Canceled++
		case "rejected":
			result.Rejected++
		case "expired":
			result.Expired++
		default:
			result.Unknown++
		}
	}

	return result, nil
}

// Summary returns a human-readable summary of the reconciliation result.
func (r *ReconcileResult) Summary() string {
	return fmt.Sprintf("checked=%d matched=%d filled=%d partial=%d canceled=%d rejected=%d expired=%d discrepancies=%d",
		r.TotalChecked, r.Matched, r.Filled, r.PartiallyFilled, r.Canceled, r.Rejected, r.Expired, len(r.Discrepancies))
}

func normalizeStatus(status string) string {
	switch status {
	case "FILLED", "filled":
		return "filled"
	case "PARTIALLY_FILLED", "partially_filled":
		return "partially_filled"
	case "CANCELED", "canceled":
		return "canceled"
	case "REJECTED", "rejected":
		return "rejected"
	case "EXPIRED", "expired":
		return "expired"
	case "NEW", "new":
		return "new"
	default:
		return status
	}
}
