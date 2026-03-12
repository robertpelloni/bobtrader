"""
PowerTrader AI — Alert Rules Builder
=======================================
Visual alert rule builder for the pt_hub.py chart tabs.
Users can create, edit, and manage price/indicator alerts with conditions.
"""

from __future__ import annotations
import json
import tkinter as tk
from tkinter import ttk, messagebox
from typing import List, Dict, Optional
from dataclasses import dataclass, asdict
from pathlib import Path
from datetime import datetime


DARK_BG = "#1e1e2e"
DARK_PANEL = "#2a2a3e"
DARK_FG = "#e0e0e0"
ACCENT = "#7c3aed"
RED = "#ef4444"
GREEN = "#22c55e"
AMBER = "#f59e0b"

ALERTS_FILE = "alert_rules.json"


@dataclass
class AlertRule:
    """A single alert rule."""
    rule_id: str
    coin: str
    condition: str          # "price_above", "price_below", "rsi_above", "rsi_below", "volume_spike"
    threshold: float
    enabled: bool = True
    triggered: bool = False
    created_at: str = ""
    last_triggered: str = ""
    notification_channel: str = "all"  # "all", "discord", "telegram", "email"


CONDITION_LABELS = {
    "price_above": "Price Above",
    "price_below": "Price Below",
    "rsi_above": "RSI Above",
    "rsi_below": "RSI Below",
    "volume_spike": "Volume Spike (%)",
}


