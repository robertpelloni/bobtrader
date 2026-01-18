#!/usr/bin/env python3
"""
PowerTrader AI - Backtesting Engine
===================================
Simulates trading strategy on historical data to validate performance
before risking real capital.

Usage:
    python pt_backtester.py BTC 2024-01-01 2024-12-31
    python pt_backtester.py ETH 2024-06-01 2024-12-31 --initial-capital 10000
    python pt_backtester.py BTC 2024-01-01 2024-12-31 --fee-pct 0.1
"""

import sys
import json
import argparse
from datetime import datetime, timedelta
from dataclasses import dataclass, field
from typing import List, Dict, Optional, Tuple
from pathlib import Path
import time

try:
    from kucoin.client import Market
except ImportError:
    print("ERROR: kucoin-python not installed. Run: pip install kucoin-python")
    sys.exit(1)


# =============================================================================
# CONFIGURATION - Matches pt_trader.py defaults
# =============================================================================


@dataclass
class BacktestConfig:
    """Configuration matching pt_trader.py trading logic."""

    # Entry conditions
    trade_start_level: int = 3  # Num of predicted lows price must drop below

    # DCA settings
    dca_levels: List[float] = field(
        default_factory=lambda: [-2.5, -5.0, -10.0, -20.0, -30.0, -40.0, -50.0]
    )
    max_dca_buys_per_24h: int = 2
    dca_multiplier: float = 2.0  # Each DCA doubles previous size

    # Position sizing
    start_alloc_pct: float = 0.5  # 0.5% of capital per trade start

    # Exit conditions (trailing profit margin)
    pm_start_pct_no_dca: float = 5.0
    pm_start_pct_with_dca: float = 2.5
    trailing_gap_pct: float = 0.5

    # Simulation settings
    initial_capital: float = 10000.0
    fee_pct: float = 0.075  # Robinhood Crypto spread ~0.075%
    slippage_pct: float = 0.05  # Market order slippage estimate


# =============================================================================
# DATA STRUCTURES
# =============================================================================


@dataclass
class Candle:
    """OHLCV candle data."""

    timestamp: int  # Unix seconds
    open: float
    high: float
    low: float
    close: float
    volume: float

    @property
    def datetime(self) -> datetime:
        return datetime.fromtimestamp(self.timestamp)


@dataclass
class Trade:
    """Completed trade record."""

    entry_time: datetime
    exit_time: datetime
    entry_price: float
    exit_price: float
    quantity: float
    dca_count: int
    pnl: float
    pnl_pct: float
    fees_paid: float
    holding_hours: float


@dataclass
class Position:
    """Open position tracker."""

    coin: str
    quantity: float = 0.0
    cost_basis: float = 0.0  # Total $ invested
    avg_price: float = 0.0
    entry_time: Optional[datetime] = None
    dca_count: int = 0
    dca_times: List[datetime] = field(default_factory=list)
    highest_price_since_entry: float = 0.0
    last_buy_size: float = 0.0

    @property
    def is_open(self) -> bool:
        return self.quantity > 0

    def add(self, quantity: float, price: float, timestamp: datetime):
        """Add to position (entry or DCA)."""
        cost = quantity * price
        self.cost_basis += cost
        self.quantity += quantity
        self.avg_price = self.cost_basis / self.quantity if self.quantity > 0 else 0
        if self.entry_time is None:
            self.entry_time = timestamp
        self.last_buy_size = cost

    def reset(self):
        """Reset after closing position."""
        self.quantity = 0.0
        self.cost_basis = 0.0
        self.avg_price = 0.0
        self.entry_time = None
        self.dca_count = 0
        self.dca_times = []
        self.highest_price_since_entry = 0.0
        self.last_buy_size = 0.0


