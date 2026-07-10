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

1. **Portfolio budget heuristics** — shared constants + `PortfolioBudgetHeuristic` in `optimisation/assist/portfolio_rules.go`, used by v1 (`RuleBasedPortfolioAdvisor`) and v2 (`NewPortfolioBudgetRulePolicy`).
2. **CVRP telemetry CSV** — moved from `cmd/owp/cli_telemetry.go` to `cvrp/telemetry_csv.go` (`BuildResultsCSV`, `BuildDiscoveriesCSV`).
3. **VRPTW telemetry CSV** — moved to `vrptw/telemetry_csv.go` (`BuildDiscoveriesCSV`).
4. **PFRS worker intelligence wiring** — `wirePFRSWorkerIntelligence` in `cmd/owp/cli_intelligence.go` deduplicates tune-pfrs setup.
5. **Policy registry clarity** — documented that `LoadPolicyRegistryFile` (policy_v1.json) and `LoadPolicyRegistry` (policy_registry.json) serve different schemas.

## Current package health

### `internal/optimisation` (~70 files)

**Correctly placed:** core metaheuristics (`search.go`), SI v1 assist, SI 2.0 policies (`policy_*.go`, `policy_executor.go`). No longer imports `domain/*` (legacy NRP search moved to `inrc2/legacysearch` in Phase 21).

**Production SI 2.0 path:** `PolicySearchHookRunner` (search), `AllocateBudgetsViaPolicy` (portfolio), `inrc2.HybridWorkerDecisionEngine` (PFRS).

**Parallel systems:**
- v1: `SearchHookRunner`, `RuleBasedPortfolioAdvisor` — fallback inside hybrid mode
- v2: `PolicySearchHookRunner`, learned JSON loaders, post-run learning pipeline
- PFRS: `inrc2.HybridWorkerDecisionEngine` / `RuleBasedWorkerDecisionEngine`

### `internal/infrastructure/inrc2`

**Correctly placed:** PFRS solver, beam search, domain CSV writers (`discoveries_csv.go`, `beam_tree_csv.go`, etc.), worker decision/assist for beam workers, PFRS tune orchestration (`tune_options.go`, `pfrs_tune_*_runner.go`, `tune_validate.go`), legacy work-item search (`legacysearch/` — constructive, metaheuristics, NRP objectives).

**Misplaced (documented, not moved):**
- `worker_learning_emit.go` — emits learning CSV for CVRP/JSS/VRPTW but lives in NRP package because it uses `WorkerLearningRecord` types defined here.

### `cmd/owp`

**Correctly placed:** command dispatch, flag parsing, run dir/S3 helpers, `finalizeGenericSolverRun`.

