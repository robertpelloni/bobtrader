#!/usr/bin/env python3
"""
PowerTrader AI - Multi-Exchange Data Integration
=================================================
Unified interface for fetching market data from multiple exchanges.
Supports KuCoin, Binance, and Coinbase with price aggregation.

Usage:
    from pt_exchanges import ExchangeManager, get_aggregated_price

    manager = ExchangeManager()

    # Single exchange
    price = manager.get_price("BTC", exchange="binance")
    candles = manager.get_candles("ETH", timeframe="1hour", limit=100)

    # Aggregated across exchanges
    agg_price = manager.get_aggregated_price("BTC")

    # CLI
    python pt_exchanges.py price BTC
    python pt_exchanges.py candles ETH --timeframe 1hour --limit 50
    python pt_exchanges.py compare BTC
"""

import time
import hmac
import hashlib
import requests
from abc import ABC, abstractmethod
from dataclasses import dataclass
from datetime import datetime, timedelta
from typing import List, Dict, Optional, Tuple, Any
from enum import Enum
import argparse
import json
import statistics


class ExchangeType(Enum):
    KUCOIN = "kucoin"
    BINANCE = "binance"
    COINBASE = "coinbase"


@dataclass
class Ticker:
    exchange: str
    symbol: str
    price: float
    bid: float
    ask: float
    volume_24h: float
    timestamp: datetime


@dataclass
class OHLCV:
    timestamp: int
    open: float
    high: float
    low: float
    close: float
    volume: float

    @property
    def datetime(self) -> datetime:
        return datetime.fromtimestamp(self.timestamp)


@dataclass
class OrderBook:
    exchange: str
    symbol: str
    bids: List[Tuple[float, float]]  # [(price, quantity), ...]
    asks: List[Tuple[float, float]]
    timestamp: datetime


class ExchangeBase(ABC):
    def __init__(self):
        self.rate_limit_delay = 0.1
        self.last_request_time = 0
        self.session = requests.Session()
        self.session.headers.update({"User-Agent": "PowerTrader-AI/1.0"})

    def _rate_limit(self):
        elapsed = time.time() - self.last_request_time
        if elapsed < self.rate_limit_delay:
            time.sleep(self.rate_limit_delay - elapsed)
        self.last_request_time = time.time()

    def _request(self, method: str, url: str, **kwargs) -> dict:
        self._rate_limit()
        try:
            resp = self.session.request(method, url, timeout=10, **kwargs)
            resp.raise_for_status()
            return resp.json()
        except requests.RequestException as e:
            raise ExchangeError(f"{self.__class__.__name__} request failed: {e}")

    @abstractmethod
    def get_ticker(self, symbol: str) -> Ticker:
        pass

    @abstractmethod
    def get_candles(self, symbol: str, timeframe: str, limit: int) -> List[OHLCV]:
        pass

    @abstractmethod
    def get_orderbook(self, symbol: str, depth: int) -> OrderBook:
        pass

    @abstractmethod
    def normalize_symbol(self, coin: str, quote: str) -> str:
        pass

    @abstractmethod
    def normalize_timeframe(self, tf: str) -> str:
        pass


class ExchangeError(Exception):
    pass


