"""
Worker Value Model — Batch Prediction.

Applies the trained model to worker_learning.csv records and outputs
worker_predictions.json with per-worker predictions, decision paths,
and feature contributions.

This is offline analysis only — does not change optimiser behaviour.
"""

import argparse
import json
import sys
from pathlib import Path

import numpy as np
import pandas as pd
from sklearn.tree import DecisionTreeClassifier, DecisionTreeRegressor

from .train import (
    SPAWN_FEATURES,
    TARGET_GLOBAL_BEST,
    TARGET_IMPROVED,
    TARGET_IMPROVEMENT,
    TARGET_ROI,
    load_data,
    load_data_from_runs,
    prepare_features,
)


def retrain_models(df: pd.DataFrame) -> dict:
    """Retrain all models on the full dataset (for prediction, not evaluation)."""
    X = prepare_features(df)

    y_improved = df[TARGET_IMPROVED].astype(int)
    y_global_best = df[TARGET_GLOBAL_BEST].astype(int)
    y_improvement = df[TARGET_IMPROVEMENT].fillna(0).astype(float)
    y_roi = df[TARGET_ROI].fillna(0).astype(float)

    improved_clf = DecisionTreeClassifier(
        max_depth=8, min_samples_leaf=5, min_samples_split=10,
        random_state=42, class_weight="balanced",
    )
    improved_clf.fit(X, y_improved)

    global_best_clf = DecisionTreeClassifier(
        max_depth=8, min_samples_leaf=5, min_samples_split=10,
        random_state=42, class_weight="balanced",
    )
    global_best_clf.fit(X, y_global_best)

    improvement_reg = DecisionTreeRegressor(
        max_depth=8, min_samples_leaf=5, min_samples_split=10, random_state=42,
    )
    improvement_reg.fit(X, y_improvement)

    roi_reg = DecisionTreeRegressor(
        max_depth=8, min_samples_leaf=5, min_samples_split=10, random_state=42,
    )
    roi_reg.fit(X, y_roi)

    return {
        "improved": improved_clf,
        "produced_global_best": global_best_clf,
        "improvement_amount": improvement_reg,
        "roi": roi_reg,
    }


def get_decision_path_text(tree, feature_names: list, sample: np.ndarray) -> list:
    """Extract the decision path as a list of human-readable conditions."""
    node_indicator = tree.decision_path(sample.reshape(1, -1))
    node_ids = node_indicator.indices

    path_steps = []
    for node_id in node_ids:
        if tree.tree_.children_left[node_id] == tree.tree_.children_right[node_id]:
            # Leaf node.
            continue
        feature_idx = tree.tree_.feature[node_id]
        threshold = tree.tree_.threshold[node_id]
        feature_name = feature_names[feature_idx]
        feature_value = sample[feature_idx]

        if feature_value <= threshold:
            path_steps.append({
                "feature": feature_name,
                "condition": "<=",
                "threshold": round(float(threshold), 2),
                "value": round(float(feature_value), 4),
            })
        else:
            path_steps.append({
                "feature": feature_name,
                "condition": ">",
                "threshold": round(float(threshold), 2),
                "value": round(float(feature_value), 4),
            })

    return path_steps


def compute_feature_contributions(tree, feature_names: list, sample: np.ndarray) -> dict:
    """
    Compute approximate feature contributions using the decision path.
    For each split the worker passes through, the feature used gets credit
    proportional to how much it narrows the prediction.
    """
    node_indicator = tree.decision_path(sample.reshape(1, -1))
    node_ids = node_indicator.indices

    contributions = {f: 0.0 for f in feature_names}

    for i in range(len(node_ids) - 1):
        node_id = node_ids[i]
        if tree.tree_.children_left[node_id] == tree.tree_.children_right[node_id]:
            continue
        feature_idx = tree.tree_.feature[node_id]
        feature_name = feature_names[feature_idx]
        # Weight by depth (earlier splits matter more).
        weight = 1.0 / (i + 1)
        contributions[feature_name] += weight

    # Normalise.
    total = sum(contributions.values())
    if total > 0:
        contributions = {k: v / total for k, v in contributions.items()}

    return contributions


