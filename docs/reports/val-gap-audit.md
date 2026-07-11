# val-* Experiment Matrix — Gap Audit

**Purpose:** Confirm production storage contains every canonical SI2 validation run defined in `experiment-matrix.ts`.  
**Regenerate:** `cd platform/web/pfrs-lab && npm run audit-val-matrix` (set `STORAGE_PROVIDER=s3` for production).

---

## Canonical matrix

| Tier | Configs | Policies | Seeds | Runs |
|------|---------|----------|-------|------|
| Fast | 8 | rules, hybrid, learned | 10 | 240 |
| Deep | 8 | rules, hybrid, learned | 2 | 48 |
| **Total** | **16** | — | — | **288** |

Label patterns:

- Fast: `val-<domain>-<instance>-<mode>-<policy>-s<seed>`
- Deep: `val-deep-<domain>-<instance>-<mode>-<policy>-s<seed>`

Source of truth: `platform/web/pfrs-lab/src/lib/experiment-matrix.ts`  
Execution scripts: `validate-si2.ps1` (fast), `validate-si2-deep.ps1` (deep).

---

## Production snapshot (S3)

**Audited:** 2026-07-11  
**Bucket:** `pfrs-research-lab-data`  
**Total runs in bucket:** 679  
**val-* runs:** 295  

| Metric | Value |
|--------|-------|
| Matrix expected | 288 |
| Matrix found | **288** |
| Gap | **0** |
| Status | **Complete** |

### Per-config coverage

| Config | Tier | Found / Expected |
|--------|------|------------------|
| fast-cvrp-sa | fast | 30 / 30 |
| fast-cvrp-portfolio | fast | 30 / 30 |
| fast-jss-tabu | fast | 30 / 30 |
| fast-jss-portfolio | fast | 30 / 30 |
| fast-vrptw-sa | fast | 30 / 30 |
| fast-vrptw-portfolio | fast | 30 / 30 |
| fast-nrp-sa | fast | 30 / 30 |
| fast-nrp-portfolio | fast | 30 / 30 |
| deep-cvrp-sa | deep | 6 / 6 |
| deep-cvrp-portfolio | deep | 6 / 6 |
| deep-jss-tabu | deep | 6 / 6 |
| deep-jss-portfolio | deep | 6 / 6 |
| deep-vrptw-sa | deep | 6 / 6 |
| deep-vrptw-portfolio | deep | 6 / 6 |
| deep-nrp-sa | deep | 6 / 6 |
| deep-nrp-portfolio | deep | 6 / 6 |

### Extra val-* runs (not in matrix)

These are smoke / timing runs — safe to keep, not required for matrix completeness:

| Run label | Origin |
|-----------|--------|
| `val-quick-cvrp-sa-rules-s42` | `validate-si2-quick.ps1` |
| `val-quick-cvrp-sa-hybrid-s42` | `validate-si2-quick.ps1` |
| `val-quick-cvrp-sa-learned-s42` | `validate-si2-quick.ps1` |
| `val-quick-cvrp-portfolio-rules-s42` | `validate-si2-quick.ps1` |
| `val-quick-cvrp-portfolio-hybrid-s42` | `validate-si2-quick.ps1` |
| `val-quick-cvrp-portfolio-learned-s42` | `validate-si2-quick.ps1` |
| `val-test-timing` | Ad-hoc timing probe |

---

## Local dev snapshot

A fresh clone with only seeded TSP demos shows **0 / 288** matrix coverage — expected. Local `data/runs/` contains `demo-tsp-*` only unless you sync from S3 or run validation scripts.

To audit local:

```powershell
cd platform/web/pfrs-lab
$env:STORAGE_PROVIDER='local'
npm run audit-val-matrix
```

To work against production data without copying runs:

```powershell
$env:STORAGE_PROVIDER='s3'
$env:PFRS_S3_BUCKET='pfrs-research-lab-data'
npm run dev
```

---

## Relationship to EXP-007 (320 assist runs)

EXP-007 / [statistical validation](./search-intelligence-statistical-validation.md) used **worker-decision-mode** sweep (`off`, `shadow`, `assist`, `adaptive`) with different label conventions — 320 runs, not `val-*` labels.

| Suite | Labels | Purpose |
|-------|--------|---------|
| SI assist validation (EXP-007) | Domain-specific bench labels | Prove assist/adaptive safe + compute savings |
| SI2 policy matrix (this audit) | `val-*` / `val-deep-*` | Prove rules/hybrid/learned policies across seeds |

Both suites are required for different claims. The experiment matrix page documents the `val-*` catalog; EXP-007 remains the assist-mode evidence.

---

## Remediation playbook

| Gap type | Action |
|----------|--------|
| Missing fast tier labels | `cd platform/go; .\scripts\validate-si2.ps1` (or re-run failed rows from log) |
| Missing deep tier labels | `.\scripts\validate-si2-deep.ps1` |
| Matrix complete but UI stale | `npm run rebuild-intelligence` with S3 env |
| Unsure which labels failed | Check `platform/go/scripts/regression-logs/` |

After remediation, re-run `npm run audit-val-matrix` and update the **Production snapshot** section above.
