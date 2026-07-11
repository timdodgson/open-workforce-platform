"""Tests for Step 7 neural plateau training."""

import unittest

import pandas as pd

from neural_training import PLATEAU_GLOBAL_GAIN, train_neural_where_plateau
from trajectory_training import enrich_trajectory_features


def _build_runs(n_runs: int = 50, trace_len: int = 15) -> pd.DataFrame:
    from trajectory_training import enrich_trajectory_features as _  # noqa: F401
    parts = []
    for run in range(n_runs):
        trace = pd.DataFrame([
            {
                "run_id": f"val-cvrp-a32k5-sa-s{run}",
                "algorithm": "sa",
                "candidates": (i + 1) * 8000,
                "iterations_total": 120000,
                "plateau_length": i * 900,
                "current_penalty": 950 - i * 3 + (i % 4),
                "best_penalty": 950 - i * 8,
                "initial_penalty": 1000,
                "improvement_rate": 0.05 + (i % 3) * 0.02,
                "temperature": 1.0,
                "final_best_penalty": 820 if run % 2 == 0 else 800,
            }
            for i in range(trace_len)
        ])
        parts.append(trace)
    return pd.concat(parts, ignore_index=True)


class TestNeuralPlateau(unittest.TestCase):
    def test_skips_when_trajectory_gain_strong(self):
        df = _build_runs(10)
        trajectory = {
            "status": "trained",
            "gain_vs_checkpoint": 0.05,
            "classifiers": [{"domain": "cvrp", "algorithm": "sa", "instance": "a32k5", "cv_mean": 0.9}],
        }
        result = train_neural_where_plateau(df, trajectory)
        self.assertEqual(result["status"], "skipped")
        self.assertEqual(result["reason"], "trajectory_not_plateau")

    def test_runs_on_plateau_trajectory(self):
        df = _build_runs(60)
        enriched = enrich_trajectory_features(df)
        self.assertFalse(enriched.empty)
        trajectory = {
            "status": "trained",
            "gain_vs_checkpoint": 0.002,
            "classifiers": [
                {
                    "domain": "cvrp",
                    "algorithm": "sa",
                    "instance": "a32k5",
                    "cv_mean": 0.55,
                }
            ],
        }
        self.assertLess(trajectory["gain_vs_checkpoint"], PLATEAU_GLOBAL_GAIN)
        result = train_neural_where_plateau(df, trajectory, min_samples=50)
        self.assertIn(result["status"], ("trained", "no_winner", "insufficient_data"))


if __name__ == "__main__":
    unittest.main()
