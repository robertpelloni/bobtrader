#!/usr/bin/env python3
"""
PowerTrader Thinker - Multi-Exchange Price Integration
=======================================================
Adds multi-exchange price fetching and aggregation to pt_thinker.py.

This module provides:
- ExchangeManager from pt_exchanges.py with KuCoin, Binance, Coinbase
- get_aggregated_current_price() - median/mean/VWAP across exchanges
- get_candle_from_exchanges() - fetch candles with fallback
- detect_arbitrage_opportunities() - find cross-exchange price differences

Integration pattern:
- KuCoin remains primary source for consistency
- Binance and Coinbase are fallbacks
- Robinhood current price fetch is unchanged (still used for execution price)
"""

from pt_exchanges import ExchangeManager
import time

_exchange_manager = None


def init_exchanges():
    global _exchange_manager
    if _exchange_manager is None:
        _exchange_manager = ExchangeManager(
            enabled_exchanges=["kucoin", "binance", "coinbase"]
        )
        print(
            "[Exchange] Initialized multi-exchange manager: KuCoin (primary), Binance, Coinbase"
        )


def get_aggregated_current_price(coin_symbol, method="median"):
    global _exchange_manager
    if _exchange_manager is None:
        init_exchanges()

    try:
        agg = _exchange_manager.get_aggregated_price(coin=coin_symbol, method=method)
        if agg:
            spread_pct = agg.get("spread_pct", 0.0)
            if spread_pct > 0.5:
                print(
                    f"[Arbitrage] {coin_symbol}: {spread_pct:.2f}% spread detected across exchanges"
                )
            return agg["aggregated_price"]
    except Exception as e:
        print(f"[Exchange] Error fetching aggregated price for {coin_symbol}: {e}")
        return None


def get_candle_from_exchanges(coin_symbol, timeframe, exchange="kucoin"):
    global _exchange_manager
    if _exchange_manager is None:
        init_exchanges()

    exchanges_to_try = [exchange]
    if exchange == "kucoin":
        exchanges_to_try.extend(["binance", "coinbase"])

    for ex in exchanges_to_try:
        try:
            candles = _exchange_manager.get_candles(
                coin=coin_symbol.replace("-USDT", ""),
                exchange=ex,
                timeframe=timeframe,
                limit=1,
            )
            if candles:
                return candles[0]
        except Exception as e:
            if ex == exchange:
                print(f"Warning: {ex} candle fetch failed, trying fallback...")
            continue

    return None


def detect_arbitrage_opportunities(coin_symbol, min_spread_pct=0.3):
    global _exchange_manager
    if _exchange_manager is None:
        init_exchanges()

    try:
        arb = _exchange_manager.detect_arbitrage(
            coin_symbol, min_spread_pct=min_spread_pct
        )
        if arb:
            print(f"\n[ARBITRAGE OPPORTUNITY]")
            print(f"  Coin: {arb['coin']}")
            print(f"  Buy: {arb['buy_exchange']} @ ${arb['buy_price']:,.2f}")
            print(f"  Sell: {arb['sell_exchange']} @ ${arb['sell_price']:,.2f}")
            print(f"  Spread: {arb['spread_pct']:.2f}%")
        return arb
    except Exception as e:
        print(f"[Exchange] Error checking arbitrage: {e}")
        return None
