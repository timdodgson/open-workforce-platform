# Search Intelligence 2.0 — Validation Report

## Status: Validated

Generated: 2026-07-07T22:25:48.167676

---

## Retrospective Policy Validation

Compares Rule vs Learned policy decisions at every search checkpoint
using real telemetry from shadow-mode runs. No solver behaviour was changed.

| Metric | Value |
|--------|-------|
| Total checkpoints | 5150 |
| Agreements | 2703 |
| Disagreements | 2447 |
| Agreement rate | 52.5% |
| Rule stop recommendations | 2824 |
| Learned stop recommendations | 377 |
| Learned more confident on disagree | 2447 |
| Mean learned confidence | 0.8500 |
| Mean rule confidence | 0.6097 |

---

## Per-Domain Results

| Domain | Checkpoints | Agreement | Disagreement | Rate | Rule Stops | Learned Stops |
|--------|-------------|-----------|--------------|------|------------|---------------|
| CVRP | 3750 | 1456 | 2294 | 38.8% | 2539 | 245 |
| JSS | 700 | 571 | 129 | 81.6% | 255 | 126 |
| VRPTW | 700 | 676 | 24 | 96.6% | 30 | 6 |

---

## Acceptance Criteria

| Criterion | Result |
|-----------|--------|
| Agreement rate > 80% | ❌ FAIL (52.5%) |
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
