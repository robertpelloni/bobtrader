package orders

import (
	"fmt"
	"sync"
	"time"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/exchange"
)

// OrderTypeExtended extends the basic order types with advanced variants.
type OrderTypeExtended string

const (
	StopLoss     OrderTypeExtended = "stop_loss"
	TakeProfit   OrderTypeExtended = "take_profit"
	TrailingStop OrderTypeExtended = "trailing_stop"
	StopLimit    OrderTypeExtended = "stop_limit"
)

// ConditionalOrder represents an order that triggers based on price conditions.
type ConditionalOrder struct {
	ID           string             `json:"id"`
	Symbol       string             `json:"symbol"`
	Side         exchange.OrderSide `json:"side"`
	Type         OrderTypeExtended  `json:"type"`
	Quantity     string             `json:"quantity"`
	TriggerPrice float64            `json:"trigger_price"` // Price that activates the order
	LimitPrice   float64            `json:"limit_price"`   // Limit price for stop-limit orders
	TrailPercent float64            `json:"trail_percent"` // Trail percentage for trailing stops
	HighestPrice float64            `json:"highest_price"` // Track highest seen price (trailing stop)
	LowestPrice  float64            `json:"lowest_price"`  // Track lowest seen price
	ParentID     string             `json:"parent_id"`     // Parent order for OCO groups
	GroupID      string             `json:"group_id"`      // OCO group identifier
	Active       bool               `json:"active"`
	Triggered    bool               `json:"triggered"`
	CreatedAt    time.Time          `json:"created_at"`
	TriggeredAt  *time.Time         `json:"triggered_at,omitempty"`
}

// ShouldTrigger checks if a price update should trigger this conditional order.
func (co *ConditionalOrder) ShouldTrigger(currentPrice float64) bool {
	if !co.Active || co.Triggered {
		return false
	}

	switch co.Type {
	case StopLoss:
		// Stop loss triggers when price drops below trigger (for buy positions)
		if co.Side == exchange.Buy {
			return currentPrice <= co.TriggerPrice
		}
		return currentPrice >= co.TriggerPrice

	case TakeProfit:
		// Take profit triggers when price rises above trigger (for buy positions)
		if co.Side == exchange.Buy {
			return currentPrice >= co.TriggerPrice
		}
		return currentPrice <= co.TriggerPrice

	case TrailingStop:
		// Update tracking price and check if trail is hit
		if co.Side == exchange.Buy {
			if currentPrice > co.HighestPrice {
				co.HighestPrice = currentPrice
			}
			trailTrigger := co.HighestPrice * (1 - co.TrailPercent/100)
			return currentPrice <= trailTrigger
		}
		if currentPrice < co.LowestPrice || co.LowestPrice == 0 {
			co.LowestPrice = currentPrice
		}
		trailTrigger := co.LowestPrice * (1 + co.TrailPercent/100)
		return currentPrice >= trailTrigger

	case StopLimit:
		return currentPrice >= co.TriggerPrice

	default:
		return false
	}
}

// Trigger marks the order as triggered.
func (co *ConditionalOrder) Trigger() {
	co.Triggered = true
	co.Active = false
	now := time.Now().UTC()
	co.TriggeredAt = &now
}

// Manager manages conditional and advanced orders.
type Manager struct {
	mu        sync.RWMutex
	orders    map[string]*ConditionalOrder // ID -> order
	ocoGroups map[string][]string          // groupID -> order IDs
}

// NewManager creates a new order manager.
func NewManager() *Manager {
	return &Manager{
		orders:    make(map[string]*ConditionalOrder),
		ocoGroups: make(map[string][]string),
	}
}

