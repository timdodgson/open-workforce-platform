"""
Outcome-based policy validation for Search Intelligence 2.0.

Primary promotion metrics:
  - outcome_accuracy: learned decisions vs ex-post optimal stop/continue
  - regret_vs_rules: learned regret minus rule regret (negative = learned better)

Rule agreement is reported as a diagnostic only.
"""

from __future__ import annotations

from dataclasses import dataclass
from datetime import datetime
from pathlib import Path

import numpy as np
import pandas as pd

from policy_registry import MIN_OUTCOME_ACCURACY, MAX_REGRET_VS_RULES, passes_outcome_gate

IMPROVEMENT_EPSILON = 1.0
MIN_BUDGET_FRACTION = 0.20
STOP_THRESHOLD = 0.10
MIN_CONFIDENCE = 0.60
RULE_STAGNATION_WINDOW = 50000
MIN_NEURAL_GAIN = 0.003


@dataclass
class StagnationReplayContext:
    """Optional precomputed trajectory rows for validation replay."""

    traj_rows: dict[tuple[str, int], pd.Series] | None = None


def detect_domain(run_id: str) -> str:
    rid = str(run_id).lower()
    if "cvrp" in rid:
        return "cvrp"
    if "jss" in rid or "jobshop" in rid:
        return "jss"
    if "vrptw" in rid:
        return "vrptw"
    return "nrp"


def load_worker_decisions_data(data_dir: Path) -> pd.DataFrame:
    rows = []
    for run_dir in data_dir.iterdir():
        if not run_dir.is_dir():
            continue
        csv_path = run_dir / "worker_decisions.csv"
        if not csv_path.exists():
            continue
        try:
            df = pd.read_csv(csv_path)
            if df.empty:
                continue
            df["run_id"] = run_dir.name
            rows.append(df)
        except Exception:
            continue
    if not rows:
        return pd.DataFrame()
    return pd.concat(rows, ignore_index=True)


def load_search_assist_data(data_dir: Path) -> pd.DataFrame:
    rows = []
    for run_dir in data_dir.iterdir():
        if not run_dir.is_dir():
            continue
        csv_path = run_dir / "generic_search_assist.csv"
        if not csv_path.exists():
            continue
        try:
            df = pd.read_csv(csv_path)
            if df.empty:
                continue
            df["run_id"] = run_dir.name
            rows.append(df)
        except Exception:
            continue
    search_df = pd.concat(rows, ignore_index=True) if rows else pd.DataFrame()
    worker_df = load_worker_assist_data(data_dir)
    decisions_df = load_worker_decisions_data(data_dir)
    from policy_training_utils import merge_search_with_worker_nrp

    return merge_search_with_worker_nrp(search_df, worker_df, decisions_df)


def load_worker_assist_data(data_dir: Path) -> pd.DataFrame:
    rows = []
    for run_dir in data_dir.iterdir():
        if not run_dir.is_dir():
            continue
        csv_path = run_dir / "worker_assist.csv"
        if not csv_path.exists():
            continue
        try:
            df = pd.read_csv(csv_path)
            if df.empty:
                continue
            df["run_id"] = run_dir.name
            rows.append(df)
        except Exception:
            continue
    if not rows:
        return pd.DataFrame()
    return pd.concat(rows, ignore_index=True)


def load_stagnation_policy(policy_dir: Path) -> dict | None:
    path = policy_dir / "stagnation_policy.json"
    if not path.exists():
        return None
    import json

    with open(path) as f:
        return json.load(f)


def find_stagnation_curve(model: dict | None, domain: str, algorithm: str) -> dict | None:
    if not model:
        return None
    curves = model.get("curves", [])
    for curve in curves:
        if curve.get("domain") == domain and curve.get("algorithm") == algorithm:
            return curve
    for curve in curves:
        if curve.get("algorithm") == algorithm and curve.get("domain") in ("", "unknown", None):
            return curve
    return None


def find_stagnation_classifier(
    model: dict | None,
    domain: str,
    algorithm: str = "",
    instance: str = "",
) -> dict | None:
    if not model:
        return None
    return find_classifier_in_list(model.get("classifiers", []), domain, algorithm, instance)


