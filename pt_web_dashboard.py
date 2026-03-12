"""
PowerTrader AI — Web Dashboard (FastAPI Backend)
===================================================
REST API + WebSocket server for real-time monitoring and control.

Usage:
    python3 pt_web_dashboard.py                # Start on port 8080
    python3 pt_web_dashboard.py --port 9090    # Custom port

Endpoints:
    GET  /                     → Dashboard UI (serves index.html)
    GET  /api/status           → Bot status, uptime, active coins
    GET  /api/portfolio        → Current positions and P&L
    GET  /api/trades           → Recent trade history
    GET  /api/analytics        → KPI metrics (win rate, drawdown, etc.)
    GET  /api/alerts           → Active alert rules
    POST /api/alerts           → Create new alert rule
    WS   /ws/live              → Real-time price + prediction stream
"""

from __future__ import annotations
import argparse
import asyncio
import json
import time
import os
from datetime import datetime
from pathlib import Path
from typing import List, Dict, Optional, Set

from fastapi import FastAPI, WebSocket, WebSocketDisconnect, HTTPException
from fastapi.responses import HTMLResponse, FileResponse
from fastapi.staticfiles import StaticFiles
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
import uvicorn


# =============================================================================
# APP SETUP
# =============================================================================

app = FastAPI(
    title="PowerTrader AI Dashboard",
    description="Real-time monitoring and control for PowerTrader AI",
    version="1.0.0",
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

START_TIME = time.time()

# WebSocket connection manager
class ConnectionManager:
    def __init__(self):
        self.active: Set[WebSocket] = set()

    async def connect(self, ws: WebSocket):
        await ws.accept()
        self.active.add(ws)

    def disconnect(self, ws: WebSocket):
        self.active.discard(ws)

    async def broadcast(self, data: dict):
        dead = set()
        for ws in self.active:
            try:
                await ws.send_json(data)
            except Exception:
                dead.add(ws)
        self.active -= dead

manager = ConnectionManager()


# =============================================================================
# DATA MODELS
# =============================================================================

class AlertRuleCreate(BaseModel):
    coin: str
    condition: str
    threshold: float
    notification_channel: str = "all"


class BotStatus(BaseModel):
    status: str
    uptime_seconds: float
    active_coins: List[str]
    total_trades: int
    version: str


# =============================================================================
# DATA HELPERS
# =============================================================================

def _get_coins() -> List[str]:
    """Read active coins from gui_settings.json."""
    try:
        with open("gui_settings.json") as f:
            settings = json.load(f)
        return settings.get("coins", ["BTC", "ETH"])
    except Exception:
        return ["BTC", "ETH"]


def _get_trade_history() -> List[dict]:
    """Load recent trades from trade log files."""
    trades = []
    for path in sorted(Path(".").glob("trade_history_*.json"), reverse=True)[:5]:
        try:
            with open(path) as f:
                data = json.load(f)
            if isinstance(data, list):
                trades.extend(data)
            elif isinstance(data, dict):
                trades.append(data)
        except Exception:
            pass

    # Also try analytics DB
    try:
        from pt_analytics import TradeJournal
        journal = TradeJournal()
        db_trades = journal.get_recent_trades(limit=50)
        if db_trades:
            trades.extend(db_trades)
    except Exception:
        pass

    return trades[:100]  # Cap at 100


def _get_analytics_metrics() -> dict:
    """Fetch dashboard metrics from analytics module."""
    try:
        from pt_analytics import PerformanceTracker
        tracker = PerformanceTracker()
        return tracker.get_dashboard_metrics()
    except Exception:
        return {
            "total_trades": 0,
            "win_rate": 0.0,
            "today_pnl": 0.0,
            "max_drawdown": 0.0,
            "total_pnl": 0.0,
            "avg_trade_pnl": 0.0,
            "best_trade": 0.0,
            "worst_trade": 0.0,
        }


def _get_portfolio() -> dict:
    """Get current portfolio from state files."""
    portfolio = {"positions": [], "total_value": 0.0, "cash": 0.0}
    try:
        with open("gui_settings.json") as f:
            settings = json.load(f)
        portfolio["cash"] = settings.get("available_capital", 0)

        # Read position states
        for coin in _get_coins():
            state_file = Path(f"state_{coin}.json")
            if state_file.exists():
                with open(state_file) as f:
                    state = json.load(f)
                portfolio["positions"].append({
                    "coin": coin,
                    "side": state.get("side", "none"),
                    "entry_price": state.get("entry_price", 0),
                    "quantity": state.get("quantity", 0),
                    "unrealized_pnl": state.get("unrealized_pnl", 0),
                })
    except Exception:
        pass

    return portfolio


def _get_alerts() -> List[dict]:
    """Load alert rules from file."""
    try:
        with open("alert_rules.json") as f:
            return json.load(f)
    except Exception:
        return []


# =============================================================================
# REST ENDPOINTS
# =============================================================================

@app.get("/", response_class=HTMLResponse)
async def root():
    """Serve the dashboard frontend."""
    html_path = Path(__file__).parent / "web_dashboard" / "index.html"
    if html_path.exists():
        return FileResponse(str(html_path), media_type="text/html")
    return HTMLResponse("<h1>PowerTrader AI Dashboard</h1><p>Frontend not found. Place index.html in web_dashboard/</p>")


@app.get("/api/status")
async def get_status():
    return {
        "status": "running",
        "uptime_seconds": round(time.time() - START_TIME, 1),
        "uptime_human": _format_uptime(time.time() - START_TIME),
        "active_coins": _get_coins(),
        "total_trades": len(_get_trade_history()),
        "version": "3.0.0",
        "websocket_clients": len(manager.active),
        "timestamp": datetime.now().isoformat(),
    }


@app.get("/api/portfolio")
async def get_portfolio():
    return _get_portfolio()


@app.get("/api/trades")
async def get_trades(limit: int = 50):
    trades = _get_trade_history()
    return {"trades": trades[:limit], "total": len(trades)}


@app.get("/api/analytics")
async def get_analytics():
    return _get_analytics_metrics()


@app.get("/api/alerts")
async def get_alerts():
    return {"alerts": _get_alerts()}


@app.post("/api/alerts")
async def create_alert(rule: AlertRuleCreate):
    alerts = _get_alerts()
    new_rule = {
        "rule_id": f"web_{int(time.time())}",
        "coin": rule.coin,
        "condition": rule.condition,
        "threshold": rule.threshold,
        "enabled": True,
        "triggered": False,
        "created_at": datetime.now().isoformat(),
        "last_triggered": "",
        "notification_channel": rule.notification_channel,
    }
    alerts.append(new_rule)
    with open("alert_rules.json", "w") as f:
        json.dump(alerts, f, indent=2)
    return {"status": "created", "rule": new_rule}


# =============================================================================
# MOBILE API (compact payloads for React Native / Flutter)
# =============================================================================

@app.get("/api/mobile/summary")
async def mobile_summary():
    """Single-call summary optimized for mobile apps."""
    analytics = _get_analytics_metrics()
    portfolio = _get_portfolio()
    return {
        "status": "running",
        "uptime": _format_uptime(time.time() - START_TIME),
        "coins": _get_coins(),
        "pnl": analytics.get("total_pnl", 0),
        "today_pnl": analytics.get("today_pnl", 0),
        "win_rate": analytics.get("win_rate", 0),
        "positions": len(portfolio.get("positions", [])),
        "alerts": len(_get_alerts()),
        "version": "3.0.0",
    }


class PushConfig(BaseModel):
    device_token: str
    platform: str = "ios"
    enabled: bool = True


@app.post("/api/mobile/push-config")
async def register_push(config: PushConfig):
    """Register a mobile device for push notifications."""
    push_file = Path("push_devices.json")
    devices = json.loads(push_file.read_text()) if push_file.exists() else []
    devices = [d for d in devices if d.get("device_token") != config.device_token]
    devices.append({"device_token": config.device_token, "platform": config.platform,
                     "enabled": config.enabled, "registered_at": datetime.now().isoformat()})
    push_file.write_text(json.dumps(devices, indent=2))
    return {"status": "registered", "devices": len(devices)}


# =============================================================================
# MARKETPLACE API
# =============================================================================

@app.get("/api/marketplace/strategies")
async def marketplace_strategies(tag: str = "", sort: str = "rating"):
    """List marketplace strategies."""
    try:
        from pt_marketplace import MarketplaceManager
        mkt = MarketplaceManager()
        return {"strategies": mkt.list_strategies(tag=tag, sort_by=sort)}
    except Exception:
        return {"strategies": [], "error": "Marketplace not available"}


@app.get("/api/marketplace/leaderboard")
async def marketplace_leaderboard():
    """Get strategy backtesting leaderboard."""
    try:
        from pt_marketplace import MarketplaceManager, BacktestLeaderboard
        from dataclasses import asdict
        mkt = MarketplaceManager()
        board = BacktestLeaderboard()
        entries = board.build_from_catalog(mkt.list_strategies())
        return {"leaderboard": [asdict(e) for e in entries]}
    except Exception:
        return {"leaderboard": [], "error": "Leaderboard not available"}


@app.post("/api/marketplace/install/{name}")
async def marketplace_install(name: str):
    """Install a strategy from the marketplace."""
    try:
        from pt_marketplace import MarketplaceManager
        mkt = MarketplaceManager()
        return mkt.install_strategy(name)
    except Exception as e:
        raise HTTPException(status_code=404, detail=str(e))


# =============================================================================
# WEBSOCKET
# =============================================================================

@app.websocket("/ws/live")
async def websocket_live(ws: WebSocket):
    """Real-time data stream: prices, predictions, alerts."""
    await manager.connect(ws)
    try:
        while True:
            # Send periodic updates
            data = {
                "type": "tick",
                "timestamp": datetime.now().isoformat(),
                "coins": {},
            }

            # Read latest prices from state files
            for coin in _get_coins():
                try:
                    state_file = Path(f"predicted_bounds_{coin}.json")
                    if state_file.exists():
                        with open(state_file) as f:
                            bounds = json.load(f)
                        data["coins"][coin] = bounds
                except Exception:
                    pass

            await ws.send_json(data)
            await asyncio.sleep(2)  # Update every 2 seconds

    except WebSocketDisconnect:
        manager.disconnect(ws)
    except Exception:
        manager.disconnect(ws)


# =============================================================================
# UTILITIES
# =============================================================================

def _format_uptime(seconds: float) -> str:
    hours = int(seconds // 3600)
    minutes = int((seconds % 3600) // 60)
    secs = int(seconds % 60)
    return f"{hours}h {minutes}m {secs}s"


# =============================================================================
# MAIN
# =============================================================================

def main():
    parser = argparse.ArgumentParser(description="PowerTrader AI Web Dashboard")
    parser.add_argument("--host", default="0.0.0.0", help="Host to bind to")
    parser.add_argument("--port", type=int, default=8080, help="Port to serve on")
    args = parser.parse_args()

    print(f"\n{'='*60}")
    print(f"  PowerTrader AI — Web Dashboard v1.0.0")
    print(f"  http://localhost:{args.port}")
    print(f"{'='*60}\n")

    uvicorn.run(app, host=args.host, port=args.port, log_level="info")


if __name__ == "__main__":
    main()
