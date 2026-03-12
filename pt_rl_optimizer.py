"""
PowerTrader AI — Reinforcement Learning Strategy Optimizer
============================================================
Uses tabular Q-Learning and policy gradient methods to optimize
trading strategy parameters through simulated market interaction.

Features:
    1. TradingEnvironment — Gym-like environment for RL agents
    2. QLearningAgent — Tabular Q-learning for discrete actions
    3. PolicyGradientAgent — REINFORCE algorithm for continuous params
    4. StrategyOptimizer — Wraps RL agents for strategy tuning

No external ML dependencies — pure Python + numpy implementation.

Usage:
    from pt_rl_optimizer import StrategyOptimizer

    optimizer = StrategyOptimizer()
    best_params = optimizer.optimize(
        candles=candle_data,
        param_space={"sma_fast": (5, 20), "sma_slow": (20, 100)},
        episodes=500
    )
"""

from __future__ import annotations
import json
import math
import random
import time
from dataclasses import dataclass, field
from datetime import datetime
from typing import List, Dict, Optional, Tuple, Any
from pathlib import Path

try:
    import numpy as np
except ImportError:
    np = None  # Fallback to pure Python


# =============================================================================
# TRADING ENVIRONMENT (Gym-like)
# =============================================================================

class TradingEnvironment:
    """
    Simulated trading environment for RL agents.
    State: [position, unrealized_pnl, rsi, volatility, trend]
    Actions: 0=hold, 1=buy, 2=sell
    Reward: realized P&L + position value change
    """

    def __init__(self, prices: List[float], features: Optional[List[List[float]]] = None):
        self.prices = prices
        self.features = features or self._compute_features(prices)
        self.n_steps = len(prices)
        self.reset()

    def reset(self) -> List[float]:
        self.step_idx = 0
        self.position = 0       # -1 short, 0 flat, 1 long
        self.entry_price = 0.0
        self.total_pnl = 0.0
        self.trade_count = 0
        self.equity_curve = [0.0]
        return self._get_state()

    def step(self, action: int) -> Tuple[List[float], float, bool]:
        """Execute action, return (next_state, reward, done)."""
        price = self.prices[self.step_idx]
        reward = 0.0

        if action == 1 and self.position <= 0:  # BUY
            if self.position == -1:  # Close short
                reward = self.entry_price - price
                self.total_pnl += reward
                self.trade_count += 1
            self.position = 1
            self.entry_price = price

        elif action == 2 and self.position >= 0:  # SELL
            if self.position == 1:  # Close long
                reward = price - self.entry_price
                self.total_pnl += reward
                self.trade_count += 1
            self.position = -1
            self.entry_price = price

        # Holding reward (unrealized change)
        if self.position == 1 and self.step_idx > 0:
            reward += (price - self.prices[self.step_idx - 1]) * 0.1
        elif self.position == -1 and self.step_idx > 0:
            reward += (self.prices[self.step_idx - 1] - price) * 0.1

        self.equity_curve.append(self.total_pnl)
        self.step_idx += 1
        done = self.step_idx >= self.n_steps - 1

        # Force close at end
        if done and self.position != 0:
            if self.position == 1:
                self.total_pnl += price - self.entry_price
            else:
                self.total_pnl += self.entry_price - price
            self.position = 0

        return self._get_state(), reward, done

    def _get_state(self) -> List[float]:
        """Return current state vector."""
        idx = min(self.step_idx, len(self.features) - 1)
        base = self.features[idx]
        return [float(self.position)] + base

    def _compute_features(self, prices: List[float]) -> List[List[float]]:
        """Compute basic features from price series."""
        features = []
        for i in range(len(prices)):
            # Returns
            ret = (prices[i] / prices[max(0, i-1)] - 1) * 100 if i > 0 else 0
            # SMA ratio
            window = prices[max(0, i-20):i+1]
            sma = sum(window) / len(window) if window else prices[i]
            sma_ratio = prices[i] / sma - 1 if sma > 0 else 0
            # Volatility (std of last 20 returns)
            if i >= 20:
                rets = [(prices[j] / prices[j-1] - 1) for j in range(max(1, i-20), i+1)]
                vol = (sum(r*r for r in rets) / len(rets)) ** 0.5 if rets else 0
            else:
                vol = 0
            # RSI-like
            if i >= 14:
                gains = [max(0, prices[j] - prices[j-1]) for j in range(max(1, i-14), i+1)]
                losses = [max(0, prices[j-1] - prices[j]) for j in range(max(1, i-14), i+1)]
                avg_gain = sum(gains) / len(gains) if gains else 0
                avg_loss = sum(losses) / len(losses) if losses else 1
                rsi = 100 - 100 / (1 + avg_gain / (avg_loss + 1e-10))
            else:
                rsi = 50
            # Trend (price vs SMA-50)
            w50 = prices[max(0, i-50):i+1]
            sma50 = sum(w50) / len(w50) if w50 else prices[i]
            trend = 1.0 if prices[i] > sma50 else -1.0

            features.append([ret, sma_ratio, vol, rsi / 100.0, trend])
        return features