@dataclass
class BacktestResult:
    """Backtest performance summary."""

    coin: str
    start_date: str
    end_date: str
    initial_capital: float
    final_capital: float
    total_return_pct: float
    total_trades: int
    winning_trades: int
    losing_trades: int
    win_rate: float
    avg_trade_pnl: float
    avg_trade_pnl_pct: float
    max_drawdown_pct: float
    sharpe_ratio: float
    avg_holding_hours: float
    total_fees_paid: float
    avg_dca_per_trade: float
    trades: List[Trade] = field(default_factory=list)
    equity_curve: List[Tuple[datetime, float]] = field(default_factory=list)


# =============================================================================
# PATTERN PREDICTOR (Simplified from pt_thinker.py)
# =============================================================================


class PatternPredictor:
    """
    Simplified pattern-based predictor for backtesting.
    Uses historical price patterns to predict support/resistance levels.
    """

    def __init__(self, coin: str, memory_dir: str = "."):
        self.coin = coin
        self.memory_dir = Path(memory_dir)
        self.timeframes = ["1hour", "4hour", "1day"]  # Subset for speed
        self.lookback_candles = 10  # Pattern length
        self.memories: Dict[str, List[List[float]]] = {}
        self.weights: Dict[str, List[float]] = {}

    def load_memories(self) -> bool:
        """Load pre-trained pattern memories if available."""
        loaded = False
        for tf in self.timeframes:
            mem_file = self.memory_dir / f"memories_{tf}.txt"
            weight_file = self.memory_dir / f"memory_weights_{tf}.txt"

            if mem_file.exists() and weight_file.exists():
                try:
                    with open(mem_file, "r") as f:
                        content = f.read().strip()
                        if content:
                            patterns = content.split("~")
                            self.memories[tf] = [
                                [float(x) for x in p.split(",") if x.strip()]
                                for p in patterns
                                if p.strip()
                            ]
                    with open(weight_file, "r") as f:
                        self.weights[tf] = [
                            float(x) for x in f.read().strip().split(",") if x.strip()
                        ]
                    loaded = True
                except Exception as e:
                    print(f"Warning: Could not load {tf} memories: {e}")
        return loaded

    def predict_levels(
        self, recent_candles: List[Candle], current_price: float
    ) -> Dict[str, float]:
        """
        Predict support levels based on pattern matching.
        Returns dict with predicted low levels.

        For backtesting without memories, uses simple technical levels.
        """
        if not recent_candles:
            return {}

        # Simple technical levels as fallback
        lows = [c.low for c in recent_candles[-20:]]
        highs = [c.high for c in recent_candles[-20:]]

        if not lows:
            return {}

        min_low = min(lows)
        max_high = max(highs)
        avg_low = sum(lows) / len(lows)

        # Generate support levels (simplified)
        range_size = max_high - min_low
        levels = {
            "support_1": current_price - (range_size * 0.02),  # -2% of range
            "support_2": current_price - (range_size * 0.05),  # -5% of range
            "support_3": avg_low,
            "support_4": min_low,
            "support_5": min_low - (range_size * 0.05),
        }

        return levels

    def count_levels_below_price(self, levels: Dict[str, float], price: float) -> int:
        """Count how many predicted levels are above current price (price dropped below them)."""
        return sum(1 for level in levels.values() if price < level)


# =============================================================================
# DATA FETCHER
# =============================================================================


