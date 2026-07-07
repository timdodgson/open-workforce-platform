# Dead Code Audit — updated for SI 2.0 production

## Removed (Sprint 9)

| Symbol | File | Reason |
|--------|------|--------|
| `HybridExecutor` | `hybrid_executor.go` | Duplicate of `PolicySearchHookRunner`; tests only, never on hot path |

## Production-wired (not dead)

| Item | Hot path |
|------|----------|
| `PolicySearchHookRunner` | `search.go` via `newSearchHooks` when `--policy-mode` set |
| `RunPortfolioWithAssist` + `AllocateBudgetsViaPolicy` | portfolio mode + `--policy-mode` |
| `inrc2.HybridWorkerDecisionEngine` | `tune-pfrs` + `--policy-mode` |
| `RunPostRunPolicyPipeline` | `finalizeGenericSolverRun` after solves |
| `WritePolicyEvaluationCSV` | emitted with `policy_decisions.csv` |

## Optional / library (kept, not CLI-wired)

| Item | Notes |
|------|-------|
| `PolicyHierarchy`, `PolicyProvider` | Tested registry helpers |
| `CounterfactualRecorder` | Research CSV; no CLI emission today |
| `validation_suite.go` stats API | Tested; no CLI runner |
| SI v1 `SearchHookRunner` | Still used as hybrid fallback inside `PolicySearchHookRunner` |

## Dashboard

Intelligence UI at `/intelligence`. SI 2.0 telemetry: Policies tab (`policy_decisions.csv`), SI Validation tab (`val-*` / `si2-*` runs).
