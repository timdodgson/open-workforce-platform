# PFRS Lab Dashboard Refactor

Phased cleanup after Go platform refactor (SI 2.0 validated, unified `owp solve` CLI).

## Phase 1 — Shell & routing (done)

| Item | Status |
|------|--------|
| `useRunNavigation` — single run mode source (RunMeta → API fallback) | Done |
| Intelligence `tab-registry.ts` — one tab config | Done |
| Domain summary views extracted from monolithic page | Done |
| `REFACTOR.md` roadmap | Done |

## Phase 2 — Intelligence data layer (done)

| Item | Status |
|------|--------|
| Split `intelligence-data.ts` by section | Done |
| Artifact-first loading with staleness indicator in admin | Done |
| Server-driven tab sections (reduce client fetch state) | Done |

## Phase 3 — Run pages & UX

| Item | Status |
|------|--------|
| Standardize all run pages on `RunPageShell` | Pending |
| Shared `StatGrid` / consolidate `MetricBox` usage | Pending |
| Landing page content extraction | Pending |
| Resolve or link `/knowledge` orphan route | Pending |

## Phase 4 — Docs & ops

| Item | Status |
|------|--------|
| README `STORAGE_PROVIDER` fix | Pending |
| Update root README / ARCHITECTURE to `owp solve` | Pending |
| CI hook: rebuild intelligence artifacts after policy train | Pending |
