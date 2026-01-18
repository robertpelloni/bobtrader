#!/usr/bin/env python3
"""
PowerTrader AI - Analytics & Performance Tracking
==================================================
Persistent trade logging, performance metrics, and reporting.

Usage:
    # As module - integrate with pt_trader.py
    from pt_analytics import TradeJournal, PerformanceTracker

    journal = TradeJournal()
    journal.log_entry(coin="BTC", price=50000, quantity=0.01, ...)
    journal.log_exit(trade_id, price=52000, ...)

    tracker = PerformanceTracker(journal)
    print(tracker.daily_summary())
    print(tracker.weekly_report())

    # CLI for reports
    python pt_analytics.py summary
    python pt_analytics.py report --period weekly
    python pt_analytics.py export --format csv
"""

import sqlite3
import json
from datetime import datetime, timedelta
from dataclasses import dataclass, asdict
from typing import List, Dict, Optional, Tuple, Any
from pathlib import Path
from contextlib import contextmanager
import argparse
import sys

DB_PATH = Path("hub_data/analytics.db")


@dataclass
class TradeRecord:
    id: Optional[int]
    coin: str
    side: str  # 'entry', 'dca', 'exit'
    price: float
    quantity: float
    cost_usd: float
    timestamp: datetime
    trade_group_id: str  # Links entry/DCA/exit together
    dca_level: Optional[int] = None  # Which DCA (-2.5%, -5%, etc.)
    trigger_reason: str = ""  # 'neural_level', 'dca_threshold', 'trailing_pm'
    fees: float = 0.0
    notes: str = ""


@dataclass
class ClosedTrade:
    trade_group_id: str
    coin: str
    entry_time: datetime
    exit_time: datetime
    entry_price: float
    exit_price: float
    total_quantity: float
    total_cost: float
    total_proceeds: float
    pnl: float
    pnl_pct: float
    dca_count: int
    holding_seconds: int
    total_fees: float


@dataclass
class PerformanceSnapshot:
    timestamp: datetime
    total_trades: int
    winning_trades: int
    losing_trades: int
    win_rate: float
    total_pnl: float
    total_pnl_pct: float
    avg_trade_pnl: float
    avg_holding_hours: float
    max_drawdown_pct: float
    best_trade_pnl: float
    worst_trade_pnl: float
    avg_dca_per_trade: float
    total_fees: float
    coins_traded: List[str]


