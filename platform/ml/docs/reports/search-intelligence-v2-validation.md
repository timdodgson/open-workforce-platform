# Search Intelligence 2.0 — Validation Report

## Status: Validated

Generated: 2026-07-11T16:23:37.163545

---

## Outcome-Based Promotion (Primary)

Ex-post optimal stop/continue vs learned and rule decisions.

| Metric | Value |
|--------|-------|
| Total checkpoints | 22188 |
| Learned outcome accuracy | 97.3% |
| Rule outcome accuracy | 58.2% |
| Regret vs rules | -164.9524 |
| Rule agreement (diagnostic) | 58.5% |

---

## Per-Domain Stagnation

| Domain | Samples | Outcome Acc | Regret vs Rules | Agreement | Promotion |
|--------|---------|-------------|-----------------|-----------|-----------|
| CVRP | 7230 | 99.3% | -0.8159 | 90.8% | ✅ |
| JSS | 1500 | 100.0% | -2.1362 | 64.2% | ✅ |
| NRP | 11058 | 95.1% | -327.2943 | 30.5% | ✅ |
| VRPTW | 2400 | 100.0% | -13.1837 | 86.2% | ✅ |

---

## Acceptance Criteria

| Criterion | Result |
|-----------|--------|
| Outcome accuracy >= 80% | ✅ PASS (97.3%) |
| Regret vs rules <= 0.0 | ✅ PASS (-164.9524) |
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

