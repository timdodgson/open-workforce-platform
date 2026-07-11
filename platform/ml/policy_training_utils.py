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


def worker_decisions_to_search_frame(decisions_df: pd.DataFrame) -> pd.DataFrame:
    """Map NRP worker_decisions.csv rows to generic_search_assist schema for SI training."""
    if decisions_df.empty:
        return pd.DataFrame()

    budget = pd.to_numeric(decisions_df.get("allocated_iters"), errors="coerce")
    budget = budget.fillna(pd.to_numeric(decisions_df.get("suggested_budget"), errors="coerce"))
    budget = budget.fillna(200000).astype(int)

    return pd.DataFrame({
        "run_id": decisions_df["run_id"],
        "algorithm": decisions_df.get("algorithm", "sa"),
        "candidates": budget,
        "iterations_total": budget,
        "plateau_length": pd.to_numeric(decisions_df.get("distance_from_best"), errors="coerce").fillna(0).clip(lower=0).astype(int),
        "current_penalty": pd.to_numeric(decisions_df.get("parent_objective"), errors="coerce").fillna(0).astype(int),
        "best_penalty": pd.to_numeric(decisions_df.get("global_best"), errors="coerce").fillna(0).astype(int),
        "initial_penalty": pd.to_numeric(decisions_df.get("parent_objective"), errors="coerce").fillna(0).astype(int),
        "improvement_rate": 0.0,
        "temperature": 0.0,
        "final_best_penalty": pd.to_numeric(decisions_df.get("final_objective"), errors="coerce").fillna(0).astype(int),
        "_worker_derived": True,
    })


def build_nrp_worker_search_frame(
    worker_df: pd.DataFrame,
    decisions_df: pd.DataFrame,
) -> pd.DataFrame:
    """Prefer worker_assist per run; fall back to worker_decisions for shadow-only runs."""
    assist = worker_assist_to_search_frame(worker_df)
    decisions = worker_decisions_to_search_frame(decisions_df)

    assist_runs = set(assist["run_id"].unique()) if not assist.empty else set()
    if not decisions.empty:
        decisions = decisions[~decisions["run_id"].isin(assist_runs)]

    parts = []
    if not assist.empty:
        parts.append(assist)
    if not decisions.empty:
        parts.append(decisions)
    if not parts:
        return pd.DataFrame()
    return pd.concat(parts, ignore_index=True)


def merge_search_with_worker_nrp(
    search_df: pd.DataFrame,
    worker_df: pd.DataFrame,
    decisions_df: pd.DataFrame | None = None,
) -> pd.DataFrame:
    """
    Unified search-assist frame for training/validation.

    NRP: never mix Go-adapted generic_search_assist with worker telemetry for the same
    run — the adapter duplicates worker rows with bad final_best_penalty on many spawns.
    Use worker_assist / worker_decisions only when present.
    """
    decisions_df = decisions_df if decisions_df is not None else pd.DataFrame()
    nrp_worker = build_nrp_worker_search_frame(worker_df, decisions_df)
    nrp_worker_run_ids = set(nrp_worker["run_id"].unique()) if not nrp_worker.empty else set()

    parts: list[pd.DataFrame] = []

    if not search_df.empty:
        non_nrp = search_df[search_df["run_id"].apply(lambda r: detect_domain(r) != "nrp")]
        non_nrp = non_nrp[non_nrp.apply(is_valid_search_checkpoint, axis=1)]
        if not non_nrp.empty:
            parts.append(non_nrp)

        if nrp_worker_run_ids:
            nrp_fallback = search_df[
                search_df["run_id"].apply(lambda r: detect_domain(r) == "nrp")
                & ~search_df["run_id"].isin(nrp_worker_run_ids)
                & search_df.apply(is_valid_search_checkpoint, axis=1)
            ]
            if not nrp_fallback.empty:
                parts.append(nrp_fallback)

    if not nrp_worker.empty:
        parts.append(nrp_worker)

    if not parts:
        if search_df.empty:
            return pd.DataFrame()
        return search_df[search_df.apply(is_valid_search_checkpoint, axis=1)].copy()

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


def infer_instance_from_run_id(run_id: str) -> str:
    """Extract benchmark instance slug from val-* run labels."""
    rid = str(run_id).lower()
    for token in (
        "n012w8", "n030w4", "a80k10", "a32k5", "a45k6", "a60k9",
        "la01", "ft10", "ft06", "c101",
    ):
        if token in rid:
            return token
    parts = str(run_id).split("-")
    for p in parts:
        pl = p.lower()
        if pl.startswith(("a", "n", "la", "ft", "c")) and any(ch.isdigit() for ch in pl):
            return pl
    return ""


