"""
Step 4 — Counterfactual offline evaluation before policy promotion.

Uses ex-post optimal stop/continue labels (primary) plus emitted
counterfactual_learning.csv telemetry (secondary) to gate deploy.
"""

from __future__ import annotations

from datetime import datetime
from pathlib import Path

import numpy as np
import pandas as pd

from policy_registry import MAX_FALSE_STOP_RATE, passes_counterfactual_gate
from policy_validation import (
    StagnationReplayContext,
    _promotion_search_df,
    build_trajectory_row_lookup,
    detect_domain,
    ex_post_should_stop,
    learned_would_stop,
    load_search_assist_data,
    load_stagnation_policy,
    rule_would_stop,
)

EARLY_STOP_ACTIONS = frozenset({"early_stop", "stop", "stagnation_stop"})


def load_counterfactual_data(data_dir: Path) -> pd.DataFrame:
    rows: list[pd.DataFrame] = []
    for run_dir in data_dir.iterdir():
        if not run_dir.is_dir():
            continue
        csv_path = run_dir / "counterfactual_learning.csv"
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


def _col(df: pd.DataFrame, *names: str) -> pd.Series:
    for name in names:
        if name in df.columns:
            return df[name]
    return pd.Series([""] * len(df), index=df.index)


def summarize_counterfactual_telemetry(cf_df: pd.DataFrame) -> dict:
    """Aggregate emitted counterfactual rows (secondary signal)."""
    if cf_df.empty:
        return {"status": "no_data", "samples": 0}

    actual = _col(cf_df, "actual_action", "ActualAction").astype(str).str.lower()
    regret = pd.to_numeric(_col(cf_df, "regret", "Regret"), errors="coerce").fillna(0.0)
    domain = _col(cf_df, "domain", "Domain").astype(str)

    early = actual.isin(EARLY_STOP_ACTIONS)
    false_from_regret = early & (regret > 0)

    by_domain: dict[str, dict] = {}
    for dom in sorted(domain.unique()):
        mask = domain == dom
        n = int(mask.sum())
        if n == 0:
            continue
        dom_early = int((mask & early).sum())
        dom_false = int((mask & false_from_regret).sum())
        dom_regret = regret[mask]
        by_domain[dom] = {
            "samples": n,
            "early_stops": dom_early,
            "false_stops": dom_false,
            "false_stop_rate": round(dom_false / dom_early, 4) if dom_early else 0.0,
            "mean_regret": round(float(dom_regret.mean()), 4),
            "regret_rate": round(float((dom_regret > 0).mean()), 4),
        }

    total = len(cf_df)
    early_n = int(early.sum())
    false_n = int(false_from_regret.sum())
    return {
        "status": "evaluated",
        "samples": total,
        "early_stops": early_n,
        "false_stops": false_n,
        "false_stop_rate": round(false_n / early_n, 4) if early_n else 0.0,
        "mean_regret": round(float(regret.mean()), 4),
        "regret_rate": round(float((regret > 0).mean()), 4),
        "by_domain": by_domain,
    }


