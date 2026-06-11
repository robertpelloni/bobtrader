# Go Phase 57: Multi-Exchange Price Aggregation Enhancement

## Summary
Enhance the existing PriceAggregator with live multi-exchange feeds, cross-exchange arbitrage detection, and unified price API.

## Current State
- `internal/exchange/aggregator/aggregator.go` — Basic PriceAggregator with Median/VWAP/BestBidAsk/Mean methods
- `internal/strategy/demo/cross_exchange_arbitrage.go` — Strategy that detects price differences

## Enhancements Needed

### 1. Live Multi-Exchange Feed
- Register multiple exchange adapters (Binance, Coinbase, Kraken)
- Concurrent price fetching with health monitoring
- Automatic failover if one exchange goes down

### 2. Cross-Exchange Arbitrage Engine
- Real-time spread calculation between exchanges
- Fee-adjusted profit calculation
- Execution routing to cheapest exchange

### 3. Unified Price API
- Single endpoint for best price across all exchanges
- Weighted average based on volume
- Stale price detection and filtering

## Implementation
```go
// New: internal/exchange/aggregator/live_feed.go
type LiveMultiExchangeFeed struct {
    aggregator *PriceAggregator
    exchanges  map[string]exchange.Adapter
    health     map[string]bool
    mu         sync.RWMutex
}

func (f *LiveMultiExchangeFeed) GetBestPrice(ctx context.Context, symbol string) (PriceQuote, error)
func (f *LiveMultiExchangeFeed) GetArbitrageOpportunities(ctx context.Context, symbol string) []ArbitrageOpportunity
func (f *LiveMultiExchangeFeed) GetExchangeHealth() map[string]bool
```

## Tests
- `TestLiveMultiExchangeFeed_GetBestPrice`
- `TestLiveMultiExchangeFeed_ArbitrageDetection`
- `TestLiveMultiExchangeFeed_Failover`
- `TestLiveMultiExchangeFeed_HealthMonitoring`
