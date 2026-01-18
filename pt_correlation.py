"""
Multi-Asset Correlation Analysis Module for PowerTrader AI

This module provides correlation analysis between trading pairs to help avoid
excessive exposure to correlated assets and diversify portfolio risk.

Author: PowerTrader AI Team
Version: 2.0.0
Created: 2026-01-18
License: Apache 2.0
"""

import sqlite3
import pandas as pd
import numpy as np
from datetime import datetime, timedelta
from typing import List, Dict, Tuple, Optional
from dataclasses import dataclass


@dataclass
class CorrelationMetrics:
    """Data class for correlation metrics."""

    symbol_a: str
    symbol_b: str
    correlation: float
    p_value: float
    timestamp: datetime
    timeframe: str


@dataclass
class CorrelationAlert:
    """Alert when correlation exceeds threshold."""

    symbol_a: str
    symbol_b: str
    correlation: float
    threshold: float
    timestamp: datetime
    alert_type: str  # HIGH_CORRELATION, DIVERSIFICATION_ALERT


class CorrelationAnalyzer:
    """Analyzes correlation between multiple trading pairs."""

    def __init__(self, db_path: str):
        """
        Initialize the correlation analyzer.

        Args:
            db_path: Path to analytics SQLite database
        """
        self.db_path = db_path
        self.conn = None
        self.cursor = None

        # Connect to database
        self._connect()

    def _connect(self) -> None:
        """Connect to SQLite database."""
        try:
            self.conn = sqlite3.connect(self.db_path)
            self.cursor = self.conn.cursor()
        except Exception as e:
            print(f"[CorrelationAnalyzer] Error connecting to database: {e}")
            self.conn = None
            self.cursor = None

    def calculate_correlation_matrix(
        self, symbols: List[str], timeframe_days: int = 30, min_data_points: int = 20
    ) -> Dict[str, Dict[str, float]]:
        """
        Calculate correlation matrix between all symbol pairs.

        Args:
            symbols: List of trading symbols (e.g., ['BTC', 'ETH', 'SOL'])
            timeframe_days: Number of days of data to analyze
            min_data_points: Minimum data points required

        Returns:
            Dictionary of symbol_a -> {symbol_b: correlation}
        """
        if not self.cursor:
            self._connect()

        correlation_matrix = {}

        for i, symbol_a in enumerate(symbols):
            correlation_matrix[symbol_a] = {}
            for j, symbol_b in enumerate(symbols):
                if i == j:
                    continue

                try:
                    # Query historical price data
                    query = f"""
                        SELECT timestamp, close_price
                        FROM trade_exits
                        WHERE symbol = '{symbol_a}'
                           AND timestamp >= datetime('now', '-{timeframe_days} days')
                        ORDER BY timestamp DESC
                        LIMIT {min_data_points}
                    """

                    self.cursor.execute(query, (timeframe_days,))

                    # Query for symbol_b
                    self.cursor.execute(query, (timeframe_days,))

                    # Merge data and calculate correlation
                    df_a = pd.DataFrame(
                        self.cursor.fetchall(), columns=["timestamp", "close_price"]
                    )
                    df_b = pd.DataFrame(
                        self.cursor.fetchall(), columns=["timestamp", "close_price"]
                    )

                    # Merge on timestamps
                    df_merged = pd.merge(df_a, df_b, on="timestamp", how="inner")
                    df_merged = df_merged.dropna()

                    # Calculate correlation if enough data
                    if len(df_merged) >= min_data_points:
                        # Calculate returns
                        df_a_pct = df_a["close_price"].pct_change().dropna()
                        df_b_pct = df_b["close_price"].pct_change().dropna()

                        # Calculate correlation
                        correlation = df_a_pct["close_price"].corr(
                            df_b_pct["close_price"]
                        )

                        correlation_matrix[symbol_a][symbol_b] = (
                            correlation if not pd.isna(correlation) else 0.0
                        )
                    else:
                        correlation_matrix[symbol_a][symbol_b] = 0.0

                except Exception as e:
                    print(
                        f"[CorrelationAnalyzer] Error calculating correlation between {symbol_a} and {symbol_b}: {e}"
                    )
                    correlation_matrix[symbol_a][symbol_b] = 0.0

        return correlation_matrix

    def get_current_correlations(
        self, symbols: List[str], threshold: float = 0.8, lookback_days: int = 30
    ) -> List[CorrelationAlert]:
        """
        Get current correlation levels and alert if threshold exceeded.

        Args:
            symbols: List of trading symbols
            threshold: Correlation threshold to check (default 0.8)
            lookback_days: Days of historical data to analyze

        Returns:
            List of correlation alerts for high correlations
        """
        if not self.cursor:
            self._connect()

        alerts = []

        # Calculate current correlations
        correlation_matrix = self.calculate_correlation_matrix(symbols, lookback_days)

        # Check for high correlations
        for symbol_a in symbols:
            for symbol_b in symbols:
                if symbol_a == symbol_b:
                    continue

                correlation = correlation_matrix[symbol_a].get(symbol_b, 0.0)

                if correlation and correlation >= threshold:
                    alert = CorrelationAlert(
                        symbol_a=symbol_a,
                        symbol_b=symbol_b,
                        correlation=correlation,
                        threshold=threshold,
                        timestamp=datetime.now(),
                        timeframe=f"{lookback_days} days",
                        alert_type="HIGH_CORRELATION",
                    )
                    alerts.append(alert)

        return alerts

    def get_correlation_history(
        self, symbol_a: str, symbol_b: str, period_days: int = 30
    ) -> pd.DataFrame:
        """
        Get historical correlation data for two symbols.

        Args:
            symbol_a: First trading symbol
            symbol_b: Second trading symbol
            period_days: Number of days to analyze

        Returns:
            DataFrame with historical correlation data
        """
        if not self.cursor:
            self._connect()

        try:
            # Query historical price data
            query = f"""
                        SELECT timestamp, close_price
                        FROM trade_exits
                        WHERE symbol IN ('{symbol_a}', '{symbol_b}')
                           AND timestamp >= datetime('now', '-{period_days} days')
                        ORDER BY timestamp DESC
                        LIMIT 1000
                    """

            self.cursor.execute(query, (period_days,))

            # Get both datasets
            df_a = pd.DataFrame(
                self.cursor.fetchall(), columns=["timestamp", "close_price"]
            )
            df_b = pd.DataFrame(
                self.cursor.fetchall(), columns=["timestamp", "close_price"]
            )

            # Merge on timestamps
            df_merged = pd.merge(df_a, df_b, on="timestamp", how="inner")
            df_merged = df_merged.dropna()

            # Calculate returns for each period
            df_merged["return_a"] = df_merged["close_price"].pct_change()
            df_merged["return_b"] = df_merged["close_price"].pct_change()

            # Calculate rolling correlation
            result = df_merged.copy()
            result["correlation"] = (
                df_merged["return_a"].rolling(20).corr(df_merged["return_b"])
            )

            return result.dropna(subset=["correlation"])

        except Exception as e:
            print(
                f"[CorrelationAnalyzer] Error getting correlation history for {symbol_a}/{symbol_b}: {e}"
            )
            return pd.DataFrame()

    def detect_diversification_alert(
        self,
        portfolio_symbols: List[str],
        new_symbol: str,
        correlation_threshold: float = 0.7,
    ) -> Optional[CorrelationAlert]:
        """
        Alert if adding a new symbol would increase portfolio correlation too much.

        Args:
            portfolio_symbols: Currently held symbols
            new_symbol: Symbol being considered for addition
            correlation_threshold: Maximum acceptable correlation with existing portfolio

        Returns:
            Alert if correlation threshold would be exceeded
        """
        if not self.cursor:
            self._connect()

        try:
            # Calculate correlations with new symbol
            test_symbols = portfolio_symbols + [new_symbol]
            correlation_matrix = self.calculate_correlation_matrix(test_symbols, 30)

            # Check if new symbol correlates too highly with any existing symbol
            for existing_symbol in portfolio_symbols:
                correlation = correlation_matrix[new_symbol].get(existing_symbol, 0.0)
                if correlation and correlation >= correlation_threshold:
                    alert = CorrelationAlert(
                        symbol_a=new_symbol,
                        symbol_b=existing_symbol,
                        correlation=correlation,
                        threshold=correlation_threshold,
                        timestamp=datetime.now(),
                        timeframe="30 days",
                        alert_type="DIVERSIFICATION_ALERT",
                    )
                    self._close()
                    return alert

        except Exception as e:
            print(
                f"[CorrelationAnalyzer] Error checking diversification for {new_symbol}: {e}"
            )
            return None

    def log_correlation_metrics(self, metrics: List[CorrelationMetrics]) -> None:
        """
        Log correlation metrics to analytics database.

        Args:
            metrics: List of correlation metrics to log
        """
        if not self.cursor:
            self._connect()

        try:
            for metric in metrics:
                self.cursor.execute(
                    f"""
                    INSERT INTO correlation_history
                    (symbol_a, symbol_b, correlation, p_value, timestamp, timeframe, alert_type)
                    VALUES (?, ?, ?, ?, ?, ?, ?, ?)
                """,
                    (
                        metric.symbol_a,
                        metric.symbol_b,
                        metric.correlation,
                        metric.p_value,
                        metric.timestamp,
                        metric.timeframe,
                        metric.alert_type,
                    ),
                )

            self.conn.commit()

        except Exception as e:
            print(f"[CorrelationAnalyzer] Error logging correlation metrics: {e}")

    def _close(self) -> None:
        """Close database connection."""
        if self.conn:
            self.conn.close()
            self.conn = None
            self.cursor = None