class KuCoinDataFetcher:
    """Fetches historical OHLCV data from KuCoin."""

    def __init__(self):
        self.market = Market(url="https://api.kucoin.com")
        self.rate_limit_delay = 0.2  # seconds between API calls

    def fetch_candles(
        self,
        coin: str,
        start_date: datetime,
        end_date: datetime,
        timeframe: str = "1hour",
    ) -> List[Candle]:
        """
        Fetch historical candles for a coin.

        Args:
            coin: Coin symbol (e.g., 'BTC', 'ETH')
            start_date: Start of backtest period
            end_date: End of backtest period
            timeframe: Candle interval ('1hour', '4hour', '1day', etc.)

        Returns:
            List of Candle objects sorted by timestamp ascending
        """
        symbol = f"{coin}-USDT"
        all_candles = []

        # KuCoin returns max 1500 candles per request
        # Calculate chunk size based on timeframe
        tf_minutes = {
            "1min": 1,
            "5min": 5,
            "15min": 15,
            "30min": 30,
            "1hour": 60,
            "2hour": 120,
            "4hour": 240,
            "8hour": 480,
            "12hour": 720,
            "1day": 1440,
            "1week": 10080,
        }

        minutes = tf_minutes.get(timeframe, 60)
        chunk_seconds = 1500 * minutes * 60

        current_start = int(start_date.timestamp())
        end_ts = int(end_date.timestamp())

        print(f"Fetching {coin} data from {start_date.date()} to {end_date.date()}...")

        while current_start < end_ts:
            chunk_end = min(current_start + chunk_seconds, end_ts)

            try:
                # KuCoin API: startAt and endAt are in seconds
                data = self.market.get_kline(
                    symbol, timeframe, startAt=current_start, endAt=chunk_end
                )

                if data:
                    for candle_data in data:
                        # KuCoin format: [timestamp, open, close, high, low, volume, turnover]
                        candle = Candle(
                            timestamp=int(candle_data[0]),
                            open=float(candle_data[1]),
                            close=float(candle_data[2]),
                            high=float(candle_data[3]),
                            low=float(candle_data[4]),
                            volume=float(candle_data[5]),
                        )
                        all_candles.append(candle)

                time.sleep(self.rate_limit_delay)

            except Exception as e:
                print(
                    f"Warning: API error at {datetime.fromtimestamp(current_start)}: {e}"
                )
                time.sleep(1)

            current_start = chunk_end

        # Sort by timestamp ascending and remove duplicates
        all_candles.sort(key=lambda c: c.timestamp)
        seen = set()
        unique_candles = []
        for c in all_candles:
            if c.timestamp not in seen:
                seen.add(c.timestamp)
                unique_candles.append(c)

        print(f"Fetched {len(unique_candles)} candles")
        return unique_candles


# =============================================================================
# BACKTEST ENGINE
# =============================================================================


