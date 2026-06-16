#!/usr/bin/env python3
"""
fully_automated_gay_luxuxy_communism — Supervisor Agent

Manages both Python and Go trading bots as subprocesses, monitors their
performance, rebalances capital, and exposes a unified comparison dashboard.

Phase 1 implementation — see AUTONOMOUS_DUAL_BOT_STRATEGY.md for full spec.
"""

from __future__ import annotations

import argparse
import http.server
import json
import logging
import subprocess
import threading
import time
from dataclasses import dataclass, field, asdict
from datetime import datetime, timezone
from pathlib import Path
from typing import Optional

# ── Logging ──────────────────────────────────────────────────────────────────

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
    datefmt="%Y-%m-%dT%H:%M:%S",
)
log = logging.getLogger("supervisor")

# ── Paths ─────────────────────────────────────────────────────────────────────

REPO_ROOT = Path(__file__).resolve().parent
ULTRA_DIR = REPO_ROOT / "ultratrader-go"
DATA_DIR = REPO_ROOT / "data" / "competition"
DATA_DIR.mkdir(parents=True, exist_ok=True)

PERFORMANCE_FILE = DATA_DIR / "performance.jsonl"
ALLOCATIONS_FILE = DATA_DIR / "allocations.jsonl"
PATCHES_FILE = DATA_DIR / "patches.jsonl"
RESTARTS_FILE = DATA_DIR / "restarts.jsonl"
DECISIONS_FILE = DATA_DIR / "decisions.jsonl"
STATUS_FILE = DATA_DIR / "supervisor_status.json"

# ── Default Config ────────────────────────────────────────────────────────────

DEFAULT_CONFIG: dict = {
    "python_allocation_pct": 25.0,
    "go_allocation_pct": 25.0,
    "reserve_pct": 50.0,
    "evaluation_window_h": 24,
    "max_drawdown_pct": 30.0,
    "supervisor_port": 8400,
    "go_bot_port": 8300,
    "python_bot_port": 8299,
    "poll_interval_s": 30,
    "go_config": str(ULTRA_DIR / "config" / "paper-live-data.json"),
    "python_config": str(REPO_ROOT / "config" / "autonomous-paper.json"),
    "go_cmd": ["go", "run", "-buildvcs=false", "./cmd/ultratrader"],
    "python_cmd": ["python3", "pt_hub.py"],
}

# ═══════════════════════════════════════════════════════════════════════════════
#  Data Models
# ═══════════════════════════════════════════════════════════════════════════════


@dataclass
class BotState:
    name: str
    process: Optional[subprocess.Popen] = None
    pid: Optional[int] = None
    health_status: str = "unknown"
    last_health_check: float = 0.0
    total_pnl: float = 0.0
    win_rate: float = 0.0
    trade_count: int = 0
    drawdown: float = 0.0
    allocation: float = 0.0
    current_value: float = 0.0
    port: int = 0
    consecutive_failures: int = 0
    started_at: Optional[float] = None
    errors: list[str] = field(default_factory=list)


@dataclass
class ComparisonSnapshot:
    timestamp: str
    python: dict
    go: dict
    winner: str
    allocation: dict
    recommendation: str = ""


# ═══════════════════════════════════════════════════════════════════════════════
#  Supervisor Engine
# ═══════════════════════════════════════════════════════════════════════════════


