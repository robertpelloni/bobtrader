#!/usr/bin/env python3
"""
PowerTrader AI - Analytics Dashboard Widget
=========================================
Tkinter widget for displaying real-time performance metrics from pt_analytics.py.
Integrates into pt_hub.py as a new Analytics tab.

Usage:
    from pt_analytics_dashboard import AnalyticsWidget

    analytics = AnalyticsWidget(parent_frame, db_path, settings_getter)
    analytics.pack(fill="both", expand=True)
    analytics.refresh()
"""

import tkinter as tk
from tkinter import ttk
import time
import os
from typing import Callable, Optional

try:
    from pt_analytics import TradeJournal, PerformanceTracker, get_dashboard_metrics

    ANALYTICS_AVAILABLE = True
except ImportError:
    ANALYTICS_AVAILABLE = False


class KPICard(ttk.Frame):
    def __init__(self, parent, title, value, subtext="", width=200, color="green"):
        super().__init__(parent)
        self.pack_propagate(False)

        self.title = title
        self.value_label = tk.StringVar()
        self.subtext_label = tk.StringVar(value=subtext)

        self._build(color)

    def _build(self, color):
        main_frame = ttk.Frame(self)
        main_frame.pack(fill="both", expand=True, padx=10, pady=5)

        ttk.Label(main_frame, text=self.title, font=("Helvetica", 9, "bold")).pack(
            anchor="w"
        )

        ttk.Label(
            main_frame, textvariable=self.value_label, font=("Helvetica", 16)
        ).pack(pady=(5, 0))

        if self.subtext_label.get():
            ttk.Label(
                main_frame,
                textvariable=self.subtext_label,
                font=("Helvetica", 8),
                foreground="gray",
            ).pack(anchor="w")

    def update(self, value, subtext=""):
        self.value_label.set(str(value))
        if subtext:
            self.subtext_label.set(subtext)
        else:
            self.subtext_label.set("")


class PerformanceTable(ttk.Frame):
    def __init__(self, parent):
        super().__init__(parent)

        columns = ("period", "trades", "win_rate", "total_pnl", "pnl_pct")

        self.tree = ttk.Treeview(self, columns=columns, show="headings", height=6)
        self.tree.pack(fill="both", expand=True)

        headings = {
            "period": "Period",
            "trades": "Trades",
            "win_rate": "Win Rate",
            "total_pnl": "Total P&L",
            "pnl_pct": "P&L %",
        }

        for col in columns:
            self.tree.heading(col, text=headings[col])
            self.tree.column(col, width=100, anchor="center")

    def update(self, data):
        for item in self.tree.get_children():
            self.tree.delete(item)

        for row in data:
            self.tree.insert("", "end", values=row)

    def clear(self):
        for item in self.tree.get_children():
            self.tree.delete(item)