class BacktestEngine:
    """
    Simulates trading strategy on historical data.
    Replicates pt_trader.py logic exactly.
    """

    def __init__(self, config: BacktestConfig):
        self.config = config
        self.predictor = PatternPredictor("")

    def _apply_fee_and_slippage(self, price: float, is_buy: bool) -> float:
        """Apply trading fees and slippage to price."""
        total_pct = self.config.fee_pct + self.config.slippage_pct
        if is_buy:
            return price * (1 + total_pct / 100)  # Pay more when buying
        else:
            return price * (1 - total_pct / 100)  # Receive less when selling

    def _should_enter(
        self, candle: Candle, recent_candles: List[Candle], position: Position
    ) -> bool:
        """Check if entry conditions are met."""
        if position.is_open:
            return False

        # Get predicted levels
        levels = self.predictor.predict_levels(recent_candles, candle.close)
        levels_below = self.predictor.count_levels_below_price(levels, candle.close)

        # Entry when price drops below N predicted support levels
        return levels_below >= self.config.trade_start_level

    def _should_dca(
        self, candle: Candle, position: Position, current_time: datetime
    ) -> Optional[float]:
        """
        Check if DCA conditions are met.
        Returns DCA level triggered, or None.
        """
        if not position.is_open:
            return None

        # Check 24h DCA limit
        recent_dcas = [
            t for t in position.dca_times if (current_time - t).total_seconds() < 86400
        ]
        if len(recent_dcas) >= self.config.max_dca_buys_per_24h:
            return None

        # Calculate current drawdown
        drawdown_pct = ((candle.close - position.avg_price) / position.avg_price) * 100

        # Check DCA levels (must be lower than last triggered level)
        triggered_levels = set()
        for level in self.config.dca_levels:
            if drawdown_pct <= level:
                triggered_levels.add(level)

        if not triggered_levels:
            return None

        # Only trigger if this is a new, lower level
        min_level = min(triggered_levels)
        already_triggered = (
            len([l for l in self.config.dca_levels if l >= min_level])
            <= position.dca_count
        )

        if not already_triggered:
            return min_level

        return None

    def _should_exit(self, candle: Candle, position: Position) -> bool:
        """Check if trailing profit margin exit conditions are met."""
        if not position.is_open:
            return False

        # Update highest price since entry
        position.highest_price_since_entry = max(
            position.highest_price_since_entry, candle.high
        )

        # Determine profit margin threshold based on DCA count
        if position.dca_count > 0:
            pm_threshold = self.config.pm_start_pct_with_dca
        else:
            pm_threshold = self.config.pm_start_pct_no_dca

        # Calculate current profit %
        profit_pct = ((candle.close - position.avg_price) / position.avg_price) * 100

        # Calculate highest profit %
        highest_profit_pct = (
            (position.highest_price_since_entry - position.avg_price)
            / position.avg_price
        ) * 100

        # Exit conditions:
        # 1. Current profit >= threshold AND
        # 2. Price dropped trailing_gap_pct from highest
        if profit_pct >= pm_threshold:
            trailing_trigger = position.highest_price_since_entry * (
                1 - self.config.trailing_gap_pct / 100
            )
            if candle.close <= trailing_trigger:
                return True

        return False

    def run(self, coin: str, candles: List[Candle]) -> BacktestResult:
        """
        Run backtest simulation on historical candles.

        Args:
            coin: Coin symbol
            candles: List of historical candles (sorted ascending)

        Returns:
            BacktestResult with performance metrics
        """
        if len(candles) < 50:
            raise ValueError("Need at least 50 candles for backtest")

        self.predictor.coin = coin
        self.predictor.load_memories()

        # Initialize state
        capital = self.config.initial_capital
        position = Position(coin=coin)
        trades: List[Trade] = []
        equity_curve: List[Tuple[datetime, float]] = []
        total_fees = 0.0
        peak_equity = capital
        max_drawdown = 0.0

        # Warmup period for pattern detection
        warmup = 50

        print(f"\nRunning backtest for {coin}...")
        print(
            f"Period: {candles[warmup].datetime.date()} to {candles[-1].datetime.date()}"
        )
        print(f"Initial capital: ${capital:,.2f}")
        print("-" * 50)

        for i in range(warmup, len(candles)):
            candle = candles[i]
            recent = candles[max(0, i - 50) : i]
            current_time = candle.datetime

            # Calculate current equity
            if position.is_open:
                position_value = position.quantity * candle.close
                equity = capital + position_value
            else:
                equity = capital

            equity_curve.append((current_time, equity))

            # Track max drawdown
            if equity > peak_equity:
                peak_equity = equity
            drawdown = (peak_equity - equity) / peak_equity * 100
            max_drawdown = max(max_drawdown, drawdown)

            # Check exit first (before entry/DCA)
            if self._should_exit(candle, position):
                exit_price = self._apply_fee_and_slippage(candle.close, is_buy=False)
                proceeds = position.quantity * exit_price
                fee = proceeds * (self.config.fee_pct / 100)
                total_fees += fee
                proceeds -= fee

                pnl = proceeds - position.cost_basis
                pnl_pct = (pnl / position.cost_basis) * 100
                holding_hours = (
                    current_time - position.entry_time
                ).total_seconds() / 3600

                trade = Trade(
                    entry_time=position.entry_time,
                    exit_time=current_time,
                    entry_price=position.avg_price,
                    exit_price=exit_price,
                    quantity=position.quantity,
                    dca_count=position.dca_count,
                    pnl=pnl,
                    pnl_pct=pnl_pct,
                    fees_paid=fee,
                    holding_hours=holding_hours,
                )
                trades.append(trade)

                capital += proceeds
                position.reset()
                continue

            # Check DCA
            dca_level = self._should_dca(candle, position, current_time)
            if dca_level is not None:
                # DCA size doubles each time
                dca_size = position.last_buy_size * self.config.dca_multiplier
                dca_size = min(dca_size, capital * 0.25)  # Max 25% of remaining capital

                if dca_size > 10 and capital >= dca_size:
                    buy_price = self._apply_fee_and_slippage(candle.close, is_buy=True)
                    fee = dca_size * (self.config.fee_pct / 100)
                    total_fees += fee
                    quantity = (dca_size - fee) / buy_price

                    position.add(quantity, buy_price, current_time)
                    position.dca_count += 1
                    position.dca_times.append(current_time)
                    capital -= dca_size

            # Check entry
            elif self._should_enter(candle, recent, position):
                entry_size = self.config.initial_capital * (
                    self.config.start_alloc_pct / 100
                )
                entry_size = min(
                    entry_size, capital * 0.5
                )  # Max 50% of current capital

                if entry_size > 10 and capital >= entry_size:
                    buy_price = self._apply_fee_and_slippage(candle.close, is_buy=True)
                    fee = entry_size * (self.config.fee_pct / 100)
                    total_fees += fee
                    quantity = (entry_size - fee) / buy_price

                    position.add(quantity, buy_price, current_time)
                    position.highest_price_since_entry = candle.high
                    capital -= entry_size

        # Close any open position at end
        if position.is_open:
            final_candle = candles[-1]
            exit_price = self._apply_fee_and_slippage(final_candle.close, is_buy=False)
            proceeds = position.quantity * exit_price
            fee = proceeds * (self.config.fee_pct / 100)
            total_fees += fee
            proceeds -= fee

            pnl = proceeds - position.cost_basis
            pnl_pct = (pnl / position.cost_basis) * 100
            holding_hours = (
                final_candle.datetime - position.entry_time
            ).total_seconds() / 3600

            trade = Trade(
                entry_time=position.entry_time,
                exit_time=final_candle.datetime,
                entry_price=position.avg_price,
                exit_price=exit_price,
                quantity=position.quantity,
                dca_count=position.dca_count,
                pnl=pnl,
                pnl_pct=pnl_pct,
                fees_paid=fee,
                holding_hours=holding_hours,
            )
            trades.append(trade)
            capital += proceeds

        # Calculate metrics
        total_trades = len(trades)
        winning_trades = len([t for t in trades if t.pnl > 0])
        losing_trades = len([t for t in trades if t.pnl <= 0])

        total_return_pct = (
            (capital - self.config.initial_capital) / self.config.initial_capital
        ) * 100

        avg_pnl = sum(t.pnl for t in trades) / total_trades if trades else 0
        avg_pnl_pct = sum(t.pnl_pct for t in trades) / total_trades if trades else 0
        win_rate = (winning_trades / total_trades * 100) if total_trades > 0 else 0
        avg_holding = (
            sum(t.holding_hours for t in trades) / total_trades if trades else 0
        )
        avg_dca = sum(t.dca_count for t in trades) / total_trades if trades else 0

        # Sharpe ratio (simplified - using daily returns)
        if len(equity_curve) > 1:
            daily_returns = []
            prev_equity = equity_curve[0][1]
            for dt, eq in equity_curve[1:]:
                daily_returns.append((eq - prev_equity) / prev_equity)
                prev_equity = eq

            if daily_returns:
                avg_return = sum(daily_returns) / len(daily_returns)
                std_return = (
                    sum((r - avg_return) ** 2 for r in daily_returns)
                    / len(daily_returns)
                ) ** 0.5
                sharpe = (avg_return / std_return * (252**0.5)) if std_return > 0 else 0
            else:
                sharpe = 0
        else:
            sharpe = 0

        return BacktestResult(
            coin=coin,
            start_date=candles[warmup].datetime.strftime("%Y-%m-%d"),
            end_date=candles[-1].datetime.strftime("%Y-%m-%d"),
            initial_capital=self.config.initial_capital,
            final_capital=capital,
            total_return_pct=total_return_pct,
            total_trades=total_trades,
            winning_trades=winning_trades,
            losing_trades=losing_trades,
            win_rate=win_rate,
            avg_trade_pnl=avg_pnl,
            avg_trade_pnl_pct=avg_pnl_pct,
            max_drawdown_pct=max_drawdown,
            sharpe_ratio=sharpe,
            avg_holding_hours=avg_holding,
            total_fees_paid=total_fees,
            avg_dca_per_trade=avg_dca,
            trades=trades,
            equity_curve=equity_curve,
        )


