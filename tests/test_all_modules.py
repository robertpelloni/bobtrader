"""
PowerTrader AI — Comprehensive Test Suite
============================================
Unit tests for all v3.0.0 and v4.0.0 modules.
Run: python3 -m pytest tests/test_all_modules.py -v
  or: python3 tests/test_all_modules.py
"""

import json
import random
import sys
import os
import unittest
from pathlib import Path

# Add parent dir to path
sys.path.insert(0, str(Path(__file__).parent.parent))


# =============================================================================
# MULTI-EXCHANGE TESTS
# =============================================================================

class TestMultiExchange(unittest.TestCase):
    """Tests for pt_multi_exchange.py"""

    def test_import(self):
        from pt_multi_exchange import SmartRouter, ArbitrageExecutor, LiquidityAggregator, ExecutionTracker
        self.assertTrue(True)

    def test_execution_result_serialization(self):
        from pt_multi_exchange import ExecutionResult, ExecutionLeg
        leg = ExecutionLeg("binance", "buy", "BTC", 0.1, 60000, "filled", 60000, 0.1, 6.0)
        result = ExecutionResult("test_1", "BTC", "buy", 0.1, 60000, 6.0, [leg], "filled")
        d = result.to_dict()
        self.assertEqual(d["coin"], "BTC")
        self.assertEqual(d["total_quantity"], 0.1)
        self.assertEqual(len(d["legs"]), 1)

    def test_fee_schedule(self):
        from pt_multi_exchange import EXCHANGE_FEES, min_profitable_spread
        self.assertEqual(EXCHANGE_FEES["binance"], 0.10)
        self.assertEqual(EXCHANGE_FEES["kucoin"], 0.10)
        spread = min_profitable_spread("binance", "kucoin")
        self.assertAlmostEqual(spread, 0.20)

    def test_execution_tracker(self):
        from pt_multi_exchange import ExecutionTracker
        tracker = ExecutionTracker()
        stats = tracker.get_stats()
        self.assertIn("total_executions", stats)


# =============================================================================
# DEFI TESTS
# =============================================================================

class TestDeFi(unittest.TestCase):
    """Tests for pt_defi.py"""

    def test_import(self):
        from pt_defi import DEXRouter, YieldFarmScanner, GasOptimizer, DeFiPortfolio
        self.assertTrue(True)

    def test_token_config(self):
        from pt_defi import TOKENS, TOKEN_DECIMALS
        self.assertIn("ETH", TOKENS)
        self.assertIn("USDC", TOKENS)
        self.assertEqual(TOKEN_DECIMALS["USDC"], 6)
        self.assertEqual(TOKEN_DECIMALS["ETH"], 18)

    def test_gas_price_model(self):
        from pt_defi import GasPrice
        gas = GasPrice(15, 20, 30, 45, 18)
        costs = gas.cost_for_swap(150000)
        self.assertIn("slow", costs)
        self.assertIn("fast", costs)
        self.assertGreater(costs["fast"], costs["slow"])

    def test_defi_portfolio(self):
        from pt_defi import DeFiPortfolio
        portfolio = DeFiPortfolio()
        summary = portfolio.get_summary()
        self.assertIn("total_positions", summary)

    def test_fallback_farms(self):
        from pt_defi import YieldFarmScanner
        scanner = YieldFarmScanner()
        farms = scanner._get_fallback_farms()
        self.assertGreater(len(farms), 0)
        self.assertEqual(farms[0].protocol, "aave-v3")


# =============================================================================
# MARKETPLACE TESTS
# =============================================================================

class TestMarketplace(unittest.TestCase):
    """Tests for pt_marketplace.py"""

    def test_import(self):
        from pt_marketplace import MarketplaceManager, StrategyPackager, BacktestLeaderboard, CommunityRatings
        self.assertTrue(True)

    def test_strategy_packager(self):
        from pt_marketplace import StrategyPackager
        pkg = StrategyPackager()
        result = pkg.export_strategy("test_strat", {"sma": 20}, author="tester")
        self.assertEqual(result.name, "test_strat")
        self.assertNotEqual(result.checksum, "")

    def test_marketplace_default_catalog(self):
        from pt_marketplace import MarketplaceManager
        mkt = MarketplaceManager()
        strategies = mkt.list_strategies()
        self.assertGreater(len(strategies), 0)

    def test_marketplace_search(self):
        from pt_marketplace import MarketplaceManager
        mkt = MarketplaceManager()
        results = mkt.search("momentum")
        self.assertGreater(len(results), 0)

    def test_leaderboard_scoring(self):
        from pt_marketplace import BacktestLeaderboard
        board = BacktestLeaderboard()
        catalog = [{"name": "s1", "backtest_results": {
            "total_return": 30, "win_rate": 0.6, "max_drawdown": 10, "sharpe": 2.0, "trades": 100
        }}]
        entries = board.build_from_catalog(catalog)
        self.assertEqual(len(entries), 1)
        self.assertEqual(entries[0].rank, 1)
        self.assertGreater(entries[0].score, 0)

    def test_community_ratings(self):
        from pt_marketplace import CommunityRatings
        ratings = CommunityRatings()
        ratings.add_review("test_strat", "user1", 4, "Good")
        ratings.add_review("test_strat", "user2", 5, "Great")
        avg = ratings.get_average_rating("test_strat")
        self.assertGreaterEqual(avg, 4.0)


