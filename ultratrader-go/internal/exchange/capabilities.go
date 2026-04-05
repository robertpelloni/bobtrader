package exchange

type Capability string

const (
	CapabilitySpot      Capability = "spot"
	CapabilityMargin    Capability = "margin"
	CapabilityFutures   Capability = "futures"
	CapabilityPaper     Capability = "paper"
	CapabilityBalances  Capability = "balances"
	CapabilityOrders    Capability = "orders"
	CapabilityTrades    Capability = "trades"
	CapabilityTickers   Capability = "tickers"
	CapabilityCandles   Capability = "candles"
	CapabilityOrderBook Capability = "orderbook"
)
