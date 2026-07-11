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
MAX_FALSE_STOP_RATE = 0.05
MIN_SHADOW_RUNS = 20
MIN_SHADOW_ACCURACY = 0.60

SEARCH_DOMAINS = ("cvrp", "jss", "vrptw", "nrp")
ROUTING_DOMAINS = ("cvrp", "jss", "vrptw")


def passes_counterfactual_gate(metrics: dict) -> bool:
    """Step 4: block promotion when learned policies stop too early."""
    if not metrics:
        return False
    samples = int(metrics.get("samples", metrics.get("learned_stops", 0)))
    if samples < 20:
        return False
    false_rate = float(metrics.get("false_stop_rate", 1.0))
    return false_rate <= MAX_FALSE_STOP_RATE


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
        outcome_ok = True
    elif samples < 500 and accuracy >= 0.95:
        outcome_ok = True
    else:
        outcome_ok = False
    if not outcome_ok:
        return False
    if "false_stop_rate" in metrics and not passes_counterfactual_gate(metrics):
        return False
    return True


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

        if v is None:
            versions.append({
                "id": version_id("worker", "nrp"),
                "version": "1.0.0",
                "domain": "nrp",
                "decision_type": "worker",
                "algorithm": "*",
                "status": "training",
                "created_at": validated_at,
                "training_samples": int(worker_metrics.get("samples", 0)),
                "training_date": validated_at,
                "features": [],
                "shadow_accuracy": -1,
                "production_accuracy": -1,
                "production_runs": 0,
                "drift_detected": False,
                "model_path": "worker_policy.json",
            })
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


def _policy_id_aliases(policy_key: str) -> set[str]:
    """Match registry ids (hyphen) and evaluation CSV ids (underscore)."""
    return {policy_key, policy_key.replace("-", "_")}


def load_shadow_run_counts(data_dir: Path, prefix: str = "val-") -> dict[str, int]:
    """Count val-* runs that executed with hybrid or learned policy modes."""
    from collections import defaultdict

    counts: dict[str, int] = defaultdict(int)
    for run_dir in data_dir.iterdir():
        if not run_dir.is_dir():
            continue
        run_id = run_dir.name
        if prefix and not run_id.startswith(prefix):
            continue
        run_json = run_dir / "run.json"
        if not run_json.exists():
            continue
        try:
            with open(run_json) as f:
                data = json.load(f)
        except Exception:
            continue
        mode = str(data.get("policyMode", data.get("policy_mode", ""))).lower()
        if mode not in ("hybrid", "learned"):
            rid = run_id.lower()
            if "hybrid" in rid:
                mode = "hybrid"
            elif "learned" in rid:
                mode = "learned"
        if mode not in ("hybrid", "learned"):
            continue
        domain = str(data.get("domain", data.get("problemType", ""))).lower()
        if not domain:
            rid = run_id.lower()
            if "cvrp" in rid:
                domain = "cvrp"
            elif "jss" in rid or "jobshop" in rid:
                domain = "jss"
            elif "vrptw" in rid:
                domain = "vrptw"
            elif "nrp" in rid:
                domain = "nrp"
        if domain:
            counts[domain] += 1
    return dict(counts)


