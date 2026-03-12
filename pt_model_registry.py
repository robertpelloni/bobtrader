"""
PowerTrader AI — Model Registry & A/B Testing
================================================
Manages model versions with promotion, rollback, and live A/B testing
of champion vs challenger models.

Usage:
    python3 pt_model_registry.py       # Run self-test
"""

from __future__ import annotations
import json
import shutil
import uuid
from dataclasses import dataclass, field, asdict
from datetime import datetime
from pathlib import Path
from typing import List, Dict, Optional, Tuple

import numpy as np


# =============================================================================
# DATA STRUCTURES
# =============================================================================


@dataclass
class VersionEntry:
    """A single model version in the registry."""
    version_id: str
    created_at: str
    accuracy: float
    total_predictions: int
    feature_hash: str
    description: str
    is_active: bool = False

    def to_dict(self) -> dict:
        return asdict(self)

    @classmethod
    def from_dict(cls, d: dict) -> "VersionEntry":
        return cls(**d)


@dataclass
class ABTestResult:
    """Result of an A/B test between two model versions."""
    champion_id: str
    challenger_id: str
    champion_wins: int = 0
    challenger_wins: int = 0
    total_rounds: int = 0
    target_rounds: int = 100
    is_active: bool = False
    started_at: str = ""

    @property
    def champion_rate(self) -> float:
        return self.champion_wins / self.total_rounds if self.total_rounds > 0 else 0.5

    @property
    def challenger_rate(self) -> float:
        return self.challenger_wins / self.total_rounds if self.total_rounds > 0 else 0.5

    @property
    def is_complete(self) -> bool:
        return self.total_rounds >= self.target_rounds

    def to_dict(self) -> dict:
        return asdict(self)

    @classmethod
    def from_dict(cls, d: dict) -> "ABTestResult":
        return cls(**d)


# =============================================================================
# MODEL REGISTRY
# =============================================================================


class ModelRegistry:
    """
    Manages model versions in a local registry (JSON + model files).
    Supports version registration, rollback, comparison, and A/B testing.
    """

    def __init__(self, registry_dir: str = "ml_models"):
        self.registry_dir = Path(registry_dir)
        self.registry_dir.mkdir(parents=True, exist_ok=True)
        self.registry_file = self.registry_dir / "registry.json"
        self.versions: List[VersionEntry] = []
        self.ab_test: Optional[ABTestResult] = None
        self.promotion_threshold: float = 0.01  # 1% accuracy improvement required
        self._load_registry()

    def _load_registry(self):
        """Load registry from disk."""
        if self.registry_file.exists():
            with open(self.registry_file) as f:
                data = json.load(f)
            self.versions = [VersionEntry.from_dict(v) for v in data.get("versions", [])]
            ab_data = data.get("ab_test")
            if ab_data:
                self.ab_test = ABTestResult.from_dict(ab_data)
        else:
            self.versions = []

    def _save_registry(self):
        """Save registry to disk."""
        data = {
            "versions": [v.to_dict() for v in self.versions],
            "ab_test": self.ab_test.to_dict() if self.ab_test else None,
        }
        with open(self.registry_file, "w") as f:
            json.dump(data, f, indent=2)

    def register(
        self,
        accuracy: float,
        total_predictions: int,
        feature_hash: str,
        description: str = "",
        model_dir: Optional[Path] = None,
        auto_promote: bool = True,
    ) -> str:
        """
        Register a new model version, optionally auto-promoting if it beats the active model.
        Returns the new version_id.
        """
        version_id = f"v{len(self.versions) + 1}_{uuid.uuid4().hex[:6]}"
        entry = VersionEntry(
            version_id=version_id,
            created_at=datetime.now().isoformat(),
            accuracy=accuracy,
            total_predictions=total_predictions,
            feature_hash=feature_hash,
            description=description,
            is_active=False,
        )

        # Copy model files if provided
        if model_dir and model_dir.exists():
            dest = self.registry_dir / version_id
            if dest.exists():
                shutil.rmtree(dest)
            shutil.copytree(str(model_dir), str(dest))

        self.versions.append(entry)

        # Auto-promote if better than current active
        if auto_promote:
            active = self.get_active_version()
            if active is None or accuracy > active.accuracy + self.promotion_threshold:
                self._promote(version_id)

        self._save_registry()
        return version_id

    def _promote(self, version_id: str):
        """Set a version as the active model."""
        for v in self.versions:
            v.is_active = (v.version_id == version_id)

    def get_active_version(self) -> Optional[VersionEntry]:
        """Get the currently active model version."""
        for v in self.versions:
            if v.is_active:
                return v
        return None

    def rollback(self, version_id: str) -> bool:
        """Rollback to a specific version."""
        for v in self.versions:
            if v.version_id == version_id:
                self._promote(version_id)
                self._save_registry()
                return True
        return False

    def compare_versions(self) -> str:
        """Print a comparison table of all versions."""
        lines = []
        lines.append("=" * 80)
        lines.append("MODEL VERSION REGISTRY")
        lines.append("=" * 80)
        lines.append(f"{'ID':<20} {'Date':<20} {'Accuracy':>8} {'Preds':>8} {'Active':>8}")
        lines.append("-" * 80)
        for v in self.versions:
            active_flag = "  ★" if v.is_active else ""
            lines.append(
                f"{v.version_id:<20} {v.created_at[:19]:<20} "
                f"{v.accuracy:>7.2%} {v.total_predictions:>8}{active_flag}"
            )
        lines.append("=" * 80)
        return "\n".join(lines)

    # -------------------------------------------------------------------------
    # A/B TESTING
    # -------------------------------------------------------------------------

    def start_ab_test(
        self, champion_id: str, challenger_id: str, target_rounds: int = 100
    ) -> bool:
        """Start an A/B test between champion and challenger."""
        # Validate both versions exist
        ids = {v.version_id for v in self.versions}
        if champion_id not in ids or challenger_id not in ids:
            return False

        self.ab_test = ABTestResult(
            champion_id=champion_id,
            challenger_id=challenger_id,
            target_rounds=target_rounds,
            is_active=True,
            started_at=datetime.now().isoformat(),
        )
        self._save_registry()
        return True

    def record_ab_round(self, champion_correct: bool, challenger_correct: bool):
        """Record one prediction round for the A/B test."""
        if not self.ab_test or not self.ab_test.is_active:
            return

        self.ab_test.total_rounds += 1
        if champion_correct:
            self.ab_test.champion_wins += 1
        if challenger_correct:
            self.ab_test.challenger_wins += 1

        self._save_registry()

    def evaluate_ab_test(self) -> Optional[str]:
        """
        Evaluate the A/B test. Returns the winner's version_id if complete,
        or None if still in progress.
        """
        if not self.ab_test or not self.ab_test.is_active:
            return None

        if not self.ab_test.is_complete:
            return None

        winner_id = (
            self.ab_test.challenger_id
            if self.ab_test.challenger_rate > self.ab_test.champion_rate + self.promotion_threshold
            else self.ab_test.champion_id
        )

        # Auto-promote winner
        self._promote(winner_id)
        self.ab_test.is_active = False
        self._save_registry()

        return winner_id

    def stop_ab_test(self):
        """Cancel the current A/B test."""
        if self.ab_test:
            self.ab_test.is_active = False
            self._save_registry()

    def get_ab_status(self) -> str:
        """Get human-readable A/B test status."""
        if not self.ab_test:
            return "No A/B test configured."

        t = self.ab_test
        lines = [
            "=" * 60,
            "A/B TEST STATUS",
            "=" * 60,
            f"Champion:   {t.champion_id} ({t.champion_rate:.1%} accuracy)",
            f"Challenger: {t.challenger_id} ({t.challenger_rate:.1%} accuracy)",
            f"Progress:   {t.total_rounds}/{t.target_rounds} rounds",
            f"Status:     {'ACTIVE' if t.is_active else 'COMPLETED'}",
            "=" * 60,
        ]
        return "\n".join(lines)


