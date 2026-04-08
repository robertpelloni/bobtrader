# Go Phase-39 Binance REST Adapter

## Summary
This phase implements a production-grade Binance REST API adapter, bridging the Go ultra-project from paper trading to live exchange connectivity.

## Context & Motivation
All prior phases built a robust internal platform: indicators, strategies, backtesting, optimization, risk management, and operator surfaces. However, the only exchange adapter was `paper` — an in-memory mock. To reach production, the platform needs a real exchange connection. Binance is the world's largest crypto exchange by volume and has a well-documented REST API, making it the natural first adapter.

Reference architectures:
- **BBGO** has extensive Binance integration with websocket streams, order management, and account sync.
- **CCXT** provides a unified exchange abstraction covering 100+ exchanges.
- **PowerTrader AI** uses `ccxt` for all exchange communication.

## Delivered

### Binance REST Client (`internal/exchange/binance/adapter.go`)

#### Public Endpoints
- **GetTickerPrice(symbol)** — Returns the latest price for a symbol via `/api/v3/ticker/price`.
- **GetKlines(symbol, interval, limit)** — Returns OHLCV candle data via `/api/v3/klines`. Parses Binance's array-of-arrays format into structured `Kline` objects.
- **ListMarkets()** — Fetches `/api/v3/exchangeInfo` and filters for status=TRADING symbols.

#### Authenticated Endpoints
- **Balances()** — Fetches spot account balances via signed `/api/v3/account` endpoint. Filters out zero-balance assets.
- **PlaceOrder()** — Places market/limit orders via signed POST to `/api/v3/order`. Supports GTC time-in-force for limit orders.

#### Security
- HMAC-SHA256 request signing using the secret key.
- `X-MBX-APIKEY` header authentication.
- Timestamp and recvWindow parameters for replay protection.
- No credentials stored in logs or error messages.

#### Configuration
- `Config.APIKey` / `Config.SecretKey` — Binance API credentials.
- `Config.Testnet` — Routes to `testnet.binance.vision` for safe testing.
- `Config.BaseURL` — Override for custom endpoints.

#### Error Handling
- Parses Binance API error responses with code and message.
- Returns structured errors for missing API keys.
- HTTP status code validation for all responses.

### Testing (`internal/exchange/binance/adapter_test.go`)
All tests use `httptest.Server` for deterministic, offline validation:
- `TestAdapter_GetTickerPrice` — Validates price parsing.
- `TestAdapter_GetKlines` — Validates OHLCV array parsing.
- `TestAdapter_ListMarkets` — Validates TRADING status filtering.
- `TestAdapter_Signing` — Validates HMAC-SHA256 determinism.
- `TestAdapter_TestnetURL` — Validates testnet URL routing.
- `TestAdapter_Balances_NoAPIKey` — Validates auth enforcement.
- `TestAdapter_PlaceOrder_NoAPIKey` — Validates auth enforcement.

### App Integration
- Registered `binance` factory in `core/app/app.go` with testnet-safe defaults.
- Added `APIKey`, `SecretKey`, `Testnet` to `AccountConfig` for per-account credentials.

## Architecture

```
internal/exchange/
├── binance/
│   ├── adapter.go       # Full REST client with HMAC signing
│   └── adapter_test.go  # httptest-based integration tests
├── paper/
│   └── adapter.go       # In-memory mock adapter
├── types.go             # Shared interfaces and types
├── capabilities.go      # Capability constants
└── registry.go          # Factory-based adapter registry
```

## Next Steps
1. **Binance WebSocket Adapter** — Real-time candle and order book streaming via websocket.
2. **Binance Market Data Feed** — Implement `marketdata.StreamFeed` using Binance websocket Kline streams.
3. **Order Status Polling** — Add order status checking and fill reconciliation.
4. **Rate Limiting** — Add Binance API rate limit compliance (1200 req/min for spot).
