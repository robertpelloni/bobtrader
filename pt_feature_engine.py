"""
PowerTrader AI — Feature Engineering Pipeline
===============================================
Transforms raw OHLCV candle data into standardized feature vectors
for use in the ML ensemble predictor.

Features computed:
  Price:      close-to-close return %, normalized range, gap ratio
  Volume:     volume ratio vs SMA-10, volume trend slope
  Momentum:   RSI-14, ROC-5, MACD signal crossover flag
  Volatility: ATR-14 %, Bollinger Band width
"""

from __future__ import annotations
import math
import sys
from dataclasses import dataclass, field
from typing import List, Dict, Optional, Tuple

import numpy as np


# =============================================================================
# FEATURE DEFINITIONS
# =============================================================================

FEATURE_NAMES: List[str] = [
    # Price (3)
    "close_return_pct",
    "normalized_range",
    "gap_ratio",
    # Volume (2)
    "volume_ratio_sma10",
    "volume_trend_slope",
    # Momentum (3)
    "rsi_14",
    "roc_5",
    "macd_signal_cross",
    # Volatility (2)
    "atr_14_pct",
    "bollinger_width",
]

NUM_FEATURES = len(FEATURE_NAMES)


# =============================================================================
# HELPER MATH
# =============================================================================


def _sma(values: List[float], period: int) -> List[float]:
    """Simple Moving Average. Returns list same length as input; leading entries are NaN."""
    out = [float("nan")] * len(values)
    if period <= 0 or len(values) < period:
        return out
    running = sum(values[:period])
    out[period - 1] = running / period
    for i in range(period, len(values)):
        running += values[i] - values[i - period]
        out[i] = running / period
    return out


def _ema(values: List[float], period: int) -> List[float]:
    """Exponential Moving Average."""
    out = [float("nan")] * len(values)
    if period <= 0 or len(values) < period:
        return out
    k = 2.0 / (period + 1)
    # seed with SMA
    out[period - 1] = sum(values[:period]) / period
    for i in range(period, len(values)):
        out[i] = values[i] * k + out[i - 1] * (1 - k)
    return out


def _rsi(closes: List[float], period: int = 14) -> List[float]:
    """Relative Strength Index."""
    out = [float("nan")] * len(closes)
    if len(closes) < period + 1:
        return out
    gains = []
    losses = []
    for i in range(1, len(closes)):
        delta = closes[i] - closes[i - 1]
        gains.append(max(delta, 0.0))
        losses.append(max(-delta, 0.0))

    avg_gain = sum(gains[:period]) / period
    avg_loss = sum(losses[:period]) / period

    if avg_loss == 0:
        out[period] = 100.0
    else:
        rs = avg_gain / avg_loss
        out[period] = 100.0 - (100.0 / (1.0 + rs))

    for i in range(period, len(gains)):
        avg_gain = (avg_gain * (period - 1) + gains[i]) / period
        avg_loss = (avg_loss * (period - 1) + losses[i]) / period
        if avg_loss == 0:
            out[i + 1] = 100.0
        else:
            rs = avg_gain / avg_loss
            out[i + 1] = 100.0 - (100.0 / (1.0 + rs))
    return out


def _true_range(highs: List[float], lows: List[float], closes: List[float]) -> List[float]:
    """True Range series (first element is high-low)."""
    tr = [highs[0] - lows[0]]
    for i in range(1, len(highs)):
        tr.append(
            max(
                highs[i] - lows[i],
                abs(highs[i] - closes[i - 1]),
                abs(lows[i] - closes[i - 1]),
            )
        )
    return tr


# =============================================================================
# FEATURE ENGINE
# =============================================================================


