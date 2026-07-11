# PFRS Lab Dashboard

Next.js research dashboard for the Open Workforce Platform. Reads optimisation run telemetry from local `data/runs/` or S3.

## Development

```bash
cd platform/web/pfrs-lab
npm install
npm run dev
```

Open [http://localhost:3000](http://localhost:3000). Runs are loaded from `data/runs/<run-label>/`.

Set `STORAGE_PROVIDER=local` (default) or `STORAGE_PROVIDER=s3` with `PFRS_S3_BUCKET` and AWS credentials for production data.

## Data layout

```
data/
├── runs/<run-label>/
│   ├── run.json
│   ├── worker_learning.csv
│   ├── generic_search_assist.csv      # CVRP/JSS/VRPTW search assist
│   ├── portfolio_assist.csv           # portfolio mode
│   ├── policy_decisions.csv           # SI 2.0 (--policy-mode)
│   ├── policy_evaluation.csv          # SI 2.0 policy quality
│   └── policy_learning_report.json    # post-run learning recommendation
├── intelligence_summary.json          # precomputed /intelligence overview
├── intelligence_learning.json         # precomputed learning + decisions
├── policy_dashboard.json              # precomputed policy/SI tabs
├── worker_model.json                  # ML model (Model tab)
├── worker_predictions.json            # ML predictions (Predictions tab)
└── policy_registry.json               # policy lifecycle (Policies tab)
```

Generate runs from `platform/go` (see root README). Sync ML artifacts after training:

```bash
cd platform/ml
python train_policies.py --data-dir ../web/pfrs-lab/data/runs --output-dir policies
cp policies/policy_registry.json ../web/pfrs-lab/data/

# Rebuild intelligence artifacts (local or S3 via STORAGE_PROVIDER)
cd ../web/pfrs-lab
npm run rebuild-intelligence
```

Or use **Admin → Rebuild intelligence artifacts** in the dashboard UI.

## Search Intelligence UI

Route: `/intelligence`

| Tab | Data source |
|-----|-------------|
| Overview | Static architecture + validation summary |
| Learning | `worker_learning.csv` across runs |
| Model | `data/worker_model.json` |
| Predictions | `data/worker_predictions.json` |
| Decision Analysis | `worker_decisions.csv` + `worker_assist.csv` |
| What-If Lab | Predictions simulation |
| Assist Validation | assist CSVs (worker/search/portfolio) |
| Policies | `policy_decisions.csv`, `policy_evaluation.csv`, registry |
| SI Validation | `si2-*` run label progress |

When `intelligence_*.json` / `policy_dashboard.json` exist, tabs load from cached artifacts (faster). Admin shows staleness when run count changes.

## Deploy

Production uses S3 storage (`STORAGE_PROVIDER=s3`). Upload runs with `--storage s3` on `owp solve` / `owp tune-pfrs` commands.

After retraining policies or uploading new runs to S3, rebuild artifacts:

```bash
STORAGE_PROVIDER=s3 PFRS_S3_BUCKET=pfrs-research-lab-data npm run rebuild-intelligence
```