class Supervisor:
    """Manages both bots, monitors performance, rebalances capital."""

    def __init__(self, config: dict | None = None):
        self.cfg = {**DEFAULT_CONFIG, **(config or {})}
        self.python = BotState(name="python", port=self.cfg["python_bot_port"])
        self.go = BotState(name="go", port=self.cfg["go_bot_port"])
        self.reserve_value: float = 0.0
        self.total_portfolio: float = 0.0
        self.initial_total: float = 0.0
        self.running = False
        self._httpd: Optional[http.server.HTTPServer] = None
        self._thread: Optional[threading.Thread] = None
        self._lock = threading.Lock()
        self.history: list[ComparisonSnapshot] = []
        self.last_rebalance_ts: float = 0.0
        self.rebalance_interval_s: float = 3600.0
        self._load_state()

    # ── Process Management ────────────────────────────────────────────────────

    def start_bot(self, bot: BotState) -> bool:
        """Start one bot as a subprocess."""
        with self._lock:
            if bot.process and bot.process.poll() is None:
                log.info(f"{bot.name} already running (pid={bot.pid})")
                return True

            if bot.name == "go":
                cmd = self.cfg["go_cmd"] + ["--config", self.cfg["go_config"]]
                cwd = str(ULTRA_DIR)
            else:
                cmd = self.cfg["python_cmd"]
                cwd = str(REPO_ROOT)

            log.info(f"Starting {bot.name} bot: {' '.join(cmd)} (cwd={cwd})")
            try:
                proc = subprocess.Popen(
                    cmd,
                    cwd=cwd,
                    stdout=subprocess.PIPE,
                    stderr=subprocess.PIPE,
                    text=True,
                )
                bot.process = proc
                bot.pid = proc.pid
                bot.started_at = time.time()
                bot.consecutive_failures = 0
                bot.health_status = "starting"
                self._log_event("start", bot.name, {"pid": proc.pid})
                return True
            except Exception as e:
                log.error(f"Failed to start {bot.name} bot: {e}")
                bot.consecutive_failures += 1
                return False

    def stop_bot(self, bot: BotState, force: bool = False) -> bool:
        """Stop one bot gracefully (SIGTERM), or forcibly (SIGKILL)."""
        with self._lock:
            if not bot.process or bot.process.poll() is not None:
                bot.process = None
                bot.pid = None
                bot.health_status = "stopped"
                return True

            pid = bot.pid
            log.info(f"Stopping {bot.name} bot (pid={pid}, force={force})")
            try:
                if force:
                    bot.process.kill()
                else:
                    bot.process.terminate()
                try:
                    bot.process.wait(timeout=10)
                except subprocess.TimeoutExpired:
                    if not force:
                        log.warning(f"{bot.name} didn't stop gracefully, killing")
                        bot.process.kill()
                        bot.process.wait(timeout=5)
                bot.process = None
                bot.pid = None
                bot.health_status = "stopped"
                self._log_event("stop", bot.name, {"pid": pid, "force": force})
                return True
            except Exception as e:
                log.error(f"Error stopping {bot.name} bot: {e}")
                return False

    def restart_bot(self, bot: BotState) -> bool:
        """Restart one bot (stop + start)."""
        pid_before = bot.pid
        self.stop_bot(bot)
        time.sleep(2)
        ok = self.start_bot(bot)
        self._log_event(
            "restart",
            bot.name,
            {
                "pid_before": pid_before,
                "pid_after": bot.pid,
                "success": ok,
            },
        )
        return ok

    def start_all(self):
        """Start both bots."""
        log.info("Starting both bots...")
        self.start_bot(self.go)
        time.sleep(1)
        self.start_bot(self.python)
        self.running = True
        self._save_state()

    def stop_all(self, force: bool = False):
        """Stop both bots."""
        log.info("Stopping both bots...")
        self.stop_bot(self.python, force=force)
        self.stop_bot(self.go, force=force)
        self.running = False
        self._save_state()

    def halt_emergency(self):
        """Emergency halt — stop both bots immediately."""
        log.warning("EMERGENCY HALT — stopping both bots")
        self.stop_all(force=True)
        self.running = False
        self._log_event(
            "emergency_halt",
            "both",
            {
                "python_pnl": self.python.total_pnl,
                "go_pnl": self.go.total_pnl,
            },
        )
        self._save_state()

    # ── Health Checks ─────────────────────────────────────────────────────────

    def check_health(self, bot: BotState) -> str:
        """Check if a bot's health endpoint responds."""
        import http.client

        try:
            conn = http.client.HTTPConnection("127.0.0.1", bot.port, timeout=5)
            conn.request("GET", "/health")
            resp = conn.getresponse()
            body = resp.read().decode()

            if resp.status == 200:
                bot.health_status = "healthy"
                bot.consecutive_failures = 0
                try:
                    data = json.loads(body)
                    for k in ("total_pnl", "win_rate", "trade_count", "drawdown"):
                        if k in data:
                            setattr(
                                bot,
                                k,
                                float(data[k]) if k != "trade_count" else int(data[k]),
                            )
                except (json.JSONDecodeError, ValueError):
                    pass
            else:
                bot.health_status = "degraded"
                bot.consecutive_failures += 1
            conn.close()
        except (
            ConnectionRefusedError,
            TimeoutError,
            OSError,
            http.client.HTTPException,
        ):
            if bot.process and bot.process.poll() is not None:
                bot.health_status = "down"
                bot.consecutive_failures += 1
            else:
                bot.health_status = "starting"

        if bot.consecutive_failures >= 3 and bot.health_status == "down":
            log.warning(
                f"{bot.name} down for {bot.consecutive_failures} checks — restarting"
            )
            self.restart_bot(bot)
        bot.last_health_check = time.time()
        return bot.health_status

    def poll_once(self):
        """Poll both bots and record a comparison snapshot."""
        self.check_health(self.python)
        self.check_health(self.go)
        self._read_pnl_from_files()
        snapshot = self._build_snapshot()
        self.history.append(snapshot)
        self._append_jsonl(PERFORMANCE_FILE, asdict(snapshot))
        self._check_drawdown()
        self._save_state()

    def _read_pnl_from_files(self):
        """Fallback: read PnL from bots' data files."""
        go_log = ULTRA_DIR / "data" / "signals" / "signals.jsonl"
        if go_log.exists():
            trades = wins = 0
            for line in go_log.read_text().strip().split("\n"):
                if not line.strip():
                    continue
                try:
                    entry = json.loads(line)
                    trades += 1
                    if entry.get("is_win") or entry.get("pnl", 0) > 0:
                        wins += 1
                except (json.JSONDecodeError, KeyError):
                    pass
            if trades > 0:
                self.go.trade_count = trades
                self.go.win_rate = wins / trades

    # ── Comparison & KPIs ────────────────────────────────────────────────────

    def _build_snapshot(self) -> ComparisonSnapshot:
        now = datetime.now(timezone.utc).isoformat()
        py_pnl, go_pnl = self.python.total_pnl, self.go.total_pnl
        threshold = 0.01

        if py_pnl > go_pnl * (1 + threshold) and go_pnl != 0:
            winner = "python"
        elif go_pnl > py_pnl * (1 + threshold):
            winner = "go"
        else:
            winner = "tie"

        alloc = {
            "python": self.python.current_value or self.python.allocation,
            "go": self.go.current_value or self.go.allocation,
            "reserve": self.reserve_value,
            "total": self.total_portfolio,
        }
        t = self.total_portfolio or 1
        alloc["python_pct"] = round(alloc["python"] / t * 100, 1)
        alloc["go_pct"] = round(alloc["go"] / t * 100, 1)
        alloc["reserve_pct"] = round(alloc["reserve"] / t * 100, 1)

        return ComparisonSnapshot(
            timestamp=now,
            python={
                "health": self.python.health_status,
                "pnl": round(py_pnl, 4),
                "win_rate": round(self.python.win_rate, 4),
                "trades": self.python.trade_count,
                "drawdown": round(self.python.drawdown, 4),
                "allocation": round(self.python.allocation, 2),
                "current_value": round(self.python.current_value, 2),
                "pid": self.python.pid,
                "started_at": self.python.started_at,
            },
            go={
                "health": self.go.health_status,
                "pnl": round(go_pnl, 4),
                "win_rate": round(self.go.win_rate, 4),
                "trades": self.go.trade_count,
                "drawdown": round(self.go.drawdown, 4),
                "allocation": round(self.go.allocation, 2),
                "current_value": round(self.go.current_value, 2),
                "pid": self.go.pid,
                "started_at": self.go.started_at,
            },
            winner=winner,
            allocation=alloc,
            recommendation=(
                f"{winner} is winning — consider rebalancing" if winner != "tie" else ""
            ),
        )

    def _check_drawdown(self):
        for bot in [self.python, self.go]:
            dd = bot.drawdown
            if dd >= self.cfg["max_drawdown_pct"] * 0.66:
                log.warning(f"{bot.name} drawdown at {dd:.1f}% — approaching limit")
            if dd >= self.cfg["max_drawdown_pct"]:
                log.error(f"{bot.name} exceeded max drawdown ({dd:.1f}%) — halting")
                self.stop_bot(bot)
                self._log_event(
                    "drawdown_halt",
                    bot.name,
                    {
                        "drawdown": dd,
                        "limit": self.cfg["max_drawdown_pct"],
                    },
                )

    def _maybe_rebalance(self, snapshot: Optional[ComparisonSnapshot] = None):
        """Rebalance capital from reserve to the winning bot if conditions are met."""
        now = time.time()
        # Enforce minimum interval between rebalances (1 hour default)
        if now - self.last_rebalance_ts < self.rebalance_interval_s:
            return
        if snapshot is None:
            snapshot = self._build_snapshot()
        # Only rebalance if there's a clear winner
        if snapshot.winner == "tie":
            return
        # Only rebalance if reserve is available
        if self.reserve_value <= 0:
            return
        # Transfer 5% of reserve to the winning bot
        transfer = min(self.reserve_value * 0.05, 10.0)  # cap at $10 per rebalance
        if transfer <= 0:
            return
        self.reserve_value -= transfer
        if snapshot.winner == "python":
            self.python.allocation += transfer
            winner_name = "Python"
        else:
            self.go.allocation += transfer
            winner_name = "Go"
        # Update totals
        self.total_portfolio = (
            self.python.allocation + self.go.allocation + self.reserve_value
        )
        self.last_rebalance_ts = now
        self._log_event(
            "rebalance",
            snapshot.winner,
            {
                "transfer_amount": round(transfer, 4),
                "new_reserve": round(self.reserve_value, 4),
                "reason": f"{winner_name} winning",
            },
        )
        log.info(
            f"Rebalanced: transferred ${transfer:.2f} from reserve to {winner_name} bot"
        )

    def force_close_all(self):
        """Immediately close all open positions on both bots."""
        for bot in (self.python, self.go):
            try:
                import requests

                # Adjust the endpoint/port if bots expose a different API
                url = f"http://127.0.0.1:{bot.port}/api/close_all"
                resp = requests.get(url, timeout=5)
                if resp.status_code == 200:
                    log.info(f"All positions closed for {bot.name}")
                else:
                    log.warning(
                        f"Failed to close positions for {bot.name}: {resp.status_code}"
                    )
            except Exception as e:
                log.error(f"Error closing positions for {bot.name}: {e}")

    # ── Dashboard HTTP Server ────────────────────────────────────────────────

    def start_dashboard(self):
        """Start the comparison dashboard on a background thread."""

        class Handler(http.server.BaseHTTPRequestHandler):
            sup: Supervisor = None

            def do_GET(self):
                if self.path == "/api/comparison":
                    self._json(self.sup._build_snapshot())
                elif self.path == "/api/history":
                    self._json([asdict(s) for s in self.sup.history[-100:]])
                elif self.path == "/api/status":
                    self._json(
                        {
                            "running": self.sup.running,
                            "python": {
                                "health": self.sup.python.health_status,
                                "pid": self.sup.python.pid,
                            },
                            "go": {
                                "health": self.sup.go.health_status,
                                "pid": self.sup.go.pid,
                            },
                        }
                    )
                elif self.path == "/halt":
                    self.sup.halt_emergency()
                    self._text("Emergency halt initiated")
                elif self.path.startswith("/restart"):
                    bot_name = self._get_param("bot")
                    bot = {"python": self.sup.python, "go": self.sup.go}.get(bot_name)
                    if bot:
                        ok = self.sup.restart_bot(bot)
                        self._text(f"{'OK' if ok else 'FAIL'} Restart {bot.name}")
                    else:
                        self._text("Use ?bot=python or ?bot=go")
                elif self.path in ("/", "/dashboard"):
                    self._html()
                elif self.path == "/rebalance":
                    self.sup._maybe_rebalance()
                    self._text("Rebalance triggered")
                elif self.path == "/force_close_all":
                    self.sup.force_close_all()
                    self._text("All open positions have been closed")
                else:
                    self.send_response(404)
                    self.end_headers()
                    self.wfile.write(b"Not found")

            def _json(self, data):
                self.send_response(200)
                self.send_header("Content-Type", "application/json")
                self.send_header("Access-Control-Allow-Origin", "http://127.0.0.1")
                self.end_headers()
                self.wfile.write(json.dumps(data, default=str, indent=2).encode())

            def _text(self, msg):
                self.send_response(200)
                self.send_header("Content-Type", "text/plain")
                self.end_headers()
                self.wfile.write(msg.encode())

            def _get_param(self, name):
                from urllib.parse import urlparse, parse_qs

                qs = parse_qs(urlparse(self.path).query)
                return (qs.get(name) or [None])[0]

            def _html(self):
                s = self.sup
                snap = s._build_snapshot()
                a = snap.allocation
                py, go_ = snap.python, snap.go
                html = f"""<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Dual-Bot Supervisor</title>
<style>
* {{ margin:0; padding:0; box-sizing:border-box; }}
body {{ font-family:system-ui,sans-serif; background:#070B10; color:#C7D1DB; padding:20px; }}
h1 {{ color:#00FF66; font-size:1.4em; margin-bottom:16px; }}
.grid {{ display:grid; grid-template-columns:1fr 1fr; gap:16px; margin-bottom:16px; }}
.card {{ background:#0E1626; border:1px solid #243044; border-radius:8px; padding:16px; }}
.card h2 {{ color:#00E5FF; font-size:1em; margin-bottom:12px; }}
.kpi {{ display:flex; justify-content:space-between; padding:4px 0; font-size:0.8em; }}
.kpi .label {{ color:#8B949E; }}
.kpi .value {{ color:#C7D1DB; }}
.green {{ color:#00FF66; }} .red {{ color:#FF3355; }} .yellow {{ color:#FFAA00; }}
.badge {{ display:inline-block; padding:2px 8px; border-radius:4px; font-size:0.7em; }}
.badge.healthy {{ background:#003311; color:#00FF66; }}
.badge.starting {{ background:#332200; color:#FFAA00; }}
.badge.down {{ background:#330011; color:#FF3355; }}
.alloc-bar {{ display:flex; height:24px; border-radius:4px; overflow:hidden; margin:8px 0; }}
.alloc-bar .py {{ background:#0055FF; }} .alloc-bar .go {{ background:#00FF66; }} .alloc-bar .res {{ background:#243044; }}
.controls a {{ display:inline-block; padding:6px 14px; margin:4px; border-radius:4px; text-decoration:none; font-size:0.8em; background:#243044; color:#C7D1DB; border:1px solid #535B63; }}
.halt {{ background:#FF3355; color:white; border:none; }}
.rebalance {{ background:#00AA33; color:white; border:none; }}
.footer {{ font-size:0.7em; color:#535B63; margin-top:16px; border-top:1px solid #243044; padding-top:12px; }}
</style>
</head>
<body>
<h1>Dual-Bot Supervisor <span style="font-size:0.6em;color:#8B949E;">port {s.cfg["supervisor_port"]}</span></h1>
<div class="card">
  <h2>Allocation (Python: {a["python_pct"]}% | Go: {a["go_pct"]}% | Reserve: {a["reserve_pct"]}%)</h2>
  <div class="alloc-bar">
    <div class="py" style="width:{a["python_pct"]}%"></div>
    <div class="go" style="width:{a["go_pct"]}%"></div>
    <div class="res" style="width:{a["reserve_pct"]}%"></div>
  </div>
  <div class="kpi"><span class="label">Total Portfolio:</span><span class="value green">${a["total"]:.2f}</span></div>
</div>
<div class="grid">
  <div class="card">
    <h2>Python Bot <span class="badge {py["health"]}">{py["health"]}</span></h2>
    <div class="kpi"><span class="label">PnL:</span><span class="value {"green" if py["pnl"] >= 0 else "red"}">${py["pnl"]:.4f}</span></div>
    <div class="kpi"><span class="label">Win Rate:</span><span class="value">{py["win_rate"] * 100:.1f}%</span></div>
    <div class="kpi"><span class="label">Trades:</span><span class="value">{py["trades"]}</span></div>
    <div class="kpi"><span class="label">Drawdown:</span><span class="value {"red" if py["drawdown"] > 15 else "yellow"}">{py["drawdown"]:.1f}%</span></div>
    <div class="kpi"><span class="label">PID:</span><span class="value">{py["pid"] or "-"}</span></div>
  </div>
  <div class="card">
    <h2>Go Bot <span class="badge {go_["health"]}">{go_["health"]}</span></h2>
    <div class="kpi"><span class="label">PnL:</span><span class="value {"green" if go_["pnl"] >= 0 else "red"}">${go_["pnl"]:.4f}</span></div>
    <div class="kpi"><span class="label">Win Rate:</span><span class="value">{go_["win_rate"] * 100:.1f}%</span></div>
    <div class="kpi"><span class="label">Trades:</span><span class="value">{go_["trades"]}</span></div>
    <div class="kpi"><span class="label">Drawdown:</span><span class="value {"red" if go_["drawdown"] > 15 else "yellow"}">{go_["drawdown"]:.1f}%</span></div>
    <div class="kpi"><span class="label">PID:</span><span class="value">{go_["pid"] or "-"}</span></div>
  </div>
</div>
<div class="controls">
  <a href="/rebalance" class="rebalance">🔄 Rebalance now</a>
  <a href="/halt" class="halt">EMERGENCY HALT</a>
  <a href="/restart?bot=python">Restart Python</a>
  <a href="/restart?bot=go">Restart Go</a>
  <a href="/api/comparison">API JSON</a>
</div>
<div class="footer">
  fully_automated_gay_luxuxy_communism | {snap.timestamp[:19]}
  <br>Recommendation: {snap.recommendation or "No action needed"}
</div>
<script>setTimeout(()=>location.reload(),30000)</script>
</body>
</html>"""
                self.send_response(200)
                self.send_header("Content-Type", "text/html")
                self.end_headers()
                self.wfile.write(html.encode())

        Handler.sup = self
        try:
            self._httpd = http.server.HTTPServer(
                ("0.0.0.0", self.cfg["supervisor_port"]), Handler
            )
            self._thread = threading.Thread(
                target=self._httpd.serve_forever, daemon=True
            )
            self._thread.start()
            log.info(f"Dashboard: http://127.0.0.1:{self.cfg['supervisor_port']}/")
        except OSError as e:
            log.error(f"Failed to start dashboard: {e}")

    def stop_dashboard(self):
        if self._httpd:
            self._httpd.shutdown()
            self._httpd = None

    # ── Monitoring Loop ──────────────────────────────────────────────────────

    def monitoring_loop(self):
        log.info(f"Monitoring loop started (every {self.cfg['poll_interval_s']}s)")
        while self.running:
            try:
                self.poll_once()
            except Exception as e:
                log.error(f"Poll error: {e}")
            time.sleep(self.cfg["poll_interval_s"])

    def run(self):
        log.info("=" * 60)
        log.info("fully_automated_gay_luxuxy_communism — Supervisor")
        log.info("=" * 60)
        self.start_all()
        self.start_dashboard()
        try:
            self.monitoring_loop()
        except KeyboardInterrupt:
            log.info("Shutdown requested")
        finally:
            self.stop_dashboard()
            self.stop_all(force=True)
            self._save_state()
            log.info("Supervisor stopped")

    # ── Persistence ──────────────────────────────────────────────────────────

    def _save_state(self):
        state = {
            "running": self.running,
            "python": {
                "pid": self.python.pid,
                "health": self.python.health_status,
                "pnl": self.python.total_pnl,
                "win_rate": self.python.win_rate,
                "trades": self.python.trade_count,
                "drawdown": self.python.drawdown,
                "allocation": self.python.allocation,
            },
            "go": {
                "pid": self.go.pid,
                "health": self.go.health_status,
                "pnl": self.go.total_pnl,
                "win_rate": self.go.win_rate,
                "trades": self.go.trade_count,
                "drawdown": self.go.drawdown,
                "allocation": self.go.allocation,
            },
            "reserve": self.reserve_value,
            "total": self.total_portfolio,
            "updated_at": datetime.now(timezone.utc).isoformat(),
        }
        STATUS_FILE.write_text(json.dumps(state, indent=2))

    def _load_state(self):
        if STATUS_FILE.exists():
            try:
                state = json.loads(STATUS_FILE.read_text())
                for bot_name in ("python", "go"):
                    b = getattr(self, bot_name)
                    d = state.get(bot_name, {})
                    b.total_pnl = d.get("pnl", 0)
                    b.win_rate = d.get("win_rate", 0)
                    b.trade_count = d.get("trades", 0)
                    b.allocation = d.get("allocation", 0)
                    b.drawdown = d.get("drawdown", 0)
                self.reserve_value = state.get("reserve", 0)
                self.total_portfolio = state.get("total", 0)
                log.info("Loaded previous supervisor state")
            except (json.JSONDecodeError, KeyError) as e:
                log.warning(f"Could not load state: {e}")

    def _log_event(self, event_type: str, bot: str, details: dict):
        entry = {
            "timestamp": datetime.now(timezone.utc).isoformat(),
            "event": event_type,
            "bot": bot,
            **details,
        }
        self._append_jsonl(DECISIONS_FILE, entry)

    @staticmethod
    def _append_jsonl(path: Path, data: dict):
        with open(path, "a") as f:
            f.write(json.dumps(data, default=str) + "\n")