# =============================================================================
# REPORTING
# =============================================================================


def print_report(result: BacktestResult):
    """Print formatted backtest report."""
    print("\n" + "=" * 60)
    print(f"BACKTEST REPORT: {result.coin}")
    print("=" * 60)

    print(f"\nPeriod: {result.start_date} to {result.end_date}")
    print(f"Initial Capital:  ${result.initial_capital:>12,.2f}")
    print(f"Final Capital:    ${result.final_capital:>12,.2f}")
    print(f"Total Return:     {result.total_return_pct:>12.2f}%")

    print(f"\n{'─' * 40}")
    print("TRADE STATISTICS")
    print(f"{'─' * 40}")
    print(f"Total Trades:     {result.total_trades:>12}")
    print(f"Winning Trades:   {result.winning_trades:>12}")
    print(f"Losing Trades:    {result.losing_trades:>12}")
    print(f"Win Rate:         {result.win_rate:>11.1f}%")
    print(f"Avg P&L:          ${result.avg_trade_pnl:>11.2f}")
    print(f"Avg P&L %:        {result.avg_trade_pnl_pct:>11.2f}%")
    print(f"Avg Holding Time: {result.avg_holding_hours:>10.1f}h")
    print(f"Avg DCAs/Trade:   {result.avg_dca_per_trade:>11.1f}")

    print(f"\n{'─' * 40}")
    print("RISK METRICS")
    print(f"{'─' * 40}")
    print(f"Max Drawdown:     {result.max_drawdown_pct:>11.2f}%")
    print(f"Sharpe Ratio:     {result.sharpe_ratio:>12.2f}")
    print(f"Total Fees Paid:  ${result.total_fees_paid:>11.2f}")

    if result.trades:
        print(f"\n{'─' * 40}")
        print("RECENT TRADES (Last 10)")
        print(f"{'─' * 40}")
        for trade in result.trades[-10:]:
            status = "WIN" if trade.pnl > 0 else "LOSS"
            print(
                f"  {trade.entry_time.strftime('%Y-%m-%d')} -> "
                f"{trade.exit_time.strftime('%Y-%m-%d')}: "
                f"${trade.pnl:>8.2f} ({trade.pnl_pct:>5.1f}%) "
                f"[{status}] DCAs:{trade.dca_count}"
            )

    print("\n" + "=" * 60)


