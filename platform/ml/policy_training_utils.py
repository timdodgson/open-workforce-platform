"""Shared training helpers for SI policy classifiers."""

from __future__ import annotations

import numpy as np
import pandas as pd
from sklearn.model_selection import GroupKFold, cross_val_predict
from sklearn.ensemble import GradientBoostingClassifier
from sklearn.tree import DecisionTreeClassifier
from sklearn.metrics import accuracy_score

from policy_validation import detect_domain, ex_post_should_stop


def is_valid_search_checkpoint(row: pd.Series) -> bool:
    """Routing checkpoints must have budget progress; NRP worker-derived rows are always valid."""
    if bool(row.get("_worker_derived", False)):
        return True
    iterations = float(row.get("iterations_total", 0) or 0)
    candidates = float(row.get("candidates", 0) or 0)
    return iterations > 0 and candidates > 0


def worker_assist_to_search_frame(worker_df: pd.DataFrame) -> pd.DataFrame:
    """Map NRP worker_assist.csv rows to generic_search_assist schema for SI training."""
    if worker_df.empty:
        return pd.DataFrame()

    budget = pd.to_numeric(worker_df.get("final_budget"), errors="coerce")
    budget = budget.fillna(pd.to_numeric(worker_df.get("suggested_budget"), errors="coerce"))
    budget = budget.fillna(200000).astype(int)

    return pd.DataFrame({
        "run_id": worker_df["run_id"],
        "algorithm": worker_df.get("algorithm", "sa"),
        "candidates": budget,
        "iterations_total": budget,
        "plateau_length": pd.to_numeric(worker_df.get("distance_from_best"), errors="coerce").fillna(0).clip(lower=0).astype(int),
        "current_penalty": pd.to_numeric(worker_df.get("parent_objective"), errors="coerce").fillna(0).astype(int),
        "best_penalty": pd.to_numeric(worker_df.get("global_best"), errors="coerce").fillna(0).astype(int),
        "initial_penalty": pd.to_numeric(worker_df.get("parent_objective"), errors="coerce").fillna(0).astype(int),
        "improvement_rate": 0.0,
        "temperature": 0.0,
        "final_best_penalty": pd.to_numeric(worker_df.get("final_objective"), errors="coerce").fillna(0).astype(int),
        "_worker_derived": True,
    })


def merge_search_with_worker_nrp(search_df: pd.DataFrame, worker_df: pd.DataFrame) -> pd.DataFrame:
    """
    Unified search-assist frame for training/validation.

    NRP generic_search_assist.csv from the Go adapter often lacks budget fields.
    Replace invalid NRP rows with worker_assist-derived checkpoints.
    """
    worker_search = worker_assist_to_search_frame(worker_df)
    if worker_search.empty and search_df.empty:
        return pd.DataFrame()
    if search_df.empty:
        return worker_search

    search_df = search_df.copy()
    non_nrp = search_df[search_df["run_id"].apply(lambda r: detect_domain(r) != "nrp")]
    valid_nrp = search_df[
        search_df["run_id"].apply(lambda r: detect_domain(r) == "nrp")
        & search_df.apply(is_valid_search_checkpoint, axis=1)
    ]
    nrp_worker = worker_search[worker_search["run_id"].apply(lambda r: detect_domain(r) == "nrp")]

    parts = [non_nrp[non_nrp.apply(is_valid_search_checkpoint, axis=1)]]
    if not valid_nrp.empty:
        parts.append(valid_nrp)
    if not nrp_worker.empty:
        parts.append(nrp_worker)

    return pd.concat(parts, ignore_index=True)


STAGNATION_FEATURES = [
    "candidates",
    "iterations_total",
    "plateau_length",
    "current_penalty",
    "best_penalty",
    "initial_penalty",
    "improvement_rate",
    "temperature",
    "budget_consumed",
    "plateau_ratio",
    "distance_from_best",
]


