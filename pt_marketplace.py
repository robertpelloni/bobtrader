"""
PowerTrader AI — Trading Bot Marketplace
==========================================
Share, download, and rank trading strategies.
Includes a local marketplace with strategy packaging, backtesting leaderboard,
and community rating system.

Features:
    1. StrategyPackager — export/import strategy configs as JSON packages
    2. MarketplaceManager — local strategy store with search and install
    3. BacktestLeaderboard — ranks strategies by backtest performance
    4. CommunityRatings — rate and review strategies

Usage:
    from pt_marketplace import MarketplaceManager, StrategyPackager

    # Package a strategy
    pkg = StrategyPackager()
    pkg.export_strategy("my_strategy", config, "strategies/my_strat.json")

    # Browse marketplace
    mkt = MarketplaceManager()
    strategies = mkt.list_strategies()
    mkt.install_strategy("momentum_v2")
"""

from __future__ import annotations
import json
import time
import hashlib
import shutil
from dataclasses import dataclass, asdict, field
from datetime import datetime
from typing import List, Dict, Optional
from pathlib import Path


# =============================================================================
# DATA MODELS
# =============================================================================

@dataclass
class StrategyPackage:
    """A packaged strategy for sharing."""
    name: str
    version: str
    author: str
    description: str
    tags: List[str]
    config: dict
    backtest_results: dict = field(default_factory=dict)
    created_at: str = ""
    checksum: str = ""
    rating: float = 0.0
    downloads: int = 0

    def to_dict(self) -> dict:
        return asdict(self)


@dataclass
class LeaderboardEntry:
    """An entry in the strategy backtesting leaderboard."""
    strategy_name: str
    author: str
    total_return_pct: float
    win_rate: float
    max_drawdown: float
    sharpe_ratio: float
    total_trades: int
    period: str
    rank: int = 0
    score: float = 0.0


@dataclass
class StrategyReview:
    """A user review of a strategy."""
    strategy_name: str
    reviewer: str
    rating: int  # 1–5
    comment: str
    timestamp: str = ""


# =============================================================================
# STRATEGY PACKAGER
# =============================================================================

class StrategyPackager:
    """Export and import strategy configurations as shareable packages."""

    STRATEGIES_DIR = Path("strategies")

    def __init__(self):
        self.STRATEGIES_DIR.mkdir(exist_ok=True)

    def export_strategy(self, name: str, config: dict,
                        author: str = "anonymous",
                        description: str = "",
                        tags: Optional[List[str]] = None,
                        backtest_results: Optional[dict] = None) -> StrategyPackage:
        """Package a strategy config into a shareable JSON file."""
        pkg = StrategyPackage(
            name=name,
            version="1.0.0",
            author=author,
            description=description or f"Strategy: {name}",
            tags=tags or ["custom"],
            config=config,
            backtest_results=backtest_results or {},
            created_at=datetime.now().isoformat(),
        )

        # Generate checksum
        content = json.dumps(pkg.config, sort_keys=True)
        pkg.checksum = hashlib.sha256(content.encode()).hexdigest()[:16]

        # Save to file
        filepath = self.STRATEGIES_DIR / f"{name}.json"
        filepath.write_text(json.dumps(pkg.to_dict(), indent=2))

        return pkg

    def import_strategy(self, filepath: str) -> StrategyPackage:
        """Import a strategy package from a JSON file."""
        path = Path(filepath)
        if not path.exists():
            raise FileNotFoundError(f"Strategy file not found: {filepath}")

        data = json.loads(path.read_text())

        pkg = StrategyPackage(
            name=data.get("name", "unknown"),
            version=data.get("version", "1.0.0"),
            author=data.get("author", "anonymous"),
            description=data.get("description", ""),
            tags=data.get("tags", []),
            config=data.get("config", {}),
            backtest_results=data.get("backtest_results", {}),
            created_at=data.get("created_at", ""),
            checksum=data.get("checksum", ""),
            rating=data.get("rating", 0.0),
            downloads=data.get("downloads", 0),
        )

        # Verify checksum if present
        if pkg.checksum:
            content = json.dumps(pkg.config, sort_keys=True)
            computed = hashlib.sha256(content.encode()).hexdigest()[:16]
            if computed != pkg.checksum:
                print(f"Warning: Checksum mismatch for {pkg.name}")

        return pkg

    def list_local_strategies(self) -> List[StrategyPackage]:
        """List all locally saved strategies."""
        strategies = []
        for f in self.STRATEGIES_DIR.glob("*.json"):
            try:
                pkg = self.import_strategy(str(f))
                strategies.append(pkg)
            except Exception:
                pass
        return strategies


