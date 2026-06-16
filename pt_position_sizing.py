"""
Volatility-Adjusted Position Sizing Module for PowerTrader AI

This module implements ATR (Average True Range) position sizing based on market volatility
to optimize risk-adjusted position sizes for consistent risk percentages.

Author: PowerTrader AI Team
Version: 2.0.0
Created: 2026-01-18
License: Apache 2.0
"""

import sqlite3
import os
import pandas as pd
import numpy as np
from datetime import datetime, timedelta
from typing import Dict, Optional
from dataclasses import dataclass


def kelly_fraction(win_rate: float, avg_win: float, avg_loss: float) -> float:
    """
    Simple Kelly formula.
    win_rate – proportion of winning trades (0‑1).
    avg_win – average profit per winning trade (absolute USD).
    avg_loss – average loss per losing trade (absolute USD, positive number).
    Returns a fraction of capital to risk; capped at 0.5 (half‑Kelly) and never negative.
    """
    if avg_loss == 0:
        return 0.0
    b = avg_win / avg_loss
    f = (b * win_rate - (1 - win_rate)) / b
    return max(0.0, min(f, 0.5))


@dataclass
class VolatilityMetrics:
    """Data class for volatility metrics."""

    symbol: str
    atr: float
    atr_pct_change: float
    atr_pct_position: float
    timestamp: datetime
    timeframe: str


@dataclass
class PositionSizingResult:
    """Data class for position sizing results."""

    symbol: str
    position_size_usd: float
    position_size_pct: float
    risk_amount: float
    atr: float
    volatility_level: str


