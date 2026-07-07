# Search Intelligence 2.0 — Validation Report

## Status: Validated

Generated: 2026-07-07T20:26:28.984035

---

## Retrospective Policy Validation

Compares Rule vs Learned policy decisions at every search checkpoint
using real telemetry from shadow-mode runs. No solver behaviour was changed.

| Metric | Value |
|--------|-------|
| Total checkpoints | 950 |
| Agreements | 474 |
| Disagreements | 476 |
| Agreement rate | 49.9% |
| Rule stop recommendations | 574 |
| Learned stop recommendations | 98 |
| Learned more confident on disagree | 476 |
| Mean learned confidence | 0.8500 |
| Mean rule confidence | 0.6208 |

---

## Per-Domain Results

| Domain | Checkpoints | Agreement | Disagreement | Rate | Rule Stops | Learned Stops |
|--------|-------------|-----------|--------------|------|------------|---------------|
| CVRP | 750 | 292 | 458 | 38.9% | 541 | 83 |
| JSS | 100 | 82 | 18 | 82.0% | 33 | 15 |
| VRPTW | 100 | 100 | 0 | 100.0% | 0 | 0 |

---

## Acceptance Criteria

| Criterion | Result |
|-----------|--------|
| Agreement rate > 80% | ❌ FAIL (49.9%) |
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