def evaluate_offline_counterfactual(
    data_dir: Path,
    policy_dir: Path,
    search_df: pd.DataFrame | None = None,
    stagnation_model: dict | None = None,
) -> dict:
    """
    Simulate learned vs rule decisions on historical checkpoints.
    Primary gate: false_stop_rate (learned stopped when ex-post says continue).
    """
    if search_df is None:
        search_df = _promotion_search_df(load_search_assist_data(data_dir))
    else:
        search_df = _promotion_search_df(search_df)
    if stagnation_model is None:
        stagnation_model = load_stagnation_policy(policy_dir)

    if search_df.empty:
        return {"status": "no_data", "samples": 0, "promotion_ready": False}

    domain_stats: dict[str, dict] = {}
    replay = StagnationReplayContext(traj_rows=build_trajectory_row_lookup(search_df))
    for _, row in search_df.iterrows():
        domain = detect_domain(row.get("run_id", ""))
        learned_stop, _, _ = learned_would_stop(row, stagnation_model, replay)
        rule_stop, _, _ = rule_would_stop(row)
        should_stop = ex_post_should_stop(row)

        best_at = float(row.get("best_penalty", row.get("current_penalty", 0)) or 0)
        final_best = float(row.get("final_best_penalty", best_at) or best_at)
        improvement_after = max(0.0, best_at - final_best)

        if domain not in domain_stats:
            domain_stats[domain] = _empty_stats()
        s = domain_stats[domain]
        s["samples"] += 1
        if learned_stop:
            s["learned_stops"] += 1
            if not should_stop:
                s["false_stops"] += 1
                s["quality_lost"] += improvement_after
        if rule_stop:
            s["rule_stops"] += 1
            if not should_stop:
                s["rule_false_stops"] += 1
        if learned_stop == should_stop:
            s["learned_correct"] += 1

    domains_out: dict[str, dict] = {}
    total_samples = 0
    total_false = 0
    total_learned_stops = 0
    total_quality_lost = 0.0
    total_correct = 0

    for domain, raw in domain_stats.items():
        finalized = _finalize_domain(raw)
        domains_out[domain] = finalized
        total_samples += finalized["samples"]
        total_false += raw["false_stops"]
        total_learned_stops += raw["learned_stops"]
        total_quality_lost += raw["quality_lost"]
        total_correct += raw["learned_correct"]

    global_metrics = _finalize_domain({
        "samples": total_samples,
        "learned_stops": total_learned_stops,
        "false_stops": total_false,
        "rule_false_stops": sum(d["rule_false_stops"] for d in domain_stats.values()),
        "rule_stops": sum(d["rule_stops"] for d in domain_stats.values()),
        "quality_lost": total_quality_lost,
        "learned_correct": total_correct,
    })

    cf_df = load_counterfactual_data(data_dir)
    telemetry = summarize_counterfactual_telemetry(cf_df)

    return {
        "status": "evaluated",
        "evaluated_at": datetime.now().isoformat(),
        "samples": total_samples,
        "false_stops": total_false,
        "learned_stops": total_learned_stops,
        "false_stop_rate": global_metrics["false_stop_rate"],
        "quality_lost": round(total_quality_lost, 4),
        "outcome_accuracy": global_metrics["outcome_accuracy"],
        "promotion_ready": global_metrics["promotion_ready"],
        "promotion_gate": {
            "max_false_stop_rate": MAX_FALSE_STOP_RATE,
            "min_samples": 20,
        },
        "domains": domains_out,
        "global": global_metrics,
        "telemetry": telemetry,
    }


def _empty_stats() -> dict:
    return {
        "samples": 0,
        "learned_stops": 0,
        "false_stops": 0,
        "rule_stops": 0,
        "rule_false_stops": 0,
        "quality_lost": 0.0,
        "learned_correct": 0,
    }


def _finalize_domain(s: dict) -> dict:
    n = s["samples"]
    learned_stops = s["learned_stops"]
    false_stops = s["false_stops"]
    false_stop_rate = false_stops / learned_stops if learned_stops else 0.0
    rule_stops = s["rule_stops"]
    rule_false_rate = s["rule_false_stops"] / rule_stops if rule_stops else 0.0

    metrics = {
        "samples": n,
        "learned_stops": learned_stops,
        "false_stops": false_stops,
        "false_stop_rate": round(false_stop_rate, 4),
        "rule_false_stop_rate": round(rule_false_rate, 4),
        "quality_lost": round(s["quality_lost"], 4),
        "outcome_accuracy": round(s["learned_correct"] / n, 4) if n else 0.0,
    }
    metrics["promotion_ready"] = passes_counterfactual_gate(metrics)
    return metrics


def merge_counterfactual_into_validation(validation: dict, counterfactual: dict) -> dict:
    """Attach Step 4 counterfactual eval to validation_results.json."""
    out = dict(validation)
    out["counterfactual"] = counterfactual
    if counterfactual.get("status") == "evaluated":
        out["false_stop_rate"] = counterfactual.get("false_stop_rate", 1.0)
        out["step4_promotion_ready"] = counterfactual.get("promotion_ready", False)
        # Tighten stagnation promotion: require low false-stop rate per domain.
        for domain, metrics in counterfactual.get("domains", {}).items():
            stag = out.get("policies", {}).get("stagnation", {}).get(domain)
            if stag is None:
                continue
            stag["false_stop_rate"] = metrics.get("false_stop_rate", 1.0)
            stag["false_stops"] = metrics.get("false_stops", 0)
            if not metrics.get("promotion_ready", False):
                stag["promotion_ready"] = False
        global_stag = out.get("global", {})
        if global_stag and not counterfactual.get("promotion_ready", False):
            global_stag["promotion_ready"] = False
            out["promotion_ready"] = False
    return out
