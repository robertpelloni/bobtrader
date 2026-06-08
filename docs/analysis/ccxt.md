# Architectural Analysis: ccxt/ccxt

## Overview
CCXT (CryptoCurrency eXchange Trading Library) is the industry standard for unified crypto exchange APIs. It supports over 100 exchanges and provides a consistent interface for market data, trading, and account management.

## Key Architectural Patterns

### 1. Unified API Abstraction
- **Base Class:** Every exchange class inherits from a base `Exchange` class that defines the core interface (e.g., `fetchTicker`, `createOrder`).
- **Standardized Data Formats:** CCXT normalizes heterogeneous exchange responses (JSON) into standardized internal formats for tickers, orders, trades, and balances.
- **Assimilation Strategy for Go:** Adopt a similar normalization approach in `ultratrader-go` by defining strict, comprehensive structs in `internal/exchange/types.go` that all adapters must fill.

### 2. Error Hierarchy and Mapping
- **Exception Class Tree:** CCXT uses a deep hierarchy of custom exceptions (e.g., `InsufficientFunds`, `OrderNotFound`, `RateLimitExceeded`).
- **Automatic Mapping:** Each exchange adapter defines a mapping from its specific error codes/messages to the unified CCXT exceptions.
- **Assimilation Strategy for Go:** Implement a `TypedError` system in Go that maps exchange-specific errors to common error types, allowing strategies to handle errors like `InsufficientFunds` uniformly across Binance, KuCoin, etc.

### 3. Rate Limiting and Flow Control
- **Built-in Rate Limiter:** Each exchange instance has its own rate limiter that respects the exchange's specific limits.
- **Assimilation Strategy for Go:** Strengthen the existing `internal/exchange/ratelimit` package by incorporating per-exchange limit configurations derived from CCXT's extensive metadata.

### 4. Precision and Arithmetic
- **Precise Calculation:** Uses a custom `Precise` class (or similar logic) to handle decimal arithmetic without floating-point errors, essential for crypto trading.
- **Assimilation Strategy for Go:** Use `shopspring/decimal` or fixed-point arithmetic for all quantity and price calculations.

## Implementation Takeaways for UltraTrader Go
1. **Error Mapping:** Create `internal/exchange/errors.go` with a hierarchy similar to CCXT's `errorHierarchy.ts`.
2. **Metadata-Driven Adapters:** Use metadata (fees, precision, limits) to configure adapters, moving away from hardcoded values.
3. **Comprehensive Order States:** Expand the `Order` struct to support the full range of states normalized by CCXT (e.g., `open`, `closed`, `canceled`, `expired`, `rejected`).
