"""
Step 6 — Sequence / trajectory models on full search traces.

Builds per-run checkpoint sequences, engineers trajectory features, trains
contextual classifiers, and compares gain vs checkpoint-only stagnation models.
"""

from __future__ import annotations

from datetime import datetime

import numpy as np
import pandas as pd

from policy_training_utils import (
    STAGNATION_FEATURES,
    enrich_search_features,
    infer_instance_from_run_id,
    train_context_classifiers,
)
from policy_validation import detect_domain, ex_post_should_stop

TRAJECTORY_FEATURES = STAGNATION_FEATURES + [
    "trace_progress",
    "plateau_streak_ratio",
    "recent_slope",
    "volatility",
    "improvements_so_far",
    "remaining_budget_ratio",
    "acceptance_proxy",
]

MIN_GAIN_VS_CHECKPOINT = 0.0
MIN_TRAJECTORY_CV = 0.55


def enrich_trajectory_features(df: pd.DataFrame) -> pd.DataFrame:
    """Add sequence features by walking each run's checkpoint trace in order."""
    if df.empty:
        return df

    work = enrich_search_features(df)
    parts: list[pd.DataFrame] = []

    sort_col = "candidates" if "candidates" in work.columns else None
    for _run_id, group in work.groupby("run_id"):
        g = group.sort_values(sort_col) if sort_col else group.copy()
        g = g.reset_index(drop=True)
        n = len(g)
        if n < 2:
            continue

        budget = float(g["iterations_total"].iloc[0] or 100000)
        if budget <= 0:
            budget = 100000.0

        candidates = g["candidates"].astype(float).values
        current = g["current_penalty"].astype(float).values
        best = g["best_penalty"].astype(float).values
        final_best = float(g["final_best_penalty"].iloc[-1] if "final_best_penalty" in g else best[-1])

        trace_progress = candidates / max(candidates.max(), 1.0)
        remaining = 1.0 - trace_progress

        plateau_streak = np.zeros(n)
        improvements = np.zeros(n)
        for i in range(1, n):
            if best[i] < best[i - 1]:
                improvements[i] = 1
                plateau_streak[i] = 0
            else:
                plateau_streak[i] = plateau_streak[i - 1] + 1

        improvements_so_far = np.cumsum(improvements)

        recent_slope = np.zeros(n)
        volatility = np.zeros(n)
        for i in range(n):
            j = max(0, i - 3)
            dc = max(candidates[i] - candidates[j], 1.0)
            recent_slope[i] = (best[j] - best[i]) / dc
            w0 = max(0, i - 4)
            volatility[i] = float(np.std(current[w0 : i + 1])) if i > w0 else 0.0

        g["trace_progress"] = trace_progress
        g["remaining_budget_ratio"] = remaining
        g["plateau_streak_ratio"] = plateau_streak / max(n, 1)
        g["recent_slope"] = recent_slope
        g["volatility"] = volatility
        g["improvements_so_far"] = improvements_so_far
        g["acceptance_proxy"] = g.get("improvement_rate", pd.Series(0.0, index=g.index)).astype(float)
        g["final_best_penalty"] = final_best
        g["should_stop"] = g.apply(ex_post_should_stop, axis=1).astype(int)
        if "algorithm" not in g.columns:
            g["algorithm"] = "sa"
        g["instance"] = g["run_id"].apply(infer_instance_from_run_id)
        parts.append(g)

    if not parts:
        return pd.DataFrame()
    return pd.concat(parts, ignore_index=True)


def train_trajectory_policy(df: pd.DataFrame, min_samples: int = 100) -> dict:
    enriched = enrich_trajectory_features(df)
    if enriched.empty or len(enriched) < min_samples:
        return {"status": "insufficient_data", "samples": len(enriched)}

    classifiers = train_context_classifiers(
        enriched,
        TRAJECTORY_FEATURES,
        min_samples_context=max(80, min_samples // 4),
        min_samples_fallback=max(150, min_samples),
        use_boosting=True,
    )
    if not classifiers:
        return {"status": "insufficient_data", "samples": len(enriched)}

    checkpoint = train_context_classifiers(
        enriched,
        STAGNATION_FEATURES,
        min_samples_context=max(80, min_samples // 4),
        min_samples_fallback=max(150, min_samples),
        use_boosting=True,
    )
    traj_cv = float(np.mean([c["cv_mean"] for c in classifiers]))
    checkpoint_cv = float(np.mean([c["cv_mean"] for c in checkpoint])) if checkpoint else 0.0
    gain = traj_cv - checkpoint_cv

    promotion_ready = (
        traj_cv >= MIN_TRAJECTORY_CV
        and gain >= MIN_GAIN_VS_CHECKPOINT
        and len(classifiers) >= 2
    )

    return {
        "status": "trained",
        "version": "1.0.0",
        "trained_at": datetime.now().isoformat(),
        "samples": int(len(enriched)),
        "runs": int(enriched["run_id"].nunique()),
        "classifiers": classifiers,
        "episode_accuracy": round(traj_cv, 4),
        "checkpoint_baseline_cv": round(checkpoint_cv, 4),
        "gain_vs_checkpoint": round(gain, 4),
        "promotion_ready": promotion_ready,
        "promotion_gate": {
            "min_trajectory_cv": MIN_TRAJECTORY_CV,
            "min_gain_vs_checkpoint": MIN_GAIN_VS_CHECKPOINT,
        },
    }


def summarize_trajectory_by_domain(df: pd.DataFrame) -> dict[str, dict]:
    enriched = enrich_trajectory_features(df)
    if enriched.empty:
        return {}
    out: dict[str, dict] = {}
    for domain, group in enriched.groupby(enriched["run_id"].apply(detect_domain)):
        out[str(domain)] = {
            "checkpoints": int(len(group)),
            "runs": int(group["run_id"].nunique()),
            "mean_trace_length": round(len(group) / max(group["run_id"].nunique(), 1), 2),
        }
    return out