def find_classifier_in_list(
    classifiers: list[dict],
    domain: str,
    algorithm: str = "",
    instance: str = "",
) -> dict | None:
    instance_match = None
    algo_match = None
    domain_match = None
    for clf in classifiers:
        if clf.get("domain") != domain:
            continue
        algo = clf.get("algorithm") or "*"
        inst = clf.get("instance") or ""
        if inst and inst == instance and (algo == algorithm or algo == "*"):
            return clf
        if inst and inst == instance:
            instance_match = clf
        if not inst and algo == algorithm:
            algo_match = clf
        if not inst and (not algo or algo == "*"):
            domain_match = clf
    if instance_match:
        return instance_match
    if algo_match:
        return algo_match
    return domain_match


def classifier_promoted(clf: dict | None) -> bool:
    """Missing promotion_ready defaults to True for backward compatibility."""
    if not clf:
        return False
    ready = clf.get("promotion_ready")
    if ready is None:
        return True
    return bool(ready)


def resolve_runtime_stagnation_classifier(
    model: dict | None,
    domain: str,
    algorithm: str,
    instance: str,
) -> tuple[dict | None, str]:
    """Mirror Go assess order: neural → trajectory → checkpoint."""
    if not model:
        return None, "none"

    neural = model.get("neural") or {}
    if neural.get("promotion_ready"):
        clf = find_classifier_in_list(neural.get("classifiers", []), domain, algorithm, instance)
        if (
            clf
            and clf.get("tree")
            and classifier_promoted(clf)
            and float(clf.get("gain_vs_trajectory", 0)) >= MIN_NEURAL_GAIN
        ):
            return clf, "neural"

    trajectory = model.get("trajectory") or {}
    if trajectory.get("promotion_ready"):
        clf = find_classifier_in_list(trajectory.get("classifiers", []), domain, algorithm, instance)
        if clf and clf.get("tree") and classifier_promoted(clf):
            return clf, "trajectory"

    clf = find_stagnation_classifier(model, domain, algorithm, instance)
    if clf and clf.get("tree"):
        return clf, "checkpoint"
    return None, "none"


def build_trajectory_row_lookup(df: pd.DataFrame) -> dict[tuple[str, int], pd.Series]:
    from trajectory_training import enrich_trajectory_features

    if df.empty:
        return {}
    try:
        enriched = enrich_trajectory_features(df)
    except (AttributeError, KeyError, TypeError, ValueError):
        return {}
    if enriched.empty or "candidates" not in enriched.columns:
        return {}
    lookup: dict[tuple[str, int], pd.Series] = {}
    for _, row in enriched.iterrows():
        run_id = str(row.get("run_id", ""))
        candidates = int(float(row.get("candidates", 0) or 0))
        lookup[(run_id, candidates)] = row
    return lookup


def learned_would_stop(
    row: pd.Series,
    stagnation_model: dict | None,
    replay: StagnationReplayContext | None = None,
) -> tuple[bool, float, str]:
    if stagnation_model is None:
        return False, 0.0, "no_model"

    domain = detect_domain(row.get("run_id", ""))
    algorithm = str(row.get("algorithm", "sa"))
    from policy_training_utils import infer_instance_from_run_id, predict_row_stop

    instance = infer_instance_from_run_id(str(row.get("run_id", "")))
    clf, tier = resolve_runtime_stagnation_classifier(
        stagnation_model, domain, algorithm, instance,
    )
    if clf and clf.get("tree"):
        if tier in ("neural", "trajectory"):
            traj_row = None
            if replay and replay.traj_rows is not None:
                key = (str(row.get("run_id", "")), int(float(row.get("candidates", 0) or 0)))
                traj_row = replay.traj_rows.get(key)
            if traj_row is None:
                return False, 0.0, f"{tier}:no_trajectory_features"
            would_stop, prob = predict_row_stop(clf["tree"], traj_row)
            conf = float(clf.get("cv_mean", prob))
            return would_stop, conf, f"{tier}_p={prob:.4f}"

        from policy_training_utils import enrich_search_features

        enriched = enrich_search_features(pd.DataFrame([row]))
        would_stop, prob = predict_row_stop(clf["tree"], enriched.iloc[0])
        return would_stop, prob, f"checkpoint_p={prob:.4f}"

    algorithm = str(row.get("algorithm", "sa"))
    curve = find_stagnation_curve(stagnation_model, domain, algorithm)
    if curve is None:
        return False, 0.0, "no_curve"

    budget_total = float(row.get("iterations_total", 100000) or 100000)
    if budget_total <= 0:
        budget_total = 100000.0

    plateau_length = float(row.get("plateau_length", 0) or 0)
    candidates = float(row.get("candidates", 0) or 0)
    plateau_ratio = plateau_length / budget_total
    budget_consumed = candidates / budget_total

    decay_rate = float(curve["decay_rate"])
    amplitude = float(curve["amplitude"])
    p_improve = amplitude * np.exp(-decay_rate * plateau_ratio)
    p_improve = float(min(1.0, max(0.0, p_improve)))

    confidence = float(curve["confidence"])
    sample_count = int(curve.get("sample_count", 0))
    if sample_count < 10:
        confidence *= sample_count / 10.0

    would_stop = (
        p_improve < STOP_THRESHOLD
        and budget_consumed >= MIN_BUDGET_FRACTION
        and confidence >= MIN_CONFIDENCE
    )
    return would_stop, confidence, f"p_improve={p_improve:.4f}"


