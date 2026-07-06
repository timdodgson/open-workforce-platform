# Worker Value Model

Offline decision-tree model trained on `worker_learning.csv` data.

## Purpose

Predicts whether a worker will be useful (improve the solution or find a global best) based on spawn-time features. This is a learning system that observes historical worker outcomes and builds a predictive model — it does **not** integrate with the live optimiser.

## Outputs

- `worker_model.json` — trained model (serialised as interpretable tree structure)
- Training metrics: accuracy, precision, recall, F1, ROC-AUC, feature importance, confusion matrix

## Usage

```bash
cd platform/ml
pip install -e .
train --data-dir <path-to-runs> --output worker_model.json
```

Or point at a single CSV:

```bash
train --csv path/to/worker_learning.csv --output worker_model.json
```

## Design Decisions

- **Decision Tree**: chosen for interpretability. The dashboard can render the tree logic and feature importance directly.
- **Offline only**: no live integration. The model trains on historical data and produces a static JSON.
- **No deep learning**: the dataset is small and tabular. A decision tree is the simplest model that can capture non-linear feature interactions.
- **Scikit-learn**: standard, well-maintained, minimal transitive dependencies.
