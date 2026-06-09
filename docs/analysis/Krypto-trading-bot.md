# Architectural Analysis: ctubio/Krypto-trading-bot

## Overview
Krypto-trading-bot (also known as 'K') is a high-frequency market-making bot written primarily in C++. It is renowned for its low-latency performance and comprehensive web interface. It supports multiple quoting styles (e.g., PingPong, Boomerang, AK-47) and features a modular C++ backend.

## Key Architectural Patterns

### 1. Ultra-Low Latency C++ Core
- **Performance First:** Direct use of `libcurl` and `OpenSSL` for exchange communication.
- **Header-Only/Modular Design:** Uses large header files (e.g., `Krypto.ninja-bots.h`) to organize logic while maintaining performance.
- **Assimilation Strategy for Go:** While we use Go, we should prioritize **zero-allocation paths** and efficient JSON handling to mimic the low-latency behavior of K.

### 2. Sophisticated Quoting Styles
- **PingPong:** Alternating buy and sell orders.
- **Boomerang:** Selling high and immediately placing a buy order lower (or vice-versa) to "snap back".
- **AK-47:** Placing multiple orders in a staggered grid.
- **Assimilation Strategy for Go:** Implement a `MarketMaker` strategy in Go that supports these specific quoting patterns.

### 3. Safety and Protection
- **Quote Protection:** Uses STDEV and EWMA (Exponential Weighted Moving Average) to adjust quotes based on market volatility.
- **Trend Safety:** Includes logic to stop quoting or adjust spreads when a strong trend is detected (HamelinRat quoting mode).
- **Assimilation Strategy for Go:** Port the EWMA and STDEV calculation logic to our Go technical indicators.

### 4. Persistence with SQLite
- Uses SQLite in WAL (Write-Ahead Logging) mode for persistent storage of trades and configuration.
- **Assimilation Strategy for Go:** Our current Go project already uses SQLite; we should ensure WAL mode is enabled for maximum performance under HFT loads.

## Implementation Takeaways for UltraTrader Go
1. **Quoting Logic:** Implement the "PingPong" logic as a first-class strategy in our Go `ExecutionManager`.
2. **Volatility-Aware Spreads:** Use EWMA of price returns to dynamically widen or narrow spreads.
3. **High-Frequency Dispatcher:** Optimize the Go `ExecutionManager` to handle rapid order placement and cancellation cycles.