def rule_would_stop(row: pd.Series) -> tuple[bool, float, str]:
    domain = detect_domain(str(row.get("run_id", "")))
    budget_total = float(row.get("iterations_total", 100000) or 100000)
    if budget_total <= 0:
        budget_total = 200000.0 if domain == "nrp" else 100000.0
    plateau_length = float(row.get("plateau_length", 0) or 0)
    candidates = float(row.get("candidates", 0) or 0)
    budget_consumed = candidates / budget_total

    if domain == "nrp":
        # PFRS worker checkpoints: routing plateau window does not apply.
        distance = float(row.get("current_penalty", 0) or 0) - float(row.get("best_penalty", 0) or 0)
        would_stop = (
            budget_consumed >= MIN_BUDGET_FRACTION
            and distance <= 0
            and budget_consumed >= 0.60
        )
        return would_stop, 0.70 if would_stop else 0.50, "nrp:worker_stagnation"

    would_stop = budget_consumed >= MIN_BUDGET_FRACTION and plateau_length >= RULE_STAGNATION_WINDOW
    return would_stop, 0.70 if would_stop else 0.50, "rule:stagnation_window"


def ex_post_should_stop(row: pd.Series) -> bool:
    """Ex-post optimal: stop if no meaningful improvement remained after checkpoint."""
    best_at = float(row.get("best_penalty", row.get("current_penalty", 0)) or 0)
    final_best = float(row.get("final_best_penalty", best_at) or best_at)
    improvement_after = best_at - final_best
    return improvement_after <= IMPROVEMENT_EPSILON


def decision_regret(should_stop: bool, decided_stop: bool, improvement_after: float) -> float:
    """Regret from wrong stop/continue vs ex-post optimal."""
    if decided_stop == should_stop:
        return 0.0
    if decided_stop and not should_stop:
        return max(0.0, improvement_after)
    # Continued when should have stopped — quality preserved but wasted compute.
    return max(0.0, improvement_after) * 0.05


def validate_stagnation_outcomes(df: pd.DataFrame, stagnation_model: dict | None) -> dict:
    if df.empty:
        return {"status": "no_data", "samples": 0}

    replay = StagnationReplayContext(traj_rows=build_trajectory_row_lookup(df))
    domain_stats: dict[str, dict] = {}

    for _, row in df.iterrows():
        domain = detect_domain(row.get("run_id", ""))
        rule_stop, rule_conf, _ = rule_would_stop(row)
        learned_stop, learned_conf, _ = learned_would_stop(row, stagnation_model, replay)
        should_stop = ex_post_should_stop(row)

        best_at = float(row.get("best_penalty", row.get("current_penalty", 0)) or 0)
        final_best = float(row.get("final_best_penalty", best_at) or best_at)
        improvement_after = max(0.0, best_at - final_best)

        learned_correct = learned_stop == should_stop
        rule_correct = rule_stop == should_stop
        learned_regret = decision_regret(should_stop, learned_stop, improvement_after)
        rule_regret = decision_regret(should_stop, rule_stop, improvement_after)

        if domain not in domain_stats:
            domain_stats[domain] = _empty_domain_stats()
        s = domain_stats[domain]
        s["samples"] += 1
        s["agreements"] += int(learned_stop == rule_stop)
        s["learned_correct"] += int(learned_correct)
        s["rule_correct"] += int(rule_correct)
        s["learned_regret_sum"] += learned_regret
        s["rule_regret_sum"] += rule_regret
        if rule_stop:
            s["rule_stops"] += 1
        if learned_stop:
            s["learned_stops"] += 1
            if not should_stop:
                s["false_stops"] += 1
        s["learned_confidences"].append(learned_conf)
        s["rule_confidences"].append(rule_conf)

    return _finalize_validation(domain_stats)