def save_results(result: BacktestResult, output_dir: str = "backtest_results"):
    """Save backtest results to JSON file."""
    Path(output_dir).mkdir(exist_ok=True)

    filename = f"{output_dir}/{result.coin}_{result.start_date}_{result.end_date}.json"

    # Convert to serializable format
    data = {
        "coin": result.coin,
        "start_date": result.start_date,
        "end_date": result.end_date,
        "initial_capital": result.initial_capital,
        "final_capital": result.final_capital,
        "total_return_pct": result.total_return_pct,
        "total_trades": result.total_trades,
        "winning_trades": result.winning_trades,
        "losing_trades": result.losing_trades,
        "win_rate": result.win_rate,
        "avg_trade_pnl": result.avg_trade_pnl,
        "avg_trade_pnl_pct": result.avg_trade_pnl_pct,
        "max_drawdown_pct": result.max_drawdown_pct,
        "sharpe_ratio": result.sharpe_ratio,
        "avg_holding_hours": result.avg_holding_hours,
        "total_fees_paid": result.total_fees_paid,
        "avg_dca_per_trade": result.avg_dca_per_trade,
        "trades": [
            {
                "entry_time": t.entry_time.isoformat(),
                "exit_time": t.exit_time.isoformat(),
                "entry_price": t.entry_price,
                "exit_price": t.exit_price,
                "quantity": t.quantity,
                "dca_count": t.dca_count,
                "pnl": t.pnl,
                "pnl_pct": t.pnl_pct,
                "fees_paid": t.fees_paid,
                "holding_hours": t.holding_hours,
            }
            for t in result.trades
        ],
    }

    with open(filename, "w") as f:
        json.dump(data, f, indent=2)

    print(f"\nResults saved to: {filename}")


