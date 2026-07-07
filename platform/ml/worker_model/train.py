"""
Worker Value Model — Decision Tree Training.

Trains a decision tree on worker_learning.csv to predict:
  1. Probability of improvement (binary classification)
  2. Probability of producing a global best (binary classification)
  3. Expected improvement amount (regression)
  4. Expected ROI (regression)

Outputs worker_model.json with the trained model structure and metrics.
"""

import argparse
import json
import sys
from pathlib import Path

import numpy as np
import pandas as pd
from sklearn.metrics import (
    accuracy_score,
    classification_report,
    confusion_matrix,
    f1_score,
    precision_score,
    recall_score,
    roc_auc_score,
)
from sklearn.model_selection import train_test_split
from sklearn.tree import DecisionTreeClassifier, DecisionTreeRegressor, export_text

# Features available at spawn time (before the worker runs).
# These are the only inputs the model can use for prediction.
SPAWN_FEATURES = [
    "week",
    "depth",
    "beam_rank",
    "beam_score",
    "entropy",
    "diversity",
    "beam_health",
    "temperature",
    "iterations_alloc",
    "global_best",
    "parent_objective",
    "distance_from_best",
    "plateau_length",
    "recent_improv_rate",
    "worker_count",
    "active_families",
]

# Targets we want to predict.
TARGET_IMPROVED = "improved"
TARGET_GLOBAL_BEST = "produced_global_best"
TARGET_IMPROVEMENT = "improvement_amount"
TARGET_ROI = "roi"


def load_data(csv_path: Path) -> pd.DataFrame:
    """Load a single worker_learning.csv."""
    df = pd.read_csv(csv_path)
    return df


def load_data_from_runs(data_dir: Path) -> pd.DataFrame:
    """Load worker_learning.csv from all run directories."""
    frames = []
    runs_dir = data_dir / "runs"
    if not runs_dir.exists():
        # Try data_dir directly for flat structures.
        runs_dir = data_dir

    for csv_path in runs_dir.rglob("worker_learning.csv"):
        try:
            df = pd.read_csv(csv_path)
            df["_source"] = str(csv_path.parent.name)
            frames.append(df)
        except Exception as e:
            print(f"  Warning: skipping {csv_path}: {e}", file=sys.stderr)

    if not frames:
        return pd.DataFrame()

    return pd.concat(frames, ignore_index=True)


def prepare_features(df: pd.DataFrame) -> pd.DataFrame:
    """Extract spawn-time features, handling missing columns gracefully."""
    available = [f for f in SPAWN_FEATURES if f in df.columns]
    X = df[available].copy()

    # Fill missing values with 0 (features not available in some run types).
    X = X.fillna(0)

    return X


def train_classifier(
    X_train: pd.DataFrame,
    y_train: pd.Series,
    X_test: pd.DataFrame,
    y_test: pd.Series,
    target_name: str,
    max_depth: int = 8,
) -> dict:
    """Train a decision tree classifier and return metrics + model structure."""
    clf = DecisionTreeClassifier(
        max_depth=max_depth,
        min_samples_leaf=5,
        min_samples_split=10,
        random_state=42,
        class_weight="balanced",  # Handle class imbalance (global bests are rare).
    )
    clf.fit(X_train, y_train)

    y_pred = clf.predict(X_test)
    y_proba = clf.predict_proba(X_test)

    # Metrics.
    accuracy = accuracy_score(y_test, y_pred)
    precision = precision_score(y_test, y_pred, zero_division=0)
    recall = recall_score(y_test, y_pred, zero_division=0)
    f1 = f1_score(y_test, y_pred, zero_division=0)

    # ROC-AUC (only if both classes present in test set).
    roc_auc = 0.0
    if len(np.unique(y_test)) > 1 and y_proba.shape[1] == 2:
        roc_auc = roc_auc_score(y_test, y_proba[:, 1])

    # Confusion matrix.
    cm = confusion_matrix(y_test, y_pred, labels=[0, 1])

    # Feature importance.
    feature_importance = dict(
        zip(X_train.columns.tolist(), clf.feature_importances_.tolist())
    )
    # Sort by importance descending.
    feature_importance = dict(
        sorted(feature_importance.items(), key=lambda x: x[1], reverse=True)
    )

    # Tree text representation (interpretable).
    tree_text = export_text(clf, feature_names=X_train.columns.tolist(), max_depth=6)

    # Classification report.
    report = classification_report(y_test, y_pred, output_dict=True, zero_division=0)

    return {
        "target": target_name,
        "type": "classifier",
        "max_depth": max_depth,
        "n_train": len(X_train),
        "n_test": len(X_test),
        "metrics": {
            "accuracy": round(accuracy, 4),
            "precision": round(precision, 4),
            "recall": round(recall, 4),
            "f1": round(f1, 4),
            "roc_auc": round(roc_auc, 4),
        },
        "confusion_matrix": {
            "tn": int(cm[0][0]),
            "fp": int(cm[0][1]),
            "fn": int(cm[1][0]),
            "tp": int(cm[1][1]),
        },
        "feature_importance": feature_importance,
        "classification_report": report,
        "tree_text": tree_text,
    }