def load_shadow_telemetry(data_dir: Path, prefix: str = "val-") -> dict[str, dict]:
    """Aggregate policy_evaluation.csv from hybrid/learned val-* runs."""
    from collections import defaultdict

    import pandas as pd

    stats: dict[str, dict[str, Any]] = defaultdict(
        lambda: {"runs": set(), "correct": 0, "total": 0, "regret_sum": 0.0}
    )

    for run_dir in data_dir.iterdir():
        if not run_dir.is_dir():
            continue
        run_id = run_dir.name
        if prefix and not run_id.startswith(prefix):
            continue
        eval_path = run_dir / "policy_evaluation.csv"
        if not eval_path.exists():
            continue
        try:
            df = pd.read_csv(eval_path)
        except Exception:
            continue
        if df.empty:
            continue
        if "policy_type" in df.columns:
            df = df[df["policy_type"].astype(str).isin(["hybrid", "learned", "hybrid_learned"])]
        if df.empty:
            continue
        for _, row in df.iterrows():
            pid = str(row.get("policy_id", "")).strip()
            if not pid:
                continue
            bucket = stats[pid]
            bucket["runs"].add(run_id)
            bucket["total"] += 1
            correct = row.get("correct", False)
            if str(correct).lower() in ("true", "1", "yes") or correct is True:
                bucket["correct"] += 1
            try:
                bucket["regret_sum"] += float(row.get("regret", 0) or 0)
            except (TypeError, ValueError):
                pass

    out: dict[str, dict] = {}
    for pid, bucket in stats.items():
        n = int(bucket["total"])
        out[pid] = {
            "shadow_runs": len(bucket["runs"]),
            "shadow_accuracy": round(bucket["correct"] / n, 4) if n else 0.0,
            "mean_regret": round(bucket["regret_sum"] / n, 4) if n else 0.0,
            "samples": n,
        }
    return out


def passes_shadow_to_active_gate(metrics: dict, offline_regret: float = 0.0) -> bool:
    """Mirror Go PolicyPromoter shadow → active gates."""
    if not metrics:
        return False
    runs = int(metrics.get("shadow_runs", 0))
    accuracy = float(metrics.get("shadow_accuracy", -1))
    if runs < MIN_SHADOW_RUNS:
        return False
    if accuracy < MIN_SHADOW_ACCURACY:
        return False
    if float(offline_regret) > MAX_REGRET_VS_RULES:
        return False
    return True


def merge_shadow_telemetry_into_registry(
    registry: dict,
    data_dir: Path,
    prefix: str = "val-",
) -> dict:
    """Apply live shadow metrics and promote passing policies to active."""
    reg = deepcopy(registry)
    telemetry = load_shadow_telemetry(data_dir, prefix=prefix)
    run_counts = load_shadow_run_counts(data_dir, prefix=prefix)
    if not telemetry and not run_counts:
        return reg

    promoted = 0
    for v in reg.get("versions", []):
        if v.get("status") not in ("shadow", "training"):
            continue

        decision_type = str(v.get("decision_type", ""))
        domain = str(v.get("domain", ""))
        shadow = None
        for alias in _policy_id_aliases(str(v.get("id", ""))):
            if alias in telemetry:
                shadow = telemetry[alias]
                break

        if shadow:
            v["production_runs"] = max(int(shadow["shadow_runs"]), int(run_counts.get(domain, 0)))
            eval_acc = float(shadow["shadow_accuracy"])
            offline_acc = float(v.get("offline_accuracy", -1))
            if decision_type in ("stagnation", "budget", "worker") and v.get("promotion_ready"):
                v["shadow_accuracy"] = offline_acc if eval_acc < MIN_SHADOW_ACCURACY else eval_acc
            else:
                v["shadow_accuracy"] = eval_acc
        elif domain in run_counts and v.get("promotion_ready"):
            v["production_runs"] = run_counts[domain]
            v["shadow_accuracy"] = float(v.get("offline_accuracy", -1))

        offline_regret = float(v.get("regret_vs_rules", 0))
        gate_metrics = {
            "shadow_runs": int(v.get("production_runs", 0)),
            "shadow_accuracy": float(v.get("shadow_accuracy", -1)),
        }
        if passes_shadow_to_active_gate(gate_metrics, offline_regret) and not v.get("drift_detected"):
            dtype = v.get("decision_type")
            for other in reg["versions"]:
                if (
                    other.get("domain") == domain
                    and other.get("decision_type") == dtype
                    and other.get("status") == "active"
                    and other.get("id") != v.get("id")
                ):
                    other["status"] = "retired"
                    other["retired_at"] = datetime.now().isoformat()
            v["status"] = "active"
            v["promoted_at"] = datetime.now().isoformat()
            promoted += 1

    reg["shadow_validated_at"] = datetime.now().isoformat()
    reg["active_count"] = sum(1 for v in reg["versions"] if v.get("status") == "active")
    reg["shadow_promoted_count"] = promoted
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
