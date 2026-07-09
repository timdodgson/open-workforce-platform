# Search Intelligence 2.0 — Validation Report

## Status: Validated

Generated: 2026-07-09T17:11:08.326574

---

## Outcome-Based Promotion (Primary)

Ex-post optimal stop/continue vs learned and rule decisions.

| Metric | Value |
|--------|-------|
| Total checkpoints | 34479 |
| Learned outcome accuracy | 96.3% |
| Rule outcome accuracy | 67.7% |
| Regret vs rules | -185.1260 |
| Rule agreement (diagnostic) | 66.5% |

---

## Per-Domain Stagnation

| Domain | Samples | Outcome Acc | Regret vs Rules | Agreement | Promotion |
|--------|---------|-------------|-----------------|-----------|-----------|
| CVRP | 8080 | 99.0% | -0.5424 | 90.4% | ✅ |
| JSS | 1610 | 96.2% | -2.0685 | 65.1% | ✅ |
| NRP | 22289 | 95.0% | -284.6097 | 55.8% | ✅ |
| VRPTW | 2500 | 99.6% | -12.6324 | 86.5% | ✅ |

---

## Acceptance Criteria

| Criterion | Result |
|-----------|--------|
| Outcome accuracy >= 80% | ✅ PASS (96.3%) |
| Regret vs rules <= 0.0 | ✅ PASS (-185.1260) |
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

