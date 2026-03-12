"""
PowerTrader AI — Advanced Analytics (v4.0.0)
==============================================
Pattern recognition, market microstructure analysis,
order flow analysis, and advanced statistical arbitrage.

Features:
    1. PatternRecognizer — detects chart patterns and anomalies
    2. MicrostructureAnalyzer — bid/ask spread, depth imbalance
    3. OrderFlowAnalyzer — volume delta, CVD, aggressive trades
    4. StatArbEngine — cointegration-based pairs trading

Usage:
    from pt_advanced_analytics import PatternRecognizer, StatArbEngine

    pr = PatternRecognizer()
    patterns = pr.scan(prices)

    arb = StatArbEngine()
    signals = arb.scan_pairs(["BTC", "ETH", "SOL"])
"""

from __future__ import annotations
import json
import math
import random
from dataclasses import dataclass, asdict
from datetime import datetime
from typing import List, Dict, Optional, Tuple
from pathlib import Path

try:
    import numpy as np
except ImportError:
    np = None


# =============================================================================
# PATTERN RECOGNIZER
# =============================================================================

class PatternRecognizer:
    """Detects chart patterns and market anomalies from price data."""

    def scan(self, prices: List[float], volumes: Optional[List[float]] = None) -> List[dict]:
        """Scan price series for all known patterns."""
        patterns = []
        patterns += self._detect_double_tops_bottoms(prices)
        patterns += self._detect_head_shoulders(prices)
        patterns += self._detect_support_resistance(prices)
        if volumes:
            patterns += self._detect_volume_anomalies(prices, volumes)
        patterns += self._detect_trend_breaks(prices)
        return patterns

    def _detect_double_tops_bottoms(self, prices: List[float]) -> List[dict]:
        """Detect double top and double bottom patterns."""
        results = []
        n = len(prices)
        if n < 30:
            return results

        # Find local extremes
        window = 5
        for i in range(window, n - window):
            is_peak = all(prices[i] >= prices[i-j] and prices[i] >= prices[i+j]
                         for j in range(1, window + 1))
            is_valley = all(prices[i] <= prices[i-j] and prices[i] <= prices[i+j]
                           for j in range(1, window + 1))

            if is_peak:
                # Look for matching peak within 10-40 bars
                for j in range(i + 10, min(i + 40, n - window)):
                    is_peak2 = all(prices[j] >= prices[j-k] and prices[j] >= prices[j+k]
                                  for k in range(1, min(window + 1, n - j)))
                    if is_peak2 and abs(prices[i] - prices[j]) / prices[i] < 0.02:
                        results.append({
                            "pattern": "double_top",
                            "start_idx": i, "end_idx": j,
                            "price_level": prices[i],
                            "confidence": 0.7,
                            "signal": "bearish",
                        })
                        break

            if is_valley:
                for j in range(i + 10, min(i + 40, n - window)):
                    is_valley2 = all(j + k < n and prices[j] <= prices[j-k] and prices[j] <= prices[j+k]
                                    for k in range(1, min(window + 1, n - j)))
                    if is_valley2 and abs(prices[i] - prices[j]) / prices[i] < 0.02:
                        results.append({
                            "pattern": "double_bottom",
                            "start_idx": i, "end_idx": j,
                            "price_level": prices[i],
                            "confidence": 0.7,
                            "signal": "bullish",
                        })
                        break

        return results[:5]  # Cap results

    def _detect_head_shoulders(self, prices: List[float]) -> List[dict]:
        """Simplified head & shoulders detection."""
        results = []
        n = len(prices)
        if n < 50:
            return results

        # Scan for 3-peak pattern
        peaks = []
        for i in range(3, n - 3):
            if prices[i] > prices[i-1] and prices[i] > prices[i+1] and \
               prices[i] > prices[i-2] and prices[i] > prices[i+2]:
                peaks.append((i, prices[i]))

        for i in range(len(peaks) - 2):
            l_shoulder = peaks[i]
            head = peaks[i+1]
            r_shoulder = peaks[i+2]

            # Head must be highest, shoulders roughly equal
            if head[1] > l_shoulder[1] and head[1] > r_shoulder[1]:
                shoulder_diff = abs(l_shoulder[1] - r_shoulder[1]) / l_shoulder[1]
                if shoulder_diff < 0.03:  # Within 3%
                    results.append({
                        "pattern": "head_and_shoulders",
                        "start_idx": l_shoulder[0],
                        "head_idx": head[0],
                        "end_idx": r_shoulder[0],
                        "confidence": 0.65,
                        "signal": "bearish",
                    })

        return results[:3]

    def _detect_support_resistance(self, prices: List[float]) -> List[dict]:
        """Detect key support and resistance levels."""
        if len(prices) < 20:
            return []

        # Find price levels where price bounced multiple times
        levels = []
        price_range = max(prices) - min(prices)
        tolerance = price_range * 0.01  # 1% tolerance

        # Cluster local extremes
        extremes = []
        for i in range(2, len(prices) - 2):
            if prices[i] >= prices[i-1] and prices[i] >= prices[i+1]:
                extremes.append(("resistance", prices[i]))
            if prices[i] <= prices[i-1] and prices[i] <= prices[i+1]:
                extremes.append(("support", prices[i]))

        # Group nearby levels
        for level_type, price in extremes:
            found = False
            for level in levels:
                if abs(level["price"] - price) < tolerance:
                    level["touches"] += 1
                    found = True
                    break
            if not found:
                levels.append({"type": level_type, "price": price, "touches": 1})

        # Return strongest levels
        strong_levels = [l for l in levels if l["touches"] >= 2]
        strong_levels.sort(key=lambda l: l["touches"], reverse=True)

        return [{"pattern": f"{l['type']}_level", "price": round(l["price"], 2),
                 "touches": l["touches"], "confidence": min(0.9, 0.5 + l["touches"] * 0.1),
                 "signal": "support" if l["type"] == "support" else "resistance"}
                for l in strong_levels[:6]]

    def _detect_volume_anomalies(self, prices: List[float],
                                  volumes: List[float]) -> List[dict]:
        """Detect unusual volume spikes."""
        results = []
        if len(volumes) < 20:
            return results

        for i in range(20, len(volumes)):
            avg_vol = sum(volumes[i-20:i]) / 20
            if avg_vol > 0 and volumes[i] > avg_vol * 3:
                price_change = (prices[i] / prices[i-1] - 1) * 100 if i > 0 else 0
                results.append({
                    "pattern": "volume_anomaly",
                    "idx": i,
                    "volume_ratio": round(volumes[i] / avg_vol, 1),
                    "price_change_pct": round(price_change, 2),
                    "signal": "bullish" if price_change > 0 else "bearish",
                    "confidence": min(0.9, 0.5 + (volumes[i] / avg_vol) * 0.1),
                })

        return results[:5]

    def _detect_trend_breaks(self, prices: List[float]) -> List[dict]:
        """Detect trend line breaks."""
        results = []
        n = len(prices)
        if n < 50:
            return results

        # Simple SMA trend break
        for period in [20, 50]:
            if n < period + 5:
                continue
            sma = sum(prices[-period:]) / period
            current = prices[-1]
            prev_above = prices[-2] > sum(prices[-period-1:-1]) / period

            if current > sma and not prev_above:
                results.append({
                    "pattern": f"sma{period}_breakout",
                    "direction": "bullish",
                    "sma_value": round(sma, 2),
                    "current_price": round(current, 2),
                    "confidence": 0.6,
                    "signal": "bullish",
                })
            elif current < sma and prev_above:
                results.append({
                    "pattern": f"sma{period}_breakdown",
                    "direction": "bearish",
                    "sma_value": round(sma, 2),
                    "current_price": round(current, 2),
                    "confidence": 0.6,
                    "signal": "bearish",
                })

        return results