def train_regressor(
    X_train: pd.DataFrame,
    y_train: pd.Series,
    X_test: pd.DataFrame,
    y_test: pd.Series,
    target_name: str,
    max_depth: int = 8,
) -> dict:
    """Train a decision tree regressor and return metrics + model structure."""
    reg = DecisionTreeRegressor(
        max_depth=max_depth,
        min_samples_leaf=5,
        min_samples_split=10,
        random_state=42,
    )
    reg.fit(X_train, y_train)

    y_pred = reg.predict(X_test)

    # Metrics.
    from sklearn.metrics import mean_absolute_error, mean_squared_error, r2_score

    mae = mean_absolute_error(y_test, y_pred)
    mse = mean_squared_error(y_test, y_pred)
    rmse = np.sqrt(mse)
    r2 = r2_score(y_test, y_pred)

    # Feature importance.
    feature_importance = dict(
        zip(X_train.columns.tolist(), reg.feature_importances_.tolist())
    )
    feature_importance = dict(
        sorted(feature_importance.items(), key=lambda x: x[1], reverse=True)
    )

    # Tree text representation.
    tree_text = export_text(reg, feature_names=X_train.columns.tolist(), max_depth=6)

    return {
        "target": target_name,
        "type": "regressor",
        "max_depth": max_depth,
        "n_train": len(X_train),
        "n_test": len(X_test),
        "metrics": {
            "mae": round(mae, 4),
            "mse": round(mse, 4),
            "rmse": round(rmse, 4),
            "r2": round(r2, 4),
        },
        "feature_importance": feature_importance,
        "tree_text": tree_text,
    }


def train_all(df: pd.DataFrame, test_size: float = 0.2) -> dict:
    """Train all four models and return the complete model output."""
    X = prepare_features(df)
    features_used = X.columns.tolist()

    # Encode targets.
    y_improved = df[TARGET_IMPROVED].astype(int)
    y_global_best = df[TARGET_GLOBAL_BEST].astype(int)
    y_improvement = df[TARGET_IMPROVEMENT].fillna(0).astype(float)
    y_roi = df[TARGET_ROI].fillna(0).astype(float)

    # Train/test split (stratified for classifiers where possible).
    X_train, X_test, idx_train, idx_test = train_test_split(
        X, df.index, test_size=test_size, random_state=42
    )

    print(f"  Training data: {len(X_train)} samples")
    print(f"  Test data: {len(X_test)} samples")
    print(f"  Features: {len(features_used)}")
    print(f"  Improvement rate: {y_improved.mean():.1%}")
    print(f"  Global best rate: {y_global_best.mean():.1%}")
    print()

    # 1. Improvement classifier.
    print("  Training: improved classifier...")
    improved_model = train_classifier(
        X_train, y_improved[idx_train], X_test, y_improved[idx_test], "improved"
    )
    print(f"    Accuracy: {improved_model['metrics']['accuracy']:.3f}")
    print(f"    F1: {improved_model['metrics']['f1']:.3f}")
    print(f"    ROC-AUC: {improved_model['metrics']['roc_auc']:.3f}")
    print()

    # 2. Global best classifier.
    print("  Training: produced_global_best classifier...")
    global_best_model = train_classifier(
        X_train,
        y_global_best[idx_train],
        X_test,
        y_global_best[idx_test],
        "produced_global_best",
    )
    print(f"    Accuracy: {global_best_model['metrics']['accuracy']:.3f}")
    print(f"    F1: {global_best_model['metrics']['f1']:.3f}")
    print(f"    ROC-AUC: {global_best_model['metrics']['roc_auc']:.3f}")
    print()

    # 3. Improvement amount regressor.
    print("  Training: improvement_amount regressor...")
    improvement_model = train_regressor(
        X_train,
        y_improvement[idx_train],
        X_test,
        y_improvement[idx_test],
        "improvement_amount",
    )
    print(f"    MAE: {improvement_model['metrics']['mae']:.2f}")
    print(f"    R²: {improvement_model['metrics']['r2']:.3f}")
    print()

    # 4. ROI regressor.
    print("  Training: roi regressor...")
    roi_model = train_regressor(
        X_train, y_roi[idx_train], X_test, y_roi[idx_test], "roi"
    )
    print(f"    MAE: {roi_model['metrics']['mae']:.4f}")
    print(f"    R²: {roi_model['metrics']['r2']:.3f}")
    print()

    # Summary.
    return {
        "version": "0.1.0",
        "description": "Worker Value Model — decision tree predictions for worker utility",
        "training_samples": len(X_train),
        "test_samples": len(X_test),
        "features": features_used,
        "models": {
            "improved": improved_model,
            "produced_global_best": global_best_model,
            "improvement_amount": improvement_model,
            "roi": roi_model,
        },
        "data_summary": {
            "total_records": len(df),
            "improvement_rate": round(y_improved.mean(), 4),
            "global_best_rate": round(y_global_best.mean(), 4),
            "mean_improvement": round(y_improvement.mean(), 2),
            "mean_roi": round(y_roi.mean(), 6),
        },
    }


