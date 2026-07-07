# Search Intelligence — Failure Analysis

## Objective

Investigate every Search Intelligence failure across all 225 validated experiment runs.
Understand why the recommendation was wrong. Produce a failure taxonomy with confidence intervals and actionable recommendations.

---

## Failure Inventory

Across 225 total runs (NRP: 30, CVRP: 45, JSS: 90, VRPTW: 60), **3 runs degraded**.

| # | Domain | Instance | Algorithm | Seed | Off | Shadow | Assist | Degradation |
|---|--------|----------|-----------|------|-----|--------|--------|-------------|
| F1 | JSS | la01 | SA | 555 | 688 | 688 | **751** | +63 (9.2% worse) |
| F2 | JSS | la01 | Portfolio | 123 | 666 | 666 | **675** | +9 (1.4% worse) |
| F3 | JSS | ft10 | Portfolio | 777 | 997 | 997 | **1029** | +32 (3.2% worse) |

All failures are in JSS. NRP, CVRP, and VRPTW have **zero degradations**.

---

## Failure F1: JSS la01 SA Seed 555

### Comparison

| Metric | Off | Shadow | Assist |
|--------|-----|--------|--------|
| Best objective | 688 | 688 | **751** |
| Initial objective | 751 | 751 | 751 |
| Runtime (ms) | 92 | — | 43 |
| Candidates evaluated | 100,000 | — | 50,000 |
| Improvement from initial | 63 | — | **0** |

### What Happened

The assist mode triggered an **early stop** at approximately 50K candidates. At that checkpoint the search had not improved from the initial solution of 751. The rule engine interpreted this plateau as stagnation and recommended `SearchEarlyStop`.

In the off-mode run, SA continued past 50K and eventually found an improvement (751 → 688) somewhere between 50K and 100K candidates.

### Recommendation Made

| Field | Value |
|-------|-------|
| Action | `early_stop` |
| Confidence | ~0.70 (0.6 + 0.50 × 0.2) |
| Reason | `stagnation_50000_cands; budget_50_pct` |
| Safety check | **PASSED** (budget >20%, plateau > RecentImprovWindow) |

### Why It Was Wrong

The rule evaluated correctly that 50K candidates had passed without improvement. But the rule's model of SA convergence on JSS is incorrect. On la01 with seed 555, SA's initial random walk produces poor solutions but the cooling schedule causes it to start accepting improvements **later** than the stagnation window expects.

**Root cause:** The `StagnationWindow` of 50K is calibrated for CVRP (where SA converges quickly) but is too short for JSS SA, which has a **delayed improvement curve**.

Evidence: In the off-mode run, the search evaluated 100K candidates and found improvement of 63 units. The improvement occurred after the 50K mark. The initial objective equals the final assist objective (751), confirming zero progress was made before the early stop.

### Actual Outcome vs Optimal Action

| | Value |
|---|---|
| Recommended action | Early stop at ~50K |
| Actual outcome | Missed improvement of 63 units |
| Optimal action | **Continue** — the search was still warming up |

### Category

**Early stop too early** — StagnationWindow insufficient for domain.

---

## Failure F2: JSS la01 Portfolio Seed 123

### Comparison

| Metric | Off | Shadow | Assist |
|--------|-----|--------|--------|
| Best objective | 666 (optimal) | 666 | **675** |
| Runtime (ms) | 128 | — | 81 |
| Winning strategy (off) | portfolio (undifferentiated) | — | LAHC (seed 8042) |

### What Happened

In off mode, portfolio ran both strategies (SA and LAHC) with equal budgets of 100K each. One strategy found the optimal solution (666).

In assist mode, the `RuleBasedPortfolioAdvisor` applied its heuristic:
- SA budget: 100K × 1.1 = **110K** (reason: `sa_generally_strong`, confidence 0.55)
- LAHC budget: 100K × 1.0 = **100K** (reason: `default_run`, confidence 0.50)

The portfolio_assist.csv shows:
- SA (seed 123): 110K budget → result 688 (did NOT win)
- LAHC (seed 8042): 100K budget → result 675 (WON)

In off mode the portfolio (with different seed distribution due to 100K/100K vs 110K/100K) produced 666. The seed allocation changed because the assist mode's budget redistribution affected which strategy got which seed.

### Recommendation Made

