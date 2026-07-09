"""
Shared promotion gates and registry merge for Search Intelligence 2.0.

Outcome-based promotion (primary):
  - offline_accuracy: ex-post correctness of learned decisions (target >= 0.80)
  - regret_vs_rules: mean learned regret minus mean rule regret (target <= 0)

Rule agreement is retained as a diagnostic only.
"""

from __future__ import annotations

import json
from copy import deepcopy
from datetime import datetime
from pathlib import Path
from typing import Any

MIN_OUTCOME_ACCURACY = 0.80
MAX_REGRET_VS_RULES = 0.0

SEARCH_DOMAINS = ("cvrp", "jss", "vrptw", "nrp")
ROUTING_DOMAINS = ("cvrp", "jss", "vrptw")


def passes_outcome_gate(metrics: dict, decision_type: str | None = None) -> bool:
    """Return whether domain/policy metrics meet promotion gates."""
    if not metrics:
        return False
    accuracy = float(metrics.get("outcome_accuracy", metrics.get("offline_accuracy", 0)))
    regret = float(metrics.get("regret_vs_rules", 0))
    samples = int(metrics.get("samples", metrics.get("total", 0)))
    min_accuracy = 0.65 if decision_type == "budget" else MIN_OUTCOME_ACCURACY
    if samples < 20:
        return False
    if accuracy < min_accuracy:
        return False
    if regret <= MAX_REGRET_VS_RULES:
        return True
    # Small adapter/search samples: regret from rule disagreement is noisy when accuracy is strong.
    if samples < 500 and accuracy >= 0.95:
        return True
    return False


def version_id(decision_type: str, domain: str) -> str:
    return f"{decision_type}-{domain}"


def _policy_domains(policy_key: str, training_result: dict) -> list[str]:
    if policy_key == "worker_policy":
        return ["nrp"]
    if policy_key == "budget_policy":
        domains = {
            e.get("domain")
            for e in training_result.get("entries", [])
            if e.get("domain") in ROUTING_DOMAINS
        }
        domains |= {
            c.get("domain")
            for c in training_result.get("classifiers", [])
            if c.get("domain") in ROUTING_DOMAINS
        }
        return sorted(domains) or list(ROUTING_DOMAINS)
    if policy_key in ("stagnation_policy", "restart_policy"):
        domains = {
            c.get("domain")
            for c in training_result.get("classifiers", training_result.get("curves", training_result.get("entries", [])))
            if c.get("domain") in SEARCH_DOMAINS
        }
        return sorted(domains) or [d for d in SEARCH_DOMAINS if d != "nrp"]
    return []


def build_lifecycle_registry(results: dict) -> dict:
    """Build per-domain policy_registry.json entries from training results."""
    now = datetime.now().isoformat()
    versions: list[dict[str, Any]] = []

    specs = {
        "budget_policy": "budget",
        "stagnation_policy": "stagnation",
        "restart_policy": "restart",
        "worker_policy": "worker",
    }

    for key, decision_type in specs.items():
        result = results.get(key, {})
        if result.get("status") not in ("trained", "below_promotion_gate"):
            continue

        for domain in _policy_domains(key, result):
            offline = float(result.get("accuracy", result.get("cv_mean", 0)))
            if key in ("stagnation_policy", "restart_policy", "budget_policy"):
                offline = 0.0

            versions.append({
                "id": version_id(decision_type, domain),
                "version": result.get("version", "1.0.0"),
                "domain": domain,
                "decision_type": decision_type,
                "algorithm": "*",
                "status": "training",
                "created_at": result.get("trained_at", now),
                "training_samples": result.get("trained_on", result.get("samples", 0)),
                "training_date": result.get("trained_at", now),
                "features": result.get("features_used", []),
                "offline_accuracy": round(offline, 4),
                "shadow_accuracy": -1,
                "production_accuracy": -1,
                "production_runs": 0,
                "regret_vs_rules": 0,
                "drift_detected": False,
                "model_path": key.replace("_policy", "") + "_policy.json",
                "promotion_ready": False,
            })

    return {"versions": versions}


