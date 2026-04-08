# Go Phase-43 Order Reconciliation

## Summary
This phase adds order state verification, ensuring internal portfolio tracking matches the actual exchange state — a critical safety feature for production trading.

## Context & Motivation
In production, things can go wrong: network timeouts, partial fills, exchange errors, or race conditions can cause internal state to diverge from reality. Without reconciliation:
- The portfolio tracker might show positions that don't exist on the exchange.
- Risk guards might allow trades based on stale data.
- PnL calculations could be wrong.

BBGO and WolfBot both implement order reconciliation as a core feature. Professional trading firms consider it non-negotiable.

## Delivered

### Reconciliation Engine (`internal/trading/reconciliation/reconciler.go`)

#### Core Algorithm
1. Takes a batch of local orders.
2. For each order, queries the exchange via `OrderQuerier` interface.
3. Compares internal status against exchange status.
4. Records discrepancies (status mismatches).
5. Categorizes: matched, filled, partially-filled, canceled, rejected, expired, unknown.

#### OrderQuerier Interface
Adapters implement this to support reconciliation:
```go
type OrderQuerier interface {
    QueryOrder(ctx context.Context, symbol, orderID string) (OrderStatus, error)
}
```

#### Binance Integration
- Added `QueryOrder()` to Binance adapter via signed GET to `/api/v3/order`.
- Returns `OrderStatus` with executed quantity and transaction time.
- Rate-limited via the existing token bucket.

#### Graceful Fallback
If the adapter doesn't implement `OrderQuerier`, the reconciler falls back to counting local states without querying the exchange.

### Testing
- `TestReconcileOrders_Matched` — Validates matched state detection.
- `TestReconcileOrders_Discrepancy` — Validates mismatch detection.
- `TestReconcileOrders_NoQuerier` — Validates fallback behavior.
- `TestReconcileResult_Summary` — Validates summary formatting.

## Next Steps
1. **Periodic Reconciliation Service** — Background goroutine running reconciliation on a timer.
2. **Auto-Correction** — Automatically update internal state when discrepancies are found.
3. **Trade History Sync** — Full trade history download from exchange for audit purposes.
