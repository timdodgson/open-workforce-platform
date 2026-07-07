# Duplicate Code Audit — Sprint 6

Audit after CLI refactor Sprints 1–5. Only low-risk mechanical deduplication was applied.

## Removed (this sprint)

| Duplicate | Extraction | Files |
|-----------|------------|-------|
| JSS/VRPTW portfolio vs single-search blocks | `runSearchOrPortfolio` | `cli_runtime.go`, `command_solve_jobshop.go`, `command_solve_vrptw.go` |
| `--instance` required guard | `requireInstanceFlag` | `cli_flags.go`, solve-cvrp/jss/vrptw |
| `--show-invalid` flag loop | `parseShowInvalidFlag` | `cli_flags.go`, `command_tune_pfrs.go`, `command_benchmark.go` |
| Benchmark delta/pct formatting | `formatObjectiveDelta` | `cli_output.go`, `command_benchmark.go` |
| Improvement % display | `printImprovementPct` | `cli_output.go`, solve commands |
| Search stats display | `printSearchResultStats` | `cli_output.go`, solve-jss/vrptw |
| `tune-pfrs` direct S3 upload (×2) | `uploadRunOutput` | `command_tune_pfrs.go` |
| PFRS CSV write err/success logging (×8) | `logTelemetryFileWrite` | `cli_telemetry.go`, `command_tune_pfrs.go` |
| PFRS worker CSV emit logging | `logTelemetryFileWrite` in emitters | `cli_telemetry.go` |
| Worker 0 start penalty loop (×2) | `inrc2.Worker0StartPenalty` | `pfrs_audit.go`, `command_tune_pfrs.go` |
| `writeTelemetryBytes` / `writeTelemetryBytesErr` | delegate pattern | `cli_telemetry.go` |
| Policy JSON filenames in `loadPolicies` | `stagnationPolicyFile`, `restartPolicyFile` constants | `policy_executor.go` |

## Deferred (technical debt)

| Duplicate | Why deferred |
|-----------|--------------|
| Full generic solver command framework | CVRP has portfolio table, tabu/LAHC flags, domain CSVs; domains differ intentionally |
| `runBenchmarkINRC2` instance loading vs `loadINRC2Instance` | Different default instance resolution and profile wiring; behavior risk |
| PFRS beam telemetry orchestration (~140 lines) | Domain logic (depthMap, parentMap, branchCounts); belongs in `inrc2` export helper |
| Valid/invalid league tables (tune-pfrs vs benchmark) | Different result types and sort rules |
| PFRS week execution loop (tune vs benchmark) | Different multi-algorithm benchmark semantics |
| `LoadPolicyRegistryFile` wiring into `loadPolicies` | Registry schema differs from hardcoded filenames; behavior change if `policy_v1.json` absent |
| Generic JSON model loaders (`Load*Model`) | Shared loader would change error messages and empty-file semantics |
| ILP bespoke S3 artifact upload | Individual files + manifest; not directory sync |
| `officialValidate` vs beam validation loop | Beam path mutates winning path during validation |
| SI2 `validation_suite.go` runner + CLI command | No executor exists; new feature surface |
| CSV writer pattern across `optimisation` + `inrc2` | Many schemas; generic builder is wide refactor |
| Hand-formatted PFRS `run.json` vs `writeRunMetadata` | Byte-layout tests lock exact formatting |
| `applySearchIntelligenceFlags` vs `wirePFRSWorkerIntelligence` | Different target types (`SearchConfig` vs `inrc2` engines) |
| Dashboard duplicate empty-state strings | Out of scope unless explicitly requested |
