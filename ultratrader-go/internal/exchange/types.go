package exchange

import "context"

type Market struct {
	Symbol        string
	BaseAsset     string
	QuoteAsset    string
	PriceScale    int
	QuantityScale int
}

type Balance struct {
	Asset  string
	Free   string
	Locked string
}

type OrderSide string

type OrderType string

const (
	Buy  OrderSide = "buy"
	Sell OrderSide = "sell"

	MarketOrder OrderType = "market"
	LimitOrder  OrderType = "limit"
)

type OrderRequest struct {
	Symbol   string
	Side     OrderSide
	Type     OrderType
	Quantity string
	Price    string
}

type Order struct {
	ID       string
	Symbol   string
	Side     OrderSide
	Type     OrderType
	Status   string
	Quantity string
	Price    string
}

type Adapter interface {
	Name() string
	Capabilities() []Capability
	ListMarkets(ctx context.Context) ([]Market, error)
	Balances(ctx context.Context) ([]Balance, error)
	PlaceOrder(ctx context.Context, request OrderRequest) (Order, error)
}