def calculate_portfolio_correlation(
    db_path: str,
    portfolio: Dict[str, float],  # symbol -> current position size in USD
    symbols: Optional[List[str]] = None,
    correlation_threshold: float = 0.8,
) -> Dict[str, float]:
    """
    Calculate current portfolio correlation based on position sizes.

    Args:
        db_path: Path to analytics database
        portfolio: Dictionary of symbol -> position size in USD
        symbols: Optional list of symbols (defaults to all with positions)
        correlation_threshold: Correlation threshold for warning

    Returns:
        Dictionary of symbol -> current portfolio correlation
    """
    if not symbols:
        # Get symbols with positions from portfolio
        symbols = list(portfolio.keys())

    analyzer = CorrelationAnalyzer(db_path)
    correlation_matrix = analyzer.calculate_correlation_matrix(symbols)

    # Calculate portfolio correlation as weighted average
    portfolio_correlation = {}
    total_value = sum(portfolio.values())

    for symbol in symbols:
        # Weight by position size
        weight = portfolio.get(symbol, 0) / total_value if total_value > 0 else 0

        # Average correlation with all other symbols (weighted by position)
        weighted_avg = 0.0
        weight_sum = 0.0

        for other_symbol in symbols:
            if symbol != symbol:
                correlation = correlation_matrix[symbol].get(other_symbol, 0.0)
                weighted_avg += correlation * weight
                weight_sum += weight

        if weight_sum > 0:
            portfolio_correlation[symbol] = weighted_avg / weight_sum

    return portfolio_correlation


