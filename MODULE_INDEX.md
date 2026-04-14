# PowerTrader-TS / BobTrader Ultra-Project Module Index

This document maps the legacy Python modules to their modernized Go ultra-project counterparts.

## Data & Analytics
- **`pt_analytics.py` & `pt_advanced_analytics.py`**: Ported to `ultratrader-go/internal/analytics/` (Sentiment, Correlation, Journal).
- **`pt_regime_detection.py`**: Ported to `ultratrader-go/internal/strategy/regime/`.
- **`pt_volume.py`**: Ported to `ultratrader-go/internal/indicator/` (VWAP, OBV, MFI).

## Execution & Trading
- **`pt_trader.py`**: Execution logic is modularized inside `ultratrader-go/internal/trading/execution/` and `ultratrader-go/internal/trading/orders/`.
- **`pt_rebalancer.py`**: Drift calculation and wash-sale logic ported to `ultratrader-go/internal/trading/portfolio/rebalancer.go`.
- **`pt_exchanges.py` & `pt_multi_exchange.py`**: Aggregated exchange streams ported to `ultratrader-go/internal/marketdata/aggregator.go` and `ultratrader-go/internal/exchange/`.

## Artificial Intelligence & Strategy
- **`pt_nlp_strategy.py`**: Regex parsing intent engine ported to `ultratrader-go/internal/strategy/nlp/parser.go`.
- **`pt_rl_optimizer.py` & `pt_thinker.py`**: *Partially Integrated.* The `DeepThinker` AI layer requires further architectural planning to decide if it runs as a detached Python microservice or native Go embedding.
- **`pt_backtester.py`**: Completely overhauled in `ultratrader-go/internal/backtest/` with Monte Carlo, Grid Search, and Walk-Forward Optimization.

## Notification & Control
- **`pt_notifications.py`**: Ported to `ultratrader-go/internal/notification/`.
- **`pt_web_dashboard.py`**: The API routing logic and WebSocket streams are modeled in `ultratrader-go/internal/reporting/` and the `cmd/` package APIs.
