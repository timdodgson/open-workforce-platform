# PFRS Lab — R&D Runbook

Operational guide for researchers and engineers running experiments, validating Search Intelligence, and keeping the public lab in sync.

**Audience:** R&D, internal reviewers, future contributors.  
**Related:** [ML_JOURNEY.md](./ML_JOURNEY.md) · [CITATION.md](./CITATION.md) · [EXPERIMENTS.md](./EXPERIMENTS.md) · [Experiment Matrix](https://pfrs-lab.com/experiment-matrix) · [SEARCH_INTELLIGENCE_V2.md](./SEARCH_INTELLIGENCE_V2.md) · [benchmark-suite.md](./06-engineering/benchmark-suite.md)

---

## 1. Repository map

| Path | Role |
|------|------|
| `platform/go/cmd/owp` | CLI — solve, tune-pfrs, validate-si2 |
| `platform/go/scripts/validate-si2.ps1` | Full SI2 fast matrix (240 runs) |
| `platform/go/scripts/validate-si2-deep.ps1` | SI2 deep matrix (48 runs) |
| `platform/go/scripts/validate-si2-quick.ps1` | 6-run smoke before overnight jobs |
| `platform/ml/policies/` | Learned policy JSON (versioned) |
| `platform/web/pfrs-lab/` | Next.js lab UI + artifact rebuild |
| `platform/web/pfrs-lab/src/lib/experiment-matrix.ts` | Canonical 288-run catalog |
| `docs/reports/val-gap-audit.md` | val-* coverage audit (regenerate after uploads) |

---

## 2. Environment setup

### Go solver (required)

```powershell
cd platform/go
go test ./... -short
```

Use `go run ./cmd/owp` or build `owp.exe` locally. Scripts prefer `owp.exe` when present.

### Dashboard (optional, for local inspection)

```powershell
cd platform/web/pfrs-lab
npm install
npm run dev
```

| Variable | Values | Purpose |
|----------|--------|---------|
| `STORAGE_PROVIDER` | `local` (default) · `s3` | Where runs are read/written |
| `PFRS_S3_BUCKET` | `pfrs-research-lab-data` | Production bucket |
| `AWS_REGION` | `eu-west-1` | S3 region |

Local runs land in `platform/web/pfrs-lab/data/runs/<run-label>/`.

---

## 3. Run a single experiment

### Pattern

1. Pick domain, instance, algorithm, and label (see [Experiment Matrix](https://pfrs-lab.com/experiment-matrix)).
2. Run from `platform/go` with `--run-label` (or `--pfrs-run-label` for NRP).
3. Upload with `--storage s3` for production visibility, or `--pfrs-storage local` for local dev.
4. Verify in lab: `/runs`, domain viewers, `/intelligence`.

### Canonical SI2 cell (documented as EXP-008)

```powershell
cd platform/go
go run ./cmd/owp solve cvrp `
  --instance ../../examples/cvrp/A-n32-k5.vrp `
  --mode sa --iterations 500000 `
  --policy-mode hybrid --policy-dir ../ml/policies `
  --seed 42 --run-label val-cvrp-a32k5-sa-hybrid-s42 `
  --storage s3
```

**Expect:** `run.json`, `policy_decisions.csv`, `policy_evaluation.csv`, `policy_learning_report.json`.  
**Inspect:** `/runs/val-cvrp-a32k5-sa-hybrid-s42/summary`, `/intelligence` Policies tab.

Full write-up: [EXPERIMENTS.md § EXP-008](./EXPERIMENTS.md).

### Benchmark ladder (no SI policy sweep)

Pure algorithm comparison — `--worker-decision-mode off`, no `--policy-mode`.  
Commands: [benchmark-suite.md](./06-engineering/benchmark-suite.md).

---

## 4. Validation matrices

### Fast tier — 240 runs

| Item | Value |
|------|-------|
| Script | `platform/go/scripts/validate-si2.ps1` |
| Configs | 8 (2 per domain × 4 domains) |
| Policies | `rules`, `hybrid`, `learned` |
| Seeds | 10 |
| Labels | `val-<domain>-<instance>-<mode>-<policy>-s<seed>` |
| Runtime | ~overnight |

```powershell
cd platform/go
powershell -ExecutionPolicy Bypass -File .\scripts\validate-si2.ps1
```

Logs: `platform/go/scripts/regression-logs/validate-si2-*.log`

### Deep tier — 48 runs

| Item | Value |
|------|-------|
| Script | `platform/go/scripts/validate-si2-deep.ps1` |
| Seeds | 2 (42, 123) |
| Labels | `val-deep-*` |
| Purpose | Richer checkpoints for policy training |

### Smoke before overnight

```powershell
.\scripts\validate-si2-quick.ps1
```

### CLI plan / analyze

```powershell
go run ./cmd/owp validate-si2 plan
go run ./cmd/owp validate-si2 analyze --prefix val- --runs-dir ../web/pfrs-lab/data/runs
```

`analyze` reads local `data/runs/` — sync from S3 or run locally first.

---

## 5. val-* gap audit

After uploading runs, confirm matrix coverage:

```powershell
cd platform/web/pfrs-lab
$env:STORAGE_PROVIDER='s3'
$env:PFRS_S3_BUCKET='pfrs-research-lab-data'
npm run audit-val-matrix
npm run compare-policy-modes
```

`compare-policy-modes` writes `docs/reports/ml-harness/latest.json` (Step 1 ML journey harness).

Or against local data (default `STORAGE_PROVIDER=local`):

```powershell
npm run audit-val-matrix
```

**Live UI:** `/experiment-matrix` shows per-config found/expected counts from the same catalog.

**Written audit:** [docs/reports/val-gap-audit.md](./reports/val-gap-audit.md) — snapshot + methodology.

---

## 6. Policy training and intelligence rebuild

### Retrain policies (after new val-* or deep runs)

```powershell
cd platform/ml
python train_policies.py --data-dir ../web/pfrs-lab/data/runs --output-dir policies
```

For S3-only data, sync runs locally first or point `--data-dir` at a synced copy.

### Rebuild dashboard artifacts

Precomputes `intelligence_summary.json`, `policy_dashboard.json`, etc.

```powershell
cd platform/web/pfrs-lab
# local
npm run rebuild-intelligence

# production bucket
$env:STORAGE_PROVIDER='s3'
$env:PFRS_S3_BUCKET='pfrs-research-lab-data'
npm run rebuild-intelligence
```

**UI:** Admin → **Rebuild intelligence artifacts** (shows staleness when run count drifts).

Rebuild after: new S3 uploads, policy retrain, or manifest changes.

---

## 7. Production sync checklist

| Step | Command / URL |
|------|----------------|
| Upload runs | `owp solve` / `tune-pfrs` with `--storage s3` |
| Audit matrix | `npm run audit-val-matrix` (S3 env) |
| Rebuild artifacts | `npm run rebuild-intelligence` (S3 env) |
| Verify lab | `/lab`, `/experiment-matrix`, `/intelligence` |
| Deploy site | `npm run sst:deploy` (if app code changed) |

---

## 8. Troubleshooting

| Symptom | Likely cause | Fix |
|---------|--------------|-----|
| Run missing in UI | Wrong storage provider or label typo | Match `STORAGE_PROVIDER` to where you uploaded; check `manifest.json` |
| Experiment matrix 0/N local | Only demo runs in `data/runs/` | Point dashboard at S3 or copy runs locally |
| Policies tab empty | Artifacts stale | `npm run rebuild-intelligence` |
| `validate-si2.ps1` failure | Missing policy dir | Ensure `platform/ml/policies/*.json` exist |
| Greedy mode shows wrong alg | Stale `owp.exe` | `go build ./cmd/owp` or use `go run` |
| Cypress structure failures | Site vs lab routes split | Home = `/`, lab sidebar = `/lab` |

---

## 9. What not to commit

- `platform/go/Highs.log`, `owp.exe`
- `platform/web/pfrs-lab/data/runs/` (large telemetry)
- `data/val-matrix-audit.json` (generated; re-run audit script)
- `platform/go/scripts/regression-logs/`

---

## 10. Release evidence chain

For R&D reviewers tracing a claim end-to-end:

1. **Hypothesis** — [EXPERIMENTS.md](./EXPERIMENTS.md) entry (e.g. EXP-008 single cell, EXP-007 full SI validation)
2. **Catalog** — [Experiment Matrix](https://pfrs-lab.com/experiment-matrix) (why each flag on/off)
3. **Coverage** — [val-gap-audit.md](./reports/val-gap-audit.md)
4. **Statistics** — [search-intelligence-statistical-validation.md](./reports/search-intelligence-statistical-validation.md)
5. **Live inspection** — `/runs/<label>`, `/intelligence`, `/statistics`
