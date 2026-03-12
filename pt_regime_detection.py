import logging
import math
from typing import Dict, List, Optional
from pt_config import ConfigManager

logger = logging.getLogger(__name__)

class RegimeDetector:
    """
    Analyzes historical price data to determine the current market regime
    (Bull/Bear/Sideways) and volatility state (High/Low).
    """

    def __init__(self, config=None):
        if config is None:
            cm = ConfigManager()
            cm.reload()
            self.config = cm.get().regime_detection
        else:
            self.config = config
            
        # Optional: store last known regimes to avoid rapid flapping
        self._last_known_regime = {}

    def calculate_sma(self, prices: List[float], period: int) -> Optional[float]:
        """Calculates Simple Moving Average."""
        if len(prices) < period or period <= 0:
            return None
        return sum(prices[-period:]) / period

    def calculate_atr_proxy(self, high_prices: List[float], low_prices: List[float], close_prices: List[float], period: int) -> Optional[float]:
        """Calculates a proxy for Average True Range if full OHLC data is available."""
        if len(high_prices) < period or len(low_prices) < period or len(close_prices) < period + 1:
            return None
            
        true_ranges = []
        # Need close_prices offset by 1 for the previous close
        for i in range(1, period + 1):
            idx = -i
            prev_idx = -(i + 1)
            
            high = high_prices[idx]
            low = low_prices[idx]
            prev_close = close_prices[prev_idx]
            
            tr1 = high - low
            tr2 = abs(high - prev_close)
            tr3 = abs(low - prev_close)
            
            true_ranges.append(max(tr1, tr2, tr3))
            
        return sum(true_ranges) / period

    def calculate_std_dev(self, prices: List[float], period: int) -> Optional[float]:
        """Calculates standard deviation of prices as a volatility measure."""
        if len(prices) < period or period <= 1:
            return None
            
        recent_prices = prices[-period:]
        mean = sum(recent_prices) / period
        
        variance = sum((x - mean) ** 2 for x in recent_prices) / (period - 1)
        return math.sqrt(variance)

    def detect_trend(self, prices: List[float]) -> str:
        """
        Determines the trend direction based on SMA positioning.
        Returns: "BULL", "BEAR", or "SIDEWAYS"
        """
        if not prices:
            return "SIDEWAYS"
            
        current_price = prices[-1]
        
        # Use config lookback, default to 50
        fast_period = max(5, self.config.trend_lookback_candles // 3)
        slow_period = self.config.trend_lookback_candles
        
        fast_sma = self.calculate_sma(prices, fast_period)
        slow_sma = self.calculate_sma(prices, slow_period)
        
        if fast_sma is None or slow_sma is None:
            return "SIDEWAYS" # Not enough data
            
        # Define a 'chop zone' threshold where moving averages are too close
        chop_threshold = slow_sma * 0.005 # 0.5% distance required to establish trend
        
        if fast_sma > slow_sma + chop_threshold and current_price > fast_sma:
            return "BULL"
        elif fast_sma < slow_sma - chop_threshold and current_price < fast_sma:
            return "BEAR"
        else:
            return "SIDEWAYS"

    def detect_volatility(self, prices: List[float]) -> str:
        """
        Determines volatility state based on standard deviation relative to price.
        Returns: "HIGH" or "LOW"
        """
        if not prices:
            return "LOW"
            
        period = self.config.volatility_lookback_candles
        std_dev = self.calculate_std_dev(prices, period)
        
        if std_dev is None:
            return "LOW"
            
        current_price = prices[-1]
        
        # Calculate Coefficient of Variation (CV) = StdDev / Mean
        cv = std_dev / (sum(prices[-period:]) / period)
        
        # Thresholds can be tuned. > 2.5% variation over the period suggests high volatility
        vol_threshold = 0.025 
        
        if cv > vol_threshold:
            return "HIGH"
        else:
            return "LOW"

    def evaluate_regime(self, symbol: str, close_prices: List[float]) -> Dict[str, str]:
        """
        Main entry point to evaluate the full regime for a given coin.
        Takes a list of closing prices (oldest to newest) and returns the state.
        """
        if not self.config.enabled:
            return {"trend": "SIDEWAYS", "volatility": "LOW", "status": "disabled"}
            
        trend = self.detect_trend(close_prices)
        volatility = self.detect_volatility(close_prices)
        
        regime_data = {
            "trend": trend,
            "volatility": volatility,
            "timestamp": "now", # Could add actual timestamp if needed
            "status": "active"
        }
        
        self._last_known_regime[symbol] = regime_data
        return regime_data

    # --- Utility Accessors for integration ---
    
    def get_dca_adjustment_factor(self, symbol: str) -> float:
        """
        Returns a multiplier for the DCA target based on regime.
        1.0 = Default behavior.
        >1.0 = Wait longer (buy deeper dips).
        <1.0 = Buy sooner (aggressive).
        """
        if not self.config.enabled or not self.config.auto_adjust_dca:
            return 1.0
            
        regime = self._last_known_regime.get(symbol, {"trend": "SIDEWAYS"})
        trend = regime.get("trend", "SIDEWAYS")
        
        if trend == "BEAR":
            # In a bear market, wait for deeper dumps (+50% distance) before catching knives
            return 1.5
        elif trend == "BULL":
            # In a bull market, buy shallower dips (-20% distance) aggressively
            return 0.8
        else:
            return 1.0
            
    def get_pm_adjustment_factor(self, symbol: str) -> float:
        """
        Returns a multiplier for the Profit Margin trigger based on regime.
        1.0 = Default behavior.
        >1.0 = Let winners run longer.
        <1.0 = Take profits sooner.
        """
        if not self.config.enabled or not self.config.auto_adjust_pm:
            return 1.0
            
        regime = self._last_known_regime.get(symbol, {"trend": "SIDEWAYS"})
        trend = regime.get("trend", "SIDEWAYS")
        
        if trend == "BEAR":
            # In a bear market, lock in any profit quickly (-30% threshold)
            return 0.7
        elif trend == "BULL":
            # In a bull market, let trailing stops run higher (+50% threshold)
            return 1.5
        else:
            return 1.0

# Simple standalone test block
if __name__ == "__main__":
    from dataclasses import dataclass
    
    @dataclass
    class MockConfig:
        enabled: bool = True
        trend_lookback_candles: int = 30
        volatility_lookback_candles: int = 10
        auto_adjust_dca: bool = True
        auto_adjust_pm: bool = True
        
    detector = RegimeDetector(config=MockConfig())
    
    print("Testing Uptrend (BULL):")
    bull_prices = [100 + i*2 for i in range(50)] # Linear uptrend
    regime = detector.evaluate_regime("BTC", bull_prices)
    print(f"Regime: {regime}")
    print(f"DCA Adj: {detector.get_dca_adjustment_factor('BTC')}")
    print(f"PM Adj: {detector.get_pm_adjustment_factor('BTC')}")
    print("---")
    
    print("Testing Downtrend (BEAR):")
    bear_prices = [200 - i*2 for i in range(50)] # Linear downtrend
    regime = detector.evaluate_regime("ETH", bear_prices)
    print(f"Regime: {regime}")
    print(f"DCA Adj: {detector.get_dca_adjustment_factor('ETH')}")
    print(f"PM Adj: {detector.get_pm_adjustment_factor('ETH')}")
    print("---")
    
    print("Testing Sideways:")
    side_prices = [100 + (i%5)*(1 if i%2==0 else -1) for i in range(50)] # Chop
    regime = detector.evaluate_regime("SOL", side_prices)
    print(f"Regime: {regime}")
