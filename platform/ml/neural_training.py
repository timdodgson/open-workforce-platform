"""
Step 7 — Neural policies only where Step 6 trajectory models plateau.

Trains small MLPs on plateau contexts, distills winners to trees for Go inference.
Skipped entirely when global trajectory gain is already strong.
"""

from __future__ import annotations

from datetime import datetime

import numpy as np
import pandas as pd
from sklearn.metrics import accuracy_score
from sklearn.model_selection import GroupKFold, cross_val_predict
from sklearn.neural_network import MLPClassifier
from sklearn.tree import DecisionTreeClassifier

from policy_registry import MAX_FALSE_STOP_RATE, MAX_REGRET_VS_RULES
from policy_training_utils import export_sklearn_tree, predict_row_stop
from trajectory_training import TRAJECTORY_FEATURES, enrich_trajectory_features

PLATEAU_GLOBAL_GAIN = 0.01
MIN_NEURAL_GAIN = 0.003
MIN_NEURAL_CV = 0.55
MIN_CONTEXT_SAMPLES = 120
MIN_PROMOTION_STOPS = 20


def _context_mask(df: pd.DataFrame, clf: dict) -> pd.Series:
    mask = df["domain"] == clf.get("domain")
    algo = clf.get("algorithm") or "*"
    if algo != "*" and "algorithm" in df.columns:
        mask &= df["algorithm"].astype(str) == str(algo)
    inst = clf.get("instance") or ""
    if inst and "instance" in df.columns:
        mask &= df["instance"].astype(str) == str(inst)
    return mask