# =============================================================================
# MARKETPLACE MANAGER
# =============================================================================

class MarketplaceManager:
    """
    Local strategy marketplace with search, install, and rating capabilities.
    """

    MARKETPLACE_DIR = Path("marketplace")
    INSTALLED_DIR = Path("strategies/installed")
    CATALOG_FILE = Path("marketplace/catalog.json")

    def __init__(self):
        self.MARKETPLACE_DIR.mkdir(exist_ok=True)
        self.INSTALLED_DIR.mkdir(parents=True, exist_ok=True)
        self.catalog: List[dict] = self._load_catalog()

    def _load_catalog(self) -> List[dict]:
        if self.CATALOG_FILE.exists():
            try:
                return json.loads(self.CATALOG_FILE.read_text())
            except Exception:
                pass
        return self._default_catalog()

    def _save_catalog(self):
        self.CATALOG_FILE.write_text(json.dumps(self.catalog, indent=2))

    def _default_catalog(self) -> List[dict]:
        """Built-in strategy catalog."""
        catalog = [
            {
                "name": "momentum_breakout",
                "version": "2.1.0",
                "author": "PowerTrader",
                "description": "Momentum breakout strategy using volume confirmation and ATR-based stops",
                "tags": ["momentum", "breakout", "volume"],
                "rating": 4.5,
                "downloads": 1200,
                "config": {
                    "strategy": "momentum",
                    "entry_signal": "price_above_sma20_and_volume_2x",
                    "stop_loss_atr": 2.0,
                    "take_profit_atr": 3.0,
                    "timeframe": "1hour",
                    "min_volume_ratio": 2.0,
                },
                "backtest_results": {
                    "total_return": 45.2,
                    "win_rate": 0.58,
                    "max_drawdown": 12.5,
                    "sharpe": 1.8,
                    "trades": 340,
                    "period": "2024-01 to 2024-12",
                },
            },
            {
                "name": "mean_reversion_rsi",
                "version": "1.5.0",
                "author": "PowerTrader",
                "description": "Mean reversion using RSI oversold/overbought with Bollinger Band confirmation",
                "tags": ["mean-reversion", "rsi", "bollinger"],
                "rating": 4.2,
                "downloads": 890,
                "config": {
                    "strategy": "mean_reversion",
                    "rsi_oversold": 25,
                    "rsi_overbought": 75,
                    "bb_period": 20,
                    "bb_std": 2.0,
                    "timeframe": "4hour",
                },
                "backtest_results": {
                    "total_return": 32.1,
                    "win_rate": 0.62,
                    "max_drawdown": 8.3,
                    "sharpe": 2.1,
                    "trades": 210,
                    "period": "2024-01 to 2024-12",
                },
            },
            {
                "name": "dca_accumulator",
                "version": "3.0.0",
                "author": "PowerTrader",
                "description": "Dollar-cost averaging with neural level detection for optimal entry timing",
                "tags": ["dca", "neural", "long-term"],
                "rating": 4.8,
                "downloads": 2100,
                "config": {
                    "strategy": "dca",
                    "interval_hours": 24,
                    "allocation_pct": 5.0,
                    "neural_bias": True,
                    "max_positions": 10,
                },
                "backtest_results": {
                    "total_return": 28.5,
                    "win_rate": 0.71,
                    "max_drawdown": 5.2,
                    "sharpe": 2.5,
                    "trades": 365,
                    "period": "2024-01 to 2024-12",
                },
            },
            {
                "name": "grid_trader",
                "version": "1.2.0",
                "author": "community",
                "description": "Grid trading bot with dynamic level spacing based on ATR volatility",
                "tags": ["grid", "range", "automation"],
                "rating": 3.9,
                "downloads": 650,
                "config": {
                    "strategy": "grid",
                    "grid_levels": 10,
                    "spacing_method": "atr",
                    "atr_multiplier": 0.5,
                    "order_size_pct": 2.0,
                },
                "backtest_results": {
                    "total_return": 18.7,
                    "win_rate": 0.74,
                    "max_drawdown": 6.8,
                    "sharpe": 1.5,
                    "trades": 820,
                    "period": "2024-06 to 2024-12",
                },
            },
            {
                "name": "trend_follower_macd",
                "version": "2.0.0",
                "author": "community",
                "description": "Trend-following strategy with MACD crossover and EMA filter",
                "tags": ["trend", "macd", "ema"],
                "rating": 4.0,
                "downloads": 750,
                "config": {
                    "strategy": "trend",
                    "macd_fast": 12,
                    "macd_slow": 26,
                    "macd_signal": 9,
                    "ema_filter": 200,
                    "timeframe": "4hour",
                },
                "backtest_results": {
                    "total_return": 38.4,
                    "win_rate": 0.52,
                    "max_drawdown": 15.1,
                    "sharpe": 1.6,
                    "trades": 180,
                    "period": "2024-01 to 2024-12",
                },
            },
        ]
        self.catalog = catalog
        self._save_catalog()
        return catalog

    def list_strategies(self, tag: str = "", sort_by: str = "rating") -> List[dict]:
        """List marketplace strategies, optionally filtered by tag."""
        strategies = self.catalog

        if tag:
            strategies = [s for s in strategies if tag in s.get("tags", [])]

        key = sort_by if sort_by in ("rating", "downloads", "name") else "rating"
        reverse = key != "name"
        strategies.sort(key=lambda s: s.get(key, 0), reverse=reverse)

        return strategies

    def search(self, query: str) -> List[dict]:
        """Search strategies by name, description, or tags."""
        q = query.lower()
        results = []
        for s in self.catalog:
            if (q in s.get("name", "").lower() or
                q in s.get("description", "").lower() or
                any(q in t.lower() for t in s.get("tags", []))):
                results.append(s)
        return results

    def install_strategy(self, name: str) -> dict:
        """Install a strategy from the marketplace."""
        strategy = None
        for s in self.catalog:
            if s["name"] == name:
                strategy = s
                break

        if not strategy:
            raise ValueError(f"Strategy '{name}' not found in marketplace")

        # Save to installed dir
        filepath = self.INSTALLED_DIR / f"{name}.json"
        filepath.write_text(json.dumps(strategy, indent=2))

        # Increment download count
        strategy["downloads"] = strategy.get("downloads", 0) + 1
        self._save_catalog()

        return {"status": "installed", "name": name, "path": str(filepath)}

    def publish_strategy(self, package: StrategyPackage) -> dict:
        """Publish a strategy to the marketplace."""
        entry = package.to_dict()
        entry["downloads"] = 0
        entry["rating"] = 0.0

        # Check for duplicates
        for i, s in enumerate(self.catalog):
            if s["name"] == package.name:
                self.catalog[i] = entry
                self._save_catalog()
                return {"status": "updated", "name": package.name}

        self.catalog.append(entry)
        self._save_catalog()
        return {"status": "published", "name": package.name}