class AnalyticsWidget(ttk.Frame):
    def __init__(self, parent, db_path, settings_getter: Callable):
        super().__init__(parent)

        self.db_path = db_path
        self.settings_getter = settings_getter
        self.last_refresh = 0
        self.last_data = None

        self._build_widgets()

    def _build_widgets(self):
        main_container = ttk.Frame(self)
        main_container.pack(fill="both", expand=True, padx=10, pady=10)

        title = ttk.Label(
            main_container, text="Analytics Dashboard", font=("Helvetica", 12, "bold")
        )
        title.pack(pady=(0, 15))

        kpi_frame = ttk.LabelFrame(main_container, text="Performance Overview")
        kpi_frame.pack(fill="x", pady=(0, 10))

        kpi_row1 = ttk.Frame(kpi_frame)
        kpi_row1.pack(fill="x", pady=5)

        kpi_row2 = ttk.Frame(kpi_frame)
        kpi_row2.pack(fill="x", pady=5)

        kpi_row3 = ttk.Frame(kpi_frame)
        kpi_row3.pack(fill="x", pady=5)

        self.kpi_alltime_trades = KPICard(kpi_row1, "All-Time Trades", 0)
        self.kpi_alltime_trades.pack(side="left", padx=5)

        self.kpi_alltime_winrate = KPICard(
            kpi_row1, "All-Time Win Rate", "0%", subtext="Winning / Total"
        )
        self.kpi_alltime_winrate.pack(side="left", padx=5)

        self.kpi_today_pnl = KPICard(
            kpi_row2,
            "Today P&L",
            "$0",
            color="green" if ANALYTICS_AVAILABLE else "gray",
        )
        self.kpi_today_pnl.pack(side="left", padx=5)

        self.kpi_today_trades = KPICard(kpi_row2, "Today Trades", 0)
        self.kpi_today_trades.pack(side="left", padx=5)

        self.kpi_max_drawdown = KPICard(
            kpi_row3, "Max Drawdown", "0%", subtext="All-Time", color="red"
        )
        self.kpi_max_drawdown.pack(side="left", padx=5)

        self.kpi_avg_holding = KPICard(
            kpi_row3, "Avg Holding", "0h", subtext="All-Time"
        )
        self.kpi_avg_holding.pack(side="left", padx=5)

        perf_frame = ttk.LabelFrame(main_container, text="Performance by Period")
        perf_frame.pack(fill="both", expand=True, pady=(10, 0))

        self.perf_table = PerformanceTable(perf_frame)
        self.perf_table.pack(fill="both", expand=True)

        if not ANALYTICS_AVAILABLE:
            msg = ttk.Label(
                main_container,
                text="pt_analytics module not available\nPlease install pt_analytics.py",
                foreground="red",
                justify="center",
            )
            msg.pack(fill="both", expand=True, pady=20)

    def refresh(self, force=False):
        if not ANALYTICS_AVAILABLE:
            return

        settings = self.settings_getter()
        refresh_interval = settings.get("analytics_refresh_seconds", 5.0)

        now = time.time()
        if not force and (now - self.last_refresh) < refresh_interval:
            return

        self.last_refresh = now

        try:
            journal = TradeJournal(self.db_path)
            metrics = get_dashboard_metrics(journal)

            all_time = metrics.get("all_time", {})
            today = metrics.get("today", {})
            last_7d = metrics.get("last_7_days", {})
            last_30d = metrics.get("last_30_days", {})

            trades_all = all_time.get("total_trades", 0)
            winrate_all = all_time.get("win_rate", 0.0)
            pnl_all = all_time.get("total_pnl", 0.0)
            drawdown = all_time.get("max_drawdown", 0.0)

            trades_today = today.get("trades", 0)
            pnl_today = today.get("pnl", 0.0)

            avg_holding = all_time.get("avg_holding_hours", 0.0)

            self.kpi_alltime_trades.update(trades_all)
            self.kpi_alltime_winrate.update(
                f"{winrate_all:.1f}%",
                subtext=f"{int(all_time.get('winning_trades', 0))}W / {trades_all}",
            )

            pnl_sign = "+" if pnl_today >= 0 else ""
            self.kpi_today_pnl.update(f"{pnl_sign}${pnl_today:,.2f}")
            self.kpi_today_trades.update(trades_today)

            self.kpi_max_drawdown.update(f"{drawdown:.2f}%")
            self.kpi_avg_holding.update(f"{avg_holding:.1f}h")

            perf_data = [
                (
                    "All-Time",
                    trades_all,
                    f"{winrate_all:.1f}%",
                    f"${pnl_all:,.2f}",
                    f"{(pnl_all / max(pnl_all, 1) * 100):.1f}%",
                ),
                (
                    "Last 7 Days",
                    last_7d.get("trades", 0),
                    f"{last_7d.get('win_rate', 0):.1f}%",
                    f"${last_7d.get('pnl', 0):,.2f}",
                    "",
                ),
                (
                    "Last 30 Days",
                    last_30d.get("trades", 0),
                    f"{last_30d.get('win_rate', 0):.1f}%",
                    f"${last_30d.get('pnl', 0):,.2f}",
                    "",
                ),
            ]

            self.perf_table.update(perf_data)

            self.last_data = metrics

        except Exception as e:
            print(f"[Analytics Dashboard] Error refreshing: {e}")
