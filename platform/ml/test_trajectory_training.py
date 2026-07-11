"""Tests for Step 6 trajectory sequence training."""

import unittest

import pandas as pd

from trajectory_training import (
    MIN_GAIN_VS_CHECKPOINT,
    enrich_trajectory_features,
    train_trajectory_policy,
)


def _sample_trace(n: int = 8) -> pd.DataFrame:
    rows = []
    for i in range(n):
        rows.append({
            "run_id": "val-cvrp-a32k5-sa-s42",
            "algorithm": "sa",
            "candidates": (i + 1) * 10000,
            "iterations_total": 100000,
            "plateau_length": i * 1000,
            "current_penalty": 900 - i * 5,
            "best_penalty": 900 - i * 10,
            "initial_penalty": 1000,
            "improvement_rate": 0.1,
            "temperature": 1.0,
            "final_best_penalty": 800,
        })
    return pd.DataFrame(rows)


class TestTrajectoryFeatures(unittest.TestCase):
    def test_enrich_adds_sequence_columns(self):
        df = _sample_trace()
        enriched = enrich_trajectory_features(df)
        self.assertFalse(enriched.empty)
        for col in ("trace_progress", "recent_slope", "volatility", "improvements_so_far"):
            self.assertIn(col, enriched.columns)


class TestTrajectoryTraining(unittest.TestCase):
    def test_trains_when_enough_checkpoints(self):
        parts = []
        for run in range(40):
            trace = _sample_trace(12).copy()
            trace["run_id"] = f"val-cvrp-a32k5-sa-s{run}"
            trace.loc[trace.index[-4]:, "best_penalty"] = 820
            trace["final_best_penalty"] = 800
            parts.append(trace)
        df = pd.concat(parts, ignore_index=True)
        result = train_trajectory_policy(df, min_samples=50)
        self.assertIn(result["status"], ("trained", "insufficient_data"))
        if result["status"] == "trained":
            self.assertGreater(len(result["classifiers"]), 0)
            self.assertIn("gain_vs_checkpoint", result)


if __name__ == "__main__":
    unittest.main()
