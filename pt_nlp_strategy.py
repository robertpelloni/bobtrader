"""
PowerTrader AI — NLP Strategy Generator & Auto-Strategy
==========================================================
Natural language to trading strategy code converter,
automated strategy generation, and transfer learning from patterns.

Features:
    1. NLPStrategyParser — parse natural language into strategy configs
    2. StrategyGenerator — auto-generate strategies from performance data
    3. TransferLearner — learn from successful strategy patterns

Usage:
    from pt_nlp_strategy import NLPStrategyParser, StrategyGenerator

    parser = NLPStrategyParser()
    config = parser.parse("Buy ETH when RSI drops below 30 and sell when above 70")

    gen = StrategyGenerator()
    strategies = gen.generate(n=5, base_performance=backtest_data)
"""

from __future__ import annotations
import json
import random
import re
from dataclasses import dataclass, asdict
from datetime import datetime
from typing import List, Dict, Optional, Tuple
from pathlib import Path


# =============================================================================
# NLP STRATEGY PARSER
# =============================================================================

class NLPStrategyParser:
    """
    Parses natural language strategy descriptions into executable configs.
    Uses keyword matching and pattern recognition (no external NLP libs).
    """

    # Keyword patterns for strategy components
    INDICATOR_PATTERNS = {
        r"rsi\s*(?:drops?\s*)?(?:below|under|<)\s*(\d+)": ("rsi_below", float),
        r"rsi\s*(?:rises?\s*)?(?:above|over|>)\s*(\d+)": ("rsi_above", float),
        r"(?:price\s*)?(?:above|over|breaks?\s*above)\s*(?:the\s*)?sma\s*(\d+)": ("price_above_sma", int),
        r"(?:price\s*)?(?:below|under|breaks?\s*below)\s*(?:the\s*)?sma\s*(\d+)": ("price_below_sma", int),
        r"macd\s*(?:cross(?:es|over)?)\s*(?:above|up)": ("macd_cross_above", None),
        r"macd\s*(?:cross(?:es|over)?)\s*(?:below|down)": ("macd_cross_below", None),
        r"volume\s*(?:is\s*)?(?:above|over|>|exceeds?)\s*(\d+(?:\.\d+)?)\s*[xX]": ("volume_above_x", float),
        r"bollinger\s*(?:band)?\s*(?:lower|bottom)": ("bb_lower_touch", None),
        r"bollinger\s*(?:band)?\s*(?:upper|top)": ("bb_upper_touch", None),
        r"(?:stop\s*loss|sl)\s*(?:at\s*)?(\d+(?:\.\d+)?)\s*%": ("stop_loss_pct", float),
        r"(?:take\s*profit|tp)\s*(?:at\s*)?(\d+(?:\.\d+)?)\s*%": ("take_profit_pct", float),
        r"atr\s*(?:stop)?\s*(\d+(?:\.\d+)?)\s*[xX]": ("atr_stop_multiplier", float),
    }

    COIN_PATTERN = r"\b(BTC|ETH|SOL|ADA|DOT|MATIC|LINK|AVAX|XRP|DOGE)\b"
    TIMEFRAME_PATTERN = r"(\d+)\s*(?:min(?:ute)?|hour|day|week)"

    ACTION_PATTERNS = {
        r"\b(?:buy|long|enter\s+long|go\s+long)\b": "buy",
        r"\b(?:sell|short|exit|close|go\s+short)\b": "sell",
    }

    def parse(self, description: str) -> dict:
        """Parse a natural language strategy description into a config."""
        desc_lower = description.lower()
        config = {
            "name": self._generate_name(description),
            "description": description,
            "coins": self._extract_coins(description),
            "timeframe": self._extract_timeframe(desc_lower),
            "entry_conditions": [],
            "exit_conditions": [],
            "risk_management": {},
            "generated_at": datetime.now().isoformat(),
            "source": "nlp",
        }

        # Extract buy/sell conditions
        sentences = re.split(r"[.;]|(?:\band\b|\bthen\b)", desc_lower)
        current_action = "buy"

        for sentence in sentences:
            # Detect action
            for pattern, action in self.ACTION_PATTERNS.items():
                if re.search(pattern, sentence):
                    current_action = action
                    break

            # Extract indicators
            for pattern, (indicator, cast_fn) in self.INDICATOR_PATTERNS.items():
                match = re.search(pattern, sentence)
                if match:
                    condition = {"indicator": indicator}
                    if cast_fn and match.groups():
                        condition["value"] = cast_fn(match.group(1))

                    if indicator.startswith("stop_loss") or indicator.startswith("take_profit") or indicator.startswith("atr_stop"):
                        config["risk_management"][indicator] = condition.get("value", 0)
                    elif current_action == "buy":
                        config["entry_conditions"].append(condition)
                    else:
                        config["exit_conditions"].append(condition)

        # Default risk management
        if "stop_loss_pct" not in config["risk_management"]:
            config["risk_management"]["stop_loss_pct"] = 5.0
        if "take_profit_pct" not in config["risk_management"]:
            config["risk_management"]["take_profit_pct"] = 10.0

        return config

    def _extract_coins(self, text: str) -> List[str]:
        matches = re.findall(self.COIN_PATTERN, text.upper())
        return list(set(matches)) if matches else ["BTC", "ETH"]

    def _extract_timeframe(self, text: str) -> str:
        match = re.search(self.TIMEFRAME_PATTERN, text)
        if match:
            n = int(match.group(1))
            unit = match.group(0).split()[-1]
            if "min" in unit: return f"{n}min"
            if "hour" in unit: return f"{n}hour"
            if "day" in unit: return f"{n}day"
            if "week" in unit: return f"{n}week"
        return "1hour"

    def _generate_name(self, description: str) -> str:
        words = description.lower().split()[:5]
        keywords = [w for w in words if len(w) > 2 and w not in ("the", "and", "when", "with")]
        return "_".join(keywords[:3]) or "custom_strategy"