def _train_mlp_distilled(
    subset: pd.DataFrame,
    feature_cols: list[str],
    baseline_cv: float,
) -> dict | None:
    cols = [c for c in feature_cols if c in subset.columns]
    if len(cols) < 5 or len(subset) < MIN_CONTEXT_SAMPLES:
        return None

    X = subset[cols].fillna(0).astype(float).values
    y = subset["should_stop"].astype(int).values
    groups = subset["run_id"].values if "run_id" in subset.columns else np.arange(len(subset))
    if len(np.unique(y)) < 2:
        return None

    n_splits = min(5, max(2, len(np.unique(groups)) // 10))
    if n_splits < 2:
        return None

    mlp = MLPClassifier(
        hidden_layer_sizes=(48, 24),
        activation="relu",
        alpha=1e-4,
        max_iter=250,
        early_stopping=True,
        validation_fraction=0.15,
        random_state=42,
    )
    cv = GroupKFold(n_splits=n_splits)
    try:
        mlp_pred = cross_val_predict(mlp, X, y, cv=cv, groups=groups, method="predict")
        cv_mlp = float(accuracy_score(y, mlp_pred))
    except Exception:
        return None

    if cv_mlp < baseline_cv + MIN_NEURAL_GAIN or cv_mlp < MIN_NEURAL_CV:
        return None

    mlp.fit(X, y)
    distill = DecisionTreeClassifier(
        max_depth=12,
        min_samples_leaf=12,
        class_weight="balanced",
        random_state=42,
    )
    distill.fit(X, mlp.predict(X))

    return {
        "domain": str(subset["domain"].iloc[0]),
        "algorithm": str(subset["algorithm"].iloc[0]) if "algorithm" in subset.columns else "*",
        "instance": str(subset["instance"].iloc[0]) if "instance" in subset.columns else "",
        "features_used": cols,
        "samples": int(len(y)),
        "cv_mean": round(cv_mlp, 4),
        "cv_trajectory": round(baseline_cv, 4),
        "cv_neural": round(cv_mlp, 4),
        "gain_vs_trajectory": round(cv_mlp - baseline_cv, 4),
        "trainer": "neural_distilled",
        "positive_rate": round(float(y.mean()), 4),
        "tree": export_sklearn_tree(distill, cols),
    }


def annotate_neural_promotion(enriched: pd.DataFrame, classifiers: list[dict]) -> list[dict]:
    from policy_validation import decision_regret, rule_would_stop

    annotated: list[dict] = []
    for clf in classifiers:
        entry = dict(clf)
        subset = enriched[_context_mask(enriched, clf)]
        learned_stops = 0
        false_stops = 0
        learned_regret_sum = 0.0
        rule_regret_sum = 0.0
        rows = 0
        for _, row in subset.iterrows():
            stop, _ = predict_row_stop(clf["tree"], row)
            should_stop = int(row.get("should_stop", 0)) == 1
            rule_stop, _, _ = rule_would_stop(row)
            best_at = float(row.get("best_penalty", row.get("current_penalty", 0)) or 0)
            final_best = float(row.get("final_best_penalty", best_at) or best_at)
            improvement_after = max(0.0, best_at - final_best)
            learned_regret_sum += decision_regret(should_stop, stop, improvement_after)
            rule_regret_sum += decision_regret(should_stop, rule_stop, improvement_after)
            rows += 1
            if not stop:
                continue
            learned_stops += 1
            if not should_stop:
                false_stops += 1
        false_rate = false_stops / learned_stops if learned_stops else 0.0
        regret_vs_rules = (learned_regret_sum - rule_regret_sum) / rows if rows else 0.0
        entry["learned_stops"] = learned_stops
        entry["false_stops"] = false_stops
        entry["false_stop_rate"] = round(false_rate, 4)
        entry["regret_vs_rules"] = round(regret_vs_rules, 4)
        entry["promotion_ready"] = (
            learned_stops >= MIN_PROMOTION_STOPS
            and false_rate <= MAX_FALSE_STOP_RATE
            and regret_vs_rules <= MAX_REGRET_VS_RULES
        )
        annotated.append(entry)
    return annotated


def train_neural_where_plateau(
    df: pd.DataFrame,
    trajectory: dict | None,
    min_samples: int = 100,
) -> dict:
    if not trajectory or trajectory.get("status") != "trained":
        return {"status": "skipped", "reason": "no_trajectory", "samples": 0}

    global_gain = float(trajectory.get("gain_vs_checkpoint", 0))
    if global_gain >= PLATEAU_GLOBAL_GAIN:
        return {
            "status": "skipped",
            "reason": "trajectory_not_plateau",
            "global_trajectory_gain": round(global_gain, 4),
            "samples": 0,
        }

    enriched = enrich_trajectory_features(df)
    if enriched.empty or len(enriched) < min_samples:
        return {"status": "insufficient_data", "samples": len(enriched)}

    promoted: list[dict] = []
    skipped_contexts = 0
    traj_classifiers = trajectory.get("classifiers", [])

    for traj_clf in traj_classifiers:
        mask = _context_mask(enriched, traj_clf)
        subset = enriched[mask]
        baseline_cv = float(traj_clf.get("cv_mean", 0))
        neural_clf = _train_mlp_distilled(subset, TRAJECTORY_FEATURES, baseline_cv)
        if neural_clf:
            promoted.append(neural_clf)
        else:
            skipped_contexts += 1

    promoted = annotate_neural_promotion(enriched, promoted)
    promoted = [c for c in promoted if c.get("promotion_ready")]

    if not promoted:
        return {
            "status": "no_winner",
            "reason": "mlp_did_not_beat_trajectory",
            "global_trajectory_gain": round(global_gain, 4),
            "plateau_contexts_tried": len(traj_classifiers),
            "skipped_contexts": skipped_contexts,
            "samples": int(len(enriched)),
        }

    gains = [c["gain_vs_trajectory"] for c in promoted]
    mean_gain = float(np.mean(gains))
    return {
        "status": "trained",
        "version": "1.0.0",
        "trained_at": datetime.now().isoformat(),
        "plateau_detected": True,
        "global_trajectory_gain": round(global_gain, 4),
        "classifiers": promoted,
        "promoted_contexts": len(promoted),
        "skipped_contexts": skipped_contexts,
        "gain_vs_trajectory": round(mean_gain, 4),
        "promotion_ready": mean_gain >= MIN_NEURAL_GAIN and len(promoted) >= 1,
        "samples": int(len(enriched)),
        "promotion_gate": {
            "plateau_global_gain_lt": PLATEAU_GLOBAL_GAIN,
            "min_neural_gain": MIN_NEURAL_GAIN,
        },
    }