class TradeJournal:
    def __init__(self, db_path: Path = DB_PATH):
        self.db_path = db_path
        self.db_path.parent.mkdir(parents=True, exist_ok=True)
        self._init_db()

    @contextmanager
    def _get_conn(self):
        conn = sqlite3.connect(self.db_path, detect_types=sqlite3.PARSE_DECLTYPES)
        conn.row_factory = sqlite3.Row
        try:
            yield conn
            conn.commit()
        finally:
            conn.close()

    def _init_db(self):
        with self._get_conn() as conn:
            conn.execute("""
                CREATE TABLE IF NOT EXISTS trades (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    coin TEXT NOT NULL,
                    side TEXT NOT NULL,
                    price REAL NOT NULL,
                    quantity REAL NOT NULL,
                    cost_usd REAL NOT NULL,
                    timestamp TIMESTAMP NOT NULL,
                    trade_group_id TEXT NOT NULL,
                    dca_level INTEGER,
                    trigger_reason TEXT,
                    fees REAL DEFAULT 0,
                    notes TEXT,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                )
            """)

            conn.execute("""
                CREATE TABLE IF NOT EXISTS closed_trades (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    trade_group_id TEXT UNIQUE NOT NULL,
                    coin TEXT NOT NULL,
                    entry_time TIMESTAMP NOT NULL,
                    exit_time TIMESTAMP NOT NULL,
                    entry_price REAL NOT NULL,
                    exit_price REAL NOT NULL,
                    total_quantity REAL NOT NULL,
                    total_cost REAL NOT NULL,
                    total_proceeds REAL NOT NULL,
                    pnl REAL NOT NULL,
                    pnl_pct REAL NOT NULL,
                    dca_count INTEGER NOT NULL,
                    holding_seconds INTEGER NOT NULL,
                    total_fees REAL NOT NULL,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                )
            """)

            conn.execute("""
                CREATE TABLE IF NOT EXISTS daily_snapshots (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    date DATE UNIQUE NOT NULL,
                    total_trades INTEGER,
                    winning_trades INTEGER,
                    losing_trades INTEGER,
                    total_pnl REAL,
                    total_fees REAL,
                    snapshot_json TEXT,
                    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
                )
            """)

            conn.execute("CREATE INDEX IF NOT EXISTS idx_trades_coin ON trades(coin)")
            conn.execute(
                "CREATE INDEX IF NOT EXISTS idx_trades_timestamp ON trades(timestamp)"
            )
            conn.execute(
                "CREATE INDEX IF NOT EXISTS idx_trades_group ON trades(trade_group_id)"
            )
            conn.execute(
                "CREATE INDEX IF NOT EXISTS idx_closed_coin ON closed_trades(coin)"
            )
            conn.execute(
                "CREATE INDEX IF NOT EXISTS idx_closed_exit ON closed_trades(exit_time)"
            )

    def generate_trade_group_id(self, coin: str) -> str:
        return f"{coin}_{datetime.now().strftime('%Y%m%d_%H%M%S_%f')}"

    def log_entry(
        self,
        coin: str,
        price: float,
        quantity: float,
        cost_usd: float,
        trigger_reason: str = "neural_level",
        fees: float = 0.0,
        notes: str = "",
        trade_group_id: Optional[str] = None,
        timestamp: Optional[datetime] = None,
    ) -> str:
        if trade_group_id is None:
            trade_group_id = self.generate_trade_group_id(coin)
        if timestamp is None:
            timestamp = datetime.now()

        with self._get_conn() as conn:
            conn.execute(
                """
                INSERT INTO trades (coin, side, price, quantity, cost_usd, timestamp,
                                   trade_group_id, trigger_reason, fees, notes)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            """,
                (
                    coin,
                    "entry",
                    price,
                    quantity,
                    cost_usd,
                    timestamp,
                    trade_group_id,
                    trigger_reason,
                    fees,
                    notes,
                ),
            )

        return trade_group_id

    def log_dca(
        self,
        trade_group_id: str,
        coin: str,
        price: float,
        quantity: float,
        cost_usd: float,
        dca_level: int,
        trigger_reason: str = "dca_threshold",
        fees: float = 0.0,
        notes: str = "",
        timestamp: Optional[datetime] = None,
    ):
        if timestamp is None:
            timestamp = datetime.now()

        with self._get_conn() as conn:
            conn.execute(
                """
                INSERT INTO trades (coin, side, price, quantity, cost_usd, timestamp,
                                   trade_group_id, dca_level, trigger_reason, fees, notes)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            """,
                (
                    coin,
                    "dca",
                    price,
                    quantity,
                    cost_usd,
                    timestamp,
                    trade_group_id,
                    dca_level,
                    trigger_reason,
                    fees,
                    notes,
                ),
            )

    def log_exit(
        self,
        trade_group_id: str,
        coin: str,
        price: float,
        quantity: float,
        proceeds_usd: float,
        trigger_reason: str = "trailing_pm",
        fees: float = 0.0,
        notes: str = "",
        timestamp: Optional[datetime] = None,
    ) -> ClosedTrade:
        if timestamp is None:
            timestamp = datetime.now()

        with self._get_conn() as conn:
            conn.execute(
                """
                INSERT INTO trades (coin, side, price, quantity, cost_usd, timestamp,
                                   trade_group_id, trigger_reason, fees, notes)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            """,
                (
                    coin,
                    "exit",
                    price,
                    quantity,
                    proceeds_usd,
                    timestamp,
                    trade_group_id,
                    trigger_reason,
                    fees,
                    notes,
                ),
            )

            rows = conn.execute(
                """
                SELECT * FROM trades WHERE trade_group_id = ? ORDER BY timestamp
            """,
                (trade_group_id,),
            ).fetchall()

        entry_trades = [r for r in rows if r["side"] == "entry"]
        dca_trades = [r for r in rows if r["side"] == "dca"]
        exit_trades = [r for r in rows if r["side"] == "exit"]

        if not entry_trades or not exit_trades:
            raise ValueError(
                f"Invalid trade group {trade_group_id}: missing entry or exit"
            )

        entry_time = (
            datetime.fromisoformat(entry_trades[0]["timestamp"])
            if isinstance(entry_trades[0]["timestamp"], str)
            else entry_trades[0]["timestamp"]
        )
        exit_time = (
            datetime.fromisoformat(exit_trades[-1]["timestamp"])
            if isinstance(exit_trades[-1]["timestamp"], str)
            else exit_trades[-1]["timestamp"]
        )

        all_buys = entry_trades + dca_trades
        total_quantity = sum(r["quantity"] for r in all_buys)
        total_cost = sum(r["cost_usd"] for r in all_buys)
        total_proceeds = sum(r["cost_usd"] for r in exit_trades)
        total_fees = sum(r["fees"] for r in rows)

        entry_price = total_cost / total_quantity if total_quantity > 0 else 0
        exit_price = total_proceeds / total_quantity if total_quantity > 0 else 0

        pnl = total_proceeds - total_cost - total_fees
        pnl_pct = (pnl / total_cost * 100) if total_cost > 0 else 0
        holding_seconds = int((exit_time - entry_time).total_seconds())

        closed = ClosedTrade(
            trade_group_id=trade_group_id,
            coin=coin,
            entry_time=entry_time,
            exit_time=exit_time,
            entry_price=entry_price,
            exit_price=exit_price,
            total_quantity=total_quantity,
            total_cost=total_cost,
            total_proceeds=total_proceeds,
            pnl=pnl,
            pnl_pct=pnl_pct,
            dca_count=len(dca_trades),
            holding_seconds=holding_seconds,
            total_fees=total_fees,
        )

        with self._get_conn() as conn:
            conn.execute(
                """
                INSERT OR REPLACE INTO closed_trades 
                (trade_group_id, coin, entry_time, exit_time, entry_price, exit_price,
                 total_quantity, total_cost, total_proceeds, pnl, pnl_pct, dca_count,
                 holding_seconds, total_fees)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
            """,
                (
                    closed.trade_group_id,
                    closed.coin,
                    closed.entry_time,
                    closed.exit_time,
                    closed.entry_price,
                    closed.exit_price,
                    closed.total_quantity,
                    closed.total_cost,
                    closed.total_proceeds,
                    closed.pnl,
                    closed.pnl_pct,
                    closed.dca_count,
                    closed.holding_seconds,
                    closed.total_fees,
                ),
            )

        return closed

    def get_open_positions(self) -> Dict[str, List[dict]]:
        with self._get_conn() as conn:
            closed_groups = set(
                r["trade_group_id"]
                for r in conn.execute(
                    "SELECT trade_group_id FROM closed_trades"
                ).fetchall()
            )

            all_trades = conn.execute("""
                SELECT * FROM trades ORDER BY timestamp
            """).fetchall()

        open_positions = {}
        for trade in all_trades:
            group_id = trade["trade_group_id"]
            if group_id in closed_groups:
                continue

            coin = trade["coin"]
            if coin not in open_positions:
                open_positions[coin] = []
            open_positions[coin].append(dict(trade))

        return open_positions

    def get_closed_trades(
        self,
        coin: Optional[str] = None,
        start_date: Optional[datetime] = None,
        end_date: Optional[datetime] = None,
        limit: int = 100,
    ) -> List[ClosedTrade]:
        query = "SELECT * FROM closed_trades WHERE 1=1"
        params = []

        if coin:
            query += " AND coin = ?"
            params.append(coin)
        if start_date:
            query += " AND exit_time >= ?"
            params.append(start_date)
        if end_date:
            query += " AND exit_time <= ?"
            params.append(end_date)

        query += " ORDER BY exit_time DESC LIMIT ?"
        params.append(limit)

        with self._get_conn() as conn:
            rows = conn.execute(query, params).fetchall()

        return [
            ClosedTrade(
                trade_group_id=r["trade_group_id"],
                coin=r["coin"],
                entry_time=datetime.fromisoformat(r["entry_time"])
                if isinstance(r["entry_time"], str)
                else r["entry_time"],
                exit_time=datetime.fromisoformat(r["exit_time"])
                if isinstance(r["exit_time"], str)
                else r["exit_time"],
                entry_price=r["entry_price"],
                exit_price=r["exit_price"],
                total_quantity=r["total_quantity"],
                total_cost=r["total_cost"],
                total_proceeds=r["total_proceeds"],
                pnl=r["pnl"],
                pnl_pct=r["pnl_pct"],
                dca_count=r["dca_count"],
                holding_seconds=r["holding_seconds"],
                total_fees=r["total_fees"],
            )
            for r in rows
        ]


