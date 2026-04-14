# PowerTrader-TS / BobTrader Ultra-Project Submodules & Libraries

## Legacy Python Dependencies
The original Python backend (`pt_*.py`) heavily relied on standard data science libraries:
- **`pandas`, `numpy`:** Handled dataframe manipulation and numerical analysis.
- **`scikit-learn`, `xgboost`:** Explored for ML extensions.
- **`yagmail`, `discord-webhook`, `python-telegram-bot`:** Notification services.
- **`matplotlib`, `tkinter`:** Legacy local GUI visualization.

## Go Port Core Libraries (`ultratrader-go`)
The Go backend transitions these responsibilities into highly concurrent, compiled implementations.

**Standard Library Equivalents:**
- `math`, `sort`, `sync`: Handle concurrent state and calculations previously managed by NumPy.
- `net/http`: Replaces FastAPI/Express for REST routing and web dashboard connections.
- `regexp`: Powers the `nlp.Parser` transitioning from `re`.

**Planned External Integrations:**
- **`github.com/mattn/go-sqlite3`:** SQLite engine for `pt_analytics.py` database equivalence.
- **`github.com/gorilla/websocket`:** WebSocket streaming for live dashboard updates replacing generic WS nodes.

## Project Structure
- `/` (Root): Contains legacy python files, configuration, and documentation.
- `ultratrader-go/`: The active Go migration.
  - `internal/marketdata/`: Exchange connectors, feeds, and aggregation.
  - `internal/trading/`: Execution pipelines, risk management guards, and portfolio trackers.
  - `internal/backtest/`: Walk-forward optimizers, Grid Search, Monte Carlo simulations.
  - `internal/analytics/`: ML Ensembles, Q-Learning Reinforcement Learning, Sentiment engines.
  - `internal/reporting/`: API Handlers for Web Dashboards.
