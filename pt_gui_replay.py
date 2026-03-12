"""
PowerTrader AI — Trade Replay Widget
=======================================
Replays historical trades on a candle chart with annotations.
Integrates into the Backtesting / Research tab in pt_hub.py.
"""

from __future__ import annotations
import json
import tkinter as tk
from tkinter import ttk, filedialog
from typing import List, Dict, Optional
from pathlib import Path
from datetime import datetime

try:
    import matplotlib
    matplotlib.use("TkAgg")
    from matplotlib.backends.backend_tkagg import FigureCanvasTkAgg
    from matplotlib.figure import Figure
    import matplotlib.dates as mdates
    HAS_MPL = True
except ImportError:
    HAS_MPL = False

try:
    import numpy as np
    HAS_NP = True
except ImportError:
    HAS_NP = False

DARK_BG = "#1e1e2e"
DARK_PANEL = "#2a2a3e"
DARK_FG = "#e0e0e0"
RED = "#ef4444"
GREEN = "#22c55e"
ACCENT = "#7c3aed"
AMBER = "#f59e0b"


class TradeReplay(ttk.Frame):
    """
    Replays historical trades step-by-step on an annotated price chart.
    Loads backtest results JSON and lets users step through trades.
    """

    def __init__(self, parent: tk.Widget, **kwargs):
        super().__init__(parent, **kwargs)
        self.trades: List[dict] = []
        self.prices: List[dict] = []  # {time, price}
        self.current_idx = 0
        self._build_ui()

    def _build_ui(self):
        # Header
        hdr = ttk.Frame(self)
        hdr.pack(fill="x", padx=10, pady=(10, 5))
        ttk.Label(hdr, text="🔄 Trade Replay", font=("Segoe UI", 14, "bold")).pack(side="left")
        ttk.Button(hdr, text="Load Results", command=self._load_file).pack(side="right", padx=4)

        # Chart area
        self.chart_frame = ttk.Frame(self)
        self.chart_frame.pack(fill="both", expand=True, padx=10, pady=5)

        # Playback controls
        controls = ttk.Frame(self)
        controls.pack(fill="x", padx=10, pady=(0, 5))

        self.prev_btn = ttk.Button(controls, text="◀ Prev", command=self._prev_trade)
        self.prev_btn.pack(side="left", padx=4)

        self.play_btn = ttk.Button(controls, text="▶ Play", command=self._play)
        self.play_btn.pack(side="left", padx=4)

        self.next_btn = ttk.Button(controls, text="Next ▶", command=self._next_trade)
        self.next_btn.pack(side="left", padx=4)

        self.speed_var = tk.StringVar(value="1x")
        ttk.Label(controls, text="Speed:").pack(side="left", padx=(20, 4))
        for s in ["0.5x", "1x", "2x", "5x"]:
            ttk.Radiobutton(controls, text=s, variable=self.speed_var, value=s).pack(side="left", padx=2)

        # Trade info
        self.info_frame = ttk.Frame(self)
        self.info_frame.pack(fill="x", padx=10, pady=(0, 10))

        self.trade_label = ttk.Label(self.info_frame, text="No trades loaded",
                                     font=("Consolas", 10))
        self.trade_label.pack(fill="x")

        self.progress_label = ttk.Label(self.info_frame, text="",
                                        font=("Consolas", 9))
        self.progress_label.pack(fill="x")

        self._playing = False

    def _load_file(self):
        """Load a backtest results JSON file."""
        path = filedialog.askopenfilename(
            title="Select Backtest Results",
            filetypes=[("JSON files", "*.json"), ("All files", "*.*")],
            initialdir="."
        )
        if not path:
            return

        try:
            with open(path) as f:
                data = json.load(f)

            # Extract trades
            if isinstance(data, dict):
                self.trades = data.get("trades", [])
                self.prices = data.get("equity_curve", [])
            elif isinstance(data, list):
                self.trades = data
                self.prices = []

            self.current_idx = 0
            self._draw_chart()
            self.trade_label.configure(text=f"Loaded {len(self.trades)} trades from {Path(path).name}")
        except Exception as e:
            self.trade_label.configure(text=f"Error loading: {e}")

    def _draw_chart(self):
        """Draw the chart with trades up to current_idx highlighted."""
        if not HAS_MPL or not HAS_NP:
            return

        for w in self.chart_frame.winfo_children():
            w.destroy()

        fig = Figure(figsize=(8, 4), dpi=100, facecolor=DARK_BG)
        ax = fig.add_subplot(111)
        ax.set_facecolor(DARK_PANEL)

        # Plot equity curve if available
        if self.prices:
            times = list(range(len(self.prices)))
            values = [p.get("equity", p.get("value", 0)) if isinstance(p, dict) else p
                      for p in self.prices]
            ax.plot(times, values, color=ACCENT, linewidth=1.5, alpha=0.7, label="Equity")

        # Plot trades up to current index
        visible_trades = self.trades[:self.current_idx + 1] if self.trades else []

        buy_x, buy_y = [], []
        sell_x, sell_y = [], []

        for i, trade in enumerate(visible_trades):
            entry_price = trade.get("entry_price", trade.get("buy_price", 0))
            exit_price = trade.get("exit_price", trade.get("sell_price", 0))

            if entry_price:
                buy_x.append(i)
                buy_y.append(entry_price)
            if exit_price:
                sell_x.append(i)
                sell_y.append(exit_price)

        if not self.prices and (buy_y or sell_y):
            # If no equity curve, plot entry/exit prices as a line
            all_prices = buy_y + sell_y
            ax.plot(range(len(all_prices)), all_prices, color=ACCENT, linewidth=1, alpha=0.5)

        if buy_x:
            ax.scatter(buy_x, buy_y, color=GREEN, marker="^", s=60, zorder=5, label="Buy")
        if sell_x:
            ax.scatter(sell_x, sell_y, color=RED, marker="v", s=60, zorder=5, label="Sell")

        # Highlight current trade
        if self.trades and 0 <= self.current_idx < len(self.trades):
            trade = self.trades[self.current_idx]
            entry = trade.get("entry_price", trade.get("buy_price", 0))
            exit_p = trade.get("exit_price", trade.get("sell_price", 0))
            pnl = trade.get("pnl", trade.get("return_pct", 0))

            if entry:
                ax.axhline(y=entry, color=AMBER, linestyle="--", alpha=0.5, linewidth=1)
            if exit_p:
                ax.axhline(y=exit_p, color=AMBER, linestyle="--", alpha=0.5, linewidth=1)

        ax.set_title("Trade Replay", color=DARK_FG, fontsize=12, fontweight="bold")
        ax.tick_params(colors=DARK_FG)
        ax.spines["bottom"].set_color(DARK_FG)
        ax.spines["left"].set_color(DARK_FG)
        ax.spines["top"].set_visible(False)
        ax.spines["right"].set_visible(False)
        ax.legend(loc="upper left", fontsize=8, framealpha=0.3,
                  facecolor=DARK_PANEL, edgecolor=DARK_FG, labelcolor=DARK_FG)

        fig.tight_layout()

        canvas = FigureCanvasTkAgg(fig, self.chart_frame)
        canvas.draw()
        canvas.get_tk_widget().pack(fill="both", expand=True)

        self._update_info()

    def _update_info(self):
        """Update trade info labels."""
        if not self.trades:
            return

        n = len(self.trades)
        idx = min(self.current_idx, n - 1)
        trade = self.trades[idx]

        coin = trade.get("coin", "?")
        entry = trade.get("entry_price", trade.get("buy_price", 0))
        exit_p = trade.get("exit_price", trade.get("sell_price", 0))
        pnl = trade.get("pnl", trade.get("return_pct", 0))
        side = trade.get("side", "long")

        pnl_color = "green" if pnl >= 0 else "red"
        self.trade_label.configure(
            text=f"Trade {idx + 1}/{n}: {coin} {side.upper()}  "
                 f"Entry: ${entry:,.2f}  Exit: ${exit_p:,.2f}  "
                 f"P&L: {pnl:+.2f}%"
        )
        self.progress_label.configure(text=f"Progress: {idx + 1} / {n}")

    def _next_trade(self):
        if self.trades and self.current_idx < len(self.trades) - 1:
            self.current_idx += 1
            self._draw_chart()

    def _prev_trade(self):
        if self.current_idx > 0:
            self.current_idx -= 1
            self._draw_chart()

    def _play(self):
        """Auto-play through trades."""
        if self._playing:
            self._playing = False
            self.play_btn.configure(text="▶ Play")
            return

        self._playing = True
        self.play_btn.configure(text="⏸ Pause")
        self._auto_advance()

    def _auto_advance(self):
        if not self._playing or not self.trades:
            return
        if self.current_idx >= len(self.trades) - 1:
            self._playing = False
            self.play_btn.configure(text="▶ Play")
            return

        self.current_idx += 1
        self._draw_chart()

        speed_map = {"0.5x": 2000, "1x": 1000, "2x": 500, "5x": 200}
        delay = speed_map.get(self.speed_var.get(), 1000)
        self.after(delay, self._auto_advance)