**Remaining in CLI (acceptable):** display formatters for PFRS tune (`pfrs_tune_display.go`); flag parsing (`pfrs_tune_flags.go` embeds `inrc2.TuneOptions`).

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
| **BYOD / BYOA SDK** | — | Future — see below; keep `searchdef.Problem`, telemetry contract, injectable runners |

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
  → searchdef, assist, policy only
  ✗ should not import infrastructure/*
  ✗ no domain/* imports (achieved Phase 21 for legacy NRP extraction)

telemetry/workerlearning
  ✗ should not import optimisation (achieved Phase 24 — uses SearchOutcome; CLI bridges from SearchResult)
```

## BYOD SDK (`internal/sdk`)

External domains and custom search modes plug in via a small registry without forking `cmd/owp` or `optimisation` internals.

**Package layout:**

```
internal/sdk
  RegisterProblem(desc)   — domain loader + CLI defaults
  RegisterSearch(mode, runner) — optional custom search mode
  GetProblem / Problems / RegisteredSearchModes
  RunSearch / ResolveSearchRunner — custom mode or optimisation.RunSearch fallback

internal/sdk/builtin
  init() registers cvrp, vrptw, jobshop (imported from cmd/owp)
```

**Extension points** (preserve when refactoring):

| Extension point | Contract | Registration |
|-----------------|----------|--------------|
| **Domain** | `searchdef.Problem` + `ProblemLoader` | `sdk.RegisterProblem` |
| **Algorithm** | `SearchRunner(problem, SearchConfig) SearchResult` | `sdk.RegisterSearch` (built-in: sa, lahc, tabu, portfolio, adaptive) |
| **SI / policy** | `assist`, `policy`, `SearchConfig` | Unchanged — passed through `SearchConfig` |
| **Telemetry** | CSV + `run.json` contract | CLI finalize hooks (per-domain commands today) |
| **CLI** | `owp list-solvers` | Registry discovery; per-domain `solve-*` commands remain |

**Do not** when refactoring:

- Couple `optimisation` to a specific infrastructure package (fixed in Phase 19: `siadapter` under `inrc2`).
- Grow domain-specific logic in `optimisation` root — new domains belong in `infrastructure/<domain>`, `sdk/builtin`, or external modules.
- Hide `Problem` / `SearchConfig` behind cmd-only types — `searchdef` and `sdk.Problem` are the stable engine contract.

**Example:** `examples/byod-tsp` — minimal TSP domain registered via `owp-sdk` and wired into `cmd/owp`.

**Extracted module:** `platform/owp-sdk` (`searchdef` + problem registry). Search runner registration remains in `platform/go/internal/sdk` (depends on `optimisation`).

NRP legacy work-item types now live entirely in `inrc2/legacysearch` (`domain_assignment.go`, `domain_workitem.go`, `domain_plan.go`). Phase 23 removed `internal/domain/*` and `infrastructure/loader`.

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

## Legacy stack (removed / quarantined)

| Component | Status | Notes |
|-----------|--------|-------|
| `owp optimise` | **Removed** (Phase 22) | Use `solve-*` / `tune-pfrs` |
| `owp benchmark` | **Removed** (Phase 22) | Use `benchmark-inrc2`, `benchmark-*-ilp` |
| `internal/legacy/application` | **Removed** (Phase 22) | Was only used by deprecated CLI |
| `internal/domain/*` | **Removed** (Phase 23) | Types inlined into `inrc2/legacysearch` |
| `infrastructure/loader` | **Removed** (Phase 23) | JSON dataset loader; no CLI consumer after Phase 22 |

New features belong in `infrastructure/<domain>` + `optimisation`.

### SI v1 assist (`optimisation/assist` + `optimisation/searchdef`)

Core search types live in `searchdef` (no dependency on `search.go`). SI v1 hook implementations live in `assist`. The parent `optimisation` package re-exports via `searchdef_exports.go` and wires policy hooks via `search_hooks_bridge.go`.

| Package / file | Role |
|----------------|------|
| `searchdef/` | `Problem`, `SearchAssistConfig`, `SearchProgress`, checkpoint types |
| `assist/` | `SearchHookRunner`, `RuleBasedSearchAssist`, `AdaptiveSearchAssist`, CSV writer |
| `search_hooks_bridge.go` | `newSearchHooks` / `finalizeSearchHooks` (avoids assist ↔ search cycle) |
| `assist/portfolio.go`, `assist/portfolio_rules.go`, `assist/portfolio_runner.go` | PortfolioAssist types, advisor, safety, CSV, `ExecutePortfolio` |
| `portfolio_bridge.go`, `portfolio_advice.go` | `RunPortfolioWithAssist` bridge + advice resolution (`RunSearch` injected) |
| `policy/assist_hook.go` | `PolicySearchHookRunner` (SI 2.0 search checkpoints) |
| `search_intelligence.go` | WorkerAssist / PortfolioAssist docs; SearchAssist types re-exported |

## cmd/owp layout (Sprint 6+)

| File group | Commands |
|------------|----------|
| `cli_parse.go`, `cli_profile_flags.go`, `pfrs_cli_flags.go` | Shared flag parsing (replaces `cli_flags.go`) |
| `cli_storage.go`, `cli_intelligence.go`, `cli_search_defaults.go` | Run output, SI flags, search defaults (replaces `cli_runtime.go` body) |
| `command_tune_pfrs.go`, `pfrs_tune_flags.go`, `pfrs_tune_display.go` | `tune-pfrs`, `visualise-pfrs` (flags + display; orchestration in `inrc2`) |
| `benchmark_*.go` | `benchmark`, `benchmark-inrc2`, `benchmark-ilp`, `benchmark-*-ilp` |
| `command_solve_*.go`, `solve_*.go` | Domain metaheuristic solvers |
| `command_inrc2_*.go`, `inrc2_display.go` | `validate-inrc2`, `solve-inrc2` |
| `command_nrp_convert.go` | `convert-nrp` |
| `command_validate_si2.go` | `validate-si2 plan`, `validate-si2 analyze` |
| `scripts/regression-post-refactor.ps1` | Post-refactor gate (go test + 24 live runs) |

### `telemetry/workerlearning` (Phase 24)

Domain-agnostic worker learning CSV contract (`Record`, `WriteCSV`). Single-search emission uses `SearchOutcome` + `SingleWorkerConfig`; `cmd/owp` maps `optimisation.SearchResult` at the CLI boundary. NRP beam workers use `inrc2.EmitNRPWorkerLearning` → `workerlearning.Record` directly.

| File | Role |
|------|------|
| `record.go` | `Record`, CSV header/row writers |
| `outcome.go` | `SearchOutcome` (engine-agnostic stats) |
| `emit_single.go` | `EmitSingleWorkerLearning` for CVRP/JSS/VRPTW |

### `inrc2/siadapter`

PFRS worker telemetry adapters (worker CSV ↔ generic SI CSV). Under `inrc2` because it maps domain recorder types to `optimisation` CSV schemas.

### `inrc2` PFRS tune (Phase 20)

| File | Role |
|------|------|
| `tune_options.go` | `TuneOptions`, grid builder, beam/single-config detection |
| `pfrs_tune_sweep_runner.go` | `RunTuneSweep`, `FinalizeTuneSweep` |
| `pfrs_tune_beam_runner.go` | `RunTuneBeam` with hook callbacks for CLI progress |
| `tune_validate.go` | `OfficialValidateBeamPath` (pre-refinement scoring) |

### `inrc2/legacysearch` (Phase 21)

Legacy work-item / INRC-II metaheuristic stack (moved from `optimisation`):

| File group | Role |
|------------|------|
| `algorithm.go`, `profile.go` | Algorithm registry, `AlgorithmProfile` |
| `domain_assignment.go`, `domain_workitem.go`, `domain_plan.go` | Work-item plan types (inlined from `internal/domain/*`, Phase 23) |
| `constructive.go`, `hillclimbing.go`, `annealing.go`, `tabusearch.go`, `lns.go` | Metaheuristics |
| `neighbourhood.go`, `context.go`, `types.go` | Moves, context, NRP input types |
| `objective.go`, `nrp_objectives.go` | Objective scoring + INRC-II soft penalties |

Used by `inrc2.SolveWeek` and `benchmark-inrc2`.