# =============================================================================
# BACKTESTING LEADERBOARD
# =============================================================================

class BacktestLeaderboard:
    """Ranks strategies by backtest performance with a composite score."""

    LEADERBOARD_FILE = Path("marketplace/leaderboard.json")

    def __init__(self):
        self.entries: List[LeaderboardEntry] = []

    def build_from_catalog(self, catalog: List[dict]) -> List[LeaderboardEntry]:
        """Build leaderboard from marketplace catalog backtest results."""
        self.entries = []

        for s in catalog:
            bt = s.get("backtest_results", {})
            if not bt:
                continue

            entry = LeaderboardEntry(
                strategy_name=s.get("name", "unknown"),
                author=s.get("author", "anonymous"),
                total_return_pct=bt.get("total_return", 0),
                win_rate=bt.get("win_rate", 0),
                max_drawdown=bt.get("max_drawdown", 0),
                sharpe_ratio=bt.get("sharpe", 0),
                total_trades=bt.get("trades", 0),
                period=bt.get("period", ""),
            )

            # Composite score: weighted combination
            entry.score = (
                entry.total_return_pct * 0.3 +
                entry.win_rate * 100 * 0.2 +
                entry.sharpe_ratio * 10 * 0.25 +
                (100 - entry.max_drawdown) * 0.15 +
                min(entry.total_trades / 10, 10) * 0.1
            )

            self.entries.append(entry)

        # Rank by composite score
        self.entries.sort(key=lambda e: e.score, reverse=True)
        for i, entry in enumerate(self.entries):
            entry.rank = i + 1

        self._save()
        return self.entries

    def _save(self):
        self.LEADERBOARD_FILE.parent.mkdir(exist_ok=True)
        data = [asdict(e) for e in self.entries]
        self.LEADERBOARD_FILE.write_text(json.dumps(data, indent=2))

    def get_top(self, n: int = 10) -> List[LeaderboardEntry]:
        return self.entries[:n]

    def print_leaderboard(self):
        print(f"\n{'Rank':>4} {'Strategy':<25} {'Return':>8} {'Win%':>6} {'Sharpe':>7} {'DD':>6} {'Score':>7}")
        print("-" * 70)
        for e in self.entries:
            print(f"{e.rank:>4} {e.strategy_name:<25} {e.total_return_pct:>7.1f}% {e.win_rate*100:>5.1f}% {e.sharpe_ratio:>7.2f} {e.max_drawdown:>5.1f}% {e.score:>7.1f}")


