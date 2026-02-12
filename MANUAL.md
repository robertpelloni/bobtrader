# PowerTrader AI - User Manual

**Version:** 3.2.0
**Last Updated:** 2026-02-10

---

## Table of Contents

1.  [Introduction](#introduction)
2.  [Getting Started](#getting-started)
    *   [Installation](#installation)
    *   [First Run & Setup](#first-run--setup)
3.  [User Interface Overview](#user-interface-overview)
    *   [Main Dashboard](#main-dashboard)
    *   [Charts Tab](#charts-tab)
    *   [Analytics Tab](#analytics-tab)
    *   [Volume Tab](#volume-tab)
    *   [Risk Tab](#risk-tab)
4.  [Configuration](#configuration)
    *   [Trading Settings](#trading-settings)
    *   [Notifications](#notifications)
    *   [Exchanges](#exchanges)
    *   [Analytics & Logging](#analytics--logging)
    *   [Risk Management](#risk-management)
5.  [Trading Strategy](#trading-strategy)
    *   [Neural Levels](#neural-levels)
    *   [DCA System](#dca-system)
    *   [Trailing Profit Margin](#trailing-profit-margin)
6.  [Advanced Features](#advanced-features)
    *   [Multi-Exchange Support](#multi-exchange-support)
    *   [Volume Analysis](#volume-analysis)
    *   [Correlation & Diversification](#correlation--diversification)
7.  [Troubleshooting](#troubleshooting)

---

## Introduction

**PowerTrader AI** is a fully automated crypto trading bot powered by a custom kNN-based price prediction AI ("The Thinker") and a structured/tiered Dollar Cost Averaging (DCA) execution engine ("The Trader").

Designed for spot trading on Robinhood Crypto (with price data from KuCoin/Binance/Coinbase), it emphasizes long-term growth, resilience, and "no loss selling" logic suitable for high-conviction assets like BTC and ETH.

---

## Getting Started

### Installation

1.  **Install Python 3.10+** (Ensure "Add to PATH" is checked).
2.  **Download** the PowerTrader AI repository.
3.  **Install Dependencies**:
    ```bash
    cd PowerTrader_AI
    pip install -r requirements.txt
    ```

### First Run & Setup

1.  **Launch the Hub**:
    ```bash
    python pt_hub.py
    ```
2.  **Open Settings**: Click `Settings...` in the menu.
3.  **Configure Robinhood API**:
    *   Go to the "Exchanges" or "Trading" tab (depending on version).
    *   Click "Robinhood API Setup".
    *   Follow the wizard to generate keys and paste them into Robinhood.
    *   **Permission Required**: "Read & Trade".
4.  **Select Coins**: Enter symbols (e.g., `BTC, ETH, SOL`) in the Trading settings.
5.  **Save**.

---

## User Interface Overview

The **PowerTrader Hub** (`pt_hub.py`) is your mission control. It manages the AI (`pt_thinker.py`), the Trader (`pt_trader.py`), and provides real-time visualization.

### Main Dashboard

*   **Left Panel**:
    *   **Controls**: Start/Stop buttons for the AI Runner and Trader.
    *   **Account Status**: Real-time Total Value, Buying Power, and PnL.
    *   **Neural Levels**: Live "Long" and "Short" signal bars (0-7) for all monitored coins.
    *   **Live Output**: Console logs from the Runner, Trader, and Trainer.
*   **Right Panel**:
    *   **Tabs**: Switch between Charts, Analytics, Volume, and Risk.
    *   **Current Trades**: Table of active positions, showing PnL, DCA stage, and next triggers.
    *   **Trade History**: Scrollable log of recent buys/sells.

### Charts Tab

*   **Candlestick Chart**: Real-time price action.
*   **Overlays**:
    *   **Blue Lines**: Neural "Long" prediction levels.
    *   **Orange Lines**: Neural "Short" prediction levels.
    *   **Green Line**: Sell target (Trailing Profit Margin).
    *   **Red Line**: Next DCA buy trigger.
    *   **Yellow Line**: Average Cost Basis.
*   **Trade Dots**: Red (Buy), Purple (DCA), Green (Sell) dots showing execution history.

### Analytics Tab

*   **KPI Cards**:
    *   **Win Rate**: % of profitable trades.
    *   **Total PnL**: Realized profit in USD.
    *   **Profit Factor**: Ratio of gross profit to gross loss.
    *   **Sharpe Ratio**: Risk-adjusted return metric.
*   **Performance Table**: Compare results across different time periods (Today, 7 Days, 30 Days, All Time).

### Volume Tab

*   **Volume Profile**: Statistics on trading volume (Average, Median, Percentiles).
*   **Metrics**:
    *   **Volume Ratio**: Current volume vs. Moving Average.
    *   **Trend**: Is volume increasing or decreasing?
    *   **Z-Score**: Statistical anomaly detection (e.g., "High Volume Spike").
*   **Analysis**: AI interpretation of whether volume supports the current price move.

### Risk Tab

*   **Correlation Matrix**: Visual grid showing how much your coins move together.
    *   **Green**: Low correlation (Good for diversification).
    *   **Red**: High correlation (Risk concentration).
*   **Portfolio Score**: Overall diversification score.
*   **Position Sizing**: Recommended position sizes based on volatility (ATR).

---

## Configuration

Access via `Settings...` in the top menu. Settings are saved to `config.yaml` (automatically migrated from `gui_settings.json`).

### Trading Settings

*   **Coins**: Comma-separated list (e.g., `BTC, ETH`).
*   **Trade Start Level**: Neural signal strength (1-7) required to open a new trade (Default: 3).
*   **Start Allocation %**: % of account value for the initial buy (Default: 0.5%).
*   **DCA Multiplier**: How much larger each DCA buy is compared to the previous position size (Default: 2.0x).
*   **DCA Levels**: List of % drops to trigger hard DCA buys (e.g., `-2.5, -5.0, -10.0`).

### Notifications

*   **Enable/Disable**: Global toggle.
*   **Platforms**:
    *   **Email**: Gmail address & App Password.
    *   **Discord**: Webhook URL.
    *   **Telegram**: Bot Token & Chat ID.
*   **Rate Limits**: Max messages per minute to prevent spam.

### Exchanges

*   **Robinhood**: Primary trading execution (managed via Setup Wizard).
*   **Data Sources**: API keys for **KuCoin**, **Binance**, and **Coinbase** (Optional, but recommended for robust price data fallback).

### Analytics & Logging

*   **Retention**: How long to keep trade logs (Default: 365 days).
*   **Log Level**: Detail level for system logs (INFO/DEBUG).

### Risk Management

*   **Correlation Alert**: Threshold (0.0-1.0) to warn about high correlation (Default: 0.8).
*   **Position Sizing**:
    *   **Max Risk %**: Max account % to risk per trade.
    *   **Volatility Factor**: Adjust size based on ATR (High volatility = smaller size).

---

## New in v2.6.0

### Backtesting & Optimization

PowerTrader now includes a powerful **Historical Backtesting Engine** and **Genetic Optimizer**.

#### Running a Backtest
Currently accessible via API (UI integration in progress):
`POST /api/strategy/backtest`
```json
{
  "strategy": "RSIStrategy",
  "symbol": "BTC",
  "timeframe": "1h",
  "startDate": "2023-01-01",
  "initialBalance": 10000
}
```

#### Genetic Optimization (HyperOpt)
To automatically tune strategy parameters (e.g., RSI Period, Thresholds):
`POST /api/hyperopt/run`
```json
{
  "strategy": "RSIStrategy",
  "pair": "BTC-USD",
  "generations": 10,
  "populationSize": 20,
  "parameterSpace": {
    "period": { "min": 5, "max": 30, "step": 1 },
    "buyThreshold": { "min": 20, "max": 40, "step": 1 }
  }
}
```

### Paper Trading Mode

You can now simulate trading without risking real funds, using real market data.
In `config.yaml`:
```yaml
trading:
  execution_mode: "paper"  # vs "live"
```
The system will start with a virtual $10,000 balance and simulate fees (0.1%) and order fills.

### Notification System

The TypeScript backend now supports multi-channel notifications.
Configure in `config.yaml` or through the Settings UI:

```yaml
notifications:
  enabled: true
  platforms:
    email: true
    discord: true
    telegram: true
  discord_webhook_url: "YOUR_WEBHOOK_URL"
  email_address: "your@gmail.com"
  email_app_password: "your_app_password"
  telegram_bot_token: "YOUR_BOT_TOKEN"
  telegram_chat_id: "YOUR_CHAT_ID"
```

### Multi-Exchange Trading

PowerTrader v2.6.0 now supports trading on **KuCoin** and **Binance.US** in addition to Robinhood.
To switch exchanges, update `config.yaml`:

```yaml
trading:
  active_exchange: "kucoin" # options: robinhood, kucoin, binance, paper

exchanges:
  kucoin:
    key: "YOUR_API_KEY"
    secret: "YOUR_API_SECRET"
    passphrase: "YOUR_API_PASSPHRASE"
  binance:
    key: "YOUR_API_KEY"
    secret: "YOUR_API_SECRET"
```

---

## New in v3.0.0-alpha

### AI Evolution: DeepThinker

PowerTrader now features an advanced LSTM neural network engine ("DeepThinker") capable of learning complex price patterns.

#### AI Lab
A new "AI Lab" page allows you to train models directly from the browser:
1.  Navigate to **AI Lab**.
2.  Enter a symbol (e.g., `ETH`).
3.  Click **Start Training Loop**.
4.  Watch the Loss Curve decrease as the model learns.
5.  Click **Test Inference** to see real-time price predictions.

---

## New in v3.1.0

### DeFi Integration

PowerTrader now supports decentralized trading on **Uniswap V3** (Polygon/Ethereum).

#### Configuration
Update `config.yaml` to enable the Uniswap connector:

```yaml
trading:
  active_exchange: "uniswap"

exchanges:
  uniswap:
    # No API Key needed, but RPC URL and Private Key are required
    dummy: "placeholder"

defi:
  rpc_url: "https://polygon-rpc.com"
  private_key: "YOUR_WALLET_PRIVATE_KEY"
  chain_id: 137 # Polygon Mainnet
```

**Note:** Ensure your wallet has MATIC for gas fees.

---

## New in v3.2.0

### Advanced DeFi: Liquidity Manager

PowerTrader now allows you to become a Market Maker on Uniswap V3 (Polygon).

#### Liquidity Dashboard
Navigate to **DeFi > Liquidity** in the frontend to access the new dashboard.
*   **Mint Position**: Add liquidity to a pool (e.g., WETH/USDC).
    *   **Auto-Ranging**: The system automatically calculates the optimal price range (Upper/Lower ticks) using Bollinger Bands (20, 2) on the last 20 candles.
    *   **Input**: Simply provide the amount of Token0 and Token1 you wish to deposit.
*   **Monitor Positions**: View your active NFT positions.
    *   **Stats**: See current liquidity, price range, and **Unclaimed Fees**.
*   **Collect & Remove**:
    *   **Remove**: Withdraw 100% of liquidity and collect all accrued fees in one transaction.

#### Strategy: Bollinger Band Ranging
The `LiquidityManager` uses statistical volatility to set ranges:
*   **Lower Tick**: Corresponds to the Lower Bollinger Band.
*   **Upper Tick**: Corresponds to the Upper Bollinger Band.
*   **Goal**: Capture fees while the price remains within "normal" volatility bounds, minimizing Impermanent Loss risk compared to full-range positions.

---

## Trading Strategy

### Neural Levels

The AI ("Thinker") analyzes historical patterns using a kNN (k-Nearest Neighbors) approach. It outputs 7 predicted low levels (LONG) and 7 predicted high levels (SHORT).
*   **Signal Strength**: Higher level (e.g., 7) = Stronger signal.
*   **Entry**: Trade starts when LONG signal >= `Trade Start Level` (default 3) AND SHORT signal == 0.

### DCA System

If the price drops after entry, the bot buys more to lower the average cost basis.
*   **Triggers**: DCA happens if price hits a specific Neural Level OR a hardcoded % drop (whichever comes first).
*   **Safety**: Max 2 DCA buys per coin per rolling 24 hours.

### Trailing Profit Margin

The bot sells when the price is profitable.
*   **Activation**: When profit > `Start %` (5% initially, 2.5% if DCA occurred).
*   **Trailing**: Once active, the sell line follows the price up, staying `Gap %` (default 0.5%) behind the peak.
*   **Execution**: Sells immediately when price drops below the trailing line.

---

## Advanced Features

### Multi-Exchange Support

PowerTrader aggregates prices from KuCoin, Binance, and Coinbase to ensure accuracy and detect anomalies/arbitrage before trading on Robinhood.

### Volume Analysis

The system analyzes volume trends to confirm trade entries.
*   **Low Volume**: May reject trades if volume is too low (weak signal).
*   **Anomalies**: Detects pump-and-dump signatures via Z-Score analysis.

### Correlation & Diversification

The Risk module monitors your portfolio. If you hold multiple coins that move identically (Correlation > 0.8), it warns you, helping you avoid "all eggs in one basket" risk.

---

## Troubleshooting

*   **"Robinhood API credentials not found"**: Run the Setup Wizard in Settings.
*   **"Not Trained"**: Click "Train All" in the hub.
*   **Trades not starting**: Check if `Buying Power` > $0 and Neural Levels are reaching the threshold (blue bars in the UI).
*   **Logs**: Check `hub_data/powertrader.log` for detailed errors.

---

**Disclaimer**: This software is for educational purposes. Use at your own risk. Crypto trading involves significant risk of loss.
