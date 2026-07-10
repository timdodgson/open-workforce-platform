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

**Correctly placed:** core metaheuristics, `search.go`, SI v1 assist (`search_assist_hooks.go`, `portfolio_assist.go`), SI 2.0 policies (`policy_*.go`, `policy_executor.go`).

**Production SI 2.0 path:** `PolicySearchHookRunner` (search), `AllocateBudgetsViaPolicy` (portfolio), `inrc2.HybridWorkerDecisionEngine` (PFRS).

**Parallel systems:**
- v1: `SearchHookRunner`, `RuleBasedPortfolioAdvisor` — fallback inside hybrid mode
- v2: `PolicySearchHookRunner`, learned JSON loaders, post-run learning pipeline
- PFRS: `inrc2.HybridWorkerDecisionEngine` / `RuleBasedWorkerDecisionEngine`

### `internal/infrastructure/inrc2`

**Correctly placed:** PFRS solver, beam search, domain CSV writers (`discoveries_csv.go`, `beam_tree_csv.go`, etc.), worker decision/assist for beam workers.

**Misplaced (documented, not moved):**
- `worker_learning_emit.go` — emits learning CSV for CVRP/JSS/VRPTW but lives in NRP package because it uses `WorkerLearningRecord` types defined here.

### `cmd/owp`

**Correctly placed:** command dispatch, flag parsing, run dir/S3 helpers, `finalizeGenericSolverRun`.

**Remaining in CLI (acceptable):** PFRS hand-formatted `run.json` builders — tightly coupled to dashboard layout; could move to `inrc2` in a future sprint.

## v5 technical debt (remaining)

| Item | Risk | Status |
|------|------|--------|
| Wire `PolicySearchHookRunner` into `search.go` | — | **Done** |
| Wire `worker_policy.json` to PFRS | — | **Done** |
| Portfolio via `--policy-dir` | — | **Done** |
| Post-run learning pipeline | — | **Done** |
| Dashboard SI 2.0 tabs | — | **Done** |
| Full `validate-si2.ps1` (240 runs) | Low | Operator task |
| Unify `inrc2.WorkerDecisionEngine` with `optimisation.WorkerAssist` interface | High | Future |
| Retire SI v1-only paths | High | v1 is hybrid fallback today |
| Split `optimisation` into subpackages | High | Large churn |

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

## Naming glossary (Sprint 5)

Consistent terms for Search Intelligence before release.

### Product / architecture

| Term | Meaning |
|------|---------|
| **Search Intelligence (SI)** | Umbrella: AI advises solvers without replacing core search |
| **SI v1** | Production paths: `SearchHookRunner`, `RuleBasedPortfolioAdvisor`, `inrc2.WorkerDecisionEngine` |
| **SI 2.0** | `PolicySearchHookRunner`, learned JSON loaders, post-run pipeline — see `docs/SEARCH_INTELLIGENCE_V2.md` |

### Integration styles (three layers)

| Term | Scope | Production implementation |
|------|-------|---------------------------|
| **WorkerAssist** | Beam / parallel worker spawn (PFRS) | `inrc2.WorkerDecisionEngine` (parallel to `WorkerAssist` interface) |
| **SearchAssist** | Single-algorithm checkpoint hooks | `SearchHookRunner` + `RuleBasedSearchAssist` / `AdaptiveSearchAssist` |
| **PortfolioAssist** | Multi-strategy budget allocation | `RunPortfolioWithAssist` + `RuleBasedPortfolioAdvisor` |

### CLI flags (do not rename)

| Flag | Field | Values | Notes |
|------|-------|--------|-------|
| `--worker-decision-mode` | `SearchConfig.AssistMode` (CVRP/JSS/VRPTW) or PFRS worker wiring | `off`, `shadow`, `assist`, `adaptive` | Historical name; controls all SI layers, not workers only |
| `--policy-mode` | `SearchConfig.PolicyMode` | `rules`, `hybrid`, `learned` | SI 2.0; defaults `--policy-dir` to `../ml/policies` |

### Mode semantics (shared across layers)

| Mode | Behaviour |
|------|-----------|
| **off** | No SI hooks; zero overhead |
| **shadow** | Record predictions; no search behaviour change |
| **assist** | Act on recommendations with safety overrides |
| **adaptive** | Live-updating assist (`AdaptiveSearchAssist` for search; PFRS uses assist path with adaptive messaging) |

### SI 2.0 policy types

| Type | Role |
|------|------|
| **RulePolicy** | Deterministic rules; preserves v1 heuristic behaviour |
| **LearnedPolicy** | JSON model via `PolicyModel`; defers when low confidence |
| **HybridPolicy** | Learned when confident; `RulePolicy` fallback |
| **PolicySearchHookRunner** | Production search-loop policy execution (`policy_executor.go`) |
| **PolicyHierarchy** | Optional registry resolution (tests / future) |

### Telemetry CSVs by layer

| CSV | Layer |
|-----|-------|
| `worker_decisions.csv` | WorkerAssist shadow (PFRS) |
| `worker_assist.csv` | WorkerAssist assist/adaptive (PFRS) |
| `generic_search_assist.csv` | SearchAssist |
| `portfolio_assist.csv` | PortfolioAssist |
| `policy_decisions.csv` | SI 2.0 `PolicySearchHookRunner` |
| `policy_evaluation.csv` | SI 2.0 post-run evaluation |
| `worker_learning.csv` | Cross-layer training observations |

## Legacy stack (deprecated — do not extend)

The original generic NRP path predates domain-specific solvers. It remains for backward compatibility only.

| Component | Status | Replacement |
|-----------|--------|-------------|
| `owp optimise` | Deprecated | `solve-cvrp`, `solve-vrptw`, `solve-jobshop`, `tune-pfrs` |
| `owp benchmark` | Deprecated | `benchmark-inrc2`, `benchmark-*-ilp`, domain solve commands |
| `internal/application` | Legacy | Domain packages under `infrastructure/*` |
| `internal/domain/*` | Legacy | Problem-specific models in `infrastructure/*` |

New features belong in `infrastructure/<domain>` + `optimisation`, not `application` or `domain`.

### SI v1 assist (logical group — parent `optimisation` package)

These files implement SearchAssist and PortfolioAssist; a physical `optimisation/assist` subpackage is deferred to avoid import cycles with `search.go` (`RunSearch` ↔ hooks).

| File | Role |
|------|------|
| `search_assist_hooks.go` | `SearchHookRunner`, rule-based checkpoints |
| `adaptive_search_assist.go` | Adaptive checkpoint policy |
| `portfolio_assist.go` | Portfolio budget allocation |
| `portfolio_budget_rules.go` | Shared budget heuristics |
| `search_intelligence.go` | SI type definitions and documentation |

## cmd/owp layout (Sprint 6+)

| File group | Commands |
|------------|----------|
| `command_tune_pfrs.go`, `pfrs_tune_*.go` | `tune-pfrs`, `visualise-pfrs` |
| `benchmark_*.go` | `benchmark`, `benchmark-inrc2`, `benchmark-ilp`, `benchmark-*-ilp` |
| `command_solve_*.go`, `solve_*.go` | Domain metaheuristic solvers |
| `command_inrc2_*.go`, `inrc2_display.go` | `validate-inrc2`, `solve-inrc2` |
| `command_nrp_convert.go` | `convert-nrp` |
| `deprecation.go` | Shared legacy warnings |
