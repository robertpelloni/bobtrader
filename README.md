# PowerTrader_AI
Fully automated crypto trading powered by a custom price prediction AI and a structured/tiered DCA system.

DO NOT TRUST THE POWERTRADER FORK FROM Drizztdowhateva!!!

This is my personal trading bot that I decided to make open source. I made this strategy to match my personal goals. This system is meant to be a foundation/framework for you to build your dream bot!

I know there are "commonly essential" trading features that are missing (like no stop loss for example). This is by design because many of those things would just not work with this system's strategy as it stands, for my personal reasons below:

I do not believe in selling worthwhile coins at a loss (and why would you trade anything besides worthwhile coins with a trading bot, anyways???).

I DO believe in crypto. I'd rather just wait and maybe add more money to my account if need be so that the bot can buy even more of the coin while the price is down.

I personally feel like many of those common things people use, like stop loss, are actually a trick or something, and I personally have absolutely no problem adding more money to my account to afford more DCA or having to wait for extended periods of time, if need be. In my opinion, anything else is just greedy and desperate, which is the exact OPPOSITE of needed attributes for long term growth. Plus, this is just spot trading... there's no worry of liquidation and it feels to me like many "risk management" tactics are really only meant for futures trading but people blindly apply them to spot trading when it just plain isn't necessary.

I know the AI and the trading strategy are extremely simple because I'm the one that designed and made them. I've been developing this specific trading strategy for almost a decade and the design of the AI system for the last few years. The overall strategy is based on what ACTUALLY works from real trading experience, not just stuff I read in LLM responses or search engine results.


Ok now that all of that is out of the way...

I am not selling anything. This trading bot is not a product. This system is for experimentation and education. The only reason you would EVER send me money is if you are voluntarily donating (donation routes can be found at the bottom of this readme :) ). Do not fall for any scams! PowerTrader AI is COMPLETELY FREE FOREVER!

IMPORTANT: This software places real trades automatically. You are responsible for everything it does to your money and your account. Keep your API keys private. I am not giving financial advice. I am not responsible for any losses incurred or any security breaches to your computer (the code is entirely open source and can be confirmed non-malicious). You are fully responsible for doing your own due diligence to learn and understand this trading system and to use it properly. You are fully responsible for all of your money and all of the bot's actions, and any gains or losses.

“It’s an instance-based (kNN/kernel-style) predictor with online per-instance reliability weighting, used as a multi-timeframe trading signal.” - ChatGPT on the type of AI used in this trading bot.

So what exactly does that mean?

When people think AI, they usually think about LLM style AIs and neural networks. What many people don't realize is there are many types of Artificial Intelligence and Machine Learning - and the one in my trading system falls under the "Other" category.

When training for a coin, it goes through the entire history for that coin on multiple timeframes and saves each pattern it sees, along with what happens on the next candle AFTER the pattern. It uses these saved patterns to generate a predicted candle by taking a weighted average of the closest matches in memory to the current pattern in time. This weighted average output is done once for each timeframe, from 1 hour up to 1 week. Each timeframe gets its own predicted candle. The low and high prices from these candles are what are shown as the blue and orange horizontal lines on the price charts. 

After a candle closes, it checks what happened against what it predicted, and adjusts the weight for each "memory pattern" that was used to generate the weighted average, depending on how accurate each pattern was compared to what actually happened.

Yes, it is EXTREMELY simple. Yes, it is STILL considered AI.

Here is how the trading bot utilizes the price prediction ai to automatically make trades:

For determining when to start trades, the AI's Thinker script sends a signal to start a trade for a coin if the ask price for the coin drops below at least 3 of the the AI's predicted low prices for the coin (it predicts the currently active candle's high and low prices for each timeframe across all timeframes from 1hr to 1wk).

For determining when to DCA, it uses either the current price level from the AI that is tied to the current amount of DCA buys that have been done on the trade (for example, right after a trade starts when 3 blue lines get crossed, its first DCA wont happen until the price crosses the 4th line, so on so forth), or it uses the hardcoded drawdown % for its current level, whichever it hits first. It only allows a max of 2 DCAs within a rolling 24hr window to keep from dumping all of your money in too quickly on coins that are having an extended downtrend. Other risk management features can easily be added, as well, with just a bit of Python code!

For determining when to sell, the bot uses a trailing profit margin to maximize the potential gains. The margin line is set at either 5% gain if no DCA has happened on the trade, or 2.5% gain if any DCA has happened. The trailing margin gap is 0.5% (this is the amount the price has to go over the profit margin to begin raising the profit margin up to TRAIL after the price and maximize how much profit is gained once the price drops below the profit margin again and the bot sells the trade.


# Setup & First-Time Use (Windows)

THESE INSTRUCTIONS WERE WRITTEN BY AI! PLEASE LET ME KNOW IF THERE ARE ANY ERRORS OR ISSUES WITH THIS SETUP PROCESS!

If you have any crypto holdings in Robinhood currently, either transfer them out of your Robinhood account or sell them to dollars BEFORE going through this setup process!

This page walks you through installing PowerTrader AI from start to finish, in the exact order a first-time user should do it.  
No coding knowledge needed.  
These instructions are Windows-based but PowerTrader AI *should* be able to run on any OS.

IMPORTANT: This software places real trades automatically. You are responsible for everything it does to your money and your account. Keep your API keys private. I am not giving financial advice. I am not responsible for any losses incurred or any security breaches to your computer (the code is entirely open source and can be confirmed non-malicious). You are fully responsible for doing your own due diligence to learn and understand this trading system and to use it properly. You are fully responsible for all of your money and all of the bot's actions, and any gains or losses.

---

## Step 1 — Install Python

1. Go to **python.org** and download Python for Windows.
2. Run the installer.
3. **Check the box** that says: **“Add Python to PATH”**.
4. Click **Install Now**.

---

## Step 2 — Download PowerTrader AI

1. Do not download the zip file of the repo! There is an issue I have to fix.
2. Create a folder on your computer, like: `C:\PowerTraderAI\`
3. On the PowerTrader_AI repo page, go to the code page for pt_hub.py, click the "Download Raw File" button, save it into the folder you just created.
4. Repeat that for all files in the repo (except the readme and the license).

---

## Step 3 — Install PowerTrader AI (one command)

1. Open **Command Prompt** (Windows key → type **cmd** → Enter).
2. Go into your PowerTrader AI folder. Example:

   `cd C:\PowerTraderAI`

3. If using Python 3.12 or higher, run this command:

   `python -m pip install setuptools`

4. Install everything PowerTrader AI needs:

   `python -m pip install -r requirements.txt`

---

## Step 4 — Start PowerTrader AI

From the same Command Prompt window (inside your PowerTrader folder), run:

`python pt_hub.py`

The app that opens is the **PowerTrader Hub**.  
This is the only thing you need to run day-to-day.

---

## Step 5 — Set your folder, coins, and Robinhood keys (inside the Hub)

### Open Settings

In the Hub, open **Settings** and do this in order:

- **Main Neural Folder**: set this to the same folder that contains `pt_hub.py` (recommended easiest).
- **Choose which coins to trade**: start with **BTC**.
- **While you are still in Settings**, click **Robinhood API Setup** and do this:

1. Click **Generate Keys**.
2. Copy the **Public Key** shown in the wizard.
3. On Robinhood, add a new API key and paste that Public Key.
4. Set permissions to allow trading (the wizard tells you what to select).
5. Robinhood will show your API Key (often starts with `rh`). Copy it.
6. Paste the API Key back into the wizard and click **Save**.
7. Close the wizard and go back to the **Settings** screen.
8. **NOW** click **Save** in Settings.

After saving, you will have two files in your PowerTrader AI folder:  
`r_key.txt` and `r_secret.txt`  
Keep them private.

PowerTrader AI uses a simple folder style:  
**BTC uses the main folder**, and other coins use their own subfolders (like `ETH\`).

---

## Step 6 — Train (inside the Hub)

Training builds the system’s coin “memory” so it can generate signals.

1. In the Hub, click **Train All**.
2. Wait until training finishes.

---

## Step 7 — Start the system (inside the Hub)

When all coins have completed training, click:

1. **Start All**

The Hub will:  
**start pt_thinker.py**, wait until it is ready, then it will **start pt_trader.py**.  
You don’t need to manually start separate programs. The hub handles everything!

---

## Neural Levels (the LONG/SHORT numbers)

- These are signal strength levels from low to high.
- They are predicted high and low prices for all timeframes from 1hr to 1wk.
- They are used to show how stretched a coin's price is and for determining when to start trades and potentially when to DCA for the first few levels of DCA (Whichever price is higher, Neural level or hardcoded drawdown % for the current DCA level).
- Higher number = stronger signal.
- LONG = buy-direction signal. SHORT = No-start signal

A TRADE WILL START FOR A COIN IF THAT COIN REACHES A LONG LEVEL OF 3 OR HIGHER WHILE HAVING A SHORT LEVEL OF 0! This is adjustable in the settings.

---

## Features (Version 2.0.0 - 2026-01-18)

### New in v2.0.0: Comprehensive Dashboard & Risk Management
- **Volume Analysis Dashboard**: Visualize volume trends, anomalies, and profiles per coin.
- **Risk Management Dashboard**: Correlation matrix and portfolio diversification analysis.
- **Enhanced Settings**: Centralized configuration for all modules via a tabbed interface.
- **Advanced Documentation**: See [MANUAL.md](MANUAL.md) for a complete guide.

### Analytics Integration System
- **Persistent Trade Journal**: SQLite-based database logging for all trades
- **Performance Tracking**: Real-time metrics including win rate, P&L, Sharpe ratio, max drawdown
- **Trade Group IDs**: Automatic linking of entries, DCAs, and exits for complete trade tracking
- **Dashboard Widgets**: Real-time KPI cards and period comparison tables
- **Integration**: Single-point integration into pt_trader.py with graceful fallback

**Modules**: pt_analytics.py, pt_analytics_dashboard.py

### Analytics Dashboard
- **KPI Cards**: Total trades, win rate, today's P&L, max drawdown
- **Performance Tables**: Period comparisons (all-time, 7 days, 30 days)
- **Real-time Updates**: Auto-refresh with 5-second cache interval
- **GUI Integration**: Dedicated ANALYTICS tab in pt_hub.py

---

### Multi-Exchange Price Aggregation
- **Unified Interface**: ExchangeManager for KuCoin, Binance, and Coinbase
- **Cross-Exchange Price**: Median/VWAP across multiple exchanges
- **Arbitrage Monitoring**: Automatic detection of price spreads between exchanges
- **Fallback Chain**: KuCoin → Binance → Coinbase for reliability
- **Real-time Verification**: Price data verification before trading decisions

**Modules**: pt_exchanges.py (1006 lines), pt_thinker_exchanges.py (100 lines)

---

### Notification System
- **Multi-Platform Support**: Email (Gmail), Discord (webhooks), Telegram (bot)
- **Unified Interface**: Single NotificationManager for all platforms
- **Rate Limiting**: Platform-specific limits (Email: 2/hr, Discord: 30/min, Telegram: 20/min)
- **Notification Levels**: INFO, WARNING, ERROR, CRITICAL with color coding
- **Configuration**: JSON-based configuration file
- **SQLite Logging**: All sent notifications logged for audit trail
- **Async Support**: Non-blocking notifications via asyncio

**Modules**: pt_notifications.py (876 lines)

---

### Volume Analysis System
- **Technical Indicators**: SMA, EMA, VWAP calculations
- **Volume Trend Detection**: Increasing, decreasing, or stable analysis
- **Anomaly Detection**: Z-score based statistical outlier detection
- **CLI Tools**: Backtesting utilities for volume-based strategies

**Modules**: pt_volume.py (237 lines)

---

### Multi-Asset Correlation Analysis
- **Portfolio Correlation Calculation**: Position-weighted correlation analysis
- **Historical Correlation Tracking**: 7/30/90-day correlation periods
- **Diversification Alerts**: Automatic alerts when correlations exceed thresholds (>0.8)
- **Correlation Matrix**: Multi-asset correlation computation
- **Integration Ready**: Integration points for pt_thinker.py and pt_analytics.py

**Modules**: pt_correlation.py (447 lines)

---

### Volatility-Adjusted Position Sizing
- **ATR Calculation**: 14-period Average True Range for volatility measurement
- **True Range Calculation**: Accurate volatility assessment using high/low/close
- **Risk-Adjusted Sizing**: Position sizes based on configurable risk (1%-10% of account)
- **Volatility Factor Adjustment**: Dynamic position sizing based on ATR %
  - Low volatility (<1%): 1.5x position size
  - Medium volatility (1-2%): 1.25x position size
  - High volatility (>5%): 0.75x position size
  - Very high volatility (>8%): 0.5x position size
- **Market Volatility Data**: Volatility metrics retrieval from analytics database
- **Complete Recommendation System**: Position sizing with volatility level classification

**Modules**: pt_position_sizing.py (414 lines)

---

### Version Management
- **Single Source of Truth**: VERSION.md contains project version number
- **Dynamic Display**: Version number shown in GUI header (v2.0.0)
- **Automated Bumping**: Version increments with each release
- **Change Tracking**: Comprehensive CHANGELOG.md documenting all changes
- **Documentation**: ROADMAP.md and MODULE_INDEX.md for project inventory

**Files**: VERSION.md (current version), CHANGELOG.md, ROADMAP.md, MODULE_INDEX.md

---

### Configuration Management System
- **Unified Configuration**: Centralized configuration management with YAML format
- **Config Validation**: Schema validation and constraint checking (trading, notifications, system)
- **Environment Variables**: POWERTRADER_ prefix overrides (KUCOIN_API_KEY, EMAIL_ADDRESS, etc.)
- **Hot-Reload Support**: File watcher for automatic configuration reloading
- **Migration Path**: Seamless migration from existing gui_settings.json
- **Configuration Dataclasses**: TradingConfig, NotificationConfig, ExchangeConfig, AnalyticsConfig, SystemConfig
- **ConfigManager Singleton**: Global access with get_config() function
- **Callback System**: Configuration change notifications for GUI integration
- **Export Methods**: dict and JSON export for GUI settings panel
- **Retention Policies**: Configurable log file rotation and backup retention

**Modules**: pt_config.py (628 lines)

---

### Structured Logging System
- **Structured JSON Logging**: JSON-formatted log entries with timestamps, levels, modules
- **Console & File Handlers**: Dual output (human-readable console, structured file)
- **Log Rotation**: Automatic rotation by file size (configurable max size)
- **Retention Policies**: Backup log retention (configurable count)
- **Critical Notifications**: Integration with pt_notifications.py for critical log events
- **Specialized Loggers**: Module-specific loggers (main, trader, thinker, analytics, notifications, exchanges)
- **Search Functionality**: Query logs by level, module, or text
- **Summary Generation**: By-level and by-module log summaries for dashboard
- **Color-Coded Console**: DEBUG, INFO, WARNING, ERROR, CRITICAL with colors
- **Performance Tracking**: API call timing and performance metrics logging

**Modules**: pt_logging.py (538 lines)

---

## Documentation

- **[MANUAL.md](MANUAL.md)**: **COMPLETE USER MANUAL** - Start here!
- **README.md**: This file - main project documentation, setup, and usage
- **CHANGELOG.md**: Complete version history with all changes documented
- **ROADMAP.md**: Current status and future feature planning
- **MODULE_INDEX.md**: Complete inventory of all modules with versions and locations
- **UNIVERSAL_LLM_INSTRUCTIONS.md**: Universal guidelines for all AI agents
- **Model-Specific Files**: CLAUDE.md (Claude), GEMINI.md (Gemini), GPT.md (GPT), copilot-instructions.md (Copilot)
- **AGENTS.md**: Comprehensive agent instruction documentation
- **MCP_SERVERS_RESEARCH.md**: Research on 25+ MCP servers and financial libraries for future integration

---

## Project Structure

**Core System (5 files)**:
- pt_hub.py (5,835 lines) - Main GUI and orchestration hub
- pt_thinker.py (1,381 lines) - Price prediction AI
- pt_trader.py (2,421 lines) - Trade execution engine
- pt_trainer.py (1,625 lines) - AI training system
- pt_backtester.py (876 lines) - Historical strategy testing

**Analytics System (3 files)**:
- pt_analytics.py (770 lines) - SQLite trade journal
- pt_analytics_dashboard.py (262 lines) - Dashboard widgets

**Exchange System (2 files)**:
- pt_exchanges.py (1006 lines) - Multi-exchange manager
- pt_thinker_exchanges.py (100 lines) - Exchange integration wrapper

**Notification System (2 files)**:
- pt_notifications.py (876 lines) - Unified notification system

**Volume Analysis (1 file)**:
- pt_volume.py (237 lines) - Volume metrics and analysis

**Risk Management (2 files)**:
- pt_correlation.py (447 lines) - Multi-asset correlation analysis
- pt_position_sizing.py (414 lines) - Volatility-adjusted position sizing

**Configuration Management (1 file)**:
- pt_config.py (628 lines) - Unified configuration management with hot-reload

**Logging System (1 file)**:
- pt_logging.py (538 lines) - Structured JSON logging with rotation

**Total**: 17 Python modules, ~17,530 lines of code

---

## License

PowerTrader AI is released under **Apache 2.0** license.

---

## Adding more coins (later)

1. Open **Settings**
2. Add one new coin
3. Save
4. Click **Train All**, wait for training to complete
5. Click **Start All**

---

## Donate

PowerTrader AI is COMPLETELY free and open source! If you want to support the project, you can donate or become a member:

- Cash App: **$garagesteve**
- PayPal: **@garagesteve**
- Facebook (Subscribe to my Facebook page for only $1/month): **https://www.facebook.com/stephen.bryant.hughes**

---

## License

PowerTrader AI is released under the **Apache 2.0** license.

---

IMPORTANT: This software places real trades automatically. You are responsible for everything it does to your money and your account. Keep your API keys private. I am not giving financial advice. I am not responsible for any losses incurred or any security breaches to your computer (the code is entirely open source and can be confirmed non-malicious). You are fully responsible for doing your own due diligence to learn and understand this trading system and to use it properly. You are fully responsible for all of your money and all of the bot's actions, and any gains or losses.
