"""
PowerTrader AI — Multi-Exchange Trade Executor
=================================================
Executes trades across multiple exchanges with smart routing,
arbitrage execution, and liquidity aggregation for large orders.

Features:
    1. SmartRouter — routes orders to the exchange with best price
    2. ArbitrageExecutor — detects and executes cross-exchange arb
    3. LiquidityAggregator — splits large orders across exchanges
    4. ExecutionTracker — logs all executions with timestamps

Usage:
    from pt_multi_exchange import SmartRouter, ArbitrageExecutor

    router = SmartRouter()
    order = router.route_order("BTC", "buy", 0.1)

    arb = ArbitrageExecutor()
    arb.scan_and_execute(coins=["BTC", "ETH", "SOL"])
"""

from __future__ import annotations
import json
import time
from dataclasses import dataclass, asdict, field
from datetime import datetime
from typing import List, Dict, Optional, Tuple
from pathlib import Path
from enum import Enum

from pt_exchanges import ExchangeManager, ExchangeError, Ticker, OrderBook


# =============================================================================
# DATA MODELS
# =============================================================================

class OrderSide(Enum):
    BUY = "buy"
    SELL = "sell"


class OrderStatus(Enum):
    PENDING = "pending"
    PARTIAL = "partial"
    FILLED = "filled"
    FAILED = "failed"
    CANCELLED = "cancelled"


@dataclass
class ExecutionLeg:
    """A single leg of an order placed on one exchange."""
    exchange: str
    side: str
    coin: str
    quantity: float
    price: float
    status: str = "pending"
    fill_price: float = 0.0
    fill_quantity: float = 0.0
    fee: float = 0.0
    timestamp: str = ""
    order_id: str = ""


@dataclass
class ExecutionResult:
    """Result of an order execution (may span multiple exchanges)."""
    order_id: str
    coin: str
    side: str
    total_quantity: float
    avg_fill_price: float
    total_fees: float
    legs: List[ExecutionLeg]
    status: str
    slippage_pct: float = 0.0
    timestamp: str = ""

    def to_dict(self) -> dict:
        return {
            "order_id": self.order_id,
            "coin": self.coin,
            "side": self.side,
            "total_quantity": self.total_quantity,
            "avg_fill_price": self.avg_fill_price,
            "total_fees": self.total_fees,
            "status": self.status,
            "slippage_pct": self.slippage_pct,
            "timestamp": self.timestamp,
            "legs": [asdict(l) for l in self.legs],
        }


@dataclass
class ArbitrageOpportunity:
    """A detected arbitrage opportunity."""
    coin: str
    buy_exchange: str
    buy_price: float
    sell_exchange: str
    sell_price: float
    spread_pct: float
    estimated_profit: float
    max_quantity: float
    timestamp: str = ""


# =============================================================================
# FEE SCHEDULE
# =============================================================================

# Default taker fees per exchange (in %)
EXCHANGE_FEES = {
    "kucoin": 0.10,
    "binance": 0.10,
    "coinbase": 0.40,
}

# Minimum profitable spread must exceed 2x fees
def min_profitable_spread(buy_exchange: str, sell_exchange: str) -> float:
    buy_fee = EXCHANGE_FEES.get(buy_exchange, 0.10)
    sell_fee = EXCHANGE_FEES.get(sell_exchange, 0.10)
    return buy_fee + sell_fee


# =============================================================================
# SMART ORDER ROUTER
# =============================================================================

