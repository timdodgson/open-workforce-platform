"""
Step 5 — Offline contextual bandits for portfolio and worker budget decisions.

Trains per-context arm statistics from logged telemetry and exports JSON
for Go runtime lookup (no Python at solve time).
"""

from __future__ import annotations

from datetime import datetime

import numpy as np
import pandas as pd

from policy_training_utils import infer_instance_from_run_id

PORTFOLIO_ARMS = [0.5, 0.75, 1.0, 1.25, 1.5]
WORKER_ARMS = ("skip", "default", "boost")
WORKER_ARM_BUDGET = {"skip": 0.0, "default": 1.0, "boost": 1.25}
MAX_EPISODE_REGRET = 0.10
MIN_ARM_SAMPLES = 3


def _portfolio_instance_slug(row: pd.Series) -> str:
    inst = str(row.get("instance", "")).strip()
    if inst:
        return inst.lower().replace("-", "").replace("_", "")[:12]
    return infer_instance_from_run_id(str(row.get("run_id", "")))


def _nearest_portfolio_arm(mult: float) -> float:
    return min(PORTFOLIO_ARMS, key=lambda a: abs(a - mult))


def _portfolio_reward(row: pd.Series) -> float:
    if "strategy_won" in row and pd.notna(row["strategy_won"]):
        return float(bool(int(row["strategy_won"])))
    return 0.0


def _observed_portfolio_mult(row: pd.Series) -> float:
    orig = float(row.get("original_budget", 0) or 0)
    final = float(row.get("final_budget", row.get("recommended_budget", orig)) or orig)
    if orig <= 0:
        return 1.0
    return float(np.clip(final / orig, 0.25, 2.0))


def train_portfolio_bandit(df: pd.DataFrame, min_samples: int = 20) -> dict:
    if df.empty or len(df) < min_samples:
        return {"status": "insufficient_data", "samples": len(df)}

    work = df.copy()
    if "run_id" not in work.columns:
        work["run_id"] = work.index.astype(str)
    work["instance_slug"] = work.apply(_portfolio_instance_slug, axis=1)
    work["observed_mult"] = work.apply(_observed_portfolio_mult, axis=1)
    work["observed_arm"] = work["observed_mult"].apply(_nearest_portfolio_arm)
    work["reward"] = work.apply(_portfolio_reward, axis=1)

    entries: list[dict] = []
    regrets: list[float] = []

    group_cols = ["domain", "instance_slug", "strategy"]
    for (domain, instance, strategy), group in work.groupby(group_cols):
        if len(group) < MIN_ARM_SAMPLES:
            continue

        arm_stats: list[dict] = []
        for arm in PORTFOLIO_ARMS:
            arm_rows = group[group["observed_arm"] == arm]
            n = len(arm_rows)
            mean_reward = float(arm_rows["reward"].mean()) if n else 0.0
            arm_stats.append({
                "arm": str(arm),
                "value": float(arm),
                "mean_reward": round(mean_reward, 4),
                "samples": int(n),
            })

        observed = arm_stats
        best = max(observed, key=lambda a: (a["mean_reward"], a["samples"]))
        baseline = next((a for a in observed if a["value"] == 1.0), None)
        baseline_reward = baseline["mean_reward"] if baseline else 0.0
        episode_regret = max(0.0, best["mean_reward"] - baseline_reward)
        regrets.append(episode_regret)

        confidence = min(0.95, 0.45 + len(group) / 120.0)
        entries.append({
            "domain": str(domain),
            "instance": str(instance) if instance else "",
            "strategy": str(strategy),
            "arms": observed,
            "best_arm": best["arm"],
            "best_value": best["value"],
            "confidence": round(confidence, 4),
            "episode_regret": round(episode_regret, 4),
            "samples": int(len(group)),
        })

    if not entries:
        return {"status": "insufficient_data", "samples": len(df)}

    global_regret = float(np.mean(regrets)) if regrets else 1.0
    return {
        "status": "trained",
        "version": "1.0.0",
        "trained_at": datetime.now().isoformat(),
        "arms": PORTFOLIO_ARMS,
        "entries": entries,
        "episode_regret": round(global_regret, 4),
        "promotion_ready": global_regret <= MAX_EPISODE_REGRET and len(entries) >= 2,
        "samples": int(len(df)),
    }


