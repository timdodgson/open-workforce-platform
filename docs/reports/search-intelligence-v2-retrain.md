# Search Intelligence 2.0 — Validation Report

## Status: Validated

Generated: 2026-07-09T15:03:00.638067

---

## Outcome-Based Promotion (Primary)

Ex-post optimal stop/continue vs learned and rule decisions.

| Metric | Value |
|--------|-------|
| Total checkpoints | 12318 |
| Learned outcome accuracy | 99.1% |
| Rule outcome accuracy | 85.5% |
| Regret vs rules | -3.2744 |
| Rule agreement (diagnostic) | 85.4% |

---

## Per-Domain Stagnation

| Domain | Samples | Outcome Acc | Regret vs Rules | Agreement | Promotion |
|--------|---------|-------------|-----------------|-----------|-----------|
| CVRP | 7980 | 99.0% | -0.7316 | 89.6% | ✅ |
| JSS | 1600 | 98.8% | -1.8958 | 63.3% | ✅ |
| NRP | 238 | 96.2% | 0.4954 | 81.1% | ❌ |
| VRPTW | 2500 | 99.6% | -12.6324 | 86.5% | ✅ |

---

## Acceptance Criteria

| Criterion | Result |
|-----------|--------|
| Outcome accuracy >= 80% | ✅ PASS (99.1%) |
| Regret vs rules <= 0.0 | ✅ PASS (-3.2744) |
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

