# Architectural Analysis: whittlem/pycryptobot

## Overview
PyCryptoBot is a popular Python-based trading bot that features a rich set of strategy parameters and complex risk management logic. It is particularly notable for its sophisticated sell triggers and trailing stop-loss implementations.

## Key Architectural Patterns

### 1. Complex Sell Decision Logic (`is_sell_trigger`)
- **Multiple Exit Paths:** PyCryptoBot doesn't just look for a signal reversal; it has several independent exit conditions:
  - **Trailing Stop Loss (TSL):** Dynamic or fixed percentage-based trailing.
  - **Prevent Loss:** Selling before margin drops below 0% if certain thresholds are hit.
  - **Profit Bank:** Selling when a high upper percentage profit is reached.
  - **Sell at Resistance:** Selling if the price hits a resistance level and margin is sufficient.
  - **Fibonacci Low Failsafe:** Using Fibonacci levels as dynamic stop-loss points.
- **Assimilation Strategy for Go:** Implement a `CompositeExitStrategy` that evaluates multiple exit rules in priority order.

### 2. Dynamic Trailing Stop Loss
- **Trigger and Multiplier:** TSL can be dynamic, where the trigger level and the stop-loss percentage itself can be adjusted as the margin grows.
- **Bailout Logic:** Immediate sell if the price drops by a certain percentage from the wait price.
- **Assimilation Strategy for Go:** Port the `DynamicTSL` logic to Go, allowing for parameterized multipliers on the stop distance.

### 3. Signal Filtering and Exclusion
- **Bull Only Mode:** Option to only trade when a "golden cross" (EMA crossover) is present.
- **High Price Exclusion:** Preventing buys within a certain percentage of the recent high to avoid buying the "top".
- **Assimilation Strategy for Go:** Implement these filters as `ExecutionGuards` or `SignalFilters` in the Go pipeline.

## Implementation Takeaways for UltraTrader Go
1. **DynamicTSL Strategy:** Implement a stateful trailing stop-loss that tracks the highest price and adjusts its distance dynamically.
2. **PreventLoss Logic:** Add a "safety" check that triggers a sell if the profit margin starts rapidly evaporating toward zero.
3. **Indicator Integration:** PyCryptoBot heavily relies on Pandas DataFrames for technical indicators (EMA, MACD, BBands). Our Go platform should ensure its indicator library remains consistent with these Python implementations.
