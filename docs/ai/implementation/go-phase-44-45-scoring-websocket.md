# Go Phase-44 Risk-Adjusted Scoring Functions

## Summary
Expanded the optimizer's evaluation capabilities beyond raw PnL with Sharpe Ratio, Profit Factor, Win Rate, and Composite scoring functions.

## Delivered
- **SharpeScorer**: Risk-adjusted return metric penalizing high variance.
- **ProfitFactorScorer**: Returns relative to capital deployed.
- **WinRateScorer**: Proportion of winning trades from buy/sell pair analysis.
- **CompositeScorer**: Fluent builder for weighted combination of multiple scorers.

## Phase-45 Binance WebSocket Feed

## Summary
Implemented a zero-dependency WebSocket client for Binance real-time market data streams.

## Delivered
- Pure-Go WebSocket client (no external dependencies).
- Real-time ticker and kline subscriptions.
- Automatic reconnection with exponential backoff.
- Ping/pong keep-alive handling.
- Testnet URL routing.
- Complete frame parser supporting all WebSocket opcodes.

## Architecture
```
internal/marketdata/binance/
├── feed.go          # REST-based StreamFeed (polling)
├── feed_test.go     # REST feed tests
├── ws_feed.go       # WebSocket StreamFeed (real-time)
└── ws_feed_test.go  # WebSocket parsing tests
```

The WebSocket feed is a drop-in replacement for the REST feed, providing sub-second latency instead of polling intervals.
