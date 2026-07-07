# Search Intelligence 2.0 — Policy Training Pipeline

Trains learned policies from historical telemetry data.

## Setup

```bash
cd platform/ml
python -m venv .venv
.venv\Scripts\activate   # Windows
pip install -r requirements.txt
```

## Train All Policies

```bash
python train_policies.py --data-dir ../web/pfrs-lab/data/runs --output-dir policies
```

## Output

```
policies/
├── budget_policy.json          # Portfolio budget allocation
├── stagnation_policy.json      # Early stop / stagnation detection
├── restart_policy.json         # Restart timing and algorithm selection
├── worker_policy.json          # Worker value prediction (NRP)
├── policy_v1.json              # Combined policy registry
└── training_report.json        # Full training metrics
```

## What It Does

1. **Scans** all run directories for telemetry CSVs
2. **Engineers features** from raw data (budget consumed, plateau ratio, improvement rate, etc.)
3. **Trains** decision tree models per domain per policy type
4. **Cross-validates** with 5-fold stratified CV
5. **Computes** accuracy, feature importance, calibration
6. **Exports** JSON models in the format Go code expects
7. **Generates** training report with all metrics

## No Fabrication

All model weights come from real telemetry data. If insufficient data exists for a domain/policy combination, that policy is marked as "insufficient_data" and not trained.