# =============================================================================
# COMMUNITY RATINGS
# =============================================================================

class CommunityRatings:
    """Rate and review marketplace strategies."""

    REVIEWS_FILE = Path("marketplace/reviews.json")

    def __init__(self):
        self.reviews: List[dict] = self._load()

    def _load(self) -> List[dict]:
        if self.REVIEWS_FILE.exists():
            try:
                return json.loads(self.REVIEWS_FILE.read_text())
            except Exception:
                return []
        return []

    def _save(self):
        self.REVIEWS_FILE.parent.mkdir(exist_ok=True)
        self.REVIEWS_FILE.write_text(json.dumps(self.reviews, indent=2))

    def add_review(self, strategy_name: str, reviewer: str,
                   rating: int, comment: str) -> dict:
        """Add a review (rating 1–5)."""
        rating = max(1, min(5, rating))
        review = {
            "strategy_name": strategy_name,
            "reviewer": reviewer,
            "rating": rating,
            "comment": comment,
            "timestamp": datetime.now().isoformat(),
        }
        self.reviews.append(review)
        self._save()
        return review

    def get_reviews(self, strategy_name: str) -> List[dict]:
        return [r for r in self.reviews if r["strategy_name"] == strategy_name]

    def get_average_rating(self, strategy_name: str) -> float:
        reviews = self.get_reviews(strategy_name)
        if not reviews:
            return 0.0
        return sum(r["rating"] for r in reviews) / len(reviews)


# =============================================================================
# SELF-TEST
# =============================================================================

if __name__ == "__main__":
    print("=" * 60)
    print("Trading Bot Marketplace — Self-Test")
    print("=" * 60)

    # 1. Strategy Packager
    print("\n1. StrategyPackager — export...")
    packager = StrategyPackager()
    pkg = packager.export_strategy(
        "test_momentum",
        config={"strategy": "momentum", "sma_period": 20, "volume_multiplier": 1.5},
        author="test_user",
        description="Test momentum strategy",
        tags=["momentum", "test"],
    )
    print(f"   Exported: {pkg.name} v{pkg.version} (checksum: {pkg.checksum})")

    # 2. Marketplace
    print("\n2. MarketplaceManager — listing...")
    mkt = MarketplaceManager()
    strategies = mkt.list_strategies()
    for s in strategies:
        print(f"   ★{s.get('rating', 0):.1f}  {s['name']:<25} ({s.get('downloads', 0)} downloads)")

    # 3. Search
    print("\n3. MarketplaceManager — search 'trend'...")
    results = mkt.search("trend")
    for r in results:
        print(f"   → {r['name']}: {r['description'][:50]}...")

    # 4. Leaderboard
    print("\n4. BacktestLeaderboard...")
    board = BacktestLeaderboard()
    board.build_from_catalog(strategies)
    board.print_leaderboard()

    # 5. Community Ratings
    print("\n5. CommunityRatings...")
    ratings = CommunityRatings()
    ratings.add_review("dca_accumulator", "tester", 5, "Excellent strategy!")
    avg = ratings.get_average_rating("dca_accumulator")
    print(f"   dca_accumulator avg rating: {avg:.1f}/5")

    print("\n✅ Self-test complete")