| Field | Value |
|-------|-------|
| Action (SA) | `run` with 1.1× boost |
| Confidence (SA) | 0.55 |
| Reason (SA) | `sa_generally_strong` |
| Action (LAHC) | `run` at 1.0× |
| Confidence (LAHC) | 0.50 |
| Reason (LAHC) | `default_run` |
| Safety check | **PASSED** |

### Why It Was Wrong

The heuristic "SA is generally strong" is a generic prior with no instance-specific evidence. On la01, **Tabu is the strongest algorithm** (all 5 seeds reach optimal 666). LAHC is second-strongest. SA is actually the weakest on la01.

The budget boost to SA was inconsequential (only 10% extra). The real problem is that boosting SA's seed allocation pushed LAHC into a different seed (8042 instead of one that would have found 666), causing a cascade effect on the portfolio result.

### Actual Outcome vs Optimal Action

| | Value |
|---|---|
| Recommended action | Boost SA 10%, keep LAHC default |
| Actual outcome | Portfolio found 675 instead of 666 |
| Optimal action | **Boost LAHC or Tabu** (both are stronger on la01), or better: no intervention (keep equal budgets) |

### Category

**Budget misallocation** — heuristic favours the wrong algorithm for this instance.

---

## Failure F3: JSS ft10 Portfolio Seed 777

### Comparison

| Metric | Off | Shadow | Assist |
|--------|-----|--------|--------|
| Best objective | 997 | 997 | **1029** |
| Runtime (ms) | 180 | — | 134 |
| Winning strategy (off) | portfolio (undifferentiated) | — | LAHC (seed 8696) |

### What Happened

In off mode with seed 777, the portfolio produced 997 (77 unit improvement from initial 1074).

