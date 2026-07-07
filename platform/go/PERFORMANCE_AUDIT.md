# Performance and Maintainability Audit (Sprint 8)

Lightweight audit of obvious inefficiencies. No algorithm, solver output, or statistical result changes.

## Fixed in this sprint

### Dashboard (`platform/web/pfrs-lab`)

| Issue | Fix | Risk |
|-------|-----|------|
| Benchmarks/statistics re-fetched `run.json` after `listRunsAsync` already loaded it | Reuse `run.metadata`; call `loadRunSummary` only when objective is missing from metadata | Low |
| `listRunsAsync` fetched runs sequentially | Parallel `Promise.all` over run IDs | Low |
| Manifest timestamps unused; duplicate manifest parsing | Cached `readManifestIndex()`; attach `timestamp` to `RunListEntry` | Low |
| Only `listRuns` was cached | `loadRunMetadata`, `loadWeeks`, manifest index now use `cached()` | Low |
| Intelligence page parsed `worker_learning.csv` twice per run | Parse once; reuse rows for `learning` and `decisionLearning` | Low |
| `EfficiencyDashboard` scatter plot recomputed max inside `.map()` (O(n²)) | Precompute `scatterPlot` in `useMemo` | Low |
| Stale `/api/runs` returned single pseudo-run | Route now returns all runs from `listRunsAsync` | Low |
| Unused `loadRunSummary` import on experiments page | Removed | Low |

### CLI / Go (`platform/go`)

| Issue | Fix | Risk |
|-------|-----|------|
| `officialValidate` re-read week files on every path step | Preload week data once per file | Low |
| `benchmark-inrc2` parsed `--algorithm` twice | Single `parseStringFlag` call | Low |
| `extractPriorities` ignored JSON unmarshal errors | Return error (matches `extractCapacities`) | Low |

## Deferred — v5 technical debt

These are real opportunities but need design work or carry behaviour risk.

### Dashboard

- **Manifest-only list views** — Skip per-run `run.json` on list/home/trends when manifest fields suffice; detail pages still fetch full JSON.
- **Intelligence eager CSV loading** — Load tab-scoped files instead of all assist/learning CSVs for every run on page load.
- **`force-dynamic` on run pages** — Consider `revalidate` where data is not truly live.
- **CSV parser consolidation** — Several near-duplicate parsers in intelligence routes.
- **Broader loader caching** — `loadTree`, `loadDiscoveries`, etc. still uncached.

### CLI / Go

- **Wire SI 2.0 `PolicySearchHookRunner`** into `search.go` (scaffold exists, not on hot path).
- **`runBenchmarkINRC2` → shared `loadINRC2Instance`** — dedupe instance loading; verify benchmark parity first.
- **CVRP portfolio via `runSearchOrPortfolio`** — special per-strategy table output blocks naive merge.
- **ILP bespoke S3 upload path** — consolidate with `uploadRunOutput`.
- **NRP triple JSON unmarshal** in benchmark path — needs trace before collapsing.

## Verification

```powershell
Set-Location "c:\Users\Tim\open-workforce-platform\platform\go"
go test ./... -short

Set-Location "c:\Users\Tim\open-workforce-platform\platform\web\pfrs-lab"
npm run build
```

## Behaviour guarantees

- No solver algorithms changed.
- No statistical formulas changed (Welch t-test, Cohen's d, etc. unchanged).
- Dashboard displays same data; fewer redundant I/O and parse operations.
- `extractPriorities` now fails fast on corrupt item JSON (previously produced zero-valued inputs silently).