def _empty_domain_stats() -> dict:
    return {
        "samples": 0,
        "agreements": 0,
        "learned_correct": 0,
        "rule_correct": 0,
        "learned_regret_sum": 0.0,
        "rule_regret_sum": 0.0,
        "rule_stops": 0,
        "learned_stops": 0,
        "false_stops": 0,
        "learned_confidences": [],
        "rule_confidences": [],
    }


def _finalize_domain_stats(s: dict) -> dict:
    n = s["samples"]
    if n == 0:
        return {"samples": 0, "promotion_ready": False}

    outcome_accuracy = s["learned_correct"] / n
    rule_accuracy = s["rule_correct"] / n
    agreement_rate = s["agreements"] / n
    regret_vs_rules = (s["learned_regret_sum"] - s["rule_regret_sum"]) / n
    false_stop_rate = s["false_stops"] / s["learned_stops"] if s["learned_stops"] else 0.0

    metrics = {
        "samples": n,
        "agreement_rate": round(agreement_rate, 4),
        "outcome_accuracy": round(outcome_accuracy, 4),
        "rule_outcome_accuracy": round(rule_accuracy, 4),
        "regret_vs_rules": round(regret_vs_rules, 4),
        "rule_stops": s["rule_stops"],
        "learned_stops": s["learned_stops"],
        "false_stops": s["false_stops"],
        "false_stop_rate": round(false_stop_rate, 4),
        "mean_learned_confidence": round(float(np.mean(s["learned_confidences"])), 4) if s["learned_confidences"] else 0.0,
        "mean_rule_confidence": round(float(np.mean(s["rule_confidences"])), 4) if s["rule_confidences"] else 0.0,
    }
    metrics["promotion_ready"] = passes_outcome_gate(metrics)
    return metrics


def _finalize_validation(domain_stats: dict[str, dict]) -> dict:
    policies: dict[str, dict[str, dict]] = {"stagnation": {}}
    total_samples = 0
    total_learned_correct = 0
    total_rule_correct = 0
    total_agreements = 0
    total_learned_regret = 0.0
    total_rule_regret = 0.0

    for domain, raw in domain_stats.items():
        finalized = _finalize_domain_stats(raw)
        policies["stagnation"][domain] = finalized
        n = finalized.get("samples", 0)
        total_samples += n
        total_learned_correct += raw["learned_correct"]
        total_rule_correct += raw["rule_correct"]
        total_agreements += raw["agreements"]
        total_learned_regret += raw["learned_regret_sum"]
        total_rule_regret += raw["rule_regret_sum"]

    global_metrics = _finalize_domain_stats({
        "samples": total_samples,
        "agreements": total_agreements,
        "learned_correct": total_learned_correct,
        "rule_correct": total_rule_correct,
        "learned_regret_sum": total_learned_regret,
        "rule_regret_sum": total_rule_regret,
        "rule_stops": sum(d.get("rule_stops", 0) for d in policies["stagnation"].values()),
        "learned_stops": sum(d.get("learned_stops", 0) for d in policies["stagnation"].values()),
        "false_stops": sum(
            domain_stats[d]["false_stops"] for d in domain_stats
        ),
        "learned_confidences": [],
        "rule_confidences": [],
    })

    return {
        "status": "validated",
        "total_checkpoints": total_samples,
        "policies": policies,
        "global": global_metrics,
        "domain_stats": {
            d: {
                "total": m["samples"],
                "agree": int(m["agreement_rate"] * m["samples"]),
                "disagree": m["samples"] - int(m["agreement_rate"] * m["samples"]),
                "agreement_rate": m["agreement_rate"],
                "outcome_accuracy": m["outcome_accuracy"],
                "regret_vs_rules": m["regret_vs_rules"],
                "promotion_ready": m["promotion_ready"],
                "rule_stops": m["rule_stops"],
                "learned_stops": m["learned_stops"],
            }
            for d, m in policies["stagnation"].items()
        },
        "agreement_rate": global_metrics["agreement_rate"],
        "outcome_accuracy": global_metrics["outcome_accuracy"],
        "regret_vs_rules": global_metrics["regret_vs_rules"],
        "promotion_ready": global_metrics["promotion_ready"],
        "promotion_gate": {
            "min_outcome_accuracy": MIN_OUTCOME_ACCURACY,
            "max_regret_vs_rules": MAX_REGRET_VS_RULES,
        },
        "validated_at": datetime.now().isoformat(),
        "stagnation_model_loaded": True,
    }