class PerformanceTracker:
    def __init__(self, journal: TradeJournal):
        self.journal = journal

    def calculate_snapshot(
        self, start_date: Optional[datetime] = None, end_date: Optional[datetime] = None
    ) -> PerformanceSnapshot:
        trades = self.journal.get_closed_trades(
            start_date=start_date, end_date=end_date, limit=10000
        )

        if not trades:
            return PerformanceSnapshot(
                timestamp=datetime.now(),
                total_trades=0,
                winning_trades=0,
                losing_trades=0,
                win_rate=0.0,
                total_pnl=0.0,
                total_pnl_pct=0.0,
                avg_trade_pnl=0.0,
                avg_holding_hours=0.0,
                max_drawdown_pct=0.0,
                best_trade_pnl=0.0,
                worst_trade_pnl=0.0,
                avg_dca_per_trade=0.0,
                total_fees=0.0,
                coins_traded=[],
            )

        total_trades = len(trades)
        winning = [t for t in trades if t.pnl > 0]
        losing = [t for t in trades if t.pnl <= 0]

        total_pnl = sum(t.pnl for t in trades)
        total_cost = sum(t.total_cost for t in trades)
        total_pnl_pct = (total_pnl / total_cost * 100) if total_cost > 0 else 0

        pnls = [t.pnl for t in trades]
        cumulative = []
        running = 0
        for p in sorted(trades, key=lambda x: x.exit_time):
            running += p.pnl
            cumulative.append(running)

        peak = cumulative[0] if cumulative else 0
        max_dd = 0
        for val in cumulative:
            if val > peak:
                peak = val
            dd = (peak - val) / peak * 100 if peak > 0 else 0
            max_dd = max(max_dd, dd)

        return PerformanceSnapshot(
            timestamp=datetime.now(),
            total_trades=total_trades,
            winning_trades=len(winning),
            losing_trades=len(losing),
            win_rate=(len(winning) / total_trades * 100) if total_trades > 0 else 0,
            total_pnl=total_pnl,
            total_pnl_pct=total_pnl_pct,
            avg_trade_pnl=total_pnl / total_trades if total_trades > 0 else 0,
            avg_holding_hours=sum(t.holding_seconds for t in trades)
            / total_trades
            / 3600
            if total_trades > 0
            else 0,
            max_drawdown_pct=max_dd,
            best_trade_pnl=max(pnls) if pnls else 0,
            worst_trade_pnl=min(pnls) if pnls else 0,
            avg_dca_per_trade=sum(t.dca_count for t in trades) / total_trades
            if total_trades > 0
            else 0,
            total_fees=sum(t.total_fees for t in trades),
            coins_traded=list(set(t.coin for t in trades)),
        )

    def daily_summary(self, date: Optional[datetime] = None) -> Dict[str, Any]:
        if date is None:
            date = datetime.now()

        start = date.replace(hour=0, minute=0, second=0, microsecond=0)
        end = start + timedelta(days=1)

        snapshot = self.calculate_snapshot(start_date=start, end_date=end)

        return {
            "date": start.strftime("%Y-%m-%d"),
            "trades": snapshot.total_trades,
            "wins": snapshot.winning_trades,
            "losses": snapshot.losing_trades,
            "win_rate": f"{snapshot.win_rate:.1f}%",
            "pnl": f"${snapshot.total_pnl:,.2f}",
            "fees": f"${snapshot.total_fees:,.2f}",
            "coins": snapshot.coins_traded,
        }

    def weekly_report(self, weeks_back: int = 0) -> Dict[str, Any]:
        now = datetime.now()
        end = now - timedelta(weeks=weeks_back)
        start = end - timedelta(weeks=1)

        snapshot = self.calculate_snapshot(start_date=start, end_date=end)

        daily_summaries = []
        for i in range(7):
            day = start + timedelta(days=i)
            daily_summaries.append(self.daily_summary(day))

        return {
            "period": f"{start.strftime('%Y-%m-%d')} to {end.strftime('%Y-%m-%d')}",
            "summary": {
                "total_trades": snapshot.total_trades,
                "win_rate": f"{snapshot.win_rate:.1f}%",
                "total_pnl": f"${snapshot.total_pnl:,.2f}",
                "total_pnl_pct": f"{snapshot.total_pnl_pct:.2f}%",
                "avg_trade_pnl": f"${snapshot.avg_trade_pnl:,.2f}",
                "max_drawdown": f"{snapshot.max_drawdown_pct:.2f}%",
                "best_trade": f"${snapshot.best_trade_pnl:,.2f}",
                "worst_trade": f"${snapshot.worst_trade_pnl:,.2f}",
                "avg_holding_hours": f"{snapshot.avg_holding_hours:.1f}h",
                "avg_dca_per_trade": f"{snapshot.avg_dca_per_trade:.1f}",
                "total_fees": f"${snapshot.total_fees:,.2f}",
                "coins_traded": snapshot.coins_traded,
            },
            "daily": daily_summaries,
        }

    def coin_breakdown(self) -> Dict[str, Dict[str, Any]]:
        all_trades = self.journal.get_closed_trades(limit=10000)

        by_coin: Dict[str, List[ClosedTrade]] = {}
        for t in all_trades:
            if t.coin not in by_coin:
                by_coin[t.coin] = []
            by_coin[t.coin].append(t)

        result = {}
        for coin, trades in by_coin.items():
            wins = len([t for t in trades if t.pnl > 0])
            total_pnl = sum(t.pnl for t in trades)
            total_cost = sum(t.total_cost for t in trades)

            result[coin] = {
                "trades": len(trades),
                "wins": wins,
                "losses": len(trades) - wins,
                "win_rate": f"{wins / len(trades) * 100:.1f}%" if trades else "0%",
                "total_pnl": f"${total_pnl:,.2f}",
                "total_pnl_pct": f"{total_pnl / total_cost * 100:.2f}%"
                if total_cost > 0
                else "0%",
                "avg_holding_hours": sum(t.holding_seconds for t in trades)
                / len(trades)
                / 3600
                if trades
                else 0,
            }

        return result

    def export_csv(self, filepath: str):
        trades = self.journal.get_closed_trades(limit=100000)

        with open(filepath, "w") as f:
            f.write(
                "trade_group_id,coin,entry_time,exit_time,entry_price,exit_price,"
                "quantity,cost,proceeds,pnl,pnl_pct,dca_count,holding_hours,fees\n"
            )

            for t in trades:
                f.write(
                    f"{t.trade_group_id},{t.coin},{t.entry_time.isoformat()},"
                    f"{t.exit_time.isoformat()},{t.entry_price:.8f},{t.exit_price:.8f},"
                    f"{t.total_quantity:.8f},{t.total_cost:.2f},{t.total_proceeds:.2f},"
                    f"{t.pnl:.2f},{t.pnl_pct:.2f},{t.dca_count},"
                    f"{t.holding_seconds / 3600:.2f},{t.total_fees:.2f}\n"
                )

        print(f"Exported {len(trades)} trades to {filepath}")