# =============================================================================
# MARKET MICROSTRUCTURE ANALYZER
# =============================================================================

class MicrostructureAnalyzer:
    """Analyzes market microstructure: spread, depth, imbalance."""

    def analyze_orderbook(self, bids: List[Tuple[float, float]],
                          asks: List[Tuple[float, float]]) -> dict:
        """Analyze order book structure."""
        if not bids or not asks:
            return {"error": "Empty order book"}

        best_bid = bids[0][0]
        best_ask = asks[0][0]
        spread = best_ask - best_bid
        spread_pct = (spread / best_bid) * 100

        bid_depth = sum(p * q for p, q in bids[:10])
        ask_depth = sum(p * q for p, q in asks[:10])
        total_depth = bid_depth + ask_depth

        imbalance = (bid_depth - ask_depth) / total_depth if total_depth > 0 else 0

        # Detect spoofing (large orders far from mid)
        mid = (best_bid + best_ask) / 2
        spoof_threshold = mid * 0.005  # 0.5% from mid
        suspicious = []
        for price, qty in bids + asks:
            if abs(price - mid) > spoof_threshold and qty * price > bid_depth * 0.3:
                suspicious.append({"price": price, "qty": qty, "type": "potential_spoof"})

        return {
            "spread": round(spread, 4),
            "spread_pct": round(spread_pct, 4),
            "spread_bps": round(spread_pct * 100, 2),
            "bid_depth_usd": round(bid_depth, 2),
            "ask_depth_usd": round(ask_depth, 2),
            "imbalance": round(imbalance, 4),
            "imbalance_signal": "bullish" if imbalance > 0.2 else "bearish" if imbalance < -0.2 else "neutral",
            "suspicious_orders": suspicious[:3],
        }