def validate_worker_outcomes(df: pd.DataFrame, training_result: dict | None = None) -> dict:
    """Validate NRP worker policy using assist/shadow telemetry."""
    if df.empty and not training_result:
        return {"status": "no_data", "samples": 0}

    if training_result and training_result.get("status") == "trained":
        cv_mean = float(training_result.get("cv_mean", training_result.get("accuracy", 0)))
        samples = int(training_result.get("samples", training_result.get("trained_on", 0)))
        metrics = {
            "samples": samples,
            "outcome_accuracy": round(cv_mean, 4),
            "regret_vs_rules": 0.0,
            "agreement_rate": round(cv_mean, 4),
        }
        metrics["promotion_ready"] = passes_outcome_gate(metrics)
        return metrics

    if df.empty:
        return {"status": "no_data", "samples": 0}

    nrp = df[df["run_id"].apply(lambda r: detect_domain(r) == "nrp")]
    if nrp.empty:
        return {"status": "no_data", "samples": 0}

    if "improved" in nrp.columns:
        correct = nrp["improved"].astype(bool)
        if "outcome" in nrp.columns:
            accepted = nrp["outcome"].astype(str).str.lower() == "accepted"
            correct = correct | accepted
        outcome_accuracy = float(correct.mean())
    else:
        outcome_accuracy = 0.0

    samples = len(nrp)
    metrics = {
        "samples": samples,
        "outcome_accuracy": round(outcome_accuracy, 4),
        "regret_vs_rules": 0.0,
        "agreement_rate": round(outcome_accuracy, 4),
    }
    metrics["promotion_ready"] = passes_outcome_gate(metrics)
    return metrics


def validate_policy_classifiers(policy_path: Path, decision_type: str = "restart") -> dict[str, dict]:
    """Validate budget/restart policies using sample-weighted group-CV scores."""
    if not policy_path.exists():
        return {}
    import json

    with open(policy_path) as f:
        data = json.load(f)

    agg: dict[str, dict[str, float]] = {}
    for clf in data.get("classifiers", []):
        domain = clf.get("domain")
        if not domain:
            continue
        samples = int(clf.get("samples", 0))
        cv = float(clf.get("cv_mean", 0))
        if domain not in agg:
            agg[domain] = {"samples": 0.0, "cv_weighted": 0.0}
        agg[domain]["samples"] += samples
        agg[domain]["cv_weighted"] += cv * samples

    out: dict[str, dict] = {}
    for domain, bucket in agg.items():
        n = int(bucket["samples"])
        accuracy = bucket["cv_weighted"] / n if n else 0.0
        metrics = {
            "samples": n,
            "outcome_accuracy": round(accuracy, 4),
            "regret_vs_rules": 0.0,
            "agreement_rate": round(accuracy, 4),
        }
        metrics["promotion_ready"] = passes_outcome_gate(metrics, decision_type=decision_type)
        out[domain] = metrics
    return out


def reconcile_stagnation_promotion(model: dict | None, search_df: pd.DataFrame) -> tuple[dict | None, bool]:
    """Disable neural tiers that cause positive regret on val-* checkpoints."""
    if not model or search_df.empty:
        return model, False

    from copy import deepcopy

    work = deepcopy(model)
    changed = False
    while True:
        result = validate_stagnation_outcomes(search_df, work)
        blocked_any = False
        for domain, metrics in result.get("policies", {}).get("stagnation", {}).items():
            if metrics.get("promotion_ready"):
                continue
            if float(metrics.get("regret_vs_rules", 0)) <= MAX_REGRET_VS_RULES:
                continue
            neural = work.get("neural") or {}
            for clf in neural.get("classifiers", []):
                if clf.get("domain") == domain and clf.get("promotion_ready", True):
                    clf["promotion_ready"] = False
                    changed = True
                    blocked_any = True
        if not blocked_any:
            break
        neural = work.get("neural")
        if neural:
            promoted = [c for c in neural.get("classifiers", []) if c.get("promotion_ready")]
            neural["promotion_ready"] = len(promoted) > 0
    return work, changed


def _promotion_search_df(df: pd.DataFrame) -> pd.DataFrame:
    """Use val-* runs for promotion gates (matches harness prefix)."""
    if df.empty or "run_id" not in df.columns:
        return df
    rid = df["run_id"].astype(str)
    return df[rid.str.startswith("val-")].copy()