# =============================================================================
# Q-LEARNING AGENT
# =============================================================================

class QLearningAgent:
    """
    Tabular Q-Learning agent with discretized state space.
    Good for learning simple trading rules.
    """

    def __init__(self, state_bins: int = 10, n_actions: int = 3,
                 lr: float = 0.1, gamma: float = 0.95, epsilon: float = 0.1):
        self.state_bins = state_bins
        self.n_actions = n_actions
        self.lr = lr
        self.gamma = gamma
        self.epsilon = epsilon
        self.q_table: Dict[tuple, List[float]] = {}
        self.training_rewards: List[float] = []

    def _discretize(self, state: List[float]) -> tuple:
        """Discretize continuous state into bins."""
        binned = []
        for v in state:
            b = int(v * self.state_bins)
            b = max(-self.state_bins, min(self.state_bins, b))
            binned.append(b)
        return tuple(binned)

    def get_action(self, state: List[float], training: bool = True) -> int:
        """Epsilon-greedy action selection."""
        if training and random.random() < self.epsilon:
            return random.randint(0, self.n_actions - 1)

        key = self._discretize(state)
        if key not in self.q_table:
            self.q_table[key] = [0.0] * self.n_actions
        return max(range(self.n_actions), key=lambda a: self.q_table[key][a])

    def update(self, state: List[float], action: int,
               reward: float, next_state: List[float], done: bool):
        """Q-Learning update rule."""
        key = self._discretize(state)
        next_key = self._discretize(next_state)

        if key not in self.q_table:
            self.q_table[key] = [0.0] * self.n_actions
        if next_key not in self.q_table:
            self.q_table[next_key] = [0.0] * self.n_actions

        target = reward
        if not done:
            target += self.gamma * max(self.q_table[next_key])

        self.q_table[key][action] += self.lr * (target - self.q_table[key][action])

    def train(self, env: TradingEnvironment, episodes: int = 500) -> dict:
        """Train the agent on the environment."""
        self.training_rewards = []

        for ep in range(episodes):
            state = env.reset()
            total_reward = 0.0

            while True:
                action = self.get_action(state, training=True)
                next_state, reward, done = env.step(action)
                self.update(state, action, reward, next_state, done)
                state = next_state
                total_reward += reward

                if done:
                    break

            self.training_rewards.append(total_reward)

            # Decay epsilon
            self.epsilon = max(0.01, self.epsilon * 0.995)

        return {
            "episodes": episodes,
            "final_pnl": env.total_pnl,
            "trades": env.trade_count,
            "avg_reward_last_50": sum(self.training_rewards[-50:]) / min(50, len(self.training_rewards)),
            "q_table_size": len(self.q_table),
        }


# =============================================================================
# POLICY GRADIENT AGENT (REINFORCE)
# =============================================================================

class PolicyGradientAgent:
    """
    Simple policy gradient agent using REINFORCE.
    Uses a single-layer softmax policy (no neural network needed).
    """

    def __init__(self, state_dim: int = 6, n_actions: int = 3,
                 lr: float = 0.01, gamma: float = 0.99):
        self.state_dim = state_dim
        self.n_actions = n_actions
        self.lr = lr
        self.gamma = gamma

        # Linear policy weights
        if np is not None:
            self.weights = np.random.randn(state_dim, n_actions) * 0.01
        else:
            self.weights = [[random.gauss(0, 0.01) for _ in range(n_actions)]
                           for _ in range(state_dim)]

        self.episode_log: List[Tuple] = []  # (state, action, reward)
        self.training_rewards: List[float] = []

    def _softmax(self, logits: List[float]) -> List[float]:
        max_l = max(logits)
        exp_l = [math.exp(l - max_l) for l in logits]
        total = sum(exp_l)
        return [e / total for e in exp_l]

    def _forward(self, state: List[float]) -> List[float]:
        """Compute action probabilities."""
        if np is not None:
            logits = np.dot(state, self.weights).tolist()
        else:
            logits = [sum(s * w for s, w in zip(state, col))
                     for col in zip(*self.weights)]
        return self._softmax(logits)

    def get_action(self, state: List[float]) -> int:
        probs = self._forward(state)
        r = random.random()
        cumulative = 0.0
        for i, p in enumerate(probs):
            cumulative += p
            if r <= cumulative:
                return i
        return len(probs) - 1

    def store_transition(self, state: List[float], action: int, reward: float):
        self.episode_log.append((state, action, reward))

    def update(self):
        """REINFORCE update at end of episode."""
        if not self.episode_log:
            return

        # Compute discounted returns
        returns = []
        G = 0.0
        for _, _, r in reversed(self.episode_log):
            G = r + self.gamma * G
            returns.insert(0, G)

        # Normalize returns
        mean_r = sum(returns) / len(returns)
        std_r = (sum((r - mean_r)**2 for r in returns) / len(returns))**0.5 + 1e-8
        returns = [(r - mean_r) / std_r for r in returns]

        # Update weights
        for (state, action, _), G in zip(self.episode_log, returns):
            probs = self._forward(state)
            for a in range(self.n_actions):
                for s_idx in range(min(len(state), self.state_dim)):
                    grad = state[s_idx] * ((1.0 if a == action else 0.0) - probs[a])
                    if np is not None:
                        self.weights[s_idx][a] += self.lr * grad * G
                    else:
                        self.weights[s_idx][a] += self.lr * grad * G

        self.episode_log = []

    def train(self, env: TradingEnvironment, episodes: int = 500) -> dict:
        """Train using REINFORCE."""
        self.training_rewards = []

        for ep in range(episodes):
            state = env.reset()
            total_reward = 0.0

            while True:
                action = self.get_action(state)
                next_state, reward, done = env.step(action)
                self.store_transition(state, action, reward)
                state = next_state
                total_reward += reward

                if done:
                    break

            self.update()
            self.training_rewards.append(total_reward)

        return {
            "episodes": episodes,
            "final_pnl": env.total_pnl,
            "trades": env.trade_count,
            "avg_reward_last_50": sum(self.training_rewards[-50:]) / min(50, len(self.training_rewards)),
        }