// PlaceConditional adds a conditional order.
func (m *Manager) PlaceConditional(order ConditionalOrder) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	if order.ID == "" {
		order.ID = fmt.Sprintf("CO-%d", len(m.orders)+1)
	}
	if order.CreatedAt.IsZero() {
		order.CreatedAt = time.Now().UTC()
	}
	order.Active = true

	// Initialize tracking prices for trailing stops
	if order.Type == TrailingStop {
		order.HighestPrice = order.TriggerPrice
		order.LowestPrice = order.TriggerPrice
	}

	m.orders[order.ID] = &order

	// Track OCO group
	if order.GroupID != "" {
		m.ocoGroups[order.GroupID] = append(m.ocoGroups[order.GroupID], order.ID)
	}

	return order.ID
}

// PlaceBracketOrder creates a pair of stop-loss + take-profit orders.
func (m *Manager) PlaceBracketOrder(symbol string, side exchange.OrderSide, quantity string, stopLoss, takeProfit float64) (stopID, tpID string) {
	groupID := fmt.Sprintf("BRACKET-%d", time.Now().UnixNano())

	stopID = m.PlaceConditional(ConditionalOrder{
		Symbol:       symbol,
		Side:         oppositeSide(side),
		Type:         StopLoss,
		Quantity:     quantity,
		TriggerPrice: stopLoss,
		GroupID:      groupID,
	})

	tpID = m.PlaceConditional(ConditionalOrder{
		Symbol:       symbol,
		Side:         oppositeSide(side),
		Type:         TakeProfit,
		Quantity:     quantity,
		TriggerPrice: takeProfit,
		GroupID:      groupID,
	})

	return stopID, tpID
}

// PlaceTrailingStop creates a trailing stop order.
func (m *Manager) PlaceTrailingStop(symbol string, side exchange.OrderSide, quantity string, trailPercent float64, currentPrice float64) string {
	return m.PlaceConditional(ConditionalOrder{
		Symbol:       symbol,
		Side:         side,
		Type:         TrailingStop,
		Quantity:     quantity,
		TriggerPrice: currentPrice,
		TrailPercent: trailPercent,
		HighestPrice: currentPrice,
		LowestPrice:  currentPrice,
	})
}

// CheckPrice evaluates all active orders against a price update.
// Returns the IDs of triggered orders and their corresponding market orders.
func (m *Manager) CheckPrice(symbol string, currentPrice float64) []ConditionalOrder {
	m.mu.Lock()
	defer m.mu.Unlock()

	var triggered []ConditionalOrder

	for _, order := range m.orders {
		if order.Symbol != symbol || !order.Active {
			continue
		}
		if order.ShouldTrigger(currentPrice) {
			order.Trigger()
			triggered = append(triggered, *order)

			// Handle OCO: cancel other orders in the same group
			if order.GroupID != "" {
				m.cancelGroup(order.GroupID, order.ID)
			}
		}
	}

	return triggered
}

// Cancel removes a conditional order.
func (m *Manager) Cancel(orderID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	order, ok := m.orders[orderID]
	if !ok || !order.Active {
		return false
	}
	order.Active = false
	return true
}

// Get returns a conditional order by ID.
func (m *Manager) Get(orderID string) (ConditionalOrder, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	order, ok := m.orders[orderID]
	if !ok {
		return ConditionalOrder{}, false
	}
	return *order, true
}

// ActiveOrders returns all active conditional orders.
func (m *Manager) ActiveOrders() []ConditionalOrder {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []ConditionalOrder
	for _, order := range m.orders {
		if order.Active {
			result = append(result, *order)
		}
	}
	return result
}

// cancelGroup cancels all orders in an OCO group except the triggering order.
func (m *Manager) cancelGroup(groupID, triggerID string) {
	ids, ok := m.ocoGroups[groupID]
	if !ok {
		return
	}
	for _, id := range ids {
		if id != triggerID {
			if order, ok := m.orders[id]; ok {
				order.Active = false
			}
		}
	}
}

func oppositeSide(side exchange.OrderSide) exchange.OrderSide {
	if side == exchange.Buy {
		return exchange.Sell
	}
	return exchange.Buy
}