def generate_explanation(
    contributions: dict,
    feature_values: dict,
    prediction_label: str,
    predicted_value: float,
) -> str:
    """Generate a natural-language explanation for a prediction."""
    # Get top contributing features (those with > 5% contribution).
    top = sorted(contributions.items(), key=lambda x: x[1], reverse=True)
    significant = [(f, c) for f, c in top if c > 0.05][:4]

    if not significant:
        return f"The model predicted {prediction_label} based on the overall distribution of training data."

    # Feature-specific explanations.
    EXPLANATIONS = {
        "distance_from_best": lambda v: "parent gap was small" if v <= 500 else "parent gap was large",
        "parent_objective": lambda v: f"parent objective was {int(v)}",
        "global_best": lambda v: f"global best was {int(v)}",
        "depth": lambda v: f"worker depth was {'low' if v <= 5 else 'moderate' if v <= 20 else 'high'}",
        "week": lambda v: f"this was week {int(v)}",
        "beam_rank": lambda v: f"beam rank was {int(v)}",
        "beam_health": lambda v: f"beam health was {'high' if v >= 0.7 else 'moderate' if v >= 0.4 else 'low'}",
        "entropy": lambda v: f"entropy was {'high' if v >= 2.0 else 'moderate' if v >= 1.0 else 'low'}",
        "temperature": lambda v: f"temperature was {'high' if v >= 5.0 else 'moderate' if v >= 1.0 else 'low'}",
        "iterations_alloc": lambda v: f"iteration budget was {int(v):,}",
        "worker_count": lambda v: f"worker count was {int(v)}",
        "active_families": lambda v: f"there were {int(v)} active families",
        "plateau_length": lambda v: f"plateau length was {int(v)}",
        "recent_improv_rate": lambda v: f"recent improvement rate was {'high' if v > 1.0 else 'low'}",
        "diversity": lambda v: f"diversity was {'high' if v >= 0.5 else 'low'}",
        "beam_score": lambda v: f"beam score was {int(v)}",
    }

    reasons = []
    for feature, _ in significant:
        value = feature_values.get(feature, 0)
        if feature in EXPLANATIONS:
            reasons.append(EXPLANATIONS[feature](value))
        else:
            reasons.append(f"{feature} was {value}")

    intro = f"This worker was predicted to have {prediction_label} because:"
    bullets = "\n".join(f"  • {r}" for r in reasons)
    return f"{intro}\n{bullets}"


def predict_all(df: pd.DataFrame) -> list:
    """Generate predictions for all workers in the dataset."""
    print("  Retraining models on full dataset...")
    models = retrain_models(df)

    X = prepare_features(df)
    feature_names = X.columns.tolist()
    X_array = X.values

    predictions = []
    total = len(df)

    for i in range(total):
        if i % 100 == 0:
            print(f"  Predicting: {i}/{total}...", end="\r")

        row = df.iloc[i]
        sample = X_array[i]

        # Predictions.
        improved_proba = models["improved"].predict_proba(sample.reshape(1, -1))[0]
        p_improved = float(improved_proba[1]) if len(improved_proba) > 1 else 1.0

        gb_proba = models["produced_global_best"].predict_proba(sample.reshape(1, -1))[0]
        p_global_best = float(gb_proba[1]) if len(gb_proba) > 1 else 0.0

        expected_improvement = float(models["improvement_amount"].predict(sample.reshape(1, -1))[0])
        expected_roi = float(models["roi"].predict(sample.reshape(1, -1))[0])

        # Actuals.
        actual_improved = bool(row.get("improved", False))
        actual_global_best = bool(row.get("produced_global_best", False))
        actual_improvement = float(row.get("improvement_amount", 0))
        actual_roi = float(row.get("roi", 0))

        # Decision path for global_best model (most interesting).
        decision_path = get_decision_path_text(
            models["produced_global_best"], feature_names, sample
        )

        # Feature contributions for global_best model.
        contributions = compute_feature_contributions(
            models["produced_global_best"], feature_names, sample
        )

        # Feature values for this worker.
        feature_values = {f: float(sample[j]) for j, f in enumerate(feature_names)}

        # Explanation.
        explanation = generate_explanation(
            contributions, feature_values, "a high probability of improvement", p_improved
        )

        # Prediction error.
        improvement_error = expected_improvement - actual_improvement
        roi_error = expected_roi - actual_roi

        predictions.append({
            # Identity.
            "index": i,
            "run_id": str(row.get("_source", row.get("runId", ""))),
            "problem_type": str(row.get("problem_type", "")),
            "instance": str(row.get("instance", "")),
            "algorithm": str(row.get("algorithm", "")),
            "seed": int(row.get("run_seed", 0)),
            "week": int(row.get("week", 0)),
            "depth": int(row.get("depth", 0)),
            # Actual outcomes.
            "actual": {
                "improved": actual_improved,
                "produced_global_best": actual_global_best,
                "improvement_amount": round(actual_improvement, 2),
                "roi": round(actual_roi, 6),
            },
            # Predictions.
            "predicted": {
                "p_improved": round(p_improved, 4),
                "p_global_best": round(p_global_best, 4),
                "expected_improvement": round(expected_improvement, 2),
                "expected_roi": round(expected_roi, 6),
            },
            # Errors.
            "error": {
                "improvement": round(improvement_error, 2),
                "roi": round(roi_error, 6),
            },
            # Explanation.
            "decision_path": decision_path,
            "feature_contributions": {
                k: round(v, 4) for k, v in contributions.items() if v > 0.01
            },
            "feature_values": {k: round(v, 4) for k, v in feature_values.items()},
            "explanation": explanation,
        })

    print(f"  Predicting: {total}/{total} — done.")
    return predictions