# =============================================================================
# ORDER FLOW ANALYZER
# =============================================================================

class OrderFlowAnalyzer:
    """Analyzes order flow: volume delta, CVD, aggressive trades."""

    def compute_volume_delta(self, prices: List[float],
                              volumes: List[float]) -> List[float]:
        """Estimate volume delta (buy vol - sell vol) from price movement."""
        deltas = [0.0]
        for i in range(1, len(prices)):
            if prices[i] > prices[i-1]:
                deltas.append(volumes[i] * 0.6)  # Price up = net buying
            elif prices[i] < prices[i-1]:
                deltas.append(-volumes[i] * 0.6)
            else:
                deltas.append(0.0)
        return deltas

    def compute_cvd(self, deltas: List[float]) -> List[float]:
        """Compute Cumulative Volume Delta."""
        cvd = []
        total = 0.0
        for d in deltas:
            total += d
            cvd.append(total)
        return cvd

    def detect_divergence(self, prices: List[float],
                          cvd: List[float], lookback: int = 20) -> Optional[dict]:
        """Detect price/CVD divergence."""
        if len(prices) < lookback or len(cvd) < lookback:
            return None

        price_trend = prices[-1] - prices[-lookback]
        cvd_trend = cvd[-1] - cvd[-lookback]

        if price_trend > 0 and cvd_trend < 0:
            return {"type": "bearish_divergence", "signal": "bearish",
                    "price_change": round(price_trend, 2),
                    "cvd_change": round(cvd_trend, 2), "confidence": 0.65}
        elif price_trend < 0 and cvd_trend > 0:
            return {"type": "bullish_divergence", "signal": "bullish",
                    "price_change": round(price_trend, 2),
                    "cvd_change": round(cvd_trend, 2), "confidence": 0.65}

        return None

    def analyze(self, prices: List[float], volumes: List[float]) -> dict:
        """Full order flow analysis."""
        deltas = self.compute_volume_delta(prices, volumes)
        cvd = self.compute_cvd(deltas)
        divergence = self.detect_divergence(prices, cvd)

        recent_delta = sum(deltas[-10:]) if len(deltas) >= 10 else sum(deltas)
        return {
            "recent_delta": round(recent_delta, 2),
            "delta_signal": "buying" if recent_delta > 0 else "selling",
            "cvd_current": round(cvd[-1], 2) if cvd else 0,
            "cvd_trend": "rising" if len(cvd) > 1 and cvd[-1] > cvd[-2] else "falling",
            "divergence": divergence,
        }


# =============================================================================
# STATISTICAL ARBITRAGE ENGINE
# =============================================================================

