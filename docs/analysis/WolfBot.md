# Architectural Analysis: Ekliptor/WolfBot

## Overview
WolfBot is a sophisticated TypeScript-based trading bot that features a vast library of trading strategies and technical indicators. It emphasizes **extensibility through inheritance and mixins** and provides a robust framework for backtesting and live trading.

## Key Architectural Patterns

### 1. Strategy Inheritance Hierarchy
- **AbstractBase:** `AbstractStrategy` defines the core lifecycle (initialization, data processing, signal emission).
- **Technical Layers:** `TechnicalStrategy` extends the base to provide easy access to indicators.
- **Mixins:** Uses mixins for shared behaviors like `AbstractTrailingStop`, `AbstractTakeProfitStrategy`, etc.
- **Assimilation Strategy for Go:** While Go lacks class-based inheritance, we can use **composition and interfaces** (e.g., the existing `Strategy` interface in `ExecutionManager`) to achieve similar modularity.

### 2. Rich Indicator Integration
- WolfBot has over 50 strategy implementations (e.g., `BollingerBands`, `Ichimoku`, `Wyckoff`).
- Strategies can easily "add" indicators which are automatically updated as new market data arrives.
- **Assimilation Strategy for Go:** Enhance the Go strategy context to allow strategies to request and consume indicators from a shared registry.

### 3. Advanced Execution Logic (The Bollinger Case)
- Unlike simple mean-reversion Bollinger strategies, WolfBot's `BollingerBands.ts` includes **breakout detection**.
- If the price stays at the upper/lower band for a configured number of candles (`breakout` parameter), it assumes the trend will continue rather than reverse.
- This adds a layer of "regime awareness" to a standard technical indicator.

## Implementation Takeaways for UltraTrader Go
1. **WolfBotBollinger:** Implement a Bollinger strategy that distinguishes between "mean reversion" (reaching a band) and "trend breakout" (staying at a band).
2. **Strategy Metadata:** WolfBot strategies expose many parameters (N, K, breakout). Our `ExecutionManager` should eventually support parameterized strategy instantiation.
3. **Signal Weighting:** WolfBot uses `defaultWeight` for signals. We should consider adding a `Confidence` or `Weight` field to our Go signals.