# =============================================================================
# RL OPTIMIZER TESTS
# =============================================================================

class TestRLOptimizer(unittest.TestCase):
    """Tests for pt_rl_optimizer.py"""

    def setUp(self):
        random.seed(42)
        self.prices = [100.0]
        for _ in range(99):
            self.prices.append(self.prices[-1] * (1 + random.gauss(0, 0.01)))

    def test_import(self):
        from pt_rl_optimizer import TradingEnvironment, QLearningAgent, PolicyGradientAgent, StrategyOptimizer
        self.assertTrue(True)

    def test_environment_reset(self):
        from pt_rl_optimizer import TradingEnvironment
        env = TradingEnvironment(self.prices)
        state = env.reset()
        self.assertEqual(len(state), 6)  # position + 5 features
        self.assertEqual(env.position, 0)

    def test_environment_step(self):
        from pt_rl_optimizer import TradingEnvironment
        env = TradingEnvironment(self.prices)
        env.reset()
        state, reward, done = env.step(1)  # BUY
        self.assertEqual(env.position, 1)

    def test_qlearning_train(self):
        from pt_rl_optimizer import TradingEnvironment, QLearningAgent
        env = TradingEnvironment(self.prices)
        agent = QLearningAgent()
        result = agent.train(env, episodes=10)
        self.assertEqual(result["episodes"], 10)
        self.assertIn("final_pnl", result)

    def test_policy_gradient_train(self):
        from pt_rl_optimizer import TradingEnvironment, PolicyGradientAgent
        env = TradingEnvironment(self.prices)
        agent = PolicyGradientAgent()
        result = agent.train(env, episodes=10)
        self.assertEqual(result["episodes"], 10)

    def test_strategy_optimizer(self):
        from pt_rl_optimizer import StrategyOptimizer
        opt = StrategyOptimizer()
        result = opt.optimize(self.prices, "qlearning", episodes=10)
        self.assertEqual(result["method"], "qlearning")


# =============================================================================
# NLP STRATEGY TESTS
# =============================================================================

class TestNLPStrategy(unittest.TestCase):
    """Tests for pt_nlp_strategy.py"""

    def test_import(self):
        from pt_nlp_strategy import NLPStrategyParser, StrategyGenerator, TransferLearner
        self.assertTrue(True)

    def test_parse_rsi_strategy(self):
        from pt_nlp_strategy import NLPStrategyParser
        parser = NLPStrategyParser()
        config = parser.parse("Buy ETH when RSI drops below 30")
        self.assertIn("ETH", config["coins"])
        self.assertGreater(len(config["entry_conditions"]), 0)
        self.assertEqual(config["entry_conditions"][0]["indicator"], "rsi_below")

    def test_parse_coins(self):
        from pt_nlp_strategy import NLPStrategyParser
        parser = NLPStrategyParser()
        config = parser.parse("Trade BTC and SOL with SMA 200")
        self.assertIn("BTC", config["coins"])
        self.assertIn("SOL", config["coins"])

    def test_parse_risk_management(self):
        from pt_nlp_strategy import NLPStrategyParser
        parser = NLPStrategyParser()
        config = parser.parse("Buy ETH with stop loss at 3% and take profit at 8%")
        self.assertEqual(config["risk_management"]["stop_loss_pct"], 3.0)
        self.assertEqual(config["risk_management"]["take_profit_pct"], 8.0)

    def test_strategy_generator(self):
        from pt_nlp_strategy import StrategyGenerator
        gen = StrategyGenerator()
        strategies = gen.generate(n=5, seed=42)
        self.assertEqual(len(strategies), 5)
        # Should be sorted by fitness
        self.assertGreaterEqual(strategies[0]["fitness_score"], strategies[-1]["fitness_score"])

    def test_transfer_learner(self):
        from pt_nlp_strategy import TransferLearner
        learner = TransferLearner()
        learner.learn_from_results("test", {"timeframe": "1hour"}, {"sharpe": 2.0})
        best = learner.get_best_patterns("sharpe", 1)
        self.assertEqual(len(best), 1)


# =============================================================================
# ADVANCED ANALYTICS TESTS
# =============================================================================