class SmartRouter:
    """
    Routes orders to the exchange offering the best price.
    Considers: best bid/ask, fees, available liquidity.
    """

    def __init__(self, manager: Optional[ExchangeManager] = None):
        self.manager = manager or ExchangeManager()
        self.execution_log: List[dict] = []

    def find_best_exchange(self, coin: str, side: str) -> Tuple[str, float]:
        """
        Find the exchange with the best price for the given side.
        Buy → lowest ask. Sell → highest bid.
        """
        tickers = self.manager.get_all_tickers(coin)
        if not tickers:
            raise ExchangeError(f"No tickers available for {coin}")

        best_exchange = None
        best_price = float("inf") if side == "buy" else 0.0

        for ex_name, ticker in tickers.items():
            fee_pct = EXCHANGE_FEES.get(ex_name, 0.10) / 100.0

            if side == "buy":
                effective_price = ticker.ask * (1 + fee_pct)
                if effective_price < best_price:
                    best_price = effective_price
                    best_exchange = ex_name
            else:
                effective_price = ticker.bid * (1 - fee_pct)
                if effective_price > best_price:
                    best_price = effective_price
                    best_exchange = ex_name

        if best_exchange is None:
            raise ExchangeError(f"Could not find best exchange for {coin} {side}")

        return best_exchange, best_price

    def route_order(self, coin: str, side: str, quantity: float,
                    max_slippage_pct: float = 0.5) -> ExecutionResult:
        """
        Route an order to the best exchange.
        Simulates execution (paper trading mode by default).
        """
        order_id = f"smart_{int(time.time() * 1000)}"
        best_exchange, best_price = self.find_best_exchange(coin, side)

        fee_pct = EXCHANGE_FEES.get(best_exchange, 0.10) / 100.0
        fees = quantity * best_price * fee_pct

        leg = ExecutionLeg(
            exchange=best_exchange,
            side=side,
            coin=coin,
            quantity=quantity,
            price=best_price,
            status="filled",
            fill_price=best_price,
            fill_quantity=quantity,
            fee=fees,
            timestamp=datetime.now().isoformat(),
            order_id=order_id,
        )

        result = ExecutionResult(
            order_id=order_id,
            coin=coin,
            side=side,
            total_quantity=quantity,
            avg_fill_price=best_price,
            total_fees=fees,
            legs=[leg],
            status="filled",
            slippage_pct=0.0,
            timestamp=datetime.now().isoformat(),
        )

        self._log_execution(result)
        return result

    def compare_routes(self, coin: str, side: str, quantity: float) -> List[dict]:
        """Compare execution cost across all exchanges."""
        tickers = self.manager.get_all_tickers(coin)
        routes = []

        for ex_name, ticker in tickers.items():
            fee_pct = EXCHANGE_FEES.get(ex_name, 0.10) / 100.0
            price = ticker.ask if side == "buy" else ticker.bid

            total_cost = quantity * price
            fees = total_cost * fee_pct
            effective_cost = total_cost + fees if side == "buy" else total_cost - fees

            routes.append({
                "exchange": ex_name,
                "price": price,
                "fee_pct": fee_pct * 100,
                "fees": fees,
                "total_cost": effective_cost,
                "volume_24h": ticker.volume_24h,
            })

        routes.sort(key=lambda r: r["total_cost"], reverse=(side == "sell"))
        return routes

    def _log_execution(self, result: ExecutionResult):
        self.execution_log.append(result.to_dict())
        # Persist to file
        log_path = Path("execution_log.json")
        try:
            existing = json.loads(log_path.read_text()) if log_path.exists() else []
            existing.append(result.to_dict())
            log_path.write_text(json.dumps(existing[-500:], indent=2))  # keep last 500
        except Exception:
            pass


# =============================================================================
# ARBITRAGE EXECUTOR
# =============================================================================

class ArbitrageExecutor:
    """
    Detects and executes cross-exchange arbitrage opportunities.
    Safety checks: minimum spread threshold, fee deduction, max position size.
    """

    def __init__(self, manager: Optional[ExchangeManager] = None,
                 min_spread_pct: float = 0.3,
                 max_position_usd: float = 1000.0):
        self.manager = manager or ExchangeManager()
        self.min_spread_pct = min_spread_pct
        self.max_position_usd = max_position_usd
        self.opportunities: List[ArbitrageOpportunity] = []
        self.executed: List[dict] = []

    def scan(self, coins: List[str]) -> List[ArbitrageOpportunity]:
        """Scan all coins for arbitrage opportunities."""
        self.opportunities = []

        for coin in coins:
            opp = self._check_coin(coin)
            if opp:
                self.opportunities.append(opp)

        return self.opportunities

    def _check_coin(self, coin: str) -> Optional[ArbitrageOpportunity]:
        """Check a single coin for arbitrage across exchanges."""
        try:
            tickers = self.manager.get_all_tickers(coin)
        except Exception:
            return None

        if len(tickers) < 2:
            return None

        # Find best buy (lowest ask) and best sell (highest bid)
        best_buy_ex = None
        best_buy_price = float("inf")
        best_sell_ex = None
        best_sell_price = 0.0

        for ex_name, ticker in tickers.items():
            if ticker.ask > 0 and ticker.ask < best_buy_price:
                best_buy_price = ticker.ask
                best_buy_ex = ex_name
            if ticker.bid > 0 and ticker.bid > best_sell_price:
                best_sell_price = ticker.bid
                best_sell_ex = ex_name

        if best_buy_ex is None or best_sell_ex is None:
            return None
        if best_buy_ex == best_sell_ex:
            return None

        # Calculate spread after fees
        gross_spread_pct = ((best_sell_price - best_buy_price) / best_buy_price) * 100
        fee_cost_pct = min_profitable_spread(best_buy_ex, best_sell_ex)
        net_spread_pct = gross_spread_pct - fee_cost_pct

        if net_spread_pct < self.min_spread_pct:
            return None

        # Max quantity limited by position size
        max_qty = self.max_position_usd / best_buy_price
        estimated_profit = max_qty * (best_sell_price - best_buy_price) * (1 - fee_cost_pct / 100)

        return ArbitrageOpportunity(
            coin=coin,
            buy_exchange=best_buy_ex,
            buy_price=best_buy_price,
            sell_exchange=best_sell_ex,
            sell_price=best_sell_price,
            spread_pct=net_spread_pct,
            estimated_profit=estimated_profit,
            max_quantity=max_qty,
            timestamp=datetime.now().isoformat(),
        )

    def execute_opportunity(self, opp: ArbitrageOpportunity) -> dict:
        """
        Execute an arbitrage opportunity (paper trading).
        In production, this would place real orders via exchange APIs.
        """
        buy_fee = EXCHANGE_FEES.get(opp.buy_exchange, 0.10) / 100.0
        sell_fee = EXCHANGE_FEES.get(opp.sell_exchange, 0.10) / 100.0

        buy_cost = opp.max_quantity * opp.buy_price * (1 + buy_fee)
        sell_revenue = opp.max_quantity * opp.sell_price * (1 - sell_fee)
        net_profit = sell_revenue - buy_cost

        result = {
            "type": "arbitrage",
            "coin": opp.coin,
            "buy_exchange": opp.buy_exchange,
            "buy_price": opp.buy_price,
            "sell_exchange": opp.sell_exchange,
            "sell_price": opp.sell_price,
            "quantity": opp.max_quantity,
            "buy_cost": buy_cost,
            "sell_revenue": sell_revenue,
            "net_profit": net_profit,
            "spread_pct": opp.spread_pct,
            "status": "simulated",
            "timestamp": datetime.now().isoformat(),
        }

        self.executed.append(result)
        return result

    def scan_and_execute(self, coins: List[str]) -> List[dict]:
        """Scan all coins and execute profitable arbitrage."""
        opportunities = self.scan(coins)
        results = []

        for opp in opportunities:
            if opp.estimated_profit > 1.0:  # Min $1 profit
                result = self.execute_opportunity(opp)
                results.append(result)

        return results