class KuCoinExchange(ExchangeBase):
    BASE_URL = "https://api.kucoin.com"

    TIMEFRAME_MAP = {
        "1min": "1min",
        "5min": "5min",
        "15min": "15min",
        "30min": "30min",
        "1hour": "1hour",
        "2hour": "2hour",
        "4hour": "4hour",
        "8hour": "8hour",
        "12hour": "12hour",
        "1day": "1day",
        "1week": "1week",
    }

    def normalize_symbol(self, coin: str, quote: str = "USDT") -> str:
        return f"{coin.upper()}-{quote.upper()}"

    def normalize_timeframe(self, tf: str) -> str:
        return self.TIMEFRAME_MAP.get(tf, "1hour")

    def get_ticker(self, symbol: str) -> Ticker:
        data = self._request(
            "GET", f"{self.BASE_URL}/api/v1/market/stats", params={"symbol": symbol}
        )

        if data.get("code") != "200000":
            raise ExchangeError(f"KuCoin error: {data.get('msg', 'Unknown error')}")

        stats = data["data"]
        return Ticker(
            exchange="kucoin",
            symbol=symbol,
            price=float(stats["last"]),
            bid=float(stats["buy"] or 0),
            ask=float(stats["sell"] or 0),
            volume_24h=float(stats["vol"]),
            timestamp=datetime.now(),
        )

    def get_candles(
        self,
        symbol: str,
        timeframe: str = "1hour",
        limit: int = 100,
        start_time: Optional[int] = None,
        end_time: Optional[int] = None,
    ) -> List[OHLCV]:
        params = {"symbol": symbol, "type": self.normalize_timeframe(timeframe)}

        if start_time:
            params["startAt"] = start_time
        if end_time:
            params["endAt"] = end_time

        data = self._request(
            "GET", f"{self.BASE_URL}/api/v1/market/candles", params=params
        )

        if data.get("code") != "200000":
            raise ExchangeError(f"KuCoin error: {data.get('msg', 'Unknown error')}")

        candles = []
        for c in data["data"][:limit]:
            candles.append(
                OHLCV(
                    timestamp=int(c[0]),
                    open=float(c[1]),
                    close=float(c[2]),
                    high=float(c[3]),
                    low=float(c[4]),
                    volume=float(c[5]),
                )
            )

        return sorted(candles, key=lambda x: x.timestamp)

    def get_orderbook(self, symbol: str, depth: int = 20) -> OrderBook:
        data = self._request(
            "GET",
            f"{self.BASE_URL}/api/v1/market/orderbook/level2_20",
            params={"symbol": symbol},
        )

        if data.get("code") != "200000":
            raise ExchangeError(f"KuCoin error: {data.get('msg', 'Unknown error')}")

        book = data["data"]
        return OrderBook(
            exchange="kucoin",
            symbol=symbol,
            bids=[(float(b[0]), float(b[1])) for b in book["bids"][:depth]],
            asks=[(float(a[0]), float(a[1])) for a in book["asks"][:depth]],
            timestamp=datetime.now(),
        )


