"""
Advanced Risk Management for PowerTrader AI (v3.0.0)

This module provides portfolio-level risk limits, concentration limits, 
and volume-based liquidity checks. It acts as a safety layer before trades are placed.

Author: PowerTrader AI Team
Version: 3.0.0
License: Apache 2.0
"""

import logging
from pt_config import ConfigManager

try:
    from pt_volume import VolumeAnalyzer
    VOLUME_AVAILABLE = True
except ImportError:
    VOLUME_AVAILABLE = False
    logging.warning("[RiskManager] pt_volume module not available - liquidity checks disabled")


class RiskManager:
    """Manages trading risk bounds to protect capital."""

    def __init__(self, config=None):
        if config is None:
            config = ConfigManager().get().risk_management
        self.config = config

    def check_portfolio_drawdown(self, current_value: float, peak_value: float) -> bool:
        """
        Check if the portfolio is within the max drawdown limit.
        Returns False if the current value has drawn down more than the allowed percentage.
        """
        if not self.config.enabled:
            return True

        if current_value is None or peak_value is None or peak_value <= 0:
            return True

        drawdown_pct = ((peak_value - current_value) / peak_value) * 100.0

        if drawdown_pct > self.config.max_portfolio_drawdown_pct:
            logging.critical(f"[RiskManager] SECURITY HALT: Portfolio drawdown {drawdown_pct:.2f}% exceeds max {self.config.max_portfolio_drawdown_pct:.2f}% (Peak: ${peak_value:.2f}, Current: ${current_value:.2f})")
            return False
            
        return True

    def check_concentration_limit(self, symbol: str, current_holdings_value_usd: float, proposed_trade_value_usd: float, total_account_value: float) -> tuple[bool, float]:
        """
        Check if adding the proposed_trade_value_usd to the current holdings 
        exceeds the max_coin_concentration_pct limit.
        
        Returns: (is_safe: bool, new_concentration_pct: float)
        """
        if not self.config.enabled:
            pass_val = 0.0
            if total_account_value > 0:
                pass_val = ((current_holdings_value_usd + proposed_trade_value_usd) / total_account_value) * 100
            return True, pass_val

        if total_account_value <= 0:
            return True, 0.0

        new_total_value = current_holdings_value_usd + proposed_trade_value_usd
        new_concentration_pct = (new_total_value / total_account_value) * 100.0

        if new_concentration_pct > self.config.max_coin_concentration_pct:
            logging.error(f"[RiskManager] Concentration limit reached for {symbol}. Proposed concentration: {new_concentration_pct:.2f}%, Max allowed: {self.config.max_coin_concentration_pct:.2f}%")
            return False, new_concentration_pct

        return True, new_concentration_pct

    def check_liquidity(self, symbol: str, proposed_order_cost_usd: float, recent_volume_usd: float = None) -> bool:
        """
        Ensures the proposed order cost doesn't exceed a safe ratio of the recent volume.
        If recent_volume_usd is heavily eclipsed (e.g., liquidity is tiny), block trade.
        """
        if not self.config.enabled:
            return True

        if recent_volume_usd is None or recent_volume_usd <= 0:
            # If no volume data is provided or found, we might want to fail safe, 
            # but usually we want to let it pass if we can't fetch it, depending on strictness.
            # In a resilient setup, we pass if data isn't fetched, to avoid locking up on API faults.
            return True

        # Ensure order isn't larger than (recent_volume_usd / min_liquidity_multiplier)
        max_order_size = recent_volume_usd / max(1.0, self.config.min_liquidity_multiplier)

        if proposed_order_cost_usd > max_order_size:
            logging.warning(f"[RiskManager] Liquidity check failed for {symbol}: proposed ${proposed_order_cost_usd:.2f} but max allowed is ${max_order_size:.2f} (volume: ${recent_volume_usd:.2f})")
            return False

        return True


def main():
    """Local simulation testing for pt_risk_management constraints."""
    import dataclasses
    
    @dataclasses.dataclass
    class DummyConfig:
        enabled: bool = True
        max_portfolio_drawdown_pct: float = 15.0
        max_coin_concentration_pct: float = 25.0
        min_liquidity_multiplier: float = 3.0
        
    cfg = DummyConfig()
    rm = RiskManager(config=cfg)
    
    print("Testing Advanced Risk Management Module...\n")
    
    # 1. Test Drawdown
    print("--- Drawdown Detection ---")
    is_safe = rm.check_portfolio_drawdown(current_value=850, peak_value=1000)
    print(f"15% Drawdown (850/1000). Passed? {is_safe} (Expected: True/borderline)")
    
    is_safe = rm.check_portfolio_drawdown(current_value=800, peak_value=1000)
    print(f"20% Drawdown (800/1000). Passed? {is_safe} (Expected: False)")

    # 2. Test Concentration
    print("\n--- Concentration Logic ---")
    acct_value = 10000.0
    
    # Existing BTC value 1000 (10%), add 1000 => 2000 (20%) -> Pass
    safe, conc = rm.check_concentration_limit("BTC", 1000.0, 1000.0, acct_value)
    print(f"BTC 20% concentration check. Passed? {safe} ({conc:.1f}%)")
    
    # Existing BTC value 2000 (20%), add 1000 => 3000 (30%) -> Fail
    safe, conc = rm.check_concentration_limit("BTC", 2000.0, 1000.0, acct_value)
    print(f"BTC 30% concentration check. Passed? {safe} ({conc:.1f}%)")
    
    # 3. Test Liquidity
    print("\n--- Liquidity Enforcement ---")
    # Volume 30,000. Order 5000. Multiplier 3 => Max order 10,000 -> Pass
    safe = rm.check_liquidity("BTC", 5000.0, 30000.0)
    print(f"BTC order $5000 against $30k volume. Passed? {safe} (Expected: True)")
    
    # Volume 30,000. Order 15000. Multiplier 3 => Max order 10,000 -> Fail
    safe = rm.check_liquidity("BTC", 15000.0, 30000.0)
    print(f"BTC order $15000 against $30k volume. Passed? {safe} (Expected: False)")

if __name__ == "__main__":
    main()