# =============================================================================
# LIQUIDITY AGGREGATOR
# =============================================================================

class LiquidityAggregator:
    """
    Splits large orders across multiple exchanges to minimize market impact.
    Uses order book depth to determine optimal allocation.
    """

    def __init__(self, manager: Optional[ExchangeManager] = None,
                 max_impact_pct: float = 0.1):
        self.manager = manager or ExchangeManager()
        self.max_impact_pct = max_impact_pct

    def get_liquidity_map(self, coin: str, side: str, depth: int = 20) -> Dict[str, float]:
        """Get available liquidity (in USD) at each exchange."""
        liquidity = {}

        for ex_name in self.manager.exchanges:
            try:
                q = "USD" if ex_name == "coinbase" else "USDT"
                ex = self.manager.exchanges[ex_name]
                symbol = ex.normalize_symbol(coin, q)
                book = ex.get_orderbook(symbol, depth)

                if side == "buy":
                    # Sum ask-side liquidity
                    total = sum(price * qty for price, qty in book.asks)
                else:
                    # Sum bid-side liquidity
                    total = sum(price * qty for price, qty in book.bids)

                liquidity[ex_name] = total
            except Exception:
                liquidity[ex_name] = 0.0

        return liquidity

    def split_order(self, coin: str, side: str, total_quantity: float,
                    total_usd: float = 0.0) -> List[ExecutionLeg]:
        """
        Split a large order across exchanges proportional to their liquidity.
        Returns a list of ExecutionLegs.
        """
        liquidity = self.get_liquidity_map(coin, side)
        total_liquidity = sum(liquidity.values())

        if total_liquidity == 0:
            raise ExchangeError(f"No liquidity available for {coin} {side}")

        legs = []
        for ex_name, liq in liquidity.items():
            if liq <= 0:
                continue

            # Allocate proportionally
            share = liq / total_liquidity
            leg_quantity = total_quantity * share

            if leg_quantity <= 0:
                continue

            # Get current price for this exchange
            try:
                ticker = self.manager.get_ticker(coin, ex_name)
                price = ticker.ask if side == "buy" else ticker.bid
            except Exception:
                continue

            fee_pct = EXCHANGE_FEES.get(ex_name, 0.10) / 100.0
            fees = leg_quantity * price * fee_pct

            legs.append(ExecutionLeg(
                exchange=ex_name,
                side=side,
                coin=coin,
                quantity=round(leg_quantity, 8),
                price=price,
                status="pending",
                fee=fees,
                timestamp=datetime.now().isoformat(),
            ))

        return legs

    def estimate_impact(self, coin: str, side: str, quantity: float) -> Dict[str, float]:
        """
        Estimate price impact of an order on each exchange.
        Returns impact percentage per exchange.
        """
        impacts = {}

        for ex_name in self.manager.exchanges:
            try:
                q = "USD" if ex_name == "coinbase" else "USDT"
                ex = self.manager.exchanges[ex_name]
                symbol = ex.normalize_symbol(coin, q)
                book = ex.get_orderbook(symbol, 50)

                levels = book.asks if side == "buy" else book.bids
                if not levels:
                    impacts[ex_name] = 999.0  # no liquidity
                    continue

                best_price = levels[0][0]
                remaining = quantity
                worst_price = best_price

                for price, qty in levels:
                    if remaining <= 0:
                        break
                    worst_price = price
                    remaining -= qty

                impact_pct = abs(worst_price - best_price) / best_price * 100
                impacts[ex_name] = round(impact_pct, 4)
            except Exception:
                impacts[ex_name] = -1  # error

        return impacts


