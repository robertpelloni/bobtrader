import logging
from typing import Dict, List, Optional, Tuple
import time

logger = logging.getLogger(__name__)

class Order:
    def __init__(self, symbol: str, side: str, amount_usd: float):
        self.symbol = symbol
        self.side = side.upper()
        self.amount_usd = amount_usd

    def __repr__(self):
        return f"Order({self.side} {self.symbol} for ${self.amount_usd:.2f})"

class Rebalancer:
    def __init__(self, config=None, db_path: str = "data/analytics.db"):
        self.config = config
        self.db_path = db_path
        
    def _get_target_allocations(self) -> Dict[str, float]:
        if not self.config or not hasattr(self.config, 'target_allocations'):
            return {}
        return self.config.target_allocations
        
    def check_drift(self, current_portfolio: Dict[str, float], total_portfolio_value: float) -> List[Dict]:
        """
        Calculates the drift of the current portfolio from target allocations.
        current_portfolio represents symbol -> value_usd
        total_portfolio_value is the total equity in USD
        Returns a list of dicts describing the required adjustment for each asset.
        """
        targets = self._get_target_allocations()
        if not targets or total_portfolio_value <= 0:
            return []
            
        # Ensure targets sum to <= 100%
        target_sum = sum(targets.values())
        if target_sum > 100.0:
            logger.warning(f"[Rebalancer] Target allocations sum > 100% ({target_sum}%). Normalizing.")
            targets = {k: (v / target_sum) * 100.0 for k, v in targets.items()}
            
        adjustments = []
        drift_threshold = getattr(self.config, 'drift_threshold_pct', 5.0) if self.config else 5.0
        
        for symbol, target_pct in targets.items():
            current_value = current_portfolio.get(symbol, 0.0)
            current_pct = (current_value / total_portfolio_value) * 100.0
            
            drift = current_pct - target_pct
            drift_abs = abs(drift)
            
            if drift_abs >= drift_threshold:
                target_value = total_portfolio_value * (target_pct / 100.0)
                diff_usd = target_value - current_value
                
                side = "BUY" if diff_usd > 0 else "SELL"
                
                adjustments.append({
                    "symbol": symbol,
                    "current_pct": current_pct,
                    "target_pct": target_pct,
                    "drift_pct": drift,
                    "diff_usd": abs(diff_usd),
                    "action": side
                })
                
        return adjustments
        
    def is_rebalance_due(self, last_rebalance_ts: float) -> bool:
        """
        Checks if a rebalance is due based on the configured time interval.
        """
        if not self.config:
            return False
            
        trigger_mode = getattr(self.config, 'trigger_mode', 'threshold').lower()
        if trigger_mode not in ['time', 'both']:
            return False
            
        interval_hours = getattr(self.config, 'rebalance_interval_hours', 168)
        interval_seconds = interval_hours * 3600
        
        now = time.time()
        return (now - last_rebalance_ts) >= interval_seconds
        
    def _is_wash_sale(self, symbol: str, current_price: float, avg_cost: float, trade_history: List[Dict]) -> bool:
        """
        Naive wash sale prevention: don't sell if the current price is less than average cost 
        AND there was a buy in the last 30 days.
        """
        if current_price >= avg_cost:
            return False # Selling for a profit is not a wash sale
            
        now = time.time()
        thirty_days_sec = 30 * 24 * 3600
        
        for trade in trade_history:
            if trade.get("symbol") == symbol and str(trade.get("side", "")).upper() == "BUY":
                trade_ts = trade.get("timestamp", 0)
                if (now - trade_ts) <= thirty_days_sec:
                    return True
                    
        return False
        
    def generate_rebalance_orders(
        self, 
        current_portfolio: Dict[str, float], 
        total_value: float,
        price_data: Dict[str, float],
        cost_basis_data: Dict[str, float],
        trade_history: List[Dict] = None
    ) -> List[Order]:
        """
        Takes the detected drift and factors in wash sales to output a final list of Orders.
        """
        if trade_history is None:
            trade_history = []
            
        avoid_wash_sales = getattr(self.config, 'avoid_wash_sales', True) if self.config else True
        adjustments = self.check_drift(current_portfolio, total_value)
        
        orders = []
        for adj in adjustments:
            symbol = adj["symbol"]
            side = adj["action"]
            diff_usd = adj["diff_usd"]
            
            if side == "SELL" and avoid_wash_sales:
                current_price = price_data.get(symbol, 0.0)
                avg_cost = cost_basis_data.get(symbol, 0.0)
                
                if current_price > 0 and avg_cost > 0:
                    if self._is_wash_sale(symbol, current_price, avg_cost, trade_history):
                        logger.warning(f"[Rebalancer] Wash sale prevented for {symbol}. Current px: {current_price}, Avg cost: {avg_cost}")
                        continue
                        
            orders.append(Order(symbol, side, diff_usd))
            
        # Optional: Sort orders so SELLs happen before BUYs (to free up capital)
        orders.sort(key=lambda o: 0 if o.side == "SELL" else 1)
            
        return orders

def main():
    logging.basicConfig(level=logging.INFO)
    print("Testing Portfolio Rebalancer...")
    
    # Mock Config
    class MockConfig:
        enabled = True
        trigger_mode = "threshold"
        drift_threshold_pct = 5.0
        target_allocations = {"BTC": 50.0, "ETH": 30.0, "SOL": 20.0}
        avoid_wash_sales = True
        
    rebalancer = Rebalancer(config=MockConfig())
    
    # Mock Portfolio
    total_value = 10000.0
    # Drifted: BTC is 70% ($7000), ETH is 20% ($2000), SOL is 10% ($1000)
    current_portfolio = {
        "BTC": 7000.0,
        "ETH": 2000.0,
        "SOL": 1000.0
    }
    
    # Prices and Costs
    price_data = {"BTC": 65000.0, "ETH": 3500.0, "SOL": 150.0}
    # BTC avg cost is high, so selling it now is a loss 
    cost_basis_data = {"BTC": 70000.0, "ETH": 3000.0, "SOL": 120.0}
    
    # Simulate a buy in the last 30 days
    recent_buy_ts = time.time() - (5 * 24 * 3600)
    trade_history = [{"symbol": "BTC", "side": "BUY", "timestamp": recent_buy_ts}]
    
    orders = rebalancer.generate_rebalance_orders(
        current_portfolio, 
        total_value, 
        price_data, 
        cost_basis_data, 
        trade_history
    )
    
    print("\nGenerated Orders (Wash Sale Prevention ON):")
    for o in orders:
        print(f"  {o}")
        
    # Test without wash sale prevention
    MockConfig.avoid_wash_sales = False
    rebalancer_no_wash = Rebalancer(config=MockConfig())
    orders_no_wash = rebalancer_no_wash.generate_rebalance_orders(
        current_portfolio, 
        total_value, 
        price_data, 
        cost_basis_data, 
        trade_history
    )
    
    print("\nGenerated Orders (Wash Sale Prevention OFF):")
    for o in orders_no_wash:
        print(f"  {o}")

if __name__ == "__main__":
    main()