def _worker_context(row: pd.Series) -> str:
    week = int(row.get("week", 0) or 0)
    depth = int(row.get("depth", 0) or 0)
    dist = float(row.get("distance_from_best", 0) or 0)
    dist_bucket = "far" if dist > 50 else "near"
    return f"week={week}|depth={depth}|dist={dist_bucket}"


def _observed_worker_arm(row: pd.Series) -> str:
    if "improved" in row and not bool(row.get("improved", True)) and float(row.get("final_budget", 1) or 1) <= 0:
        return "skip"
    suggested = float(row.get("suggested_budget", 0) or 0)
    final = float(row.get("final_budget", suggested) or suggested)
    if suggested <= 0:
        return "default"
    ratio = final / suggested
    if ratio < 0.6:
        return "skip"
    if ratio > 1.1:
        return "boost"
    return "default"


def _worker_reward(row: pd.Series) -> float:
    for col in ("improved", "produced_global_best"):
        if col in row and pd.notna(row[col]):
            return float(bool(row[col]))
    return 0.0


def train_worker_bandit(df: pd.DataFrame, min_samples: int = 20) -> dict:
    if df.empty or len(df) < min_samples:
        return {"status": "insufficient_data", "samples": len(df)}

    work = df.copy()
    work["context"] = work.apply(_worker_context, axis=1)
    work["observed_arm"] = work.apply(_observed_worker_arm, axis=1)
    work["reward"] = work.apply(_worker_reward, axis=1)

    entries: list[dict] = []
    regrets: list[float] = []

    for context, group in work.groupby("context"):
        if len(group) < MIN_ARM_SAMPLES:
            continue

        arm_stats: list[dict] = []
        for arm in WORKER_ARMS:
            arm_rows = group[group["observed_arm"] == arm]
            n = len(arm_rows)
            mean_reward = float(arm_rows["reward"].mean()) if n else 0.0
            arm_stats.append({
                "arm": arm,
                "value": WORKER_ARM_BUDGET[arm],
                "mean_reward": round(mean_reward, 4),
                "samples": int(n),
            })

        best = max(arm_stats, key=lambda a: (a["mean_reward"], a["samples"]))
        baseline = next(a for a in arm_stats if a["arm"] == "default")
        episode_regret = max(0.0, best["mean_reward"] - baseline["mean_reward"])
        regrets.append(episode_regret)

        entries.append({
            "domain": "nrp",
            "context": str(context),
            "arms": arm_stats,
            "best_arm": best["arm"],
            "best_value": best["value"],
            "confidence": round(min(0.95, 0.45 + len(group) / 200.0), 4),
            "episode_regret": round(episode_regret, 4),
            "samples": int(len(group)),
        })

    if not entries:
        return {"status": "insufficient_data", "samples": len(df)}

    global_regret = float(np.mean(regrets)) if regrets else 1.0
    return {
        "status": "trained",
        "version": "1.0.0",
        "trained_at": datetime.now().isoformat(),
        "arms": list(WORKER_ARMS),
        "entries": entries,
        "episode_regret": round(global_regret, 4),
        "promotion_ready": global_regret <= MAX_EPISODE_REGRET and len(entries) >= 3,
        "samples": int(len(df)),
    }


def evaluate_bandit_promotion(budget_bandit: dict | None, worker_bandit: dict | None) -> dict:
    budget_ok = bool(budget_bandit and budget_bandit.get("promotion_ready"))
    worker_ok = bool(worker_bandit and worker_bandit.get("promotion_ready"))
    budget_regret = float((budget_bandit or {}).get("episode_regret", 1.0))
    worker_regret = float((worker_bandit or {}).get("episode_regret", 1.0))
    return {
        "status": "evaluated",
        "portfolio_episode_regret": round(budget_regret, 4),
        "worker_episode_regret": round(worker_regret, 4),
        "portfolio_ready": budget_ok,
        "worker_ready": worker_ok,
        "promotion_ready": budget_ok or worker_ok,
        "promotion_gate": {"max_episode_regret": MAX_EPISODE_REGRET},
        "evaluated_at": datetime.now().isoformat(),
    }
