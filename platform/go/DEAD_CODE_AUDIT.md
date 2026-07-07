# Dead Code Audit — Sprint 7

Audit before release. Only symbols with zero production and zero test usage (except dedicated tests for removed symbols) were deleted.

## Removed

| Symbol | File | Reason |
|--------|------|--------|
| `LoadPolicyRegistryFile` | `policy_executor.go` | Never called; `loadPolicies` uses hardcoded filenames |
| `ValidationStatus` | `validation_suite.go` | Type defined but never referenced |
| `GenerateRunLabel` | `validation_suite.go` | Only used by its own test; no CLI runner |
| `TestGenerateRunLabel` | `validation_suite_test.go` | Tests removed symbol |
| `DefaultContinuousLearningConfig` | `continuous_learning.go` | Never called (tests use inline config) |
| `PortfolioAssistRecordV2` | `portfolio_budget_model.go` | Struct defined but never used |
| `SearchIntelligence` struct | `search_intelligence.go` | Never instantiated; interfaces retained |
| `IsActive`, `IsAssist`, `IsShadow`, `IsAdaptive` | `search_intelligence.go` | Methods on unused struct |
| Orphan section comments | `command_solve_jobshop.go`, `command_solve_cvrp.go`, `command_benchmark.go` | Leftover from main.go split; mislabeled or trailing |

## Dashboard

No dashboard files removed. Intelligence sub-routes (`decisions`, `predictions`, `what-if`, `feature-importance`, `learning`) are wired through `IntelligenceShell.tsx` tabs — intentionally reachable from `/intelligence`.

## Deferred (not dead — SI 2.0 scaffold or uncertain)

| Item | Notes |
|------|-------|
| Entire SI 2.0 policy framework (~15 files) | Test-only today; awaiting `search.go` wiring |
| `validation_suite.go` stats API | `ComputeStatistics`, `CompareGroups`, etc. — tested, no CLI runner |
| `LoadPolicyRegistry` / lifecycle / training / promotion | Orchestration for future post-run hooks |
| `--policy-mode` / `--policy-dir` on `SearchConfig` | Parsed by CLI, not read in `search.go` yet |
| `PolicySearchHookRunner`, `HybridExecutor`, `PolicyProvider` | Unit-tested; not production-wired |
| `WorkerAssist` / `PortfolioAssist` / `SearchAssist` interfaces | Documentation contracts; no Go implementors yet |
| PFRS `inrc2.WorkerDecisionEngine` | Production path parallel to `WorkerAssist` interface |
| Experiments / knowledge / chat routes | Intentionally in navigation or linked from intelligence |
