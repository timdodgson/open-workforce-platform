# Search Intelligence 2.0 — Validation Report

## Status: Validated

Generated: 2026-07-10T15:27:54.000745

---

## Outcome-Based Promotion (Primary)

Ex-post optimal stop/continue vs learned and rule decisions.

| Metric | Value |
|--------|-------|
| Total checkpoints | 46944 |
| Learned outcome accuracy | 95.5% |
| Rule outcome accuracy | 72.0% |
| Regret vs rules | -213.2936 |
| Rule agreement (diagnostic) | 70.6% |

---

## Per-Domain Stagnation

| Domain | Samples | Outcome Acc | Regret vs Rules | Agreement | Promotion |
|--------|---------|-------------|-----------------|-----------|-----------|
| CVRP | 8210 | 99.0% | -0.7330 | 89.3% | ✅ |
| JSS | 1650 | 96.1% | -1.9965 | 64.1% | ✅ |
| NRP | 34554 | 94.3% | -288.5907 | 65.3% | ✅ |
| VRPTW | 2530 | 99.2% | -12.4826 | 86.3% | ✅ |

---

## Acceptance Criteria

| Criterion | Result |
|-----------|--------|
| Outcome accuracy >= 80% | ✅ PASS (95.5%) |
| Regret vs rules <= 0.0 | ✅ PASS (-213.2936) |
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