# =============================================================================
# SELF-TEST
# =============================================================================


def _self_test():
    import tempfile
    print("=" * 60)
    print("MODEL REGISTRY SELF-TEST")
    print("=" * 60)

    with tempfile.TemporaryDirectory() as tmpdir:
        registry = ModelRegistry(registry_dir=tmpdir)

        # Register models
        v1 = registry.register(
            accuracy=0.48,
            total_predictions=100,
            feature_hash="abc123",
            description="Baseline model",
        )
        print(f"Registered: {v1}")

        v2 = registry.register(
            accuracy=0.52,
            total_predictions=200,
            feature_hash="def456",
            description="Improved features",
        )
        print(f"Registered: {v2}")

        v3 = registry.register(
            accuracy=0.50,
            total_predictions=150,
            feature_hash="ghi789",
            description="Momentum-only model",
        )
        print(f"Registered: {v3}")

        # Check active version
        active = registry.get_active_version()
        assert active is not None, "Should have an active version"
        assert active.version_id == v2, f"v2 should be active (best accuracy), got {active.version_id}"
        print(f"\nActive model: {active.version_id} (accuracy: {active.accuracy:.2%})")

        # Compare versions
        print("\n" + registry.compare_versions())

        # Rollback
        assert registry.rollback(v1), "Rollback should succeed"
        active = registry.get_active_version()
        assert active.version_id == v1, "v1 should now be active after rollback"
        print(f"After rollback: {active.version_id}")

        # Restore v2
        registry.rollback(v2)

        # A/B Test
        assert registry.start_ab_test(v2, v3, target_rounds=10), "Should start A/B test"
        print(f"\n{registry.get_ab_status()}")

        # Simulate rounds
        np.random.seed(42)
        for _ in range(10):
            champ_correct = np.random.random() < 0.55  # champion slightly better
            chal_correct = np.random.random() < 0.50
            registry.record_ab_round(champ_correct, chal_correct)

        winner = registry.evaluate_ab_test()
        assert winner is not None, "Test should be complete"
        print(f"\nA/B Test Winner: {winner}")
        print(registry.get_ab_status())

        # Persistence round-trip
        registry2 = ModelRegistry(registry_dir=tmpdir)
        assert len(registry2.versions) == 3, "Should load 3 versions"
        print(f"\n✅ Persistence round-trip passed ({len(registry2.versions)} versions)")

    print("\n✅ All assertions passed!")
    print("=" * 60)


if __name__ == "__main__":
    _self_test()