class PositionSizer:
    """Position sizer based on ATR (Average True Range)."""

    def __init__(
        self,
        db_path: str,
        default_risk_pct: float = 0.02,
        min_risk_pct: float = 0.01,
        max_risk_pct: float = 0.10,
    ):
        """
        Initialize position sizer.

        Args:
            db_path: Path to analytics database
            default_risk_pct: Default risk percentage of account (2%)
            min_risk_pct: Minimum position size (1%)
            max_risk_pct: Maximum position size (10%)
        """
        self.db_path = db_path
        self.default_risk_pct = default_risk_pct
        self.min_risk_pct = min_risk_pct
        self.max_risk_pct = max_risk_pct

        self.conn = None
        self.cursor = None
        self._connect()

    def _connect(self) -> None:
        try:
            self.conn = sqlite3.connect(self.db_path)
            self.cursor = self.conn.cursor()
        except Exception as e:
            print(f"[PositionSizer] Error connecting to database: {e}")
            self.conn = None
            self.cursor = None

    def _close(self) -> None:
        if self.conn:
            self.conn.close()
            self.conn = None
            self.cursor = None

    def calculate_true_range(self, high: float, low: float, prev_close: float) -> float:
        """
        Calculate True Range for a single period.

        True Range is the greatest of:
        - High - Low
        - abs(High - Previous Close)
        - abs(Low - Previous Close)

        Args:
            high: Current period high
            low: Current period low
            prev_close: Previous period close

        Returns:
            True Range value
        """
        tr1 = high - low
        tr2 = abs(high - prev_close)
        tr3 = abs(low - prev_close)
        return max(tr1, tr2, tr3)

    def calculate_atr(self, symbol: str, lookback_days: int = 14) -> float:
        """
        Calculate 14-period ATR for a symbol based on historical data.

        Args:
            symbol: Trading symbol
            lookback_days: Days of historical data to analyze (default 14)

        Returns:
            ATR: 14-period Average True Range
        """
        if not self.cursor:
            self._connect()

        try:
            cutoff_date = datetime.now() - timedelta(days=lookback_days * 2)

            query = """
                SELECT timestamp, close_price, high_price, low_price
                FROM trade_exits
                WHERE symbol = ?
                    AND timestamp >= ?
                ORDER BY timestamp DESC
                LIMIT 1000
            """

            self.cursor.execute(query, (symbol, cutoff_date))
            rows = self.cursor.fetchall()

            if len(rows) < 20:
                return 0.0

            df = pd.DataFrame(
                rows, columns=["timestamp", "close_price", "high_price", "low_price"]
            )
            df = df.sort_values("timestamp").reset_index(drop=True)

            true_ranges = []

            for i in range(1, len(df)):
                high = df.loc[i, "high_price"]
                low = df.loc[i, "low_price"]
                prev_close = df.loc[i - 1, "close_price"]

                tr = self.calculate_true_range(high, low, prev_close)
                true_ranges.append(tr)

            if len(true_ranges) < 14:
                return 0.0

            atr_series = (
                pd.Series(true_ranges).rolling(window=14, min_periods=14).mean()
            )

            return atr_series.iloc[-1] if not pd.isna(atr_series.iloc[-1]) else 0.0

        except Exception as e:
            print(f"[PositionSizer] Error calculating ATR for {symbol}: {e}")
            return 0.0

    def get_market_volatility(self, symbol: str, period: int = 30) -> pd.DataFrame:
        """
        Get market volatility data for a symbol.

        Args:
            symbol: Trading symbol
            period: Number of days (default 30)

        Returns:
            DataFrame with volatility metrics
        """
        if not self.cursor:
            self._connect()

        try:
            cutoff_date = datetime.now() - timedelta(days=period)

            query = """
                SELECT timestamp, close_price, high_price, low_price
                FROM trade_exits
                WHERE symbol = ?
                    AND timestamp >= ?
                ORDER BY timestamp DESC
                LIMIT 1000
            """

            self.cursor.execute(query, (symbol, cutoff_date))
            rows = self.cursor.fetchall()

            if len(rows) < 2:
                return pd.DataFrame()

            df = pd.DataFrame(
                rows, columns=["timestamp", "close_price", "high_price", "low_price"]
            )
            df = df.sort_values("timestamp").reset_index(drop=True)

            true_ranges = []

            for i in range(1, len(df)):
                high = df.loc[i, "high_price"]
                low = df.loc[i, "low_price"]
                prev_close = df.loc[i - 1, "close_price"]

                tr = self.calculate_true_range(high, low, prev_close)
                true_ranges.append(tr)

            df["true_range"] = pd.NA
            df.loc[1:, "true_range"] = true_ranges
            df["atr"] = df["true_range"].rolling(window=14, min_periods=14).mean()
            df["pct_change"] = df["close_price"].pct_change()
            df["atr_pct"] = (df["atr"] / df["close_price"]) * 100
            df["volatility"] = df["atr"] / df["close_price"]

            return df

        except Exception as e:
            print(f"[PositionSizer] Error getting volatility for {symbol}: {e}")
            return pd.DataFrame()

    def calculate_position_size(
        self,
        account_value: float,
        atr: float,
        current_price: float,
        risk_pct: Optional[float] = None,
    ) -> PositionSizingResult:
        """
        Calculate position size based on ATR.

        Args:
            account_value: Total account value in USD
            atr: 14-period ATR for symbol
            current_price: Current price of the asset
            risk_pct: Risk percentage (overrides default if not None)

        Returns:
            PositionSizingResult with position size and metrics
        """
        if atr == 0:
            atr = current_price * 0.02

        # Base risk (default 1 % of capital) – may be overridden by risk_pct
        risk_to_use = risk_pct if risk_pct is not None else self.default_risk_pct
        # Simple Kelly boost (hard‑coded 60 % win‑rate, equal avg win/loss) – capped at half‑Kelly
        kelly_adj = kelly_fraction(0.60, 1.0, 1.0)
        risk_to_use = min(risk_to_use * (1 + kelly_adj), self.max_risk_pct)

        atr_pct = (atr / current_price) * 100
        volatility_factor = 1.0

        if atr_pct < 1.0:
            volatility_factor = 1.5
        elif atr_pct < 2.0:
            volatility_factor = 1.25
        elif atr_pct > 5.0:
            volatility_factor = 0.75
        elif atr_pct > 8.0:
            volatility_factor = 0.5

        position_pct = risk_to_use * volatility_factor
        position_pct = max(self.min_risk_pct, min(position_pct, self.max_risk_pct))

        position_size_usd = account_value * position_pct
        risk_amount = position_size_usd * risk_to_use

        volatility_level = "MEDIUM"
        if atr_pct < 1.5:
            volatility_level = "LOW"
        elif atr_pct > 5.0:
            volatility_level = "HIGH"

        return PositionSizingResult(
            symbol="",
            position_size_usd=position_size_usd,
            position_size_pct=position_pct * 100,
            risk_amount=risk_amount,
            atr=atr,
            volatility_level=volatility_level,
        )

    def get_sizing_recommendation(
        self,
        symbol: str,
        account_value: float,
        current_price: float,
        risk_pct: Optional[float] = None,
    ) -> Dict:
        """
        Get complete position sizing recommendation for a symbol.

        Args:
            symbol: Trading symbol
            account_value: Total account value in USD
            current_price: Current price of the asset
            risk_pct: Risk percentage (overrides default if not None)

        Returns:
            Dictionary with complete sizing recommendation
        """
        atr = self.calculate_atr(symbol, lookback_days=14)

        result = self.calculate_position_size(
            account_value=account_value,
            atr=atr,
            current_price=current_price,
            risk_pct=risk_pct,
        )

        return {
            "symbol": symbol,
            "atr": atr,
            "atr_pct": (atr / current_price * 100) if atr > 0 else 0,
            "position_size_usd": result.position_size_usd,
            "position_size_pct": result.position_size_pct,
            "risk_amount": result.risk_amount,
            "volatility_level": result.volatility_level,
            "account_value": account_value,
        }