class TestAdvancedAnalytics(unittest.TestCase):
    """Tests for pt_advanced_analytics.py"""

    def setUp(self):
        random.seed(42)
        self.prices = [100.0]
        self.volumes = [1000.0]
        for _ in range(199):
            self.prices.append(self.prices[-1] * (1 + random.gauss(0, 0.01)))
            self.volumes.append(random.uniform(500, 2000))

    def test_import(self):
        from pt_advanced_analytics import PatternRecognizer, MicrostructureAnalyzer, OrderFlowAnalyzer, StatArbEngine
        self.assertTrue(True)

    def test_pattern_scan(self):
        from pt_advanced_analytics import PatternRecognizer
        pr = PatternRecognizer()
        patterns = pr.scan(self.prices, self.volumes)
        self.assertIsInstance(patterns, list)

    def test_support_resistance(self):
        from pt_advanced_analytics import PatternRecognizer
        pr = PatternRecognizer()
        levels = pr._detect_support_resistance(self.prices)
        for level in levels:
            self.assertIn("price", level)
            self.assertIn("touches", level)

    def test_order_flow(self):
        from pt_advanced_analytics import OrderFlowAnalyzer
        ofa = OrderFlowAnalyzer()
        result = ofa.analyze(self.prices, self.volumes)
        self.assertIn("recent_delta", result)
        self.assertIn("cvd_current", result)

    def test_volume_delta(self):
        from pt_advanced_analytics import OrderFlowAnalyzer
        ofa = OrderFlowAnalyzer()
        deltas = ofa.compute_volume_delta(self.prices, self.volumes)
        self.assertEqual(len(deltas), len(self.prices))

    def test_stat_arb_correlation(self):
        from pt_advanced_analytics import StatArbEngine
        arb = StatArbEngine()
        corr = arb.compute_correlation(self.prices, self.prices)
        self.assertAlmostEqual(corr, 1.0, places=3)

    def test_microstructure(self):
        from pt_advanced_analytics import MicrostructureAnalyzer
        ms = MicrostructureAnalyzer()
        bids = [(100 - i*0.1, 5.0) for i in range(5)]
        asks = [(100.1 + i*0.1, 5.0) for i in range(5)]
        result = ms.analyze_orderbook(bids, asks)
        self.assertGreater(result["spread"], 0)
        self.assertIn("imbalance_signal", result)


# =============================================================================
# ENTERPRISE TESTS
# =============================================================================

class TestEnterprise(unittest.TestCase):
    """Tests for pt_enterprise.py"""

    def test_import(self):
        from pt_enterprise import RBACManager, AccountManager, AuditLogger, ComplianceReporter
        self.assertTrue(True)

    def test_rbac_permissions(self):
        from pt_enterprise import RBACManager, Permission
        rbac = RBACManager()
        admin = rbac.create_user("test_admin", "admin")
        viewer = rbac.create_user("test_viewer", "viewer")
        self.assertTrue(rbac.check_permission(admin.user_id, Permission.MANAGE_USERS))
        self.assertFalse(rbac.check_permission(viewer.user_id, Permission.EXECUTE_TRADES))
        self.assertTrue(rbac.check_permission(viewer.user_id, Permission.VIEW_DASHBOARD))

    def test_account_manager(self):
        from pt_enterprise import AccountManager
        mgr = AccountManager()
        acct = mgr.create_account("Test Fund", 50000)
        self.assertEqual(acct.capital, 50000)
        self.assertEqual(acct.available, 50000)
        summary = mgr.get_all_accounts_summary()
        self.assertGreater(summary["total_capital"], 0)

    def test_audit_logger(self):
        from pt_enterprise import AuditLogger
        audit = AuditLogger()
        audit.log("test_action", "test_user", {"key": "value"})
        recent = audit.get_recent(5)
        self.assertGreater(len(recent), 0)

    def test_audit_integrity(self):
        from pt_enterprise import AuditLogger
        audit = AuditLogger()
        audit.log("integrity_test", "system")
        result = audit.verify_integrity()
        self.assertEqual(result["status"], "valid")

    def test_compliance_report(self):
        from pt_enterprise import ComplianceReporter
        reporter = ComplianceReporter()
        report = reporter.generate_report()
        self.assertIn("summary", report)
        self.assertIn("recommendations", report)


# =============================================================================
# RUNNER
# =============================================================================

if __name__ == "__main__":
    # Count tests
    loader = unittest.TestLoader()
    suite = loader.loadTestsFromModule(sys.modules[__name__])
    total = suite.countTestCases()
    print(f"\n{'='*60}")
    print(f"  PowerTrader AI — Comprehensive Test Suite ({total} tests)")
    print(f"{'='*60}\n")

    runner = unittest.TextTestRunner(verbosity=2)
    result = runner.run(suite)

    print(f"\n{'='*60}")
    if result.wasSuccessful():
        print(f"  ✅ ALL {total} TESTS PASSED")
    else:
        print(f"  ❌ {len(result.failures)} failures, {len(result.errors)} errors")
    print(f"{'='*60}")
