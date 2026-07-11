"""Tests for Step 8 research loop."""

import json
import tempfile
import unittest
from pathlib import Path

from research_loop import (
    build_experiment_command,
    build_research_queue,
    dedupe_proposals,
    evaluate_step8_gate,
    merge_research_into_validation,
    proposals_from_harness,
    proposals_from_registry,
)


class TestExperimentCommands(unittest.TestCase):
    def test_nrp_command(self):
        cmd = build_experiment_command("nrp", "sa", "hybrid", 99, "ml-exp-test")
        self.assertIn("tune-pfrs", cmd)
        self.assertIn("n012w8", cmd)
        self.assertIn("ml-exp-test", cmd)

    def test_cvrp_command(self):
        cmd = build_experiment_command("cvrp", "sa", "learned", 42, "ml-exp-cvrp")
        self.assertIn("solve cvrp", cmd)
        self.assertIn("A-n32-k5.vrp", cmd)


class TestProposalBuilders(unittest.TestCase):
    def test_registry_gap_proposal(self):
        registry = {
            "versions": [{
                "id": "stagnation-nrp",
                "domain": "nrp",
                "decision_type": "stagnation",
                "algorithm": "*",
                "promotion_ready": False,
                "offline_accuracy": 0.9,
                "regret_vs_rules": -10,
            }],
        }
        props = proposals_from_registry(registry)
        self.assertEqual(len(props), 1)
        self.assertEqual(props[0]["type"], "fill_promotion_gap")
        self.assertTrue(props[0]["requires_approval"])

    def test_active_regression_watch(self):
        from research_loop import proposals_from_active_regression

        registry = {"active_count": 12}
        harness = {
            "comparisons": [{
                "domain": "nrp",
                "instance": "n012w8",
                "algorithm": "sa",
                "modeB": "hybrid",
                "verdict": "worse",
                "objectiveDelta": 100,
                "roi": -5,
            }],
        }
        props = proposals_from_active_regression(registry, harness)
        self.assertEqual(len(props), 1)
        self.assertEqual(props[0]["signal"], "registry")

    def test_harness_regression_proposal(self):
        harness = {
            "comparisons": [{
                "domain": "nrp",
                "instance": "n012w8",
                "algorithm": "sa",
                "modeB": "hybrid",
                "verdict": "worse",
                "objectiveDelta": 100,
                "roi": -5,
            }],
        }
        props = proposals_from_harness(harness)
        self.assertEqual(len(props), 1)
        self.assertEqual(props[0]["policy_mode"], "hybrid")


class TestStep8Gate(unittest.TestCase):
    def test_loop_ok_with_multiple_signals(self):
        queue = {
            "human_approval_required": True,
            "proposals": [
                {
                    "id": "a",
                    "type": "harness_regression",
                    "requires_approval": True,
                    "command": "go run ...",
                    "signal": "harness",
                },
                {
                    "id": "b",
                    "type": "fill_promotion_gap",
                    "requires_approval": True,
                    "command": "go run ...",
                    "signal": "registry",
                },
            ],
            "summary": {"signals": ["harness", "registry"]},
        }
        gate = evaluate_step8_gate(queue)
        self.assertTrue(gate["loop_ok"])
        self.assertTrue(gate["promotion_ready"])

    def test_promote_fails_single_signal(self):
        queue = {
            "human_approval_required": True,
            "proposals": [{
                "id": "a",
                "requires_approval": True,
                "command": "go run ...",
            }],
            "summary": {"signals": ["harness"]},
        }
        gate = evaluate_step8_gate(queue)
        self.assertTrue(gate["loop_ok"])
        self.assertFalse(gate["promotion_ready"])


class TestBuildQueue(unittest.TestCase):
    def test_build_from_fixture_files(self):
        with tempfile.TemporaryDirectory() as tmp:
            policy_dir = Path(tmp)
            repo = policy_dir.parent
            (repo / "docs" / "reports" / "ml-harness").mkdir(parents=True, exist_ok=True)
            with open(policy_dir / "policy_registry.json", "w") as f:
                json.dump({
                    "promotion_ready_count": 10,
                    "promotion_total": 12,
                    "policies": [{
                        "id": "restart-jss",
                        "domain": "jss",
                        "decision_type": "restart",
                        "algorithm": "*",
                        "promotion_ready": False,
                        "offline_accuracy": 0.7,
                        "regret_vs_rules": 0,
                    }],
                }, f)
            with open(repo / "docs" / "reports" / "ml-harness" / "latest.json", "w") as f:
                json.dump({
                    "comparisons": [{
                        "domain": "cvrp",
                        "instance": "a32k5",
                        "algorithm": "sa",
                        "modeB": "learned",
                        "verdict": "worse",
                        "objectiveDelta": 1,
                        "roi": -1,
                    }],
                }, f)
            with open(policy_dir / "validation_results.json", "w") as f:
                json.dump({"global": {"total_checkpoints": 5000}}, f)

            queue = build_research_queue(policy_dir, repo)
            self.assertGreaterEqual(len(queue["proposals"]), 2)
            self.assertTrue(queue["human_approval_required"])
            merged = merge_research_into_validation({}, queue)
            self.assertIn("research", merged)
            self.assertTrue(merged["step8_loop_ok"])


class TestDedupe(unittest.TestCase):
    def test_keeps_highest_priority(self):
        props = dedupe_proposals([
            {"id": "same", "priority": 50},
            {"id": "same", "priority": 90},
            {"id": "other", "priority": 80},
        ])
        self.assertEqual(len(props), 2)
        self.assertEqual(props[0]["priority"], 90)


if __name__ == "__main__":
    unittest.main()