# =============================================================================
# EXECUTION TRACKER
# =============================================================================

class ExecutionTracker:
    """Tracks all executions and provides analytics."""

    LOG_FILE = "execution_log.json"

    def __init__(self):
        self.executions: List[dict] = self._load()

    def _load(self) -> List[dict]:
        path = Path(self.LOG_FILE)
        if path.exists():
            try:
                return json.loads(path.read_text())
            except Exception:
                return []
        return []

    def _save(self):
        Path(self.LOG_FILE).write_text(json.dumps(self.executions[-1000:], indent=2))

    def log(self, execution: dict):
        self.executions.append(execution)
        self._save()

    def get_stats(self) -> dict:
        if not self.executions:
            return {"total": 0}

        total = len(self.executions)
        by_exchange = {}
        total_fees = 0.0
        total_volume = 0.0

        for ex in self.executions:
            for leg in ex.get("legs", []):
                exchange = leg.get("exchange", "unknown")
                by_exchange[exchange] = by_exchange.get(exchange, 0) + 1
                total_fees += leg.get("fee", 0)
                total_volume += leg.get("quantity", 0) * leg.get("price", 0)

        # Arbitrage stats
        arb_executions = [e for e in self.executions if e.get("type") == "arbitrage"]
        arb_profit = sum(e.get("net_profit", 0) for e in arb_executions)

        return {
            "total_executions": total,
            "by_exchange": by_exchange,
            "total_fees": round(total_fees, 2),
            "total_volume": round(total_volume, 2),
            "arbitrage_count": len(arb_executions),
            "arbitrage_profit": round(arb_profit, 2),
        }


# =============================================================================
# SELF-TEST
# =============================================================================

if __name__ == "__main__":
    print("=" * 60)
    print("Multi-Exchange Trading — Self-Test")
    print("=" * 60)

    # Test SmartRouter
    print("\n1. SmartRouter — compare_routes (simulated)...")
    router = SmartRouter()
    try:
        routes = router.compare_routes("BTC", "buy", 0.01)
        if routes:
            for r in routes:
                print(f"   {r['exchange']:>10}: ${r['price']:>12,.2f}  fee: {r['fee_pct']:.2f}%  total: ${r['total_cost']:>10,.2f}")
        else:
            print("   No routes available (offline)")
    except Exception as e:
        print(f"   Skipped (no network): {e}")

    # Test ArbitrageExecutor
    print("\n2. ArbitrageExecutor — scan...")
    arb = ArbitrageExecutor(min_spread_pct=0.1)
    try:
        opps = arb.scan(["BTC", "ETH"])
        print(f"   Found {len(opps)} opportunities")
        for o in opps:
            print(f"   {o.coin}: buy {o.buy_exchange} @ ${o.buy_price:,.2f} → sell {o.sell_exchange} @ ${o.sell_price:,.2f} ({o.spread_pct:.3f}%)")
    except Exception as e:
        print(f"   Skipped (no network): {e}")

    # Test ExecutionTracker
    print("\n3. ExecutionTracker — stats...")
    tracker = ExecutionTracker()
    stats = tracker.get_stats()
    print(f"   {json.dumps(stats, indent=4)}")

    # Test LiquidityAggregator split logic
    print("\n4. LiquidityAggregator — order splitting...")
    agg = LiquidityAggregator()
    try:
        legs = agg.split_order("BTC", "buy", 0.5)
        for leg in legs:
            print(f"   {leg.exchange}: {leg.quantity:.4f} BTC @ ${leg.price:,.2f} (fee: ${leg.fee:.2f})")
    except Exception as e:
        print(f"   Skipped (no network): {e}")

    print("\n✅ Self-test complete")
