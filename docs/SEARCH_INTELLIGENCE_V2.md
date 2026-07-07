# Search Intelligence 2.0 — Production Guide

SI 2.0 is **production-wired**. Learned JSON policies drive search checkpoints, portfolio budgets, and PFRS worker spawn decisions.

## Quick start

```bash
cd platform/go

# CVRP with hybrid policies (default policy dir: ../ml/policies)
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n32-k5.vrp \
  --mode sa --iterations 500000 --policy-mode hybrid \
  --seed 42 --run-label my-run --pfrs-storage local

# Portfolio + policies
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n32-k5.vrp \
  --mode portfolio --iterations 500000 --policy-mode hybrid \
  --seed 42 --run-label my-portfolio --pfrs-storage local

# PFRS workers (NRP)
go run ./cmd/owp tune-pfrs --instance n012w8 \
  --worker-decision-mode assist --policy-mode hybrid \
  --pfrs-iterations-per-worker 30000 --pfrs-max-total-workers 8 \
  --pfrs-run-label my-pfrs --pfrs-storage local
```

## Flags

| Flag | Values | Purpose |
|------|--------|---------|
| `--policy-mode` | `rules`, `hybrid`, `learned` | SI 2.0 learned policy behaviour |
| `--policy-dir` | path | Policy JSON directory (default `../ml/policies`) |
| `--worker-decision-mode` | `off`, `shadow`, `assist`, `adaptive` | Assist recording + safety (all domains) |

`--policy-mode` auto-enables shadow assist when `--worker-decision-mode` is unset.

## Architecture (production path)

| Layer | CLI | Go implementation | Policy files |
|-------|-----|-------------------|--------------|
| **SearchAssist** | `solve-*` + `--policy-mode` | `PolicySearchHookRunner` in `search.go` | `stagnation_policy.json`, `restart_policy.json` |
| **PortfolioAssist** | `--mode portfolio` + `--policy-mode` | `RunPortfolioWithAssist` → `AllocateBudgetsViaPolicy` | `budget_policy.json` |
| **WorkerAssist** | `tune-pfrs` + `--policy-mode` | `inrc2.HybridWorkerDecisionEngine` | `worker_policy.json` |

Post-run: `RunPostRunPolicyPipeline` writes `policy_learning_report.json` and updates learning state.

## Train policies

```bash
cd platform/ml
python train_policies.py --data-dir ../web/pfrs-lab/data/runs --output-dir policies
```

Outputs: `budget_policy.json`, `stagnation_policy.json`, `restart_policy.json`, `worker_policy.json`, `policy_registry.json`.

## Telemetry (per run with `--run-label`)

| File | When |
|------|------|
| `policy_decisions.csv` | `--policy-mode` set |
| `policy_evaluation.csv` | `--policy-mode` set |
| `policy_learning_report.json` | `--policy-mode` set |
| `generic_search_assist.csv` | search assist active |
| `portfolio_assist.csv` | portfolio + policy/assist |
| `worker_assist.csv` | PFRS assist mode |

Dashboard: **PFRS Lab → Search Intelligence** (`/intelligence`). Tabs: Policies, SI Validation, Assist Validation, etc.

## Testing

```bash
cd platform/go
go test ./... -short                    # unit + integration (uses ../ml/policies if present)
.\scripts\validate-si2-quick.ps1      # 6-run smoke (rules/hybrid/learned × SA/portfolio)
.\scripts\validate-si2.ps1              # full 240-run regression (overnight)
```

## What is NOT on the hot path

| Component | Status |
|-----------|--------|
| `PolicyHierarchy`, `PolicyProvider` | Library/registry helpers; tested, not required for solves |
| `CounterfactualRecorder` | Optional research CSV; not emitted by CLI today |
| SI v1-only `HybridExecutor` | **Removed** — duplicate of `PolicySearchHookRunner` |

## Acceptance checklist

- [ ] `policy_decisions.csv` contains `hybrid_learned` or `restart` rows
- [ ] Portfolio run writes `portfolio_assist.csv` with `learned_win_rate_*`
- [ ] PFRS run writes `worker_assist.csv` with `hybrid_learned` reason codes
- [ ] Dashboard Policies tab shows your runs
- [ ] `go test ./... -short` passes