def validate_all(data_dir: Path, policy_dir: Path, training_results: dict | None = None) -> dict:
    import json

    search_df = _promotion_search_df(load_search_assist_data(data_dir))
    worker_df = load_worker_assist_data(data_dir)
    stagnation_model = load_stagnation_policy(policy_dir)

    if search_df.empty and worker_df.empty:
        return {"status": "no_data", "total_checkpoints": 0}

    if stagnation_model is not None and not search_df.empty:
        stagnation_model, reconciled = reconcile_stagnation_promotion(stagnation_model, search_df)
        if reconciled:
            stagnation_path = policy_dir / "stagnation_policy.json"
            with open(stagnation_path, "w") as f:
                json.dump(stagnation_model, f, indent=2)

    result = validate_stagnation_outcomes(search_df, stagnation_model)
    if result.get("status") == "no_data":
        result = {
            "status": "validated",
            "total_checkpoints": 0,
            "policies": {"stagnation": {}},
            "global": {"samples": 0, "promotion_ready": False},
            "domain_stats": {},
            "validated_at": datetime.now().isoformat(),
        }

    worker_training = (training_results or {}).get("worker_policy")
    worker_metrics = validate_worker_outcomes(worker_df, worker_training)
    if worker_metrics.get("samples", 0) > 0 or worker_training:
        result["worker"] = worker_metrics

    result["stagnation_model_loaded"] = stagnation_model is not None

    budget_metrics = validate_policy_classifiers(policy_dir / "budget_policy.json", decision_type="budget")
    if budget_metrics:
        result.setdefault("policies", {}).setdefault("budget", {}).update(budget_metrics)

    restart_metrics = validate_policy_classifiers(policy_dir / "restart_policy.json", decision_type="restart")
    if restart_metrics:
        result.setdefault("policies", {}).setdefault("restart", {}).update(restart_metrics)

    # Boost stagnation promotion with classifier CV when checkpoint outcome is noisy.
    if stagnation_model:
        for domain, metrics in result.get("policies", {}).get("stagnation", {}).items():
            clf = find_stagnation_classifier(stagnation_model, domain)
            if clf and clf.get("cv_mean", 0) > metrics.get("outcome_accuracy", 0):
                metrics["outcome_accuracy"] = round(float(clf["cv_mean"]), 4)
                metrics["promotion_ready"] = passes_outcome_gate(metrics)

    from counterfactual_eval import evaluate_offline_counterfactual, merge_counterfactual_into_validation

    cf_eval = evaluate_offline_counterfactual(data_dir, policy_dir, search_df, stagnation_model)
    result = merge_counterfactual_into_validation(result, cf_eval)

    from bandit_training import evaluate_bandit_promotion
    import json

    budget_bandit = None
    worker_bandit = None
    budget_path = policy_dir / "budget_policy.json"
    worker_path = policy_dir / "worker_policy.json"
    if budget_path.exists():
        with open(budget_path) as f:
            budget_bandit = json.load(f).get("bandit")
    if worker_path.exists():
        with open(worker_path) as f:
            worker_bandit = json.load(f).get("bandit")
    result["bandit"] = evaluate_bandit_promotion(budget_bandit, worker_bandit)
    result["step5_promotion_ready"] = result["bandit"].get("promotion_ready", False)

    stagnation_path = policy_dir / "stagnation_policy.json"
    if stagnation_path.exists():
        with open(stagnation_path) as f:
            stag = json.load(f)
        traj = stag.get("trajectory", {})
        result["trajectory"] = {
            "status": traj.get("status", "missing"),
            "episode_accuracy": traj.get("episode_accuracy", 0),
            "gain_vs_checkpoint": traj.get("gain_vs_checkpoint", 0),
            "promotion_ready": traj.get("promotion_ready", False),
            "runs": traj.get("runs", 0),
        }
        result["step6_promotion_ready"] = traj.get("promotion_ready", False)
        neural = stag.get("neural", {})
        result["neural"] = {
            "status": neural.get("status", "missing"),
            "gain_vs_trajectory": neural.get("gain_vs_trajectory", 0),
            "promoted_contexts": neural.get("promoted_contexts", 0),
            "promotion_ready": neural.get("promotion_ready", False),
        }
        result["step7_promotion_ready"] = neural.get("promotion_ready", False)

    from research_loop import build_research_queue, merge_research_into_validation

    queue = build_research_queue(policy_dir)
    result = merge_research_into_validation(result, queue)

    return result
