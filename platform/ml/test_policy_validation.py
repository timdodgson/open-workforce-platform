"""Tests for outcome-based policy validation and registry merge."""

import json
import tempfile
import unittest
from pathlib import Path

import pandas as pd

from policy_registry import (
    build_lifecycle_registry,
    merge_validation_into_registry,
    passes_outcome_gate,
)
from policy_validation import (
    ex_post_should_stop,
    validate_stagnation_outcomes,
    validate_worker_outcomes,
)


class TestOutcomeGate(unittest.TestCase):
    def test_passes_when_accuracy_and_regret_ok(self):
        self.assertTrue(passes_outcome_gate({
            "outcome_accuracy": 0.85,
            "regret_vs_rules": -0.1,
            "samples": 100,
        }))

    def test_fails_low_accuracy(self):
        self.assertFalse(passes_outcome_gate({
            "outcome_accuracy": 0.50,
            "regret_vs_rules": -0.1,
            "samples": 100,
        }))

    def test_fails_positive_regret(self):
        self.assertFalse(passes_outcome_gate({
            "outcome_accuracy": 0.90,
            "regret_vs_rules": 0.5,
            "samples": 100,
        }))


class TestStagnationValidation(unittest.TestCase):
    def test_learned_continue_when_improvement_remains(self):
        row = pd.Series({
            "run_id": "cvrp-test-1",
            "algorithm": "sa",
            "iterations_total": 100000,
            "plateau_length": 1000,
            "candidates": 30000,
            "best_penalty": 500,
            "final_best_penalty": 400,
        })
        self.assertFalse(ex_post_should_stop(row))

    def test_validate_produces_per_domain_metrics(self):
        df = pd.DataFrame([
            {
                "run_id": "cvrp-a",
                "algorithm": "sa",
                "iterations_total": 100000,
                "plateau_length": 60000,
                "candidates": 80000,
                "best_penalty": 500,
                "final_best_penalty": 500,
            },
            {
                "run_id": "vrptw-b",
                "algorithm": "lahc",
                "iterations_total": 100000,
                "plateau_length": 60000,
                "candidates": 80000,
                "best_penalty": 300,
                "final_best_penalty": 300,
            },
        ])
        model = {
            "curves": [
                {"domain": "cvrp", "algorithm": "sa", "decay_rate": 3.0, "amplitude": 0.9, "confidence": 0.85, "sample_count": 100},
                {"domain": "vrptw", "algorithm": "lahc", "decay_rate": 3.0, "amplitude": 0.9, "confidence": 0.85, "sample_count": 100},
            ]
        }
        result = validate_stagnation_outcomes(df, model)
        self.assertEqual(result["status"], "validated")
        self.assertIn("cvrp", result["domain_stats"])
        self.assertIn("vrptw", result["domain_stats"])


class TestRegistryMerge(unittest.TestCase):
    def test_merge_sets_offline_accuracy_and_promotion_ready(self):
        training = {
            "stagnation_policy": {
                "status": "trained",
                "version": "1.0.0",
                "trained_at": "2026-01-01T00:00:00",
                "trained_on": 100,
                "curves": [{"domain": "cvrp", "algorithm": "sa", "decay_rate": 3, "amplitude": 0.9, "confidence": 0.8, "sample_count": 50}],
            }
        }
        registry = build_lifecycle_registry(training)
        validation = {
            "validated_at": "2026-01-02T00:00:00",
            "policies": {
                "stagnation": {
                    "cvrp": {
                        "samples": 500,
                        "outcome_accuracy": 0.88,
                        "regret_vs_rules": -0.05,
                        "agreement_rate": 0.40,
                        "promotion_ready": True,
                    }
                }
            },
        }
        merged = merge_validation_into_registry(registry, validation, training)
        v = merged["versions"][0]
        self.assertEqual(v["offline_accuracy"], 0.88)
        self.assertEqual(v["regret_vs_rules"], -0.05)
        self.assertTrue(v["promotion_ready"])
        self.assertEqual(v["status"], "shadow")


class TestWorkerValidation(unittest.TestCase):
    def test_worker_uses_training_cv(self):
        training = {"status": "trained", "cv_mean": 0.82, "samples": 200}
        metrics = validate_worker_outcomes(pd.DataFrame(), training)
        self.assertTrue(metrics["promotion_ready"])


if __name__ == "__main__":
    unittest.main()
