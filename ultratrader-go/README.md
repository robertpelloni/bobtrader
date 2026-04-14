# UltraTrader Go

This directory contains the modernized Go port of the PowerTrader AI / BobTrader project.
The ultimate vision is a consolidated, performant, autonomous trading system merging features
from the legacy Python modules and best-in-class open source crypto bots.

## Architecture

The Go architecture is highly modular:
- **`internal/core`**: Application lifecycle, config, logging, eventlog.
- **`internal/marketdata`**: Abstractions for market data (Binance, Paper), and feed aggregation.
- **`internal/trading`**: Portfolio tracking (Tracker, Rebalancer), risk guard pipelines, account management, and execution routing.
- **`internal/strategy`**: Composition, regime detection, position sizing, scheduling, and NLP strategy parsing.
- **`internal/backtest`**: High-performance historical simulation (Walk-Forward Optimization, Grid Search, Monte Carlo, Multi-Symbol Sync).
- **`internal/analytics`**: Sentiment engines, correlation, journaling.
- **`internal/risk`**: Pluggable guard pipelines to validate order intent before execution.

## Legacy System vs Go Ultra-Project

The original Python modules (`pt_*.py`) in the root directory remain as reference architectures.
The goal is to methodically port every feature (analytics, marketplace, multi-exchange integration)
into this strictly typed Go tree.