# =============================================================================
# STRATEGY OPTIMIZER
# =============================================================================

class StrategyOptimizer:
    """
    High-level interface for optimizing strategy parameters using RL.
    Wraps Q-Learning and Policy Gradient agents.
    """

    def __init__(self):
        self.results: List[dict] = []

    def optimize(self, prices: List[float],
                 method: str = "qlearning",
                 episodes: int = 500) -> dict:
        """
        Optimize trading strategy using RL.
        
        Args:
            prices: Historical price series
            method: "qlearning" or "policy_gradient"
            episodes: Training episodes
        """
        env = TradingEnvironment(prices)

        if method == "qlearning":
            agent = QLearningAgent()
            result = agent.train(env, episodes)
            result["method"] = "qlearning"
        elif method == "policy_gradient":
            agent = PolicyGradientAgent()
            result = agent.train(env, episodes)
            result["method"] = "policy_gradient"
        else:
            raise ValueError(f"Unknown method: {method}")

        result["timestamp"] = datetime.now().isoformat()
        self.results.append(result)

        # Save results
        Path("rl_optimization_results.json").write_text(
            json.dumps(self.results, indent=2)
        )

        return result

    def compare_methods(self, prices: List[float],
                        episodes: int = 300) -> dict:
        """Compare Q-Learning vs Policy Gradient on same data."""
        ql_result = self.optimize(prices, "qlearning", episodes)
        pg_result = self.optimize(prices, "policy_gradient", episodes)

        return {
            "qlearning": ql_result,
            "policy_gradient": pg_result,
            "winner": "qlearning" if ql_result["final_pnl"] > pg_result["final_pnl"] else "policy_gradient",
        }


# =============================================================================
# SELF-TEST
# =============================================================================

if __name__ == "__main__":
    print("=" * 60)
    print("RL Strategy Optimizer — Self-Test")
    print("=" * 60)

    # Generate synthetic price data (random walk with trend)
    random.seed(42)
    prices = [100.0]
    for _ in range(499):
        change = random.gauss(0.0002, 0.02)
        prices.append(prices[-1] * (1 + change))

    # 1. Q-Learning
    print("\n1. Q-Learning Agent (300 episodes)...")
    optimizer = StrategyOptimizer()
    ql_result = optimizer.optimize(prices, "qlearning", 300)
    print(f"   Final P&L: ${ql_result['final_pnl']:.2f}")
    print(f"   Trades: {ql_result['trades']}")
    print(f"   Avg reward (last50): {ql_result['avg_reward_last_50']:.4f}")

    # 2. Policy Gradient
    print("\n2. Policy Gradient Agent (300 episodes)...")
    pg_result = optimizer.optimize(prices, "policy_gradient", 300)
    print(f"   Final P&L: ${pg_result['final_pnl']:.2f}")
    print(f"   Trades: {pg_result['trades']}")
    print(f"   Avg reward (last50): {pg_result['avg_reward_last_50']:.4f}")

    # 3. Environment test
    print("\n3. TradingEnvironment...")
    env = TradingEnvironment(prices)
    state = env.reset()
    print(f"   State dim: {len(state)}")
    print(f"   Price range: ${min(prices):.2f} - ${max(prices):.2f}")

    print("\n✅ Self-test complete")