# ═══════════════════════════════════════════════════════════════════════════════
#  CLI
# ═══════════════════════════════════════════════════════════════════════════════


def main():
    parser = argparse.ArgumentParser(
        description="fully_automated_gay_luxuxy_communism — Supervisor Agent"
    )
    parser.add_argument(
        "action",
        nargs="?",
        default="run",
        choices=["run", "start", "stop", "restart", "status", "halt"],
        help="Action to perform",
    )
    parser.add_argument(
        "--port",
        type=int,
        default=DEFAULT_CONFIG["supervisor_port"],
        help=f"Dashboard port (default: {DEFAULT_CONFIG['supervisor_port']})",
    )
    parser.add_argument(
        "--poll",
        type=int,
        default=DEFAULT_CONFIG["poll_interval_s"],
        help=f"Poll interval in seconds (default: {DEFAULT_CONFIG['poll_interval_s']})",
    )
    parser.add_argument("--config", type=str, help="Path to config JSON file")
    args = parser.parse_args()

    config = {}
    if args.config:
        config = json.loads(Path(args.config).read_text())
    config["supervisor_port"] = args.port
    config["poll_interval_s"] = args.poll

    sup = Supervisor(config)

    if args.action in ("start", "run"):
        sup.run()
    elif args.action == "stop":
        sup.stop_all()
    elif args.action == "restart":
        sup.stop_all()
        time.sleep(2)
        sup.run()
    elif args.action == "halt":
        sup.halt_emergency()
    elif args.action == "status":
        sup._load_state()
        snap = sup._build_snapshot()
        print(json.dumps(asdict(snap), default=str, indent=2))


if __name__ == "__main__":
    main()