class BinanceExchange(ExchangeBase):
    BASE_URL = "https://api.binance.com"

    TIMEFRAME_MAP = {
        "1min": "1m",
        "5min": "5m",
        "15min": "15m",
        "30min": "30m",
        "1hour": "1h",
        "2hour": "2h",
        "4hour": "4h",
        "8hour": "8h",
        "12hour": "12h",
        "1day": "1d",
        "1week": "1w",
    }

    def normalize_symbol(self, coin: str, quote: str = "USDT") -> str:
        return f"{coin.upper()}{quote.upper()}"

    def normalize_timeframe(self, tf: str) -> str:
        return self.TIMEFRAME_MAP.get(tf, "1h")

    def get_ticker(self, symbol: str) -> Ticker:
        data = self._request(
            "GET", f"{self.BASE_URL}/api/v3/ticker/24hr", params={"symbol": symbol}
        )

        return Ticker(
            exchange="binance",
            symbol=symbol,
            price=float(data["lastPrice"]),
            bid=float(data["bidPrice"]),
            ask=float(data["askPrice"]),
            volume_24h=float(data["volume"]),
            timestamp=datetime.now(),
        )

    def get_candles(
        self,
        symbol: str,
        timeframe: str = "1hour",
        limit: int = 100,
        start_time: Optional[int] = None,
        end_time: Optional[int] = None,
    ) -> List[OHLCV]:
        params = {
            "symbol": symbol,
            "interval": self.normalize_timeframe(timeframe),
            "limit": min(limit, 1000),
        }

        if start_time:
            params["startTime"] = start_time * 1000
        if end_time:
            params["endTime"] = end_time * 1000

        data = self._request("GET", f"{self.BASE_URL}/api/v3/klines", params=params)

        candles = []
        for c in data:
            candles.append(
                OHLCV(
                    timestamp=int(c[0] // 1000),
                    open=float(c[1]),
                    high=float(c[2]),
                    low=float(c[3]),
                    close=float(c[4]),
                    volume=float(c[5]),
                )
            )

        return candles

    def get_orderbook(self, symbol: str, depth: int = 20) -> OrderBook:
        data = self._request(
            "GET",
            f"{self.BASE_URL}/api/v3/depth",
            params={"symbol": symbol, "limit": depth},
        )

        return OrderBook(
            exchange="binance",
            symbol=symbol,
            bids=[(float(b[0]), float(b[1])) for b in data["bids"]],
            asks=[(float(a[0]), float(a[1])) for a in data["asks"]],
            timestamp=datetime.now(),
        )


class CoinbaseExchange(ExchangeBase):
    BASE_URL = "https://api.exchange.coinbase.com"

    TIMEFRAME_MAP = {
        "1min": 60,
        "5min": 300,
        "15min": 900,
        "30min": 1800,
        "1hour": 3600,
        "4hour": 14400,
        "1day": 86400,
    }

    def normalize_symbol(self, coin: str, quote: str = "USD") -> str:
        return f"{coin.upper()}-{quote.upper()}"

    def normalize_timeframe(self, tf: str) -> int:
        return self.TIMEFRAME_MAP.get(tf, 3600)

    def get_ticker(self, symbol: str) -> Ticker:
        ticker_data = self._request("GET", f"{self.BASE_URL}/products/{symbol}/ticker")
        stats_data = self._request("GET", f"{self.BASE_URL}/products/{symbol}/stats")

        return Ticker(
            exchange="coinbase",
            symbol=symbol,
            price=float(ticker_data["price"]),
            bid=float(ticker_data["bid"]),
            ask=float(ticker_data["ask"]),
            volume_24h=float(stats_data["volume"]),
            timestamp=datetime.now(),
        )

    def get_candles(
        self,
        symbol: str,
        timeframe: str = "1hour",
        limit: int = 100,
        start_time: Optional[int] = None,
        end_time: Optional[int] = None,
    ) -> List[OHLCV]:
        granularity = self.normalize_timeframe(timeframe)

        params = {"granularity": granularity}
        if start_time:
            params["start"] = datetime.fromtimestamp(start_time).isoformat()
        if end_time:
            params["end"] = datetime.fromtimestamp(end_time).isoformat()

        data = self._request(
            "GET", f"{self.BASE_URL}/products/{symbol}/candles", params=params
        )

        candles = []
        for c in data[:limit]:
            candles.append(
                OHLCV(
                    timestamp=int(c[0]),
                    low=float(c[1]),
                    high=float(c[2]),
                    open=float(c[3]),
                    close=float(c[4]),
                    volume=float(c[5]),
                )
            )

        return sorted(candles, key=lambda x: x.timestamp)

    def get_orderbook(self, symbol: str, depth: int = 20) -> OrderBook:
        data = self._request(
            "GET", f"{self.BASE_URL}/products/{symbol}/book", params={"level": 2}
        )

        return OrderBook(
            exchange="coinbase",
            symbol=symbol,
            bids=[(float(b[0]), float(b[1])) for b in data["bids"][:depth]],
            asks=[(float(a[0]), float(a[1])) for a in data["asks"][:depth]],
            timestamp=datetime.now(),
        )


class ExchangeManager:
    def __init__(self, enabled_exchanges: Optional[List[str]] = None):
        self.exchanges: Dict[str, ExchangeBase] = {}

        available = {
            "kucoin": KuCoinExchange,
            "binance": BinanceExchange,
            "coinbase": CoinbaseExchange,
        }

        if enabled_exchanges is None:
            enabled_exchanges = ["kucoin", "binance", "coinbase"]

        for name in enabled_exchanges:
            if name in available:
                try:
                    self.exchanges[name] = available[name]()
                except Exception as e:
                    print(f"Warning: Could not initialize {name}: {e}")

    def get_price(
        self, coin: str, exchange: str = "kucoin", quote: str = "USDT"
    ) -> float:
        if exchange not in self.exchanges:
            raise ExchangeError(f"Exchange {exchange} not available")

        ex = self.exchanges[exchange]
        q = "USD" if exchange == "coinbase" else quote
        symbol = ex.normalize_symbol(coin, q)
        ticker = ex.get_ticker(symbol)
        return ticker.price

    def get_ticker(
        self, coin: str, exchange: str = "kucoin", quote: str = "USDT"
    ) -> Ticker:
        if exchange not in self.exchanges:
            raise ExchangeError(f"Exchange {exchange} not available")

        ex = self.exchanges[exchange]
        q = "USD" if exchange == "coinbase" else quote
        symbol = ex.normalize_symbol(coin, q)
        return ex.get_ticker(symbol)

    def get_candles(
        self,
        coin: str,
        exchange: str = "kucoin",
        timeframe: str = "1hour",
        limit: int = 100,
        quote: str = "USDT",
    ) -> List[OHLCV]:
        if exchange not in self.exchanges:
            raise ExchangeError(f"Exchange {exchange} not available")

        ex = self.exchanges[exchange]
        q = "USD" if exchange == "coinbase" else quote
        symbol = ex.normalize_symbol(coin, q)
        return ex.get_candles(symbol, timeframe, limit)

    def get_orderbook(
        self, coin: str, exchange: str = "kucoin", depth: int = 20, quote: str = "USDT"
    ) -> OrderBook:
        if exchange not in self.exchanges:
            raise ExchangeError(f"Exchange {exchange} not available")

        ex = self.exchanges[exchange]
        q = "USD" if exchange == "coinbase" else quote
        symbol = ex.normalize_symbol(coin, q)
        return ex.get_orderbook(symbol, depth)

    def get_all_tickers(self, coin: str, quote: str = "USDT") -> Dict[str, Ticker]:
        results = {}
        for name, ex in self.exchanges.items():
            try:
                q = "USD" if name == "coinbase" else quote
                symbol = ex.normalize_symbol(coin, q)
                results[name] = ex.get_ticker(symbol)
            except Exception as e:
                print(f"Warning: {name} ticker failed: {e}")
        return results

    def get_aggregated_price(self, coin: str, method: str = "median") -> Dict[str, Any]:
        tickers = self.get_all_tickers(coin)

        if not tickers:
            raise ExchangeError(f"No price data available for {coin}")

        prices = [t.price for t in tickers.values()]
        volumes = [t.volume_24h for t in tickers.values()]

        if method == "median":
            agg_price = statistics.median(prices)
        elif method == "mean":
            agg_price = statistics.mean(prices)
        elif method == "vwap":
            total_volume = sum(volumes)
            if total_volume > 0:
                agg_price = sum(p * v for p, v in zip(prices, volumes)) / total_volume
            else:
                agg_price = statistics.median(prices)
        else:
            agg_price = statistics.median(prices)

        spread = max(prices) - min(prices)
        spread_pct = (spread / agg_price) * 100 if agg_price > 0 else 0

        return {
            "coin": coin,
            "aggregated_price": agg_price,
            "method": method,
            "spread": spread,
            "spread_pct": spread_pct,
            "exchange_prices": {name: t.price for name, t in tickers.items()},
            "timestamp": datetime.now().isoformat(),
        }

    def detect_arbitrage(
        self, coin: str, min_spread_pct: float = 0.5
    ) -> Optional[Dict[str, Any]]:
        tickers = self.get_all_tickers(coin)

        if len(tickers) < 2:
            return None

        sorted_by_price = sorted(tickers.items(), key=lambda x: x[1].price)
        lowest_name, lowest_ticker = sorted_by_price[0]
        highest_name, highest_ticker = sorted_by_price[-1]

        spread = highest_ticker.price - lowest_ticker.price
        spread_pct = (spread / lowest_ticker.price) * 100

        if spread_pct >= min_spread_pct:
            return {
                "coin": coin,
                "buy_exchange": lowest_name,
                "buy_price": lowest_ticker.price,
                "sell_exchange": highest_name,
                "sell_price": highest_ticker.price,
                "spread": spread,
                "spread_pct": spread_pct,
                "timestamp": datetime.now().isoformat(),
            }

        return None


def print_price_comparison(manager: ExchangeManager, coin: str):
    print(f"\n{'=' * 60}")
    print(f"PRICE COMPARISON: {coin}")
    print("=" * 60)

    agg = manager.get_aggregated_price(coin)

    print(f"\nAggregated Price (median): ${agg['aggregated_price']:,.2f}")
    print(f"Cross-exchange spread: ${agg['spread']:.2f} ({agg['spread_pct']:.3f}%)")

    print(f"\n{'Exchange':<12} {'Price':>14} {'vs Median':>12}")
    print("-" * 40)

    for ex_name, price in agg["exchange_prices"].items():
        diff = price - agg["aggregated_price"]
        diff_pct = (diff / agg["aggregated_price"]) * 100
        sign = "+" if diff >= 0 else ""
        print(f"{ex_name:<12} ${price:>13,.2f} {sign}{diff_pct:>10.3f}%")

    arb = manager.detect_arbitrage(coin)
    if arb:
        print(f"\n{'!' * 40}")
        print(f"ARBITRAGE OPPORTUNITY DETECTED!")
        print(f"Buy on {arb['buy_exchange']} at ${arb['buy_price']:,.2f}")
        print(f"Sell on {arb['sell_exchange']} at ${arb['sell_price']:,.2f}")
        print(f"Potential profit: {arb['spread_pct']:.3f}%")
        print(f"{'!' * 40}")


def main():
    parser = argparse.ArgumentParser(description="PowerTrader Multi-Exchange Interface")
    subparsers = parser.add_subparsers(dest="command", help="Commands")

    price_parser = subparsers.add_parser("price", help="Get price from exchange")
    price_parser.add_argument("coin", help="Coin symbol (BTC, ETH, etc.)")
    price_parser.add_argument(
        "--exchange", "-e", default="kucoin", choices=["kucoin", "binance", "coinbase"]
    )

    compare_parser = subparsers.add_parser(
        "compare", help="Compare prices across exchanges"
    )
    compare_parser.add_argument("coin", help="Coin symbol")

    candles_parser = subparsers.add_parser("candles", help="Get OHLCV candles")
    candles_parser.add_argument("coin", help="Coin symbol")
    candles_parser.add_argument("--exchange", "-e", default="kucoin")
    candles_parser.add_argument("--timeframe", "-t", default="1hour")
    candles_parser.add_argument("--limit", "-l", type=int, default=10)

    arb_parser = subparsers.add_parser(
        "arbitrage", help="Check for arbitrage opportunities"
    )
    arb_parser.add_argument("coins", nargs="+", help="Coin symbols to check")
    arb_parser.add_argument(
        "--min-spread", type=float, default=0.3, help="Min spread % to report"
    )

    args = parser.parse_args()

    manager = ExchangeManager()

    if args.command == "price":
        try:
            price = manager.get_price(args.coin.upper(), args.exchange)
            print(f"{args.coin.upper()} on {args.exchange}: ${price:,.2f}")
        except ExchangeError as e:
            print(f"Error: {e}")

    elif args.command == "compare":
        try:
            print_price_comparison(manager, args.coin.upper())
        except ExchangeError as e:
            print(f"Error: {e}")

    elif args.command == "candles":
        try:
            candles = manager.get_candles(
                args.coin.upper(), args.exchange, args.timeframe, args.limit
            )
            print(
                f"\n{args.coin.upper()} {args.timeframe} candles from {args.exchange}:"
            )
            print(f"{'Time':<20} {'Open':>12} {'High':>12} {'Low':>12} {'Close':>12}")
            print("-" * 70)
            for c in candles:
                print(
                    f"{c.datetime.strftime('%Y-%m-%d %H:%M'):<20} "
                    f"${c.open:>11,.2f} ${c.high:>11,.2f} "
                    f"${c.low:>11,.2f} ${c.close:>11,.2f}"
                )
        except ExchangeError as e:
            print(f"Error: {e}")

    elif args.command == "arbitrage":
        print("\nScanning for arbitrage opportunities...")
        print("=" * 60)
        found = False
        for coin in args.coins:
            try:
                arb = manager.detect_arbitrage(coin.upper(), args.min_spread)
                if arb:
                    found = True
                    print(
                        f"\n{coin.upper()}: Buy {arb['buy_exchange']} @ ${arb['buy_price']:,.2f} "
                        f"-> Sell {arb['sell_exchange']} @ ${arb['sell_price']:,.2f} "
                        f"({arb['spread_pct']:.3f}%)"
                    )
            except Exception as e:
                print(f"{coin.upper()}: Error - {e}")

        if not found:
            print(f"\nNo arbitrage opportunities found above {args.min_spread}% spread")

    else:
        parser.print_help()


if __name__ == "__main__":
    main()
