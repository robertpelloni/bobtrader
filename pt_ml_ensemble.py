"""
PowerTrader AI — ML Ensemble Predictor
========================================
Wraps the existing memory-based kNN pattern matching into a clean ML
interface, adds model ensembling with adaptive weights, and provides
permutation-based feature importance analysis.

Usage:
    python3 pt_ml_ensemble.py          # Run self-test
"""

from __future__ import annotations
import json
import math
import sys
import os
import time
from dataclasses import dataclass, field
from pathlib import Path
from typing import List, Dict, Optional, Tuple
from datetime import datetime

import numpy as np

from pt_feature_engine import FeatureEngine, NUM_FEATURES


# =============================================================================
# DATA STRUCTURES
# =============================================================================


@dataclass
class ModelVersion:
    """Metadata for a trained model snapshot."""
    version_id: str
    created_at: str
    feature_set: List[str]
    accuracy: float = 0.0
    total_predictions: int = 0
    description: str = ""


@dataclass
class PredictionResult:
    """Result from ensemble prediction."""
    predicted_high_pct: float    # predicted % move up
    predicted_low_pct: float     # predicted % move down
    confidence: float            # 0-1 confidence score
    model_votes: Dict[str, float] = field(default_factory=dict)


# =============================================================================
# kNN MODEL (wraps existing pattern matching logic)
# =============================================================================


class KNNModel:
    """
    Clean interface around a k-Nearest-Neighbors predictor.
    Stores patterns (feature vectors) and their associated outcomes.
    """

    def __init__(self, name: str, k: int = 5, feature_indices: Optional[List[int]] = None):
        self.name = name
        self.k = k
        self.feature_indices = feature_indices  # subset of feature columns to use
        self.patterns: Optional[np.ndarray] = None       # (n_patterns, n_features)
        self.outcomes_high: Optional[np.ndarray] = None   # (n_patterns,) % move high
        self.outcomes_low: Optional[np.ndarray] = None    # (n_patterns,) % move low
        self.accuracy_history: List[float] = []           # rolling accuracy
        self._rolling_window = 30

    def fit(self, features: np.ndarray, outcomes_high: np.ndarray, outcomes_low: np.ndarray):
        """
        Train the model on feature matrix and outcome arrays.
        features: (n_samples, n_features)
        outcomes_high/low: (n_samples,) percentage moves
        """
        if self.feature_indices is not None:
            features = features[:, self.feature_indices]

        # Filter out rows with any NaN
        valid_mask = ~np.any(np.isnan(features), axis=1)
        valid_mask &= ~np.isnan(outcomes_high)
        valid_mask &= ~np.isnan(outcomes_low)

        self.patterns = features[valid_mask].copy()
        self.outcomes_high = outcomes_high[valid_mask].copy()
        self.outcomes_low = outcomes_low[valid_mask].copy()

    def predict(self, feature_vector: np.ndarray) -> Tuple[float, float, float]:
        """
        Predict high/low % moves for a single feature vector.
        Returns (predicted_high_pct, predicted_low_pct, confidence).
        """
        if self.patterns is None or len(self.patterns) == 0:
            return 0.0, 0.0, 0.0

        if self.feature_indices is not None:
            feature_vector = feature_vector[self.feature_indices]

        # Skip if any NaN in input
        if np.any(np.isnan(feature_vector)):
            return 0.0, 0.0, 0.0

        # Compute distances to all stored patterns
        diffs = self.patterns - feature_vector
        distances = np.sqrt(np.sum(diffs ** 2, axis=1))

        # Get k nearest
        k = min(self.k, len(distances))
        nearest_idx = np.argpartition(distances, k)[:k]
        nearest_dist = distances[nearest_idx]

        # Inverse-distance weighting
        weights = np.where(nearest_dist > 0, 1.0 / nearest_dist, 1e6)
        weight_sum = np.sum(weights)
        if weight_sum == 0:
            return 0.0, 0.0, 0.0

        weights /= weight_sum

        pred_high = float(np.sum(weights * self.outcomes_high[nearest_idx]))
        pred_low = float(np.sum(weights * self.outcomes_low[nearest_idx]))

        # Confidence: inverse of mean distance (normalized)
        mean_dist = float(np.mean(nearest_dist))
        confidence = 1.0 / (1.0 + mean_dist)

        return pred_high, pred_low, confidence

    def score(self, features: np.ndarray, actual_high: np.ndarray, actual_low: np.ndarray) -> float:
        """
        Evaluate accuracy: fraction of predictions where direction was correct.
        """
        if self.patterns is None:
            return 0.0

        correct = 0
        total = 0
        for i in range(len(features)):
            if np.any(np.isnan(features[i])):
                continue
            pred_h, pred_l, conf = self.predict(features[i])
            if np.isnan(actual_high[i]) or np.isnan(actual_low[i]):
                continue

            # Direction accuracy: did we predict the dominant move correctly?
            actual_dir = 1 if actual_high[i] > abs(actual_low[i]) else -1
            pred_dir = 1 if pred_h > abs(pred_l) else -1
            if actual_dir == pred_dir:
                correct += 1
            total += 1

        return correct / total if total > 0 else 0.0

    def record_accuracy(self, was_correct: bool):
        """Record a single prediction outcome for rolling accuracy tracking."""
        self.accuracy_history.append(1.0 if was_correct else 0.0)
        if len(self.accuracy_history) > self._rolling_window:
            self.accuracy_history.pop(0)

    @property
    def rolling_accuracy(self) -> float:
        if not self.accuracy_history:
            return 0.5  # neutral prior
        return sum(self.accuracy_history) / len(self.accuracy_history)

    def save(self, path: Path):
        """Save model to disk."""
        path.mkdir(parents=True, exist_ok=True)
        if self.patterns is not None:
            np.save(str(path / "patterns.npy"), self.patterns)
            np.save(str(path / "outcomes_high.npy"), self.outcomes_high)
            np.save(str(path / "outcomes_low.npy"), self.outcomes_low)
        meta = {
            "name": self.name,
            "k": self.k,
            "feature_indices": self.feature_indices,
            "accuracy_history": self.accuracy_history,
        }
        with open(path / "meta.json", "w") as f:
            json.dump(meta, f, indent=2)

    @classmethod
    def load(cls, path: Path) -> "KNNModel":
        """Load model from disk."""
        with open(path / "meta.json") as f:
            meta = json.load(f)
        model = cls(name=meta["name"], k=meta["k"], feature_indices=meta.get("feature_indices"))
        model.accuracy_history = meta.get("accuracy_history", [])

        patterns_path = path / "patterns.npy"
        if patterns_path.exists():
            model.patterns = np.load(str(patterns_path))
            model.outcomes_high = np.load(str(path / "outcomes_high.npy"))
            model.outcomes_low = np.load(str(path / "outcomes_low.npy"))
        return model