class StatArbEngine:
    """
    Cointegration-based pairs trading: finds correlated pairs,
    computes spread z-score, and generates entry/exit signals.
    """

    def compute_correlation(self, series_a: List[float],
                            series_b: List[float]) -> float:
        """Pearson correlation coefficient."""
        n = min(len(series_a), len(series_b))
        if n < 10:
            return 0.0

        a, b = series_a[:n], series_b[:n]
        mean_a = sum(a) / n
        mean_b = sum(b) / n

        cov = sum((a[i] - mean_a) * (b[i] - mean_b) for i in range(n)) / n
        std_a = (sum((x - mean_a)**2 for x in a) / n) ** 0.5
        std_b = (sum((x - mean_b)**2 for x in b) / n) ** 0.5

        if std_a == 0 or std_b == 0:
            return 0.0
        return cov / (std_a * std_b)

    def compute_spread(self, series_a: List[float],
                       series_b: List[float]) -> Tuple[List[float], float, float]:
        """Compute log price spread and its z-score."""
        n = min(len(series_a), len(series_b))
        spread = [math.log(series_a[i]) - math.log(series_b[i]) for i in range(n)]

        mean_s = sum(spread) / n
        std_s = (sum((s - mean_s)**2 for s in spread) / n) ** 0.5

        return spread, mean_s, std_s

    def get_zscore(self, spread: List[float], mean: float, std: float) -> float:
        """Current z-score of the spread."""
        if std == 0:
            return 0.0
        return (spread[-1] - mean) / std

    def scan_pairs(self, price_dict: Dict[str, List[float]],
                   min_corr: float = 0.7) -> List[dict]:
        """Scan all pairs for stat arb opportunities."""
        coins = list(price_dict.keys())
        results = []

        for i in range(len(coins)):
            for j in range(i + 1, len(coins)):
                a, b = coins[i], coins[j]
                corr = self.compute_correlation(price_dict[a], price_dict[b])

                if abs(corr) < min_corr:
                    continue

                spread, mean_s, std_s = self.compute_spread(price_dict[a], price_dict[b])
                zscore = self.get_zscore(spread, mean_s, std_s)

                signal = "none"
                if zscore > 2.0:
                    signal = f"short {a}, long {b}"
                elif zscore < -2.0:
                    signal = f"long {a}, short {b}"
                elif abs(zscore) < 0.5:
                    signal = "close position (mean reversion)"

                results.append({
                    "pair": f"{a}/{b}",
                    "correlation": round(corr, 4),
                    "zscore": round(zscore, 4),
                    "signal": signal,
                    "spread_mean": round(mean_s, 6),
                    "spread_std": round(std_s, 6),
                })

        results.sort(key=lambda r: abs(r["zscore"]), reverse=True)
        return results


# =============================================================================
# SELF-TEST
# =============================================================================

if __name__ == "__main__":
    print("=" * 60)
    print("Advanced Analytics — Self-Test")
    print("=" * 60)

    random.seed(42)
    # Synthetic prices + volumes
    prices = [100.0]
    volumes = [1000.0]
    for _ in range(299):
        prices.append(prices[-1] * (1 + random.gauss(0.0002, 0.015)))
        volumes.append(random.uniform(500, 2000))
    # Inject volume spike
    volumes[150] = 8000

    # 1. Pattern Recognition
    print("\n1. PatternRecognizer...")
    pr = PatternRecognizer()
    patterns = pr.scan(prices, volumes)
    for p in patterns[:5]:
        print(f"   {p.get('pattern', '?'):>25} | signal: {p.get('signal', '?'):>8} | confidence: {p.get('confidence', 0):.2f}")

    # 2. Order Flow
    print("\n2. OrderFlowAnalyzer...")
    ofa = OrderFlowAnalyzer()
    flow = ofa.analyze(prices, volumes)
    print(f"   Delta: {flow['recent_delta']:>8.1f} ({flow['delta_signal']})")
    print(f"   CVD: {flow['cvd_current']:>8.1f} ({flow['cvd_trend']})")

    # 3. Stat Arb
    print("\n3. StatArbEngine...")
    prices2 = [p * 0.03 + random.gauss(0, 0.5) for p in prices]  # Correlated
    prices3 = [random.uniform(10, 50) for _ in prices]  # Uncorrelated
    arb = StatArbEngine()
    pairs = arb.scan_pairs({"BTC": prices, "ETH": prices2, "DOGE": prices3}, min_corr=0.5)
    for p in pairs:
        print(f"   {p['pair']:>10} | corr: {p['correlation']:>6.3f} | z: {p['zscore']:>6.3f} | {p['signal']}")

    # 4. Microstructure
    print("\n4. MicrostructureAnalyzer...")
    bids = [(100.0 - i*0.1, random.uniform(1, 10)) for i in range(10)]
    asks = [(100.1 + i*0.1, random.uniform(1, 10)) for i in range(10)]
    ms = MicrostructureAnalyzer()
    result = ms.analyze_orderbook(bids, asks)
    print(f"   Spread: {result['spread_bps']} bps | Imbalance: {result['imbalance']:.3f} ({result['imbalance_signal']})")

    print("\n✅ Self-test complete")