# =============================================================================
# STRATEGY GENERATOR
# =============================================================================

class StrategyGenerator:
    """
    Auto-generates trading strategies by combining indicators,
    timeframes, and risk parameters, then ranks them by expected fitness.
    """

    INDICATORS = [
        {"type": "rsi", "params": {"period": 14, "oversold": 30, "overbought": 70}},
        {"type": "sma_cross", "params": {"fast": 10, "slow": 50}},
        {"type": "macd", "params": {"fast": 12, "slow": 26, "signal": 9}},
        {"type": "bollinger", "params": {"period": 20, "std": 2.0}},
        {"type": "atr_breakout", "params": {"period": 14, "multiplier": 2.0}},
        {"type": "volume_spike", "params": {"threshold": 2.0, "period": 20}},
        {"type": "ema_trend", "params": {"period": 200}},
        {"type": "stochastic", "params": {"k": 14, "d": 3, "oversold": 20, "overbought": 80}},
    ]

    TIMEFRAMES = ["5min", "15min", "1hour", "4hour", "1day"]

    RISK_PROFILES = {
        "conservative": {"stop_loss": 2.0, "take_profit": 4.0, "max_positions": 3},
        "moderate": {"stop_loss": 5.0, "take_profit": 10.0, "max_positions": 5},
        "aggressive": {"stop_loss": 10.0, "take_profit": 20.0, "max_positions": 8},
    }

    def generate(self, n: int = 10, seed: int = 42) -> List[dict]:
        """Generate n random strategy configurations."""
        random.seed(seed)
        strategies = []

        for i in range(n):
            # Pick 1-3 indicators
            n_indicators = random.randint(1, 3)
            indicators = random.sample(self.INDICATORS, n_indicators)

            # Mutate parameters
            mutated = []
            for ind in indicators:
                m = {"type": ind["type"], "params": {}}
                for k, v in ind["params"].items():
                    if isinstance(v, int):
                        m["params"][k] = max(1, v + random.randint(-v//3, v//3))
                    elif isinstance(v, float):
                        m["params"][k] = round(max(0.1, v * random.uniform(0.7, 1.3)), 2)
                mutated.append(m)

            timeframe = random.choice(self.TIMEFRAMES)
            risk_name = random.choice(list(self.RISK_PROFILES.keys()))
            risk = self.RISK_PROFILES[risk_name].copy()

            name = f"auto_{mutated[0]['type']}_{timeframe}_{risk_name}_{i}"

            strategy = {
                "name": name,
                "indicators": mutated,
                "timeframe": timeframe,
                "risk_profile": risk_name,
                "risk_params": risk,
                "fitness_score": self._estimate_fitness(mutated, timeframe, risk),
                "generated_at": datetime.now().isoformat(),
                "source": "auto_generator",
            }
            strategies.append(strategy)

        strategies.sort(key=lambda s: s["fitness_score"], reverse=True)
        return strategies

    def _estimate_fitness(self, indicators: List[dict], timeframe: str, risk: dict) -> float:
        """Heuristic fitness score based on indicator diversity and risk balance."""
        score = 50.0
        # Diversity bonus (more diverse indicators = higher potential)
        types = set(ind["type"] for ind in indicators)
        score += len(types) * 10
        # Risk-reward ratio bonus
        rr = risk.get("take_profit", 10) / max(risk.get("stop_loss", 5), 0.1)
        score += min(rr * 5, 20)
        # Timeframe modifier
        tf_scores = {"5min": -5, "15min": 0, "1hour": 5, "4hour": 10, "1day": 8}
        score += tf_scores.get(timeframe, 0)
        # Noise
        score += random.gauss(0, 5)
        return round(max(0, min(100, score)), 1)


# =============================================================================
# TRANSFER LEARNER
# =============================================================================

class TransferLearner:
    """
    Learns from successful strategy patterns to inform new strategy generation.
    Analyzes which indicator combinations and parameters perform best.
    """

    PATTERNS_FILE = Path("transfer_patterns.json")

    def __init__(self):
        self.patterns: Dict[str, dict] = self._load()

    def _load(self) -> Dict[str, dict]:
        if self.PATTERNS_FILE.exists():
            try:
                return json.loads(self.PATTERNS_FILE.read_text())
            except Exception:
                pass
        return {}

    def _save(self):
        self.PATTERNS_FILE.write_text(json.dumps(self.patterns, indent=2))

    def learn_from_results(self, strategy_name: str, config: dict,
                           backtest_result: dict):
        """Record a strategy's performance for pattern learning."""
        pattern = {
            "config": config,
            "return_pct": backtest_result.get("total_return", 0),
            "win_rate": backtest_result.get("win_rate", 0),
            "sharpe": backtest_result.get("sharpe", 0),
            "max_drawdown": backtest_result.get("max_drawdown", 0),
            "timestamp": datetime.now().isoformat(),
        }
        self.patterns[strategy_name] = pattern
        self._save()

    def get_best_patterns(self, metric: str = "sharpe",
                          top_n: int = 5) -> List[dict]:
        """Get top-performing strategy patterns."""
        sorted_patterns = sorted(
            self.patterns.items(),
            key=lambda x: x[1].get(metric, 0),
            reverse=True
        )
        return [{"name": k, **v} for k, v in sorted_patterns[:top_n]]

    def suggest_improvements(self, config: dict) -> List[str]:
        """Suggest improvements based on learned patterns."""
        suggestions = []
        best = self.get_best_patterns("sharpe", 3)

        if not best:
            return ["No patterns learned yet. Run backtests to build knowledge."]

        top_sharpe = best[0].get("sharpe", 0) if best else 0
        if top_sharpe > 2.0:
            suggestions.append(f"Top Sharpe: {top_sharpe:.2f} — consider similar parameters")

        # Analyze common traits among top performers
        top_configs = [p.get("config", {}) for p in best if p.get("config")]
        if top_configs:
            timeframes = [c.get("timeframe", "") for c in top_configs]
            common_tf = max(set(timeframes), key=timeframes.count) if timeframes else None
            if common_tf:
                suggestions.append(f"Most successful timeframe: {common_tf}")

        return suggestions if suggestions else ["Keep experimenting with different configurations"]


# =============================================================================
# SELF-TEST
# =============================================================================

if __name__ == "__main__":
    print("=" * 60)
    print("NLP Strategy & Auto-Generator — Self-Test")
    print("=" * 60)

    # 1. NLP Parser
    print("\n1. NLPStrategyParser...")
    parser = NLPStrategyParser()

    tests = [
        "Buy ETH when RSI drops below 30 and sell when RSI above 70 with stop loss at 5%",
        "Go long BTC when price breaks above SMA 200 and MACD crosses above. Take profit at 15%",
        "Short SOL when volume is above 3x normal and RSI above 80. ATR stop 2x",
    ]

    for desc in tests:
        config = parser.parse(desc)
        print(f"\n   Input: \"{desc[:60]}...\"")
        print(f"   Name: {config['name']}")
        print(f"   Coins: {config['coins']}")
        print(f"   Entry: {config['entry_conditions']}")
        print(f"   Exit: {config['exit_conditions']}")
        print(f"   Risk: {config['risk_management']}")

    # 2. Strategy Generator
    print("\n\n2. StrategyGenerator — auto-generate 5...")
    gen = StrategyGenerator()
    strategies = gen.generate(n=5)
    for s in strategies:
        print(f"   Score:{s['fitness_score']:>5.1f}  {s['name']}")

    # 3. Transfer Learner
    print("\n3. TransferLearner...")
    learner = TransferLearner()
    learner.learn_from_results("test_strat", {"timeframe": "1hour"},
                               {"total_return": 25, "sharpe": 1.8, "win_rate": 0.6, "max_drawdown": 8})
    suggestions = learner.suggest_improvements({})
    for s in suggestions:
        print(f"   → {s}")

    print("\n✅ Self-test complete")
