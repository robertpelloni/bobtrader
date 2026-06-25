package exchange

import (
	"context"
	"time"
)

type Market struct {
	Symbol        string
	BaseAsset     string
	QuoteAsset    string
	PriceScale    int
	QuantityScale int
	MinNotional   string
	Status        string // "trading", "break", "halt"
}

type Balance struct {
	Asset  string
	Free   string
	Locked string
	Total  string
}

type OrderSide string
type OrderType string
type OrderStatus string

const (
	Buy  OrderSide = "buy"
	Sell OrderSide = "sell"

	MarketOrder OrderType = "market"
	LimitOrder  OrderType = "limit"

	StatusOpen     OrderStatus = "open"
	StatusClosed   OrderStatus = "closed"
	StatusCanceled OrderStatus = "canceled"
	StatusExpired  OrderStatus = "expired"
	StatusRejected OrderStatus = "rejected"
)

type OrderRequest struct {
	Symbol   string
	Side     OrderSide
	Type     OrderType
	Quantity string
	Price    string
}

type Order struct {
	ID            string
	Symbol        string
	Side          OrderSide
	Type          OrderType
	Status        OrderStatus
	Quantity      string
	Price         string
	ExecutedQty   string
	RemainingQty  string
	Cost          string
	AveragePrice  string
	Timestamp     time.Time
	LastTradeTime *time.Time
}

type Adapter interface {
	Name() string
	Capabilities() []Capability
	ListMarkets(ctx context.Context) ([]Market, error)
	Balances(ctx context.Context) ([]Balance, error)
	PlaceOrder(ctx context.Context, request OrderRequest) (Order, error)
}

type Trade struct {
	ID              string
	OrderID         string
	Symbol          string
	Side            OrderSide
	Price           string
	Quantity        string
	QuoteQuantity   string
	Commission      string
	CommissionAsset string
	Time            time.Time
	IsMaker         bool
}

type TradeHistoryQuerier interface {
	QueryTrades(ctx context.Context, symbol string, limit int) ([]Trade, error)
}
