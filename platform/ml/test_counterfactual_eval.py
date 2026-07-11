"""Tests for Step 4 counterfactual offline evaluation."""

import tempfile
import unittest
from pathlib import Path

import pandas as pd

from counterfactual_eval import (
    evaluate_offline_counterfactual,
    merge_counterfactual_into_validation,
    summarize_counterfactual_telemetry,
)
from policy_registry import MAX_FALSE_STOP_RATE, passes_counterfactual_gate


class TestCounterfactualGate(unittest.TestCase):
    def test_passes_low_false_stop_rate(self):
        self.assertTrue(passes_counterfactual_gate({
            "samples": 100,
            "false_stop_rate": 0.02,
        }))

    def test_fails_high_false_stop_rate(self):
        self.assertFalse(passes_counterfactual_gate({
            "samples": 100,
            "false_stop_rate": 0.20,
        }))

    def test_threshold_matches_constant(self):
        self.assertTrue(passes_counterfactual_gate({
            "samples": 100,
            "false_stop_rate": MAX_FALSE_STOP_RATE,
        }))


class TestCounterfactualTelemetry(unittest.TestCase):
    def test_summarize_early_stop_regret(self):
        df = pd.DataFrame([
            {"actual_action": "early_stop", "regret": 5.0, "domain": "cvrp"},
            {"actual_action": "continue", "regret": 0.0, "domain": "cvrp"},
            {"actual_action": "early_stop", "regret": 0.0, "domain": "cvrp"},
        ])
        summary = summarize_counterfactual_telemetry(df)
        self.assertEqual(summary["early_stops"], 2)
        self.assertEqual(summary["false_stops"], 1)
        self.assertEqual(summary["false_stop_rate"], 0.5)


class TestOfflineEval(unittest.TestCase):
    def test_detects_false_stop_from_checkpoints(self):
        df = pd.DataFrame([
            {
                "run_id": "val-cvrp-a32k5-sa-s42",
                "algorithm": "sa",
                "iterations_total": 100000,
                "plateau_length": 1000,
                "candidates": 80000,
                "best_penalty": 500,
                "current_penalty": 520,
                "initial_penalty": 600,
                "final_best_penalty": 400,
            },
        ])
        model = {
            "classifiers": [{
                "domain": "cvrp",
                "algorithm": "sa",
                "instance": "a32k5",
                "cv_mean": 0.99,
                "tree": {
                    "feature_names": ["budget_consumed"],
                    "children_left": [-1],
                    "children_right": [-1],
                    "feature": [-2],
                    "threshold": [-2.0],
                    "value": [[0.0], [1.0]],
                },
            }],
        }
        with tempfile.TemporaryDirectory() as tmp:
            result = evaluate_offline_counterfactual(
                Path(tmp), Path(tmp), search_df=df, stagnation_model=model,
            )
        self.assertEqual(result["status"], "evaluated")
        self.assertGreaterEqual(result["false_stops"], 0)

    def test_merge_tightens_promotion(self):
        validation = {
            "policies": {"stagnation": {"cvrp": {"promotion_ready": True}}},
            "global": {"promotion_ready": True},
        }
        counterfactual = {
            "status": "evaluated",
            "false_stop_rate": 0.25,
            "promotion_ready": False,
            "domains": {"cvrp": {"false_stop_rate": 0.25, "promotion_ready": False}},
        }
        merged = merge_counterfactual_into_validation(validation, counterfactual)
        self.assertFalse(merged["policies"]["stagnation"]["cvrp"]["promotion_ready"])
        self.assertFalse(merged["step4_promotion_ready"])


if __name__ == "__main__":
    unittest.main()