class FeatureEngine:
    """Computes a standardized feature matrix from raw OHLCV candles."""

    def __init__(self):
        self.feature_names = list(FEATURE_NAMES)

    @staticmethod
    def get_feature_names() -> List[str]:
        return list(FEATURE_NAMES)

    def compute_features(
        self,
        opens: List[float],
        highs: List[float],
        lows: List[float],
        closes: List[float],
        volumes: List[float],
    ) -> np.ndarray:
        """
        Compute feature matrix from OHLCV arrays (all same length).
        Returns np.ndarray of shape (n_candles, NUM_FEATURES).
        Leading rows will contain NaN where lookback is insufficient.
        """
        n = len(closes)
        features = np.full((n, NUM_FEATURES), np.nan)

        # --- Price features ---
        # 1. close-to-close return %
        for i in range(1, n):
            if closes[i - 1] != 0:
                features[i, 0] = ((closes[i] - closes[i - 1]) / closes[i - 1]) * 100.0

        # 2. normalized range: (high - low) / close
        for i in range(n):
            if closes[i] != 0:
                features[i, 1] = (highs[i] - lows[i]) / closes[i]

        # 3. gap ratio: (open - prev_close) / prev_close
        for i in range(1, n):
            if closes[i - 1] != 0:
                features[i, 2] = ((opens[i] - closes[i - 1]) / closes[i - 1]) * 100.0

        # --- Volume features ---
        vol_sma10 = _sma(volumes, 10)

        # 4. volume ratio vs SMA-10
        for i in range(n):
            if not math.isnan(vol_sma10[i]) and vol_sma10[i] != 0:
                features[i, 3] = volumes[i] / vol_sma10[i]

        # 5. volume trend slope (linear regression slope over last 5 candles, normalized)
        for i in range(4, n):
            window = volumes[i - 4 : i + 1]
            mean_v = sum(window) / 5
            if mean_v == 0:
                features[i, 4] = 0.0
            else:
                # simple slope: polyfit degree 1
                x = np.arange(5, dtype=float)
                y = np.array(window)
                x_mean = 2.0  # mean of [0,1,2,3,4]
                y_mean = np.mean(y)
                num = np.sum((x - x_mean) * (y - y_mean))
                den = np.sum((x - x_mean) ** 2)
                slope = num / den if den != 0 else 0.0
                features[i, 4] = slope / mean_v  # normalized

        # --- Momentum features ---
        # 6. RSI-14
        rsi_vals = _rsi(closes, 14)
        for i in range(n):
            features[i, 5] = rsi_vals[i]

        # 7. ROC-5 (rate of change over 5 periods)
        for i in range(5, n):
            if closes[i - 5] != 0:
                features[i, 6] = ((closes[i] - closes[i - 5]) / closes[i - 5]) * 100.0

        # 8. MACD signal cross flag (+1 = bullish cross, -1 = bearish, 0 = none)
        ema12 = _ema(closes, 12)
        ema26 = _ema(closes, 26)
        macd_line = [
            (ema12[i] - ema26[i]) if not (math.isnan(ema12[i]) or math.isnan(ema26[i])) else float("nan")
            for i in range(n)
        ]
        # signal line = EMA-9 of MACD
        valid_macd = [v for v in macd_line if not math.isnan(v)]
        if len(valid_macd) >= 9:
            signal_raw = _ema(
                [v if not math.isnan(v) else 0.0 for v in macd_line], 9
            )
            for i in range(1, n):
                if not math.isnan(macd_line[i]) and not math.isnan(signal_raw[i]) and not math.isnan(macd_line[i-1]) and not math.isnan(signal_raw[i-1]):
                    prev_diff = macd_line[i - 1] - signal_raw[i - 1]
                    curr_diff = macd_line[i] - signal_raw[i]
                    if prev_diff <= 0 and curr_diff > 0:
                        features[i, 7] = 1.0  # bullish cross
                    elif prev_diff >= 0 and curr_diff < 0:
                        features[i, 7] = -1.0  # bearish cross
                    else:
                        features[i, 7] = 0.0

        # --- Volatility features ---
        # 9. ATR-14 as % of close
        tr = _true_range(highs, lows, closes)
        atr_vals = _sma(tr, 14)
        for i in range(n):
            if not math.isnan(atr_vals[i]) and closes[i] != 0:
                features[i, 8] = (atr_vals[i] / closes[i]) * 100.0

        # 10. Bollinger Band width: (upper - lower) / middle
        sma20 = _sma(closes, 20)
        for i in range(19, n):
            window = closes[i - 19 : i + 1]
            std = float(np.std(window))
            mid = sma20[i]
            if not math.isnan(mid) and mid != 0:
                upper = mid + 2 * std
                lower = mid - 2 * std
                features[i, 9] = (upper - lower) / mid

        return features

    def compute_features_from_candles(self, candles) -> np.ndarray:
        """
        Convenience: accepts a list of candle objects with .open, .high, .low, .close, .volume
        """
        opens = [c.open for c in candles]
        highs = [c.high for c in candles]
        lows = [c.low for c in candles]
        closes = [c.close for c in candles]
        volumes = [c.volume for c in candles]
        return self.compute_features(opens, highs, lows, closes, volumes)


# =============================================================================
# SELF-TEST
# =============================================================================


def _self_test():
    """Generate synthetic candles and verify feature computation."""
    print("=" * 60)
    print("FEATURE ENGINE SELF-TEST")
    print("=" * 60)

    np.random.seed(42)
    n = 100
    base_price = 50000.0

    # Generate a random walk for closes
    returns = np.random.normal(0, 0.01, n)
    closes = [base_price]
    for r in returns[1:]:
        closes.append(closes[-1] * (1 + r))

    opens = [c * (1 + np.random.normal(0, 0.002)) for c in closes]
    highs = [max(o, c) * (1 + abs(np.random.normal(0, 0.005))) for o, c in zip(opens, closes)]
    lows = [min(o, c) * (1 - abs(np.random.normal(0, 0.005))) for o, c in zip(opens, closes)]
    volumes = [abs(np.random.normal(1000, 200)) for _ in range(n)]

    engine = FeatureEngine()
    features = engine.compute_features(opens, highs, lows, closes, volumes)

    print(f"\nInput:  {n} candles")
    print(f"Output: {features.shape} feature matrix ({features.shape[1]} features)")
    print(f"\nFeature names: {engine.get_feature_names()}")

    # Count valid (non-NaN) values per feature
    print(f"\n{'Feature':<25} {'Valid':>6} {'Mean':>10} {'Std':>10}")
    print("-" * 55)
    for j, name in enumerate(engine.get_feature_names()):
        col = features[:, j]
        valid = np.sum(~np.isnan(col))
        if valid > 0:
            valid_vals = col[~np.isnan(col)]
            print(f"{name:<25} {int(valid):>6} {np.mean(valid_vals):>10.4f} {np.std(valid_vals):>10.4f}")
        else:
            print(f"{name:<25} {int(valid):>6} {'N/A':>10} {'N/A':>10}")

    # Assertions
    assert features.shape == (n, NUM_FEATURES), f"Shape mismatch: {features.shape}"
    assert np.sum(~np.isnan(features[-1, :])) >= 8, "Last row should have most features computed"
    
    # RSI should be between 0 and 100
    rsi_col = features[:, 5]
    valid_rsi = rsi_col[~np.isnan(rsi_col)]
    assert np.all(valid_rsi >= 0) and np.all(valid_rsi <= 100), "RSI out of range"

    print("\n✅ All assertions passed!")
    print("=" * 60)


if __name__ == "__main__":
    _self_test()