def get_dashboard_metrics(journal: TradeJournal) -> Dict[str, Any]:
    tracker = PerformanceTracker(journal)

    all_time = tracker.calculate_snapshot()
    today = tracker.daily_summary()

    last_7_days = tracker.calculate_snapshot(
        start_date=datetime.now() - timedelta(days=7)
    )

    last_30_days = tracker.calculate_snapshot(
        start_date=datetime.now() - timedelta(days=30)
    )

    open_positions = journal.get_open_positions()

    return {
        "all_time": {
            "total_trades": all_time.total_trades,
            "win_rate": all_time.win_rate,
            "total_pnl": all_time.total_pnl,
            "max_drawdown": all_time.max_drawdown_pct,
        },
        "today": today,
        "last_7_days": {
            "trades": last_7_days.total_trades,
            "pnl": last_7_days.total_pnl,
            "win_rate": last_7_days.win_rate,
        },
        "last_30_days": {
            "trades": last_30_days.total_trades,
            "pnl": last_30_days.total_pnl,
            "win_rate": last_30_days.win_rate,
        },
        "open_positions": len(open_positions),
        "coins_with_positions": list(open_positions.keys()),
    }


def print_summary(tracker: PerformanceTracker):
    snapshot = tracker.calculate_snapshot()

    print("\n" + "=" * 60)
    print("POWERTRADER AI - PERFORMANCE SUMMARY")
    print("=" * 60)

    print(f"\nALL-TIME STATISTICS")
    print("-" * 40)
    print(f"Total Trades:      {snapshot.total_trades:>10}")
    print(f"Winning Trades:    {snapshot.winning_trades:>10}")
    print(f"Losing Trades:     {snapshot.losing_trades:>10}")
    print(f"Win Rate:          {snapshot.win_rate:>9.1f}%")

    print(f"\nPROFITABILITY")
    print("-" * 40)
    print(f"Total P&L:         ${snapshot.total_pnl:>10,.2f}")
    print(f"Total P&L %:       {snapshot.total_pnl_pct:>9.2f}%")
    print(f"Avg Trade P&L:     ${snapshot.avg_trade_pnl:>10,.2f}")
    print(f"Best Trade:        ${snapshot.best_trade_pnl:>10,.2f}")
    print(f"Worst Trade:       ${snapshot.worst_trade_pnl:>10,.2f}")

    print(f"\nRISK METRICS")
    print("-" * 40)
    print(f"Max Drawdown:      {snapshot.max_drawdown_pct:>9.2f}%")
    print(f"Avg Holding Time:  {snapshot.avg_holding_hours:>9.1f}h")
    print(f"Avg DCAs/Trade:    {snapshot.avg_dca_per_trade:>10.1f}")
    print(f"Total Fees:        ${snapshot.total_fees:>10,.2f}")

    print(f"\nCOINS TRADED")
    print("-" * 40)
    print(f"  {', '.join(snapshot.coins_traded) if snapshot.coins_traded else 'None'}")

    print("\n" + "=" * 60)


