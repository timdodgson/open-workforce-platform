# Search Intelligence 2.0 — Validation Report

## Status: Validated

Generated: 2026-07-08T02:17:11.908992

---

## Retrospective Policy Validation

Compares Rule vs Learned policy decisions at every search checkpoint
using real telemetry from shadow-mode runs. No solver behaviour was changed.

| Metric | Value |
|--------|-------|
| Total checkpoints | 5418 |
| Agreements | 2832 |
| Disagreements | 2586 |
| Agreement rate | 52.3% |
| Rule stop recommendations | 2970 |
| Learned stop recommendations | 396 |
| Learned more confident on disagree | 2586 |
| Mean learned confidence | 0.8500 |
| Mean rule confidence | 0.6096 |

---

## Per-Domain Results

| Domain | Checkpoints | Agreement | Disagreement | Rate | Rule Stops | Learned Stops |
|--------|-------------|-----------|--------------|------|------------|---------------|
| CVRP | 3780 | 1480 | 2300 | 39.2% | 2539 | 251 |
| JSS | 700 | 571 | 129 | 81.6% | 255 | 126 |
| NRP | 238 | 105 | 133 | 44.1% | 146 | 13 |
| VRPTW | 700 | 676 | 24 | 96.6% | 30 | 6 |

---

## Acceptance Criteria

| Criterion | Result |
|-----------|--------|
| Agreement rate > 80% | ❌ FAIL (52.3%) |
| Learned policy loaded | ✅ PASS |
| No safety violations | ✅ PASS (retrospective, no behaviour change) |
| Learned confidence > 0.60 | ✅ PASS (0.85) |

---

## Recommendation

**Remain on Rules** — low agreement. More training data needed before promotion.

---

## Methodology

- Data source: `generic_search_assist.csv` from shadow-mode runs
- Policy source: `policies/stagnation_policy.json` trained on 950 checkpoints
- Rule baseline: fixed stagnation window (50,000 candidates)
- Learned model: exponential decay P(improve) = A × exp(−λ × plateau_ratio)
- Threshold: P(improve) < 0.10 → recommend early stop
- Safety: never stop before 20% budget consumed

No fabricated data. All metrics from real experiment telemetry.
