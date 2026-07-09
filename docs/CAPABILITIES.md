# Capability Matrix — Search Intelligence Parity

This document tracks Search Intelligence (SI) capability parity across optimisation domains.
Every labelled run should emit the same eight telemetry CSVs (data rows when SI is active, header-only stubs otherwise).

**Dashboard:** live product matrix at `/capabilities` (runtime vs telemetry vs UX). Source of truth for UI: `platform/web/pfrs-lab/src/lib/capability-matrix.ts`.

## Three layers (read this before the matrix)

| Layer | Question | Example |
|-------|----------|---------|
| **Runtime native** | Does the domain use this assist hook during search? | NRP beam WorkerAssist; CVRP SearchAssist |
| **Telemetry CSV** | Do labelled runs emit the contract file (native or adapted)? | CVRP `worker_assist.csv` adapted from search checkpoints |
| **Dashboard UX** | Can a reviewer see domain-specific results in the UI? | Gantt for JSS; Schedule for NRP |

A cell can be ✅ at telemetry and — at runtime native when adapters bridge orthogonal views. That is intentional, not a bug.

## Product status summary (2026-07)

| Area | NRP | CVRP | VRPTW | JSS |
|------|-----|------|-------|-----|
| Solvers (SA/LAHC/Tabu/Portfolio/Adaptive) | ✅ | ✅ | ✅ | ✅ |
| SI modes + policies (CLI) | ✅ | ✅ | ✅ | ✅ |
| Offline validation (promotion gates) | ✅ | ✅ | ✅ | ✅ |
| ILP benchmark (`owp benchmark-*-ilp`) | ✅ | ✅ | ⚠️ BKS only | ⚠️ BKS only |
| Native WorkerAssist | ✅ | — | — | — |
| Native SearchAssist | ⚠️ adapter (real budgets) | ✅ | ✅ | ✅ |
| NRP policy CSV hooks in Go | ✅ worker-mapped | ✅ | ✅ | ✅ |

## Gap roadmap

| Phase | Scope | Effort |
|-------|--------|--------|
| 0 | Document runtime vs telemetry vs UX (this file + `/capabilities`) | ½ day |
| 1 | Capabilities page on dashboard | 1 day |
| 2 | VRPTW route viewer + feasibility summaries | 2–3 days |
| 3 | NRP SearchAssist fidelity + Go policy hooks | 2–4 days |
| 4 | VRPTW ILP benchmark | 1–2 weeks |
| 5 | JSS ILP benchmark | 1–2 weeks |
| 6 | Domain constraint viewer pages | 3–5 days |

**Not planned:** native WorkerAssist on CVRP/JSS/VRPTW — those domains use SearchAssist; adapted worker CSVs provide training parity.

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
| Policy decisions | Yes (worker-mapped) | Yes | Yes | Yes |
| Policy evaluation | Yes (worker-mapped) | Yes | Yes | Yes |
| Counterfactual learning | Yes (with policy mode) | Yes (with policy mode) | Yes | Yes |
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

Promotion gate: `MinLearnedPolicyAgreement = 0.80` outcome accuracy with `regret_vs_rules <= 0` (see `policy_validation.py`, `policy_registry.py`). Rule agreement is diagnostic only.

Validate and merge after training:

```bash
python platform/ml/train_policies.py --data-dir <runs> --output-dir platform/ml/policies
python platform/ml/validate_policies.py --data-dir <runs> --policy-dir platform/ml/policies --merge-registry
```

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