def main():
    """Main function for testing correlation analysis."""
    import sys
    import os

    # For testing, use project database path
    project_dir = os.path.dirname(__file__)
    db_path = os.path.join(project_dir, "hub_data", "trades.db")

    analyzer = CorrelationAnalyzer(db_path)

    # Example: Calculate correlations
    symbols = ["BTC", "ETH", "SOL"]

    print("Calculating correlation matrix...")
    correlation_matrix = analyzer.calculate_correlation_matrix(
        symbols, timeframe_days=30, min_data_points=20
    )

    print("\nCorrelation Matrix:")
    for symbol_a, row in correlation_matrix.items():
        print(f"  {symbol_a}:")
        for symbol_b, correlation in row.items():
            print(f"    {symbol_b}: {correlation:.4f}")

    # Example: Get current correlations
    print("\n\nGetting current correlations...")
    alerts = analyzer.get_current_correlations(symbols, threshold=0.8)

    for alert in alerts:
        print(
            f"ALERT: {alert.symbol_a}/{alert.symbol_b}: {alert.correlation:.4f} >= {alert.threshold:.2f}"
        )
        if alert.alert_type == "HIGH_CORRELATION":
            print(
                f"  -> Consider reducing position in one of these highly correlated assets"
            )

    # Example: Calculate portfolio correlation
    portfolio = {
        "BTC": 50000.0,  # $50,000 BTC
        "ETH": 30000.0,  # $30,000 ETH
        "SOL": 20000.0,  # $20,000 SOL
    }

    print("\n\nPortfolio correlation analysis...")
    portfolio_corr = calculate_portfolio_correlation(
        db_path, portfolio, symbols=None, correlation_threshold=0.8
    )

    for symbol, correlation in portfolio_corr.items():
        print(f"{symbol}: {correlation:.4f} (position-weighted)")


if __name__ == "__main__":
    main()