def main():
    parser = argparse.ArgumentParser(description="PowerTrader AI Analytics")
    subparsers = parser.add_subparsers(dest="command", help="Commands")

    subparsers.add_parser("summary", help="Show performance summary")

    report_parser = subparsers.add_parser("report", help="Generate report")
    report_parser.add_argument(
        "--period", choices=["daily", "weekly"], default="weekly"
    )
    report_parser.add_argument(
        "--back", type=int, default=0, help="Periods back (0=current)"
    )

    export_parser = subparsers.add_parser("export", help="Export trades")
    export_parser.add_argument("--format", choices=["csv", "json"], default="csv")
    export_parser.add_argument("--output", default="trades_export")

    coins_parser = subparsers.add_parser("coins", help="Breakdown by coin")

    args = parser.parse_args()

    journal = TradeJournal()
    tracker = PerformanceTracker(journal)

    if args.command == "summary" or args.command is None:
        print_summary(tracker)

    elif args.command == "report":
        if args.period == "daily":
            date = datetime.now() - timedelta(days=args.back)
            result = tracker.daily_summary(date)
        else:
            result = tracker.weekly_report(args.back)
        print(json.dumps(result, indent=2, default=str))

    elif args.command == "export":
        if args.format == "csv":
            tracker.export_csv(f"{args.output}.csv")
        else:
            trades = journal.get_closed_trades(limit=100000)
            with open(f"{args.output}.json", "w") as f:
                json.dump([asdict(t) for t in trades], f, indent=2, default=str)
            print(f"Exported {len(trades)} trades to {args.output}.json")

    elif args.command == "coins":
        breakdown = tracker.coin_breakdown()
        print("\nPERFORMANCE BY COIN")
        print("=" * 60)
        for coin, stats in breakdown.items():
            print(f"\n{coin}")
            print("-" * 30)
            for key, val in stats.items():
                print(f"  {key}: {val}")


if __name__ == "__main__":
    main()