def enrich_search_features(df: pd.DataFrame) -> pd.DataFrame:
    out = df.copy()
    budget = out.get("iterations_total", pd.Series(100000, index=out.index)).astype(float)
    budget = budget.replace(0, 100000)
    out["budget_consumed"] = out.get("candidates", 0).astype(float) / budget
    out["plateau_ratio"] = out.get("plateau_length", 0).astype(float) / budget
    out["distance_from_best"] = (
        out.get("current_penalty", 0).astype(float) - out.get("best_penalty", 0).astype(float)
    )
    if "run_id" not in out.columns:
        out["run_id"] = "unknown"
    out["domain"] = out["run_id"].apply(detect_domain)
    out["should_stop"] = out.apply(ex_post_should_stop, axis=1).astype(int)
    return out


def train_domain_classifier(
    df: pd.DataFrame,
    domain: str,
    feature_cols: list[str],
    label_col: str = "should_stop",
    min_samples: int = 100,
    use_boosting: bool = False,
) -> dict | None:
    subset = df[df["domain"] == domain]
    if len(subset) < min_samples:
        return None

    cols = [c for c in feature_cols if c in subset.columns]
    if len(cols) < 3:
        return None

    X = subset[cols].fillna(0).astype(float).values
    y = subset[label_col].astype(int).values
    groups = subset["run_id"].values

    if len(np.unique(y)) < 2:
        return None

    n_groups = len(np.unique(groups))
    n_splits = min(5, max(2, n_groups // 10))
    if n_splits < 2:
        return None

    model = DecisionTreeClassifier(
        max_depth=8,
        min_samples_leaf=25,
        class_weight="balanced",
        random_state=42,
    )
    boost = GradientBoostingClassifier(
        n_estimators=120,
        max_depth=4,
        learning_rate=0.08,
        random_state=42,
    )
    cv = GroupKFold(n_splits=n_splits)

    try:
        oof_pred = cross_val_predict(model, X, y, cv=cv, groups=groups, method="predict")
        cv_mean = float(accuracy_score(y, oof_pred))
        if use_boosting:
            boost_pred = cross_val_predict(boost, X, y, cv=cv, groups=groups, method="predict")
            boost_cv = float(accuracy_score(y, boost_pred))
            if boost_cv > cv_mean:
                cv_mean = boost_cv
    except Exception:
        return None

    model.fit(X, y)

    return {
        "domain": domain,
        "algorithm": "*",
        "features_used": cols,
        "samples": int(len(y)),
        "cv_mean": round(cv_mean, 4),
        "positive_rate": round(float(y.mean()), 4),
        "tree": export_sklearn_tree(model, cols),
    }


def predict_row_stop(tree: dict, row: pd.Series) -> tuple[bool, float]:
    """Walk exported sklearn tree for a single checkpoint row."""
    if not tree:
        return False, 0.0

    names = tree.get("feature_names", [])
    features = [float(row.get(n, 0) or 0) for n in names]

    node = 0
    children_left = tree["children_left"]
    children_right = tree["children_right"]
    feat_idx = tree["feature"]
    threshold = tree["threshold"]
    values = tree["value"]

    while children_left[node] != children_right[node]:
        fi = feat_idx[node]
        if fi < 0 or fi >= len(features):
            break
        if features[fi] <= threshold[node]:
            node = children_left[node]
        else:
            node = children_right[node]

    leaf = values[node]
    if len(leaf) == 1:
        prob = float(leaf[0])
        pred = prob >= 0.5
    else:
        total = sum(leaf)
        prob = float(leaf[1] / total) if total > 0 else 0.5
        pred = prob >= 0.5
    return pred, prob


def export_sklearn_tree(model, feature_names: list) -> dict:
    """Export a sklearn DecisionTreeClassifier for Go runtime inference."""
    tree = model.tree_
    values = []
    for node_vals in tree.value:
        values.append(node_vals.flatten().tolist())
    return {
        "feature_names": feature_names,
        "children_left": tree.children_left.tolist(),
        "children_right": tree.children_right.tolist(),
        "feature": tree.feature.tolist(),
        "threshold": tree.threshold.tolist(),
        "value": values,
    }
