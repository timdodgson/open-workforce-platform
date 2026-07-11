"""Tests for Step 5 contextual bandit training."""

import unittest

import pandas as pd

from bandit_training import (
    MAX_EPISODE_REGRET,
    evaluate_bandit_promotion,
    train_portfolio_bandit,
    train_worker_bandit,
)


class TestPortfolioBandit(unittest.TestCase):
    def test_trains_entries_from_telemetry(self):
        df = pd.DataFrame([
            {
                "run_id": "val-cvrp-a32k5-portfolio-s42",
                "domain": "cvrp",
                "instance": "A-n32-k5",
                "strategy": "sa",
                "original_budget": 100000,
                "final_budget": 125000,
                "strategy_won": 1,
            },
            {
                "run_id": "val-cvrp-a32k5-portfolio-s42",
                "domain": "cvrp",
                "instance": "A-n32-k5",
                "strategy": "lahc",
                "original_budget": 100000,
                "final_budget": 100000,
                "strategy_won": 0,
            },
            {
                "run_id": "val-cvrp-a32k5-portfolio-s43",
                "domain": "cvrp",
                "instance": "A-n32-k5",
                "strategy": "sa",
                "original_budget": 100000,
                "final_budget": 150000,
                "strategy_won": 1,
            },
        ] * 5)
        result = train_portfolio_bandit(df, min_samples=10)
        self.assertEqual(result["status"], "trained")
        self.assertGreater(len(result["entries"]), 0)
        self.assertIn("episode_regret", result)


class TestWorkerBandit(unittest.TestCase):
    def test_trains_context_entries(self):
        df = pd.DataFrame([
            {
                "week": 1, "depth": 2, "distance_from_best": 10,
                "suggested_budget": 1000, "final_budget": 1000, "improved": 1,
            },
            {
                "week": 1, "depth": 2, "distance_from_best": 80,
                "suggested_budget": 1000, "final_budget": 0, "improved": 0,
            },
            {
                "week": 2, "depth": 3, "distance_from_best": 5,
                "suggested_budget": 1000, "final_budget": 1250, "improved": 1,
            },
        ] * 8)
        result = train_worker_bandit(df, min_samples=10)
        self.assertEqual(result["status"], "trained")
        self.assertGreaterEqual(len(result["entries"]), 1)


class TestBanditPromotion(unittest.TestCase):
    def test_promotion_when_regret_low(self):
        promo = evaluate_bandit_promotion(
            {"promotion_ready": True, "episode_regret": 0.02},
            {"promotion_ready": True, "episode_regret": 0.03},
        )
        self.assertTrue(promo["promotion_ready"])
        self.assertEqual(promo["promotion_gate"]["max_episode_regret"], MAX_EPISODE_REGRET)


if __name__ == "__main__":
    unittest.main()