def train_domain_classifier(
    df: pd.DataFrame,
    domain: str,
    feature_cols: list[str],
    label_col: str = "should_stop",
    min_samples: int = 100,
    use_boosting: bool = True,
    algorithm: str = "*",
    instance: str = "",
) -> dict | None:
    subset = df[df["domain"] == domain] if "domain" in df.columns else df
    if algorithm != "*" and "algorithm" in subset.columns:
        subset = subset[subset["algorithm"].astype(str) == algorithm]
    if instance and "instance" in subset.columns:
        subset = subset[subset["instance"].astype(str) == instance]
    if len(subset) < min_samples:
        return None

    cols = [c for c in feature_cols if c in subset.columns]
    if len(cols) < 3:
        return None

    X = subset[cols].fillna(0).astype(float).values
    y = subset[label_col].astype(int).values
    groups = subset["run_id"].values if "run_id" in subset.columns else np.arange(len(subset))

    if len(np.unique(y)) < 2:
        return None

    n_groups = len(np.unique(groups))
    n_splits = min(5, max(2, n_groups // 10))
    if n_splits < 2:
        return None

    tree_model = DecisionTreeClassifier(
        max_depth=8,
        min_samples_leaf=25,
        class_weight="balanced",
        random_state=42,
    )
    boost_model = GradientBoostingClassifier(
        n_estimators=120,
        max_depth=4,
        learning_rate=0.08,
        random_state=42,
    )
    cv = GroupKFold(n_splits=n_splits)

    try:
        tree_pred = cross_val_predict(tree_model, X, y, cv=cv, groups=groups, method="predict")
        cv_tree = float(accuracy_score(y, tree_pred))
        cv_boost = None
        trainer = "tree"
        deploy_model = tree_model

        if use_boosting:
            boost_pred = cross_val_predict(boost_model, X, y, cv=cv, groups=groups, method="predict")
            cv_boost = float(accuracy_score(y, boost_pred))
            if cv_boost >= cv_tree:
                trainer = "boost_distilled"
                boost_model.fit(X, y)
                distill = DecisionTreeClassifier(
                    max_depth=10,
                    min_samples_leaf=15,
                    class_weight="balanced",
                    random_state=42,
                )
                distill.fit(X, boost_model.predict(X))
                deploy_model = distill
                cv_mean = cv_boost
            else:
                tree_model.fit(X, y)
                deploy_model = tree_model
                cv_mean = cv_tree
        else:
            tree_model.fit(X, y)
            deploy_model = tree_model
            cv_mean = cv_tree
    except Exception:
        return None

    return {
        "domain": domain,
        "algorithm": algorithm,
        "instance": instance,
        "features_used": cols,
        "samples": int(len(y)),
        "cv_mean": round(cv_mean, 4),
        "cv_tree": round(cv_tree, 4),
        "cv_boost": round(cv_boost, 4) if cv_boost is not None else None,
        "trainer": trainer,
        "positive_rate": round(float(y.mean()), 4),
        "tree": export_sklearn_tree(deploy_model, cols),
    }


def train_context_classifiers(
    df: pd.DataFrame,
    feature_cols: list[str],
    label_col: str = "should_stop",
    min_samples_context: int = 60,
    min_samples_fallback: int = 100,
    use_boosting: bool = True,
) -> list[dict]:
    """
    Step 3: train most-specific context first (domain × algorithm × instance),
    then fall back to domain × algorithm, then domain-wide.
    """
    work = df.copy()
    if "run_id" not in work.columns:
        work["run_id"] = work.index.astype(str)
    if "instance" not in work.columns:
        work["instance"] = work["run_id"].apply(infer_instance_from_run_id)
    if "algorithm" not in work.columns:
        work["algorithm"] = "sa"
    work["algorithm"] = work["algorithm"].astype(str)
    work["instance"] = work["instance"].astype(str)

    classifiers: list[dict] = []
    covered: set[tuple[str, str, str]] = set()

    for (domain, algorithm, instance), group in work.groupby(["domain", "algorithm", "instance"]):
        if not instance or len(group) < min_samples_context:
            continue
        clf = train_domain_classifier(
            work,
            str(domain),
            feature_cols,
            label_col=label_col,
            min_samples=min_samples_context,
            use_boosting=use_boosting,
            algorithm=str(algorithm),
            instance=str(instance),
        )
        if clf and clf["cv_mean"] >= 0.5:
            classifiers.append(clf)
            covered.add((str(domain), str(algorithm), str(instance)))

    for (domain, algorithm), group in work.groupby(["domain", "algorithm"]):
        key = (str(domain), str(algorithm), "")
        if any(k[0] == str(domain) and k[1] == str(algorithm) for k in covered):
            continue
        if len(group) < min_samples_fallback:
            continue
        clf = train_domain_classifier(
            work,
            str(domain),
            feature_cols,
            label_col=label_col,
            min_samples=min_samples_fallback,
            use_boosting=use_boosting,
            algorithm=str(algorithm),
            instance="",
        )
        if clf and clf["cv_mean"] >= 0.5:
            classifiers.append(clf)
            covered.add(key)

    for domain in sorted(work["domain"].unique()):
        if any(k[0] == str(domain) for k in covered):
            continue
        group = work[work["domain"] == domain]
        if len(group) < min_samples_fallback:
            continue
        clf = train_domain_classifier(
            work,
            str(domain),
            feature_cols,
            label_col=label_col,
            min_samples=min_samples_fallback,
            use_boosting=use_boosting,
            algorithm="*",
            instance="",
        )
        if clf and clf["cv_mean"] >= 0.5:
            classifiers.append(clf)

    return classifiers


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