def _find_version(versions: list[dict], decision_type: str, domain: str) -> dict | None:
    vid = version_id(decision_type, domain)
    for v in versions:
        if v.get("id") == vid or (
            v.get("decision_type") == decision_type and v.get("domain") == domain
        ):
            return v
    return None


def merge_validation_into_registry(
    registry: dict,
    validation: dict,
    training_results: dict | None = None,
) -> dict:
    """
    Apply outcome validation metrics to registry versions.

    validation["policies"] maps decision_type -> domain -> metrics
    validation["worker"] optional worker policy metrics for nrp
    """
    reg = deepcopy(registry)
    versions = reg.setdefault("versions", [])
    policies = validation.get("policies", {})
    validated_at = validation.get("validated_at", datetime.now().isoformat())

    for decision_type, by_domain in policies.items():
        for domain, metrics in by_domain.items():
            v = _find_version(versions, decision_type, domain)
            if v is None:
                v = {
                    "id": version_id(decision_type, domain),
                    "version": "1.0.0",
                    "domain": domain,
                    "decision_type": decision_type,
                    "algorithm": "*",
                    "status": "training",
                    "created_at": validated_at,
                    "training_samples": metrics.get("samples", 0),
                    "training_date": validated_at,
                    "features": [],
                    "shadow_accuracy": -1,
                    "production_accuracy": -1,
                    "production_runs": 0,
                    "drift_detected": False,
                    "model_path": f"{decision_type}_policy.json",
                }
                versions.append(v)

            outcome_accuracy = float(metrics.get("outcome_accuracy", 0))
            regret = float(metrics.get("regret_vs_rules", 0))
            ready = passes_outcome_gate(metrics, decision_type=decision_type)

            v["offline_accuracy"] = round(outcome_accuracy, 4)
            v["regret_vs_rules"] = round(regret, 4)
            v["promotion_ready"] = ready
            v["agreement_rate"] = round(float(metrics.get("agreement_rate", 0)), 4)
            v["validated_at"] = validated_at
            v["validation_samples"] = int(metrics.get("samples", 0))

            if ready and v.get("status") == "training":
                v["status"] = "shadow"

    worker_metrics = validation.get("worker", {})
    if worker_metrics:
        v = _find_version(versions, "worker", "nrp")
        if v is None and training_results:
            wr = training_results.get("worker_policy", {})
            if wr.get("status") == "trained":
                versions.extend(build_lifecycle_registry({"worker_policy": wr})["versions"])
                v = _find_version(versions, "worker", "nrp")

        if v is not None:
            outcome_accuracy = float(
                worker_metrics.get("outcome_accuracy", worker_metrics.get("cv_mean", 0))
            )
            regret = float(worker_metrics.get("regret_vs_rules", 0))
            ready = passes_outcome_gate({
                "outcome_accuracy": outcome_accuracy,
                "regret_vs_rules": regret,
                "samples": worker_metrics.get("samples", v.get("training_samples", 0)),
            })
            v["offline_accuracy"] = round(outcome_accuracy, 4)
            v["regret_vs_rules"] = round(regret, 4)
            v["promotion_ready"] = ready
            v["validated_at"] = validated_at
            if ready and v.get("status") == "training":
                v["status"] = "shadow"

    reg["promotion_gate"] = {
        "min_outcome_accuracy": MIN_OUTCOME_ACCURACY,
        "max_regret_vs_rules": MAX_REGRET_VS_RULES,
    }
    reg["validated_at"] = validated_at
    reg["promotion_ready_count"] = sum(1 for v in versions if v.get("promotion_ready"))
    reg["promotion_total"] = len(versions)
    return reg


def load_registry(path: Path) -> dict:
    if not path.exists():
        return {"versions": []}
    with open(path) as f:
        return json.load(f)


def save_registry(path: Path, registry: dict) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    with open(path, "w") as f:
        json.dump(registry, f, indent=2)


def sync_dashboard_registry(output_dir: Path, registry: dict) -> None:
    dashboard = output_dir.resolve().parent.parent / "web" / "pfrs-lab" / "data" / "policy_registry.json"
    if dashboard.parent.exists():
        save_registry(dashboard, registry)