def main():

    project_dir = os.path.dirname(__file__)
    db_path = os.path.join(project_dir, "hub_data", "trades.db")

    if not os.path.exists(db_path):
        print(f"Database not found at {db_path}")
        print("Creating sample data for testing...")

        os.makedirs(os.path.dirname(db_path), exist_ok=True)
        conn = sqlite3.connect(db_path)
        cursor = conn.cursor()

        cursor.execute("""
            CREATE TABLE IF NOT EXISTS trade_exits (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                timestamp TEXT NOT NULL,
                symbol TEXT NOT NULL,
                close_price REAL,
                high_price REAL,
                low_price REAL
            )
        """)

        from datetime import datetime, timedelta

        base_price = 95000.0
        for i in range(30):
            ts = datetime.now() - timedelta(days=29 - i)
            volatility = np.random.uniform(-0.03, 0.03)
            price_change = base_price * volatility
            close_price = base_price + price_change
            high_price = close_price * (1 + abs(np.random.uniform(0, 0.02)))
            low_price = close_price * (1 - abs(np.random.uniform(0, 0.02)))

            cursor.execute(
                "INSERT INTO trade_exits (timestamp, symbol, close_price, high_price, low_price) VALUES (?, ?, ?, ?, ?)",
                (
                    ts.strftime("%Y-%m-%d %H:%M:%S"),
                    "BTC",
                    close_price,
                    high_price,
                    low_price,
                ),
            )

        conn.commit()
        conn.close()

    sizer = PositionSizer(
        db_path, default_risk_pct=0.02, min_risk_pct=0.01, max_risk_pct=0.10
    )

    print("Testing volatility-adjusted position sizing...")

    account_value = 50000.0
    current_price = 95000.0

    rec = sizer.get_sizing_recommendation("BTC", account_value, current_price)

    print(f"\nSymbol: {rec['symbol']}")
    print(f"Account Value: ${rec['account_value']:,.2f}")
    print(f"Current Price: ${current_price:,.2f}")
    print(f"ATR (14-period): ${rec['atr']:.2f}")
    print(f"ATR %: {rec['atr_pct']:.2f}%")
    print(f"Volatility Level: {rec['volatility_level']}")
    print(f"Recommended Position Size: ${rec['position_size_usd']:,.2f}")
    print(f"Position Size %: {rec['position_size_pct']:.2f}%")
    print(f"Risk Amount: ${rec['risk_amount']:,.2f}")

    print("\nTesting different risk percentages...")

    for risk_pct in [0.01, 0.02, 0.05, 0.10]:
        result = sizer.calculate_position_size(
            account_value, rec["atr"], current_price, risk_pct=risk_pct
        )

        print(
            f"  {risk_pct * 100:.0f}% risk: ${result.position_size_usd:,.2f} ({result.position_size_pct:.2f}%)"
        )


if __name__ == "__main__":
    main()