# =============================================================================
# ENSEMBLE PREDICTOR
# =============================================================================


class EnsemblePredictor:
    """
    Combines multiple KNNModel instances with adaptive weighting.
    Models with higher recent accuracy get more influence.
    """

    def __init__(self):
        self.models: List[KNNModel] = []

    def add_model(self, model: KNNModel):
        self.models.append(model)

    def predict(self, feature_vector: np.ndarray) -> PredictionResult:
        """Weighted-average prediction across all models."""
        if not self.models:
            return PredictionResult(0.0, 0.0, 0.0)

        total_weight = 0.0
        weighted_high = 0.0
        weighted_low = 0.0
        weighted_conf = 0.0
        model_votes = {}

        for model in self.models:
            pred_h, pred_l, conf = model.predict(feature_vector)
            weight = model.rolling_accuracy  # adaptive weight

            weighted_high += pred_h * weight
            weighted_low += pred_l * weight
            weighted_conf += conf * weight
            total_weight += weight

            model_votes[model.name] = pred_h - abs(pred_l)  # net direction

        if total_weight == 0:
            return PredictionResult(0.0, 0.0, 0.0)

        return PredictionResult(
            predicted_high_pct=weighted_high / total_weight,
            predicted_low_pct=weighted_low / total_weight,
            confidence=weighted_conf / total_weight,
            model_votes=model_votes,
        )

    def score_all(self, features: np.ndarray, actual_high: np.ndarray, actual_low: np.ndarray) -> Dict[str, float]:
        """Score each model individually."""
        return {m.name: m.score(features, actual_high, actual_low) for m in self.models}

    def save_all(self, base_dir: Path):
        """Save all models to disk."""
        base_dir.mkdir(parents=True, exist_ok=True)
        for model in self.models:
            model.save(base_dir / model.name)
        # Save ensemble metadata
        meta = {"model_names": [m.name for m in self.models]}
        with open(base_dir / "ensemble_meta.json", "w") as f:
            json.dump(meta, f, indent=2)

    @classmethod
    def load_all(cls, base_dir: Path) -> "EnsemblePredictor":
        """Load all models from disk."""
        ensemble = cls()
        meta_path = base_dir / "ensemble_meta.json"
        if not meta_path.exists():
            return ensemble
        with open(meta_path) as f:
            meta = json.load(f)
        for name in meta["model_names"]:
            model_path = base_dir / name
            if model_path.exists():
                ensemble.add_model(KNNModel.load(model_path))
        return ensemble


# =============================================================================
# FEATURE IMPORTANCE
# =============================================================================


def compute_feature_importance(
    model: KNNModel,
    features: np.ndarray,
    actual_high: np.ndarray,
    actual_low: np.ndarray,
    n_shuffles: int = 5,
) -> Dict[str, float]:
    """
    Permutation-based feature importance.
    Shuffles each feature column, measures accuracy drop.
    """
    feature_names = FeatureEngine.get_feature_names()
    baseline_score = model.score(features, actual_high, actual_low)
    importance = {}

    n_features = features.shape[1]
    indices_to_check = model.feature_indices if model.feature_indices else list(range(n_features))

    for feat_idx in indices_to_check:
        drops = []
        for _ in range(n_shuffles):
            shuffled = features.copy()
            np.random.shuffle(shuffled[:, feat_idx])
            shuffled_score = model.score(shuffled, actual_high, actual_low)
            drops.append(baseline_score - shuffled_score)

        name = feature_names[feat_idx] if feat_idx < len(feature_names) else f"feature_{feat_idx}"
        importance[name] = float(np.mean(drops))

    # Sort by importance descending
    importance = dict(sorted(importance.items(), key=lambda x: x[1], reverse=True))
    return importance