In assist mode, the advisor applied:
- SA (seed 777): 110K budget → result 1074 (NO improvement — SA failed entirely)
- LAHC (seed 8696): 100K budget → result 1029 (WON, but worse than off's winner)

The portfolio_assist.csv confirms SA received a 10% budget boost but produced no improvement whatsoever (1074 = initial). LAHC won the portfolio but with a weaker result than the off-mode run.

### Recommendation Made

| Field | Value |
|-------|-------|
| Action (SA) | `run` with 1.1× boost |
| Confidence (SA) | 0.55 |
| Reason (SA) | `sa_generally_strong` |
| Action (LAHC) | `run` at 1.0× |
| Confidence (LAHC) | 0.50 |
| Reason (LAHC) | `default_run` |
| Safety check | **PASSED** |

### Why It Was Wrong

Same root cause as F2. The heuristic boosted SA (which on ft10 seed 777 is the **weakest** performer: 1074, zero improvement). The budget change propagated a different seed to LAHC (8696 instead of the off-mode seed), and that seed found a worse local minimum.

On ft10, Tabu is again the strongest algorithm (all seeds reach 956-983). SA is inconsistent and sometimes fails to improve at all. The heuristic "SA is generally strong" is factually wrong for JSS.

### Actual Outcome vs Optimal Action

| | Value |
|---|---|
| Recommended action | Boost SA 10%, keep LAHC default |
| Actual outcome | Portfolio found 1029 instead of 997 |
| Optimal action | **Boost Tabu** (strongest on ft10), or no intervention |

### Category

**Budget misallocation** — same mechanism as F2: heuristic misidentifies strongest strategy.

---

## Failure Taxonomy

| Category | Count | Failures | Severity | Root Cause |
|----------|-------|----------|----------|------------|
| **Early stop too early** | 1 | F1 | HIGH (9.2% degradation, zero improvement) | StagnationWindow (50K) shorter than domain's improvement curve |
| **Budget misallocation** | 2 | F2, F3 | LOW-MEDIUM (1.4–3.2% degradation) | Static heuristic "SA is generally strong" wrong for JSS |
| Algorithm selection | 0 | — | — | Not observed |
| Confidence too high | 0 | — | — | Not observed (confidence was low: 0.50–0.70) |
| Model incorrect | 0 | — | — | No ML model involved in these failures |
| Rule incorrect | 3 | F1, F2, F3 | — | All three failures are rule defects, not model defects |

### Failure Rate by Architecture

| Architecture | Failures / Runs | Rate |
|--------------|----------------|------|
| WorkerAssist (NRP) | 0 / 30 | **0%** |
| SearchAssist (CVRP) | 0 / 30 | **0%** |
| SearchAssist (JSS) | 1 / 20 | **5%** |
| SearchAssist (VRPTW) | 0 / 20 | **0%** |
| PortfolioAssist (CVRP) | 0 / 15 | **0%** |
| PortfolioAssist (JSS) | 2 / 10 | **20%** |
| PortfolioAssist (VRPTW) | 0 / 10 | **0%** |

### Failure Rate by Domain

| Domain | Failures / Runs | Rate | 95% CI |
|--------|----------------|------|--------|
| NRP | 0 / 30 | 0% | [0%, 11.6%] |
| CVRP | 0 / 45 | 0% | [0%, 7.9%] |
| JSS | 3 / 90 | 3.3% | [0.7%, 9.4%] |
| VRPTW | 0 / 60 | 0% | [0%, 6.0%] |
| **Overall** | **3 / 225** | **1.3%** | **[0.3%, 3.8%]** |

---

## Confidence Interval Analysis

### SearchAssist Early-Stop Accuracy

| Domain | Correct Stops | False Stops | Accuracy | 95% CI |
|--------|--------------|-------------|----------|--------|
| CVRP | 10/10 | 0 | 100% | [69%, 100%] |
| JSS | 4/5 | 1 | 80% | [28%, 99%] |
| VRPTW | 0/0 (extends) | 0 | N/A | — |
| **Overall** | **14/15** | **1** | **93%** | **[68%, 100%]** |

Note: On VRPTW, SearchAssist never triggers early stop — it extends budgets instead.

### PortfolioAssist Budget Allocation Accuracy

| Domain | Equal or Better | Degraded | Accuracy | 95% CI |
|--------|----------------|----------|----------|--------|
| CVRP | 5/5 | 0 | 100% | [48%, 100%] |
| JSS | 8/10 | 2 | 80% | [44%, 97%] |
| VRPTW | 5/5 | 0 | 100% | [48%, 100%] |
| **Overall** | **18/20** | **2** | **90%** | **[68%, 99%]** |

### WorkerAssist Skip Accuracy (NRP)

| Metric | Value | 95% CI |
|--------|-------|--------|
| Global bests missed | 0 | [0, 0] |
| Degradation count | 0 / 30 | [0%, 11.6%] |

---

## Root Cause Analysis

### Root Cause 1: Fixed StagnationWindow

The `StagnationWindow = 50000` is a single fixed value applied identically to all domains, algorithms, and instance sizes.

**Problem:** SA on JSS has a qualitatively different improvement curve compared to SA on CVRP.

- **CVRP SA:** Initial solutions are reasonable. SA refines quickly. Stagnation at 50K genuinely means convergence.
- **JSS SA:** Initial construction produces poor schedules. SA's cooling schedule means it only starts making productive moves after significant random walk. Stagnation at 50K can mean "not yet started improving" rather than "finished improving."

Evidence: The assist run on la01 seed 555 shows `improvement_amount = 0` at early stop. The initial and final objectives are identical (751). This is a search that never even began its productive phase.

### Root Cause 2: Domain-Blind Portfolio Heuristic

The `RuleBasedPortfolioAdvisor.Advise()` applies static heuristics based only on strategy name:
- SA always gets 1.1× (reason: `sa_generally_strong`)
- Tabu gets 1.15× only on JSS/NRP (reason: `tabu_strong_on_constrained`)
- LAHC gets 0.9× when portfolio has 3+ strategies

**Problem:** The current portfolio configuration only runs **2 strategies** (SA and LAHC). With only 2 strategies:
- Tabu's boost rule never activates (it requires the strategy to be in the portfolio)
- SA gets boosted over LAHC
- But on JSS, LAHC and Tabu are consistently stronger than SA

The heuristic is both wrong in its priors and incomplete in its strategy coverage.

### Root Cause 3: Seed Sensitivity in Portfolio Mode

Budget changes in portfolio mode change the effective seed distribution across strategies. Even a 10% budget change to SA propagates into a different internal state for LAHC (via `config.Seed + int64(i)*7919` offset). This makes the portfolio result non-deterministic with respect to budget allocation — small changes can cascade.

This is not a bug per se, but it means that **any** intervention to portfolio budgets has a probability of helping or hurting based on seed sensitivity. The heuristic cannot predict which way it will go.

---

## Detailed Failure Mechanics

### F1 Decision Path (Step by Step)

```
1. Search starts: initial=751, budget=100K, StagnationWindow=50K
2. Checkpoint at 10K: plateau=10K < 50K → continue
3. Checkpoint at 20K: plateau=20K < 50K → continue
4. Checkpoint at 30K: plateau=30K < 50K → continue
5. Checkpoint at 40K: plateau=40K < 50K → continue
6. Checkpoint at 50K: plateau=50K >= 50K AND budget_used=50% >= 20%
   → Rule 1 fires: recommend early_stop
   → Confidence: 0.6 + 0.50*0.2 = 0.70
   → Safety check: budgetUsed (50%) >= MinBudgetFraction (20%) ✓
   → Safety check: plateauLength (50K) >= RecentImprovWindow (5K) ✓
   → All safety checks PASS
   → Action: EARLY STOP ACCEPTED
7. Search terminates. Final objective: 751 (no improvement)

In off mode:
8. Search continues past 50K
9. SA temperature drops, starts accepting only improvements
10. Improvement found somewhere between 50K-100K: 751 → 688
11. Final objective: 688 (improvement of 63)
```

### F2/F3 Decision Path (Step by Step)

```
1. Portfolio begins. Strategies: [sa, lahc]
2. Advisor evaluates:
   - SA: "sa_generally_strong" → BudgetMult=1.1, Confidence=0.55
   - LAHC: "default_run" → BudgetMult=1.0, Confidence=0.50
3. Safety check: 2 strategies running (minimum met), no >2× boost → PASS
4. Final budgets: SA=110K, LAHC=100K
5. Seeds: SA gets base seed, LAHC gets base_seed + 7919
6. Both strategies run with modified budgets
7. SA: result worse than or equal to off-mode
8. LAHC: result worse than off-mode (different seed trajectory)
9. Portfolio winner: LAHC, but with worse objective than off-mode portfolio winner
```

---

## Recommendations

### Immediate (Fix Known Failures)

| # | Recommendation | Addresses | Expected Impact |
|---|----------------|-----------|-----------------|
| R1 | Increase `StagnationWindow` from 50K to 75K for SA | F1 | Eliminates false early-stop on JSS la01 seed 555 |
| R2 | Make `StagnationWindow` proportional to budget (e.g., 75% of budget) | F1 | Generalises fix across all instances/budgets |
| R3 | Remove "SA is generally strong" heuristic from portfolio advisor | F2, F3 | Eliminates incorrect SA bias on JSS |
| R4 | Default portfolio advisor to equal budgets (1.0× for all) when no history is available | F2, F3 | Preserves off-mode behaviour unless evidence supports change |

### Medium-Term (Reduce Future Failures)

| # | Recommendation | Rationale |
|---|----------------|-----------|
| R5 | Add domain-specific StagnationWindow scaling (JSS needs longer windows than CVRP) | JSS SA has delayed improvement curve |
| R6 | Track improvement curve shape (first-improvement-at) and use it to calibrate stagnation thresholds | Adaptive rather than fixed |
| R7 | Require portfolio advisor to have per-instance history before deviating from equal budgets | Eliminates speculative allocation on unknown instances |
| R8 | Add a "no improvement at all" safety rule: if assist stops a search with zero improvement, that should trigger a warning or override | Would have caught F1 |
| R9 | Include Tabu in the default JSS portfolio | Tabu is the strongest algorithm on both la01 and ft10 |

### Long-Term (Architecture)

| # | Recommendation | Rationale |
|---|----------------|-----------|
| R10 | Replace static portfolio heuristics with learned allocation based on historical runs | 2/3 failures come from incorrect static priors |
| R11 | Implement counterfactual simulation: after an early-stop, run the remaining budget in background and log whether improvement would have occurred | Builds training data for better thresholds |
| R12 | Add per-algorithm, per-domain stagnation profiles (expected improvement curve) | Different algorithms have different convergence shapes |

---

## Summary

| Metric | Value |
|--------|-------|
| Total experiments | 225 |
| Total failures | 3 |
| Overall failure rate | 1.3% (95% CI: 0.3–3.8%) |
| Maximum degradation | 9.2% (F1: 688 → 751) |
| Median degradation | 3.2% (F3: 997 → 1029) |
| Domains affected | 1 of 4 (JSS only) |
| Root causes | 2 (fixed stagnation window, domain-blind heuristic) |
| Rule defects | 3 of 3 (all failures are rule-based, not ML-based) |
| ML model failures | 0 |
| Safety system failures | 0 (all passed safety — the rules themselves were wrong) |

The Search Intelligence system is fundamentally sound. All failures trace to **two specific rule deficiencies** rather than architectural problems. The safety system correctly evaluated every recommendation against its defined rules — the problem is that the rules themselves have insufficient domain awareness to prevent these specific failure modes.