def main():
    parser = argparse.ArgumentParser(
        description="Train Worker Value Model from worker_learning.csv data"
    )
    parser.add_argument(
        "--csv",
        type=Path,
        help="Path to a single worker_learning.csv file",
    )
    parser.add_argument(
        "--data-dir",
        type=Path,
        help="Path to data directory containing runs/ with worker_learning.csv files",
    )
    parser.add_argument(
        "--output",
        type=Path,
        default=Path("worker_model.json"),
        help="Output path for the trained model JSON (default: worker_model.json)",
    )
    parser.add_argument(
        "--test-size",
        type=float,
        default=0.2,
        help="Fraction of data to hold out for testing (default: 0.2)",
    )
    parser.add_argument(
        "--storage",
        choices=["local", "s3"],
        default="local",
        help="Upload output to S3 after training (default: local only)",
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

    print("Worker Value Model — Training")
    print("=" * 50)

    # Load data.
    if args.csv:
        print(f"  Loading: {args.csv}")
        df = load_data(args.csv)
    else:
        print(f"  Loading from: {args.data_dir}")
        df = load_data_from_runs(args.data_dir)

    if df.empty:
        print("  ERROR: No training data found.", file=sys.stderr)
        sys.exit(1)

    print(f"  Records loaded: {len(df)}")
    print()

    # Train.
    model_output = train_all(df, test_size=args.test_size)

    # Write output.
    args.output.parent.mkdir(parents=True, exist_ok=True)
    with open(args.output, "w") as f:
        json.dump(model_output, f, indent=2)

    print("=" * 50)
    print(f"  Model saved: {args.output}")
    print(f"  Size: {args.output.stat().st_size / 1024:.1f} KB")

    # Upload to S3 if requested.
    if args.storage == "s3":
        _upload_to_s3(args.output, args.output.name, args.s3_bucket, args.s3_region)

    print()
    print("Done.")


def _upload_to_s3(local_path: Path, s3_key: str, bucket: str, region: str):
    """Upload a file to the S3 bucket root."""
    try:
        import boto3
        s3 = boto3.client("s3", region_name=region)
        content_type = "application/json" if s3_key.endswith(".json") else "text/csv"
        s3.upload_file(str(local_path), bucket, s3_key, ExtraArgs={"ContentType": content_type})
        print(f"  Uploaded to s3://{bucket}/{s3_key}")
    except ImportError:
        print("  WARNING: boto3 not installed — skipping S3 upload", file=sys.stderr)
        print("  Install with: pip install boto3", file=sys.stderr)
    except Exception as e:
        print(f"  WARNING: S3 upload failed: {e}", file=sys.stderr)


if __name__ == "__main__":
    main()
