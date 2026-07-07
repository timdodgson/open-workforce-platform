# Package Boundaries — Sprint 4 Review

Review date: after Search Intelligence v2 CLI refactor (Sprints 1–3).

## Intended layers

| Package | Role |
|---------|------|
| `cmd/owp` | CLI parsing, run orchestration, output finalization |
| `internal/optimisation` | Domain-agnostic search engine + Search Intelligence abstractions |
| `internal/infrastructure/inrc2` | NRP/PFRS domain solver + PFRS-specific telemetry |
| `internal/infrastructure/cvrp` | CVRP domain model + solver adapters |
| `internal/infrastructure/vrptw` | VRPTW domain model + solver adapters |
| `internal/infrastructure/jobshop` | Job shop domain model |
| `platform/ml` | Python training; Go loads JSON models only |

## Sprint 4 changes (low-risk moves)

1. **Portfolio budget heuristics** — shared constants + `portfolioBudgetHeuristic` in `optimisation/portfolio_budget_rules.go`, used by v1 (`RuleBasedPortfolioAdvisor`) and v2 (`NewPortfolioBudgetRulePolicy`).
2. **CVRP telemetry CSV** — moved from `cmd/owp/cli_telemetry.go` to `cvrp/telemetry_csv.go` (`BuildResultsCSV`, `BuildDiscoveriesCSV`).
3. **VRPTW telemetry CSV** — moved to `vrptw/telemetry_csv.go` (`BuildDiscoveriesCSV`).
4. **PFRS worker intelligence wiring** — `wirePFRSWorkerIntelligence` in `cmd/owp/cli_runtime.go` deduplicates tune-pfrs setup.
5. **Policy registry clarity** — documented that `LoadPolicyRegistryFile` (policy_v1.json) and `LoadPolicyRegistry` (policy_registry.json) serve different schemas.

## Current package health

### `internal/optimisation` (~70 files)

**Correctly placed:** core metaheuristics, `search.go`, SI v1 assist (`search_assist_hooks.go`, `portfolio_assist.go`), SI v2 policy scaffold (`policy_*.go`, `hybrid_executor.go`).

**Tension:** single flat package holds algorithms, SI v1 production paths, and SI v2 scaffold code. SI 2.0 types (`PolicySearchHookRunner`, `HybridExecutor`) are tested but not wired into `search.go` hot paths.

**Parallel systems:**
- v1: `SearchHookRunner`, `RuleBasedPortfolioAdvisor`, `RuleBasedSearchAssist`
- v2: `Policy`, `RulePolicy`, `HybridExecutor`, `PolicyHierarchy`
- PFRS: `inrc2.RuleBasedWorkerDecisionEngine` (separate from `optimisation.WorkerAssist` interface)

### `internal/infrastructure/inrc2`

**Correctly placed:** PFRS solver, beam search, domain CSV writers (`discoveries_csv.go`, `beam_tree_csv.go`, etc.), worker decision/assist for beam workers.

**Misplaced (documented, not moved):**
- `worker_learning_emit.go` — emits learning CSV for CVRP/JSS/VRPTW but lives in NRP package because it uses `WorkerLearningRecord` types defined here.

### `cmd/owp`

**Correctly placed:** command dispatch, flag parsing, run dir/S3 helpers, `finalizeGenericSolverRun`.

**Remaining in CLI (acceptable):** PFRS hand-formatted `run.json` builders — tightly coupled to dashboard layout; could move to `inrc2` in a future sprint.

## v5 technical debt (do not move without integration work)

| Item | Risk | Notes |
|------|------|-------|
| Wire `PolicySearchHookRunner` / `HybridExecutor` into `search.go` | High | `--policy-mode` is parsed but not used in search loop today |
| Unify `inrc2.WorkerDecisionEngine` with `optimisation.WorkerAssist` | High | Touches validated `pfrs_search.submitWork` path |
| Move `nrp_objectives.go` to `inrc2` | Medium | Used by `optimisation/objective.go`; would create package cycle |
| Move `WorkerLearningRecord` + emitters to `optimisation` or `internal/telemetry` | Medium | Cross-domain training schema; many CSV header tests |
| Merge `ShadowRecorder` (inrc2) with `policy_shadow.go` (optimisation) | High | Different CSV schemas and lifecycles |
| Split `optimisation` into subpackages (`search`, `si`, `policy`) | High | Large import churn across infrastructure packages |
| Retire SI v1 assist in favour of SI 2.0 only | High | v1 is what production solvers run today |
| Wire `worker_policy.json` Go loader to PFRS `DecisionEngine` | Medium | Python trains model; Go path not connected |
| Integrate `continuous_learning.go` / `policy_training.go` with post-run hooks | Medium | Lifecycle orchestration exists but CLI doesn't call it |

## Dependency rules (target state)

```
cmd/owp
  → optimisation (search, assist, policy types)
  → infrastructure/* (domain problems, domain telemetry)
  → s3upload

infrastructure/*
  → optimisation (Problem, SearchConfig, SearchResult)
  ✗ should not import cmd

optimisation
  → domain/* (assignment, plan, etc.)
  ✗ should not import infrastructure/*
```

## ML / policy model loading

All Go loaders live in `optimisation`:

| Loader | File | JSON artifact |
|--------|------|---------------|
| `LoadImprovementCurveModel` | `stagnation_policy.go` | `stagnation_policy.json` |
| `LoadRestartModel` | `restart_policy.go` | `restart_policy.json` |
| `LoadPortfolioBudgetModel` | `portfolio_budget_model.go` | `portfolio_budget_model.json` |
| `LoadPolicyRegistry` | `policy_lifecycle.go` | `policy_registry.json` |
| `LoadPolicyRegistryFile` | `policy_executor.go` | `policy_v1.json` |

Python training: `platform/ml/train_policies.py`, `platform/ml/policies/*.json`.

## Telemetry file ownership

| File | Owner package |
|------|---------------|
| `run.json`, `solution.json` | `cmd/owp` (orchestration) |
| `results.csv`, `discoveries.csv` (CVRP) | `cvrp` |
| `discoveries.csv` (VRPTW) | `vrptw` |
| `worker_learning.csv` | `inrc2` (emit), called from `cmd/owp` |
| `worker_decisions.csv`, `worker_assist.csv` (PFRS) | `inrc2` |
| `generic_search_assist.csv`, `portfolio_assist.csv` | `optimisation` |
| `policy_decisions.csv` | `optimisation` (when policy runner wired) |
| PFRS beam CSVs (tree, plateau, diversity, etc.) | `inrc2` |
