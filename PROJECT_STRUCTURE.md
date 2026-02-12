# PowerTrader AI - Project Structure & Modules

**Version:** 3.2.0
**Last Updated:** 2026-02-10

This document outlines the directory structure, modules, and their current versions.

---

## Directory Structure

```
PowerTrader_AI/
├── hub_data/                  # Runtime data (databases, logs, config)
├── powertrader-ts/            # TypeScript Web Architecture (The Future)
│   ├── backend/               # Node.js + Express + TypeScript
│   │   ├── src/
│   │   │   ├── analytics/     # AnalyticsManager (SQLite)
│   │   │   ├── api/           # Express Server routes
│   │   │   ├── config/        # ConfigManager (YAML)
│   │   │   ├── defi/          # LiquidityManager (Uniswap V3)
│   │   │   ├── engine/        # Core Interfaces, StrategyFactory, BacktestEngine
│   │   │   ├── exchanges/     # Exchange Connectors (Robinhood, KuCoin, Binance)
│   │   │   ├── extensions/    # PaperTrading, HyperOpt
│   │   │   ├── modules/       # Submodules (Cointrade)
│   │   │   ├── notifications/ # NotificationManager (Email, Discord, Telegram)
│   │   │   ├── thinker/       # AI Engine (kNN)
│   │   │   └── trader/        # Execution Engine (DCA, Trail)
│   └── frontend/              # React + Vite
│       ├── src/
│       │   ├── components/    # Reusable UI components
│       │   └── pages/         # Dashboard, Settings, Volume, Risk
│
├── pt_*.py                    # Legacy Python Desktop App (Tkinter)
├── MANUAL.md                  # User Manual
├── VISION.md                  # Project Vision & Design
├── AGENTS.md                  # AI Agent Instructions
└── VERSION.md                 # Single Source of Truth for Version
```

---

## Module Inventory

### TypeScript Backend (`powertrader-ts/backend`)

| Module | Status | Description |
| :--- | :--- | :--- |
| `config` | **Production** | Centralized YAML config management. |
| `trader` | **Production** | Core DCA and Trailing Stop logic. |
| `thinker` | **Production** | Hybrid Engine: `Thinker` (kNN) and `DeepThinker` (LSTM/TensorFlow). |
| `analytics` | **Production** | SQLite trade logging and performance metrics. |
| `exchanges` | **Production** | Robinhood, KuCoin, Binance, Coinbase, Uniswap (DeFi). |
| `defi` | **Production** | `LiquidityManager` for Uniswap V3 position management. |
| `extensions` | **Production** | PaperTrading Engine and HyperOpt Genetic Optimizer. |
| `notifications` | **Production** | Multi-channel alert system. |
| `engine` | **Production** | StrategyFactory and BacktestEngine. |

### TypeScript Frontend (`powertrader-ts/frontend`)

| Page | Status | Description |
| :--- | :--- | :--- |
| `Dashboard` | **Beta** | Real-time PnL and active trades view. |
| `Settings` | **Beta** | Configuration editor. |
| `Volume` | **Beta** | Volume Profile analysis visualization. |
| `Risk` | **Beta** | Correlation matrix and position sizing. |
| `Liquidity` | **New** | Uniswap V3 position manager and fee collector. |

### Legacy Python Core (`Root`)

| File | Status | Description |
| :--- | :--- | :--- |
| `pt_hub.py` | **Maintenance** | Tkinter GUI and orchestration. |
| `pt_trader.py` | **Maintenance** | Python implementation of trading logic. |
| `pt_thinker.py` | **Maintenance** | Python implementation of AI logic. |
| `pt_volume.py` | **Production** | Volume analysis logic. |
| `pt_correlation.py` | **Production** | Risk analysis logic. |

---

## Versioning Strategy

*   **Source of Truth:** `VERSION.md` contains the current semantic version (e.g., `2.0.0`).
*   **Update Policy:**
    *   **Major (X.0.0):** Architectural changes (e.g., Python -> TypeScript).
    *   **Minor (0.X.0):** New features (e.g., New Dashboard, New Exchange).
    *   **Patch (0.0.X):** Bug fixes and small tweaks.
*   **Automation:** CI/CD pipelines (or Agents) must bump `VERSION.md` on every release commit.
