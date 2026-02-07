import tkinter as tk
from tkinter import ttk
import time
from pt_volume import VolumeAnalyzer, VolumeDataFetcher, VolumeProfile
import threading
from datetime import datetime, timedelta

class VolumeDashboard(ttk.Frame):
    def __init__(self, parent, coin_list, *args, **kwargs):
        super().__init__(parent, *args, **kwargs)
        self.coin_list = coin_list
        self.current_coin = coin_list[0] if coin_list else "BTC"

        self._setup_ui()
        self.fetcher = VolumeDataFetcher()
        self.analyzer = VolumeAnalyzer()

    def _setup_ui(self):
        # Top control bar
        top = ttk.Frame(self)
        top.pack(fill="x", padx=10, pady=10)

        ttk.Label(top, text="Coin:").pack(side="left")
        self.coin_var = tk.StringVar(value=self.current_coin)
        self.coin_combo = ttk.Combobox(top, textvariable=self.coin_var, values=self.coin_list, state="readonly", width=10)
        self.coin_combo.pack(side="left", padx=5)
        self.coin_combo.bind("<<ComboboxSelected>>", self._on_coin_change)

        ttk.Button(top, text="Refresh", command=self.refresh).pack(side="left", padx=10)

        self.status_lbl = ttk.Label(top, text="Ready")
        self.status_lbl.pack(side="left", padx=10)

        # Main content
        content = ttk.Frame(self)
        content.pack(fill="both", expand=True, padx=10)

        # Profile Section
        profile_frame = ttk.LabelFrame(content, text="Volume Profile (30 Days)")
        profile_frame.pack(fill="x", pady=5)

        self.profile_labels = {}
        fields = ["Average Volume", "Median Volume", "Std Dev", "P90 (High)"]
        for i, f in enumerate(fields):
            lbl = ttk.Label(profile_frame, text=f"{f}: N/A")
            lbl.grid(row=0, column=i, padx=20, pady=10)
            self.profile_labels[f] = lbl

        # Analysis Section
        analysis_frame = ttk.LabelFrame(content, text="Recent Analysis")
        analysis_frame.pack(fill="both", expand=True, pady=5)

        # Treeview for recent candles
        cols = ("Time", "Volume", "Ratio", "Z-Score", "Trend", "Anomaly")
        self.tree = ttk.Treeview(analysis_frame, columns=cols, show="headings")
        for c in cols:
            self.tree.heading(c, text=c)
            self.tree.column(c, width=100)

        scrollbar = ttk.Scrollbar(analysis_frame, orient="vertical", command=self.tree.yview)
        self.tree.configure(yscrollcommand=scrollbar.set)

        self.tree.pack(side="left", fill="both", expand=True)
        scrollbar.pack(side="right", fill="y")

    def _on_coin_change(self, event):
        self.current_coin = self.coin_var.get()
        self.refresh()

    def refresh(self):
        self.status_lbl.config(text="Fetching data...")
        threading.Thread(target=self._fetch_data, daemon=True).start()

    def _fetch_data(self):
        try:
            end = datetime.now()
            start = end - timedelta(days=30)
            candles = self.fetcher.fetch_candles(self.current_coin, start, end, "1hour")

            if not candles:
                self.after(0, lambda: self.status_lbl.config(text="No data found"))
                return

            profile = self.analyzer.calculate_profile(candles)

            # Analyze recent candles
            metrics_list = []
            analyzer = VolumeAnalyzer() # Fresh analyzer for streaming
            for c in candles:
                prev_ema = metrics_list[-1].volume_ema if metrics_list else None
                m = analyzer.analyze_candle(c, prev_ema)
                metrics_list.append(m)

            self.after(0, lambda: self._update_ui(profile, metrics_list[-50:]))

        except Exception as e:
            self.after(0, lambda: self.status_lbl.config(text=f"Error: {e}"))

    def _update_ui(self, profile: VolumeProfile, recent_metrics):
        # Update profile
        self.profile_labels["Average Volume"].config(text=f"Average: {profile.avg_volume:,.0f}")
        self.profile_labels["Median Volume"].config(text=f"Median: {profile.median_volume:,.0f}")
        self.profile_labels["Std Dev"].config(text=f"Std Dev: {profile.std_volume:,.0f}")
        self.profile_labels["P90 (High)"].config(text=f"P90: {profile.p90_volume:,.0f}")

        # Update tree
        for i in self.tree.get_children():
            self.tree.delete(i)

        for m in reversed(recent_metrics):
            dt = datetime.fromtimestamp(m.timestamp).strftime('%Y-%m-%d %H:%M')
            anomaly = "YES" if m.anomaly else ""
            self.tree.insert("", "end", values=(
                dt, f"{m.volume:,.0f}", f"{m.volume_ratio:.2f}x", f"{m.z_score:.2f}", m.trend, anomaly
            ))

        self.status_lbl.config(text=f"Updated {datetime.now().strftime('%H:%M:%S')}")