def write_prediction_shards(
    predictions: list,
    data_dir: Path,
    *,
    upload_s3: bool = False,
    s3_bucket: str = "pfrs-research-lab-data",
    s3_region: str = "eu-west-1",
) -> Path:
    """Write per-run worker_predictions.json shards + root index for the dashboard API."""
    runs_dir = data_dir / "runs" if (data_dir / "runs").is_dir() else data_dir
    by_run: dict[str, list] = {}
    for pred in predictions:
        run_id = str(pred.get("run_id") or "unknown")
        by_run.setdefault(run_id, []).append(pred)

    index = {
        "version": "0.2.0",
        "total_predictions": len(predictions),
        "runs": [],
    }

    for run_id, rows in sorted(by_run.items()):
        run_path = runs_dir / run_id
        run_path.mkdir(parents=True, exist_ok=True)
        shard_path = run_path / "worker_predictions.json"
        with open(shard_path, "w") as f:
            json.dump(rows, f, indent=2)
        index["runs"].append({"runId": run_id, "count": len(rows)})

    index_path = data_dir / "worker_predictions_index.json"
    with open(index_path, "w") as f:
        json.dump(index, f, indent=2)

    if upload_s3:
        from .train import _upload_to_s3
        _upload_to_s3(index_path, index_path.name, s3_bucket, s3_region)
        for run_id in by_run:
            shard = runs_dir / run_id / "worker_predictions.json"
            key = f"runs/{run_id}/worker_predictions.json"
            _upload_to_s3(shard, key, s3_bucket, s3_region)
        print(f"  Uploaded index + {len(by_run)} run shards to S3")

    return index_path


def main():
    parser = argparse.ArgumentParser(
        description="Generate per-worker predictions from Worker Value Model"
    )
    parser.add_argument(
        "--csv", type=Path, help="Path to a single worker_learning.csv file"
    )
    parser.add_argument(
        "--data-dir", type=Path, help="Path to data directory containing runs/"
    )
    parser.add_argument(
        "--output", type=Path, default=Path("worker_predictions.json"),
        help="Output path (default: worker_predictions.json)",
    )
    parser.add_argument(
        "--storage",
        choices=["local", "s3"],
        default="local",
        help="Upload output to S3 after generation (default: local only)",
    )
    parser.add_argument(
        "--s3-bucket",
        default="pfrs-research-lab-data",
        help="S3 bucket for upload (default: pfrs-research-lab-data)",
    )
    parser.add_argument(
        "--s3-region",
        default="eu-west-1",
        help="AWS region (default: eu-west-1)",
    )
    args = parser.parse_args()

    if not args.csv and not args.data_dir:
        parser.error("Provide either --csv or --data-dir")

    print("Worker Prediction Explorer — Generating Predictions")
    print("=" * 50)

    if args.csv:
        print(f"  Loading: {args.csv}")
        df = load_data(args.csv)
    else:
        print(f"  Loading from: {args.data_dir}")
        df = load_data_from_runs(args.data_dir)

    if df.empty:
        print("  ERROR: No data found.", file=sys.stderr)
        sys.exit(1)

    print(f"  Records: {len(df)}")
    print()

    predictions = predict_all(df)

    output = {
        "version": "0.2.0",
        "total_predictions": len(predictions),
        "predictions": predictions,
    }

    out_dir = args.data_dir if args.data_dir else args.output.parent
    out_dir.mkdir(parents=True, exist_ok=True)

    args.output.parent.mkdir(parents=True, exist_ok=True)
    with open(args.output, "w") as f:
        json.dump(output, f, indent=2)

    index_path = write_prediction_shards(
        predictions,
        out_dir,
        upload_s3=args.storage == "s3",
        s3_bucket=args.s3_bucket,
        s3_region=args.s3_region,
    )
    with open(index_path) as f:
        index = json.load(f)

    print()
    print("=" * 50)
    print(f"  Full export: {args.output} ({args.output.stat().st_size / 1024:.1f} KB)")
    print(f"  Dashboard index: {index_path} ({len(index['runs'])} runs)")

    if args.storage == "s3":
        from .train import _upload_to_s3
        _upload_to_s3(args.output, args.output.name, args.s3_bucket, args.s3_region)

    print()
    print("Done.")


if __name__ == "__main__":
    main()
