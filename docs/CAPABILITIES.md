# Capability Matrix — Search Intelligence Parity

This document tracks Search Intelligence (SI) capability parity across optimisation domains.
Every labelled run should emit the same eight telemetry CSVs (data rows when SI is active, header-only stubs otherwise).

## Telemetry contract (all domains)

| CSV | Purpose |
|-----|---------|
| `worker_learning.csv` | Per-worker spawn/outcome features for value models |
| `worker_decisions.csv` | Shadow-mode worker recommendations |
| `worker_assist.csv` | Assist/adaptive-mode worker interventions |
| `generic_search_assist.csv` | Search-level checkpoint recommendations |
| `portfolio_assist.csv` | Portfolio budget allocation decisions |
| `policy_decisions.csv` | SI 2.0 policy hook decisions (`--policy-mode`) |
| `policy_evaluation.csv` | Post-run policy accuracy vs outcome |
| `counterfactual_learning.csv` | Decision alternatives and regret |

Contract enforcement: `platform/go/internal/optimisation/si_telemetry_contract.go` via `EnsureSITelemetryContract`.

## CLI flags (all solve commands + tune-pfrs)

| Flag | Values | Layer |
|------|--------|-------|
| `--worker-decision-mode` | `off`, `shadow`, `assist`, `adaptive` | WorkerAssist / SearchAssist |
| `--policy-mode` | `rules`, `hybrid`, `learned` | SI 2.0 policy hooks |
| `--policy-dir` | path (default `../ml/policies`) | Learned policy JSON location |

## Current parity matrix

| Capability | NRP (`tune-pfrs`) | CVRP | JSS | VRPTW |
|------------|-------------------|------|-----|-------|
| Worker decisions (native) | Yes | — | — | — |
| Worker decisions (adapted) | — | Yes | Yes | Yes |
| Worker assist (native) | Yes | — | — | — |
| Worker assist (adapted) | — | Yes | Yes | Yes |
| Search assist (native) | — | Yes | Yes | Yes |
| Search assist (adapted) | Yes | — | — | — |
| Portfolio assist (native) | — | Yes | Yes | Yes |
| Portfolio assist (adapted) | Yes (portfolio mode) | — | — | — |
| Worker learning | Yes (beam + standard) | Yes (single-worker) | Yes | Yes |
| Policy decisions | Stub | Yes | Yes | Yes |
| Policy evaluation | Stub | Yes | Yes | Yes |
| Counterfactual learning | Stub | Yes (with policy mode) | Yes | Yes |
| Contract stubs (missing files) | Yes | Yes | Yes | Yes |

**Adapters** (`platform/go/cmd/owp/si_adapters.go`):

- CVRP/JSS/VRPTW → `worker_decisions.csv` / `worker_assist.csv` from `generic_search_assist.csv`
- NRP → `generic_search_assist.csv` from worker decision/assist recorders
- NRP portfolio → `portfolio_assist.csv` via `BuildNRPPortfolioAssistRecords`

## Policy training (Phase 3)

| Script | Scope |
|--------|-------|
| `platform/ml/train_policies.py` | Global training across all runs |
| `platform/ml/train_domain_policies.py` | Per-domain training with 80% promotion gate |

Promotion gate: `MinLearnedPolicyAgreement = 0.80` in Go (`si_telemetry_contract.go`, `policy_promotion.go`) and `MIN_LEARNED_POLICY_AGREEMENT` in Python.

## Target state

Each domain exposes the full SI surface area:

1. All eight CSVs present on every labelled run
2. Both CLI flag families accepted and validated
3. Cross-layer adapters populate the orthogonal CSV views
4. Domain-scoped policy training with shared promotion gates

## Key implementation files

| Area | Path |
|------|------|
| Telemetry hub | `platform/go/cmd/owp/cli_telemetry.go` |
| Adapters | `platform/go/cmd/owp/si_adapters.go` |
| SI contract | `platform/go/internal/optimisation/si_telemetry_contract.go` |
| Counterfactual emit | `platform/go/internal/optimisation/counterfactual_emit.go` |
| PFRS command | `platform/go/cmd/owp/command_tune_pfrs.go` |
| Generic solvers | `command_solve_cvrp.go`, `command_solve_jobshop.go`, `command_solve_vrptw.go` |
| Flag validation | `platform/go/cmd/owp/cli_runtime.go` |