def print_feature_report(importance: Dict[str, float]):
    """Print a ranked feature importance table."""
    print("\n" + "=" * 50)
    print("FEATURE IMPORTANCE REPORT")
    print("=" * 50)
    print(f"{'Rank':>4}  {'Feature':<25} {'Importance':>10}")
    print("-" * 45)
    for rank, (name, score) in enumerate(importance.items(), 1):
        bar = "█" * max(0, int(score * 200))
        print(f"{rank:>4}  {name:<25} {score:>10.4f}  {bar}")
    print("=" * 50)


# =============================================================================
# SELF-TEST
# =============================================================================


def _self_test():
    print("=" * 60)
    print("ML ENSEMBLE SELF-TEST")
    print("=" * 60)

    np.random.seed(42)
    n = 200

    # Generate synthetic OHLCV data
    base = 50000.0
    closes = [base]
    for _ in range(n - 1):
        closes.append(closes[-1] * (1 + np.random.normal(0, 0.01)))
    opens = [c * (1 + np.random.normal(0, 0.002)) for c in closes]
    highs = [max(o, c) * (1 + abs(np.random.normal(0, 0.005))) for o, c in zip(opens, closes)]
    lows = [min(o, c) * (1 - abs(np.random.normal(0, 0.005))) for o, c in zip(opens, closes)]
    volumes = [abs(np.random.normal(1000, 200)) for _ in range(n)]

    # Compute features
    engine = FeatureEngine()
    features = engine.compute_features(opens, highs, lows, closes, volumes)
    print(f"Features: {features.shape}")

    # Generate synthetic outcomes (next-candle high/low % moves)
    outcomes_high = np.array([(highs[i] - closes[i]) / closes[i] * 100 if i < n else 0.0 for i in range(n)])
    outcomes_low = np.array([(lows[i] - closes[i]) / closes[i] * 100 if i < n else 0.0 for i in range(n)])

    # Split into train/test
    split = 150
    train_feat = features[:split]
    test_feat = features[split:]
    train_high = outcomes_high[:split]
    test_high = outcomes_high[split:]
    train_low = outcomes_low[:split]
    test_low = outcomes_low[split:]

    # Create models with different feature subsets
    model_all = KNNModel("all_features", k=5)
    model_all.fit(train_feat, train_high, train_low)

    model_price = KNNModel("price_only", k=5, feature_indices=[0, 1, 2])
    model_price.fit(train_feat, train_high, train_low)

    model_momentum = KNNModel("momentum_vol", k=5, feature_indices=[5, 6, 7, 8, 9])
    model_momentum.fit(train_feat, train_high, train_low)

    # Score individually
    print(f"\nModel Accuracy on Test Set:")
    print(f"  all_features: {model_all.score(test_feat, test_high, test_low):.2%}")
    print(f"  price_only:   {model_price.score(test_feat, test_high, test_low):.2%}")
    print(f"  momentum_vol: {model_momentum.score(test_feat, test_high, test_low):.2%}")

    # Build ensemble
    ensemble = EnsemblePredictor()
    ensemble.add_model(model_all)
    ensemble.add_model(model_price)
    ensemble.add_model(model_momentum)

    # Predict on last test sample
    last_valid = None
    for i in range(len(test_feat) - 1, -1, -1):
        if not np.any(np.isnan(test_feat[i])):
            last_valid = test_feat[i]
            break

    if last_valid is not None:
        result = ensemble.predict(last_valid)
        print(f"\nEnsemble Prediction (last test sample):")
        print(f"  Predicted High: {result.predicted_high_pct:+.4f}%")
        print(f"  Predicted Low:  {result.predicted_low_pct:+.4f}%")
        print(f"  Confidence:     {result.confidence:.4f}")
        print(f"  Model Votes:    {result.model_votes}")

    # Feature importance
    importance = compute_feature_importance(model_all, test_feat, test_high, test_low, n_shuffles=3)
    print_feature_report(importance)

    # Save/Load round-trip
    save_dir = Path("/tmp/pt_ensemble_test")
    ensemble.save_all(save_dir)
    loaded = EnsemblePredictor.load_all(save_dir)
    assert len(loaded.models) == 3, f"Expected 3 models, got {len(loaded.models)}"
    print(f"\n✅ Save/Load round-trip passed ({len(loaded.models)} models)")

    # Cleanup
    import shutil
    shutil.rmtree(save_dir, ignore_errors=True)

    print("\n✅ All assertions passed!")
    print("=" * 60)


if __name__ == "__main__":
    _self_test()