# =============================================================================
# MAIN
# =============================================================================


def main():
    parser = argparse.ArgumentParser(
        description="PowerTrader AI Backtesting Engine",
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog="""
Examples:
  python pt_backtester.py BTC 2024-01-01 2024-12-31
  python pt_backtester.py ETH 2024-06-01 2024-12-31 --initial-capital 5000
  python pt_backtester.py SOL 2024-01-01 2024-06-30 --fee-pct 0.1 --save
        """,
    )

    parser.add_argument("coin", help="Coin symbol (e.g., BTC, ETH, SOL)")
    parser.add_argument("start_date", help="Start date (YYYY-MM-DD)")
    parser.add_argument("end_date", help="End date (YYYY-MM-DD)")
    parser.add_argument(
        "--initial-capital",
        type=float,
        default=10000.0,
        help="Initial capital in USD (default: 10000)",
    )
    parser.add_argument(
        "--fee-pct",
        type=float,
        default=0.075,
        help="Trading fee percentage (default: 0.075)",
    )
    parser.add_argument(
        "--slippage-pct",
        type=float,
        default=0.05,
        help="Slippage percentage (default: 0.05)",
    )
    parser.add_argument(
        "--trade-start-level",
        type=int,
        default=3,
        help="Num support levels to trigger entry (default: 3)",
    )
    parser.add_argument(
        "--timeframe", default="1hour", help="Candle timeframe (default: 1hour)"
    )
    parser.add_argument("--save", action="store_true", help="Save results to JSON file")

    args = parser.parse_args()

    # Parse dates
    try:
        start_date = datetime.strptime(args.start_date, "%Y-%m-%d")
        end_date = datetime.strptime(args.end_date, "%Y-%m-%d")
    except ValueError:
        print("ERROR: Dates must be in YYYY-MM-DD format")
        sys.exit(1)

    if start_date >= end_date:
        print("ERROR: Start date must be before end date")
        sys.exit(1)

    # Configure backtest
    config = BacktestConfig(
        initial_capital=args.initial_capital,
        fee_pct=args.fee_pct,
        slippage_pct=args.slippage_pct,
        trade_start_level=args.trade_start_level,
    )

    # Fetch data
    fetcher = KuCoinDataFetcher()
    candles = fetcher.fetch_candles(
        args.coin.upper(), start_date, end_date, args.timeframe
    )

    if len(candles) < 50:
        print(f"ERROR: Only fetched {len(candles)} candles. Need at least 50.")
        sys.exit(1)

    # Run backtest
    engine = BacktestEngine(config)
    result = engine.run(args.coin.upper(), candles)

    # Output
    print_report(result)

    if args.save:
        save_results(result)


if __name__ == "__main__":
    main()
