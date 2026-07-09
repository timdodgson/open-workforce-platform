# Search Intelligence 2.0 — Validation Report

## Status: Validated

Generated: 2026-07-09T23:30:38.628571

---

## Outcome-Based Promotion (Primary)

Ex-post optimal stop/continue vs learned and rule decisions.

| Metric | Value |
|--------|-------|
| Total checkpoints | 46700 |
| Learned outcome accuracy | 95.8% |
| Rule outcome accuracy | 72.2% |
| Regret vs rules | -216.1346 |
| Rule agreement (diagnostic) | 70.6% |

---

## Per-Domain Stagnation

| Domain | Samples | Outcome Acc | Regret vs Rules | Agreement | Promotion |
|--------|---------|-------------|-----------------|-----------|-----------|
| CVRP | 8180 | 99.0% | -0.7357 | 89.6% | ✅ |
| JSS | 1620 | 96.2% | -2.0557 | 64.7% | ✅ |
| NRP | 34400 | 94.8% | -292.2256 | 65.2% | ✅ |
| VRPTW | 2500 | 99.6% | -12.6324 | 86.5% | ✅ |

---

## Acceptance Criteria

| Criterion | Result |
|-----------|--------|
| Outcome accuracy >= 80% | ✅ PASS (95.8%) |
| Regret vs rules <= 0.0 | ✅ PASS (-216.1346) |
| Learned policy loaded | ✅ PASS |

---

## Recommendation

**Promote to shadow** — outcome gates passed globally.

---

## Methodology

- Data: `generic_search_assist.csv` shadow checkpoints
- Ex-post label: stop if `best_penalty - final_best_penalty <= 1`
- Learned: stagnation curve P(improve) model (domain + algorithm scoped)
- Rules: 50k plateau stagnation window
- Promotion: outcome accuracy >= 80% AND regret_vs_rules <= 0

