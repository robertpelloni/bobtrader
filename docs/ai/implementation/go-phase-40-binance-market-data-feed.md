# Go Phase-40 Binance Market Data Feed

## Summary
This phase implements a complete `marketdata.StreamFeed` adapter for Binance, enabling the Go ultra-project to consume live market data from Binance's REST API using the same interfaces as the paper feed.

## Delivered

### Feed Implementation (`internal/marketdata/binance/feed.go`)

Implements all 5 methods of `marketdata.StreamFeed`:
1. **LatestTick(symbol)** — Returns a `marketdata.Tick` with the latest price from Binance.
2. **LatestCandle(symbol, interval)** — Returns a `marketdata.Candle` from the most recent kline.
3. **SubscribeTicks(symbol, interval)** — Background goroutine polling Binance at configurable intervals, emitting ticks to a buffered channel.
4. **SubscribeCandles(symbol, interval)** — Background goroutine polling klines at the natural interval duration.
5. **Feed/LatestTick/LatestCandle** — Direct passthrough to the Binance REST adapter.

### Interval Parser
Supports Binance interval strings: `1s`, `1m`, `5m`, `15m`, `1h`, `4h`, `1d`, `1w`.

### Testing
All tests use `httptest.Server` for deterministic, offline validation:
- `TestFeed_LatestTick` — Validates price parsing and source attribution.
- `TestFeed_LatestCandle` — Validates OHLCV kline-to-candle conversion.
- `TestFeed_SubscribeTicks` — Validates async tick streaming with 50ms polling.
- `TestCandleIntervalToDuration` — Validates interval string parsing.

## Architecture

```
internal/marketdata/
├── binance/
│   ├── feed.go       # StreamFeed implementation using Binance REST
│   └── feed_test.go  # httptest-based validation
├── paper/
│   ├── feed.go       # Mock feed for testing
│   └── feed_test.go
├── feed.go           # StreamFeed, CandleSubscription interfaces
└── types.go          # Tick, Candle structs
```

The Binance feed is a **drop-in replacement** for the paper feed. Any strategy, scheduler, or service that depends on `marketdata.StreamFeed` can switch from paper to live Binance data by simply changing the feed instance in config.

## Next Steps
1. **Binance WebSocket** — Replace REST polling with websocket streams for sub-second latency.
2. **Order Book Streaming** — Real-time bid/ask data for spread analysis.
3. **Rate Limiter** — Add Binance API rate limit compliance (1200 req/min for spot).