class AlertRulesBuilder(ttk.Frame):
    """
    Chart-tab page for creating and managing alert rules.
    """

    def __init__(self, parent: tk.Widget, coins: List[str], **kwargs):
        super().__init__(parent, **kwargs)
        self.coins = coins
        self.rules: List[AlertRule] = []
        self._load_rules()
        self._build_ui()
        self._refresh_list()

    def _build_ui(self):
        # Header
        hdr = ttk.Frame(self)
        hdr.pack(fill="x", padx=10, pady=(10, 5))
        ttk.Label(hdr, text="🔔 Alert Rules Builder", font=("Segoe UI", 14, "bold")).pack(side="left")
        ttk.Button(hdr, text="+ New Rule", command=self._add_rule_dialog).pack(side="right", padx=4)

        # Rules list (Treeview)
        cols = ("coin", "condition", "threshold", "enabled", "triggered")
        self.tree = ttk.Treeview(self, columns=cols, show="headings", height=12)
        self.tree.heading("coin", text="Coin")
        self.tree.heading("condition", text="Condition")
        self.tree.heading("threshold", text="Threshold")
        self.tree.heading("enabled", text="Enabled")
        self.tree.heading("triggered", text="Triggered")

        self.tree.column("coin", width=100, anchor="center")
        self.tree.column("condition", width=150, anchor="center")
        self.tree.column("threshold", width=100, anchor="center")
        self.tree.column("enabled", width=80, anchor="center")
        self.tree.column("triggered", width=80, anchor="center")

        scroll = ttk.Scrollbar(self, orient="vertical", command=self.tree.yview)
        self.tree.configure(yscrollcommand=scroll.set)

        self.tree.pack(fill="both", expand=True, padx=10, pady=5, side="left")
        scroll.pack(fill="y", side="right", pady=5, padx=(0, 10))

        # Action buttons
        btn_frame = ttk.Frame(self)
        btn_frame.pack(fill="x", padx=10, pady=(0, 10))
        ttk.Button(btn_frame, text="Toggle Enable", command=self._toggle_selected).pack(side="left", padx=4)
        ttk.Button(btn_frame, text="Delete Rule", command=self._delete_selected).pack(side="left", padx=4)
        ttk.Button(btn_frame, text="Clear Triggered", command=self._clear_triggered).pack(side="left", padx=4)

        self.status_label = ttk.Label(self, text="", font=("Consolas", 10))
        self.status_label.pack(fill="x", padx=10, pady=(0, 10))

    def _refresh_list(self):
        """Refresh the treeview with current rules."""
        for item in self.tree.get_children():
            self.tree.delete(item)

        for rule in self.rules:
            cond_label = CONDITION_LABELS.get(rule.condition, rule.condition)
            enabled_str = "✅" if rule.enabled else "❌"
            triggered_str = "🔴" if rule.triggered else "—"
            self.tree.insert("", "end", iid=rule.rule_id, values=(
                rule.coin.replace("-USDT", ""),
                cond_label,
                f"{rule.threshold:,.2f}",
                enabled_str,
                triggered_str,
            ))

        active = sum(1 for r in self.rules if r.enabled)
        triggered = sum(1 for r in self.rules if r.triggered)
        self.status_label.configure(
            text=f"Total: {len(self.rules)}  |  Active: {active}  |  Triggered: {triggered}"
        )

    def _add_rule_dialog(self):
        """Open a dialog to add a new alert rule."""
        dialog = tk.Toplevel(self)
        dialog.title("New Alert Rule")
        dialog.geometry("350x300")
        dialog.transient(self)
        dialog.grab_set()

        ttk.Label(dialog, text="Coin:").grid(row=0, column=0, padx=10, pady=5, sticky="w")
        coin_var = tk.StringVar(value=self.coins[0] if self.coins else "BTC-USDT")
        coin_cb = ttk.Combobox(dialog, textvariable=coin_var, values=self.coins, state="readonly", width=20)
        coin_cb.grid(row=0, column=1, padx=10, pady=5)

        ttk.Label(dialog, text="Condition:").grid(row=1, column=0, padx=10, pady=5, sticky="w")
        cond_var = tk.StringVar(value="price_above")
        cond_cb = ttk.Combobox(dialog, textvariable=cond_var,
                               values=list(CONDITION_LABELS.keys()),
                               state="readonly", width=20)
        cond_cb.grid(row=1, column=1, padx=10, pady=5)

        ttk.Label(dialog, text="Threshold:").grid(row=2, column=0, padx=10, pady=5, sticky="w")
        thresh_var = tk.StringVar(value="50000")
        ttk.Entry(dialog, textvariable=thresh_var, width=22).grid(row=2, column=1, padx=10, pady=5)

        ttk.Label(dialog, text="Notify via:").grid(row=3, column=0, padx=10, pady=5, sticky="w")
        notify_var = tk.StringVar(value="all")
        ttk.Combobox(dialog, textvariable=notify_var,
                     values=["all", "discord", "telegram", "email"],
                     state="readonly", width=20).grid(row=3, column=1, padx=10, pady=5)

        def _save():
            try:
                threshold = float(thresh_var.get())
            except ValueError:
                messagebox.showerror("Error", "Threshold must be a number")
                return

            rule = AlertRule(
                rule_id=f"rule_{len(self.rules) + 1}_{int(datetime.now().timestamp())}",
                coin=coin_var.get(),
                condition=cond_var.get(),
                threshold=threshold,
                created_at=datetime.now().isoformat(),
                notification_channel=notify_var.get(),
            )
            self.rules.append(rule)
            self._save_rules()
            self._refresh_list()
            dialog.destroy()

        ttk.Button(dialog, text="Save Rule", command=_save).grid(row=4, column=0, columnspan=2, pady=20)

    def _toggle_selected(self):
        sel = self.tree.selection()
        if not sel:
            return
        for rule_id in sel:
            for rule in self.rules:
                if rule.rule_id == rule_id:
                    rule.enabled = not rule.enabled
        self._save_rules()
        self._refresh_list()

    def _delete_selected(self):
        sel = self.tree.selection()
        if not sel:
            return
        self.rules = [r for r in self.rules if r.rule_id not in sel]
        self._save_rules()
        self._refresh_list()

    def _clear_triggered(self):
        for rule in self.rules:
            rule.triggered = False
            rule.last_triggered = ""
        self._save_rules()
        self._refresh_list()

    def _load_rules(self):
        path = Path(ALERTS_FILE)
        if path.exists():
            try:
                with open(path) as f:
                    data = json.load(f)
                self.rules = [AlertRule(**r) for r in data]
            except Exception:
                self.rules = []

    def _save_rules(self):
        with open(ALERTS_FILE, "w") as f:
            json.dump([asdict(r) for r in self.rules], f, indent=2)

    def check_alerts(self, current_prices: Dict[str, float]) -> List[AlertRule]:
        """
        Check all enabled rules against current prices.
        Returns list of newly triggered rules.
        Called externally by pt_hub.py during the update loop.
        """
        newly_triggered = []
        for rule in self.rules:
            if not rule.enabled or rule.triggered:
                continue

            price = current_prices.get(rule.coin, 0)
            if price == 0:
                continue

            fired = False
            if rule.condition == "price_above" and price > rule.threshold:
                fired = True
            elif rule.condition == "price_below" and price < rule.threshold:
                fired = True

            if fired:
                rule.triggered = True
                rule.last_triggered = datetime.now().isoformat()
                newly_triggered.append(rule)

        if newly_triggered:
            self._save_rules()
            self._refresh_list()

        return newly_triggered
