# Go Phase-42 Rate Limiting

## Summary
This phase adds production-grade rate limiting to the Binance adapter, preventing API ban risk and ensuring smooth operation under load.

## Context & Motivation
Binance enforces strict rate limits:
- **Spot API**: 1200 request weight per minute
- **Order endpoints**: 50 orders per 10 seconds, 160,000 per 24 hours
- **IP bans**: Triggered when limits are exceeded, lasting from 2 minutes to 3 days

Without rate limiting, the platform could get banned during rapid trading or optimization loops. WolfBot and BBGO both implement rate limiting as a core safety feature.

## Delivered

### Token Bucket Rate Limiter (`internal/exchange/ratelimit/limiter.go`)
- **Token bucket algorithm**: Maintains a pool of tokens that refill at a configurable rate.
- **`Wait(ctx)`**: Blocks until a token is available (context-aware, supports cancellation).
- **`TryAcquire()`**: Non-blocking acquisition attempt, returns immediately.
- **`Remaining()`**: Returns current token count.
- **Thread-safe**: Uses `sync.Mutex` for concurrent access.

### Preset Configurations
- **`BinanceSpotLimiter()`**: 1000 tokens, 60ms refill (~16.7 req/sec, conservative below Binance's 1200/min limit).
- **`BinanceOrderLimiter()`**: 50 tokens, 200ms refill (matches Binance's 50 orders/10s limit).

### Binance Integration
- All HTTP requests pass through `limiter.Wait()` before execution.
- Order placement additionally passes through `orderLimit.Wait()`.
- Context cancellation propagates cleanly through rate limiting.

### Testing
- `TestLimiter_AllowsUpToMax` — Validates capacity enforcement.
- `TestLimiter_Refills` — Validates token refill after consumption.
- `TestLimiter_WaitBlocks` — Validates blocking behavior.
- `TestLimiter_WaitContextCancellation` — Validates context cancellation.
- `TestLimiter_ConcurrentSafety` — 200 goroutines competing for 100 tokens.
- `TestBinanceSpotLimiter` / `TestBinanceOrderLimiter` — Preset validation.

## Architecture

```
internal/exchange/
├── ratelimit/
│   ├── limiter.go       # Token bucket rate limiter
│   └── limiter_test.go  # Comprehensive test suite
├── binance/
│   └── adapter.go       # Integrated rate limiting on all requests
├── paper/
│   └── adapter.go       # No rate limiting needed for mock
└── types.go
```

## Next Steps
1. **Binance WebSocket** — Replace REST polling for real-time streaming (reduces API load).
2. **Order Reconciliation** — Fill status checking and trade history sync.
3. **Circuit Breaker** — Automatic pause on consecutive API errors.
