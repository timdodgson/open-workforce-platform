# Experiments

Laboratory notebook for the PFRS Research Lab. Records all significant experiments with reproducible parameters and observed results.

---

## Template

```
### EXP-XXX: [Title]

**Date:** YYYY-MM-DD
**Problem:** [NRP | CVRP | JSS | VRPTW]
**Instance:** [Instance name]
**Algorithm:** [SA | LAHC | Tabu | Portfolio | Adaptive]
**Parameters:**
- Iterations: X
- Workers: X
- Temperature: X
- Cooling rate: X
- Seed(s): X

**Hardware:** [CPU, cores, RAM]
**Runtime:** Xms / Xs

**Objective:**
- Best: X
- Initial: X
- Improvement: X%
- Feasible: Yes/No
- Gap to optimal: X%

**Statistical analysis:**
- [Comparison with other runs, significance, effect size]

**Observations:**
- [What was learned]

**Next experiment:**
- [What to try based on these results]
```

---

## Completed Experiments

### EXP-001: NRP Portfolio (n012w8, full beam search)

**Date:** 2026-07-04
**Problem:** NRP
**Instance:** n012w8 (12 nurses, 8 weeks)
**Algorithm:** Portfolio (SA + LAHC + Tabu)
**Parameters:**
- Iterations per worker: 1,500,000
- Max total workers: 32
- Beam width: 12
- Beam seeds: 12
- Look-ahead weight: 0.3
- Final window: 2 weeks
- Cooling: adaptive

**Hardware:** 32-core Windows, 64GB RAM
**Runtime:** 38.5s

**Objective:**
- Best: 3,465
- Gap to ILP (3,020): +14.7%
- Hard violations: 0
- Total candidates: 283.5M

**Statistical analysis:**
- Portfolio consistently outperforms SA alone (3,465 vs 3,565 best, ~100 penalty improvement).
- Look-ahead + final window coupling contributes ~50 penalty reduction vs portfolio without.

**Observations:**
- Portfolio wins by covering more of the search space — LAHC often escapes plateaus SA gets stuck on.
- Week 8 consistently worst (1,575 penalty) due to accumulated history constraints.
- 62.2% hard reject rate — most moves violate skills or succession rules.
- 14.68% acceptance of worse solutions — SA temperature is well-calibrated.

**Next experiment:**
- Try larger iteration budget (3M per worker) to see if week 8 improves.
- Test n030w4 instance to measure scaling behaviour.

---

### EXP-002: CVRP SA (A-n32-k5, A-n45-k6, A-n60-k9, A-n80-k10)

**Date:** 2026-07-05
**Problem:** CVRP
**Instance:** A-n32-k5 (31 customers), A-n45-k6 (44), A-n60-k9 (59), A-n80-k10 (79)
**Algorithm:** Simulated Annealing
**Parameters:**
- Iterations: 500,000
- Temperature: 100.0
- Cooling: adaptive
- Seed: 42

**Hardware:** 32-core Windows
**Runtime:** 80-200ms per instance

**Objective:**

| Instance | Optimal | SA Result | Gap |
|----------|---------|-----------|-----|
| A-n32-k5 | 784 | 814 | +3.8% |
| A-n45-k6 | 944 | 969 | +2.6% |
| A-n60-k9 | 1,354 | 1,423 | +5.1% |
| A-n80-k10 | 1,763 | 1,873 | +6.2% |

**Statistical analysis:**
- Gap increases with instance size (3.8% → 6.2%) as expected.
- All solutions feasible (no capacity violations).
- Constructive baseline improvement: 27-35% across instances.

**Observations:**
- SA performs well on smaller instances but struggles on n80 where the search space is larger.
- 500K iterations in <200ms shows the implementation is efficient.
- Adaptive cooling eliminates the need to manually tune cooling rate per instance.

**Next experiment:**
- Compare with LAHC and Tabu on same instances.
- Increase iterations to 1M for larger instances.

---

### EXP-003: CVRP LAHC (A-n32-k5, A-n45-k6, A-n60-k9, A-n80-k10)

**Date:** 2026-07-05
**Problem:** CVRP
**Instance:** A-n32-k5, A-n45-k6, A-n60-k9, A-n80-k10
**Algorithm:** Late Acceptance Hill Climbing
**Parameters:**
- Iterations: 500,000
- Late acceptance length: 1,000
- Seed: 42

**Hardware:** 32-core Windows
**Runtime:** 80-200ms per instance

**Objective:**

| Instance | Optimal | LAHC Result | Gap |
|----------|---------|-------------|-----|
| A-n32-k5 | 784 | 784 | ✓ optimal |
| A-n45-k6 | 944 | 968 | +2.5% |
| A-n60-k9 | 1,354 | 1,421 | +4.9% |
| A-n80-k10 | 1,763 | 1,853 | +5.1% |

**Statistical analysis:**
- LAHC hits optimal on A-n32-k5 (SA does not).
- LAHC beats SA on every instance except A-n60-k9 where results are similar.
- Average gap 3.1% vs SA's 4.4%.

**Observations:**
- LAHC's ability to accept solutions that are only slightly worse than L steps ago gives it better exploration.
- On A-n32-k5, LAHC finds optimal — the fitness history allows systematic exploration of the small search space.
- Late acceptance length of 1000 appears well-suited for CVRP instances at 500K iterations.

**Next experiment:**
- Test Tabu search which uses a fundamentally different strategy (best-move neighbourhood).

---

### EXP-004: CVRP Tabu (A-n32-k5, A-n45-k6, A-n60-k9, A-n80-k10)

**Date:** 2026-07-05
**Problem:** CVRP
**Instance:** A-n32-k5, A-n45-k6, A-n60-k9, A-n80-k10
**Algorithm:** Tabu Search (best-move, neighbourhood sampling)
**Parameters:**
- Iterations: 500,000
- Tabu tenure: 7
- Neighbourhood sample: 100
- Seed: 42

**Hardware:** 32-core Windows
**Runtime:** 500ms-2s per instance (slower due to best-move evaluation)

**Objective:**

| Instance | Optimal | Tabu Result | Gap |
|----------|---------|-------------|-----|
| A-n32-k5 | 784 | 796 | +1.5% |
| A-n45-k6 | 944 | 1,018 | +7.8% |
| A-n60-k9 | 1,354 | 1,358 | +0.3% |
| A-n80-k10 | 1,763 | 1,826 | +3.6% |

**Statistical analysis:**
- Tabu wins on A-n60-k9 (+0.3% vs LAHC's +4.9%) and A-n80-k10 (+3.6% vs LAHC's +5.1%).
- Tabu loses badly on A-n45-k6 (+7.8%) — tenure/neighbourhood may need tuning for this instance size.
- Best-move strategy is more expensive per iteration but finds better moves when the search space is large.

**Observations:**
- Tabu's strength is on larger instances where systematic neighbourhood exploration pays off.
- The 100-sample neighbourhood is a compromise — larger samples would improve quality but increase runtime.
- Tabu tenure of 7 may be too short for A-n45-k6 (allows cycling in the mid-size space).
- Runtime is 5-10× slower than SA/LAHC due to best-move evaluation (clone + evaluate per candidate).

**Next experiment:**
- Portfolio mode to combine SA/LAHC/Tabu — get best of each.
- Consider adaptive Tabu tenure based on instance size.

---

### EXP-005: JSS ft06 (all algorithms)

**Date:** 2026-07-05
**Problem:** JSS
**Instance:** ft06 (6 jobs × 6 machines, optimal 55)
**Algorithm:** SA, LAHC, Tabu, Portfolio
**Parameters:**
- Iterations: 500,000
- Seed: 42
- SA temperature: 100, adaptive cooling
- LAHC length: 1000
- Tabu tenure: 7, neighbourhood: 50

**Hardware:** 32-core Windows
**Runtime:** 100-900ms

**Objective:**

| Algorithm | Makespan | Gap |
|-----------|----------|-----|
| SA | 57 | +3.6% |
| LAHC | 55 | ✓ optimal |
| Tabu | 57 | +3.6% |
| Portfolio | 55 | ✓ optimal |

**Statistical analysis:**
- LAHC and Portfolio both reach optimal (55) on this small instance.
- SA and Tabu get close (57) but don't find the final 2-unit improvement.

**Observations:**
- ft06 is a well-known easy instance — optimal is reachable by most metaheuristics.
- LAHC's exploration behaviour works well on small JSS instances.
- Portfolio wins by including LAHC in its strategy set.

**Next experiment:**
- Test ft10 (10×10, optimal 930) which is significantly harder.
- Test la01 (10×5, optimal 666) for a different machine/job ratio.

---

### EXP-006: VRPTW C101 (all algorithms)

**Date:** 2026-07-05
**Problem:** VRPTW
**Instance:** C101 (100 customers, clustered, best known 828)
**Algorithm:** SA, LAHC, Tabu, Portfolio
**Parameters:**
- Iterations: 500,000
- Seed: 42

**Hardware:** 32-core Windows
**Runtime:** 100-600ms

**Objective:**

| Algorithm | Distance | Vehicles | Gap to BKS |
|-----------|----------|----------|------------|
| SA | 885 | 11 | +6.9% |
| LAHC | 829 | 10 | +0.1% |
| Tabu | 957 | 12 | +15.6% |
| Portfolio | 885 | 11 | +6.9% |

**Statistical analysis:**
- LAHC achieves near-optimal (829 vs BKS 828, +0.1% gap).
- Tabu performs poorly on VRPTW — time window constraints significantly reduce the feasible neighbourhood.
- Portfolio returns SA's result (SA was the winner in this portfolio run).

**Observations:**
- LAHC excels on VRPTW C101 — the clustered customer layout means good solutions are reachable with local moves.
- Tabu's best-move strategy struggles because many moves violate time windows, reducing the effective neighbourhood.
- 10 vehicles used (LAHC) vs 25 available — significant route consolidation achieved.
- The +0.1% gap from BKS (828) shows the algorithm is highly competitive on this instance class.

**Next experiment:**
- Test R101 (random customers) where the structure is different.
- Increase iterations to 1M to see if LAHC can reach exactly 828.


---

### EXP-007: Search Intelligence Statistical Validation (All Domains)

**Date:** 2026-07-07
**Problem:** NRP, CVRP, JSS, VRPTW
**Instances:** n012w8, A-n32-k5, la01, C101
**Algorithm:** Best per domain (SA, Tabu) + Portfolio
**Modes:** off, shadow, assist, adaptive
**Parameters:**
- CVRP: 500K iterations
- JSS: 100K iterations
- VRPTW: 100K iterations
- NRP: 100K × 16 workers
- Seeds: 42, 123, 555, 777, 999, 1001, 2022, 3033, 4044, 5055 (10 per config)

**Hardware:** Windows, 32-core
**Total runs:** 320

**Key Results:**

| Domain | Assist vs Off | Adaptive vs Off | Compute Saved |
|--------|---------------|-----------------|---------------|
| CVRP SA | Identical (p=1.0) | Identical (p=1.0) | 59-60% |
| CVRP Portfolio | Identical (p=1.0) | Identical (p=1.0) | 73% |
| JSS Tabu | Identical (p=1.0) | Identical (p=1.0) | 40-41% |
| JSS Portfolio | Better (p=0.24, NS) | Better (p=0.24, NS) | 17-23% |
| NRP SA | Better (p=0.19, NS) | Neutral (p=0.18) | — |
| NRP Portfolio | **Better (p=0.036)** | Neutral (p=0.22) | — |
| VRPTW SA | Better (p=0.19, NS) | **Better (p<0.001)** | Trades for quality |
| VRPTW Portfolio | **Better (p=0.006)** | **Better (p<0.001)** | Trades for quality |

**Statistical analysis:**
- Welch t-test (two-tailed) used for all comparisons
- Effect sizes (Cohen's d): VRPTW adaptive d=-2.91 to -3.45 (very large)
- No test shows assist/adaptive worse than off at any significance level
- Shadow perfectly neutral on all deterministic solvers (CVRP, JSS, VRPTW)

**Observations:**
- Adaptive mode's primary value is on VRPTW where budget extension produces dramatically better solutions
- On converged problems (CVRP, JSS Tabu), assist and adaptive produce identical results to off but faster
- NRP variance is too high for 10 seeds to reach significance on SA, but trend is positive
- Zero feasibility violations across all 320 runs
- All safety invariants held (no missed bests, no unsafe stops)

**Verdict:** SAFE FOR RELEASE

**Next experiment:**
- 20-seed validation on larger NRP instances (n030w8) to confirm trends
- Adaptive with learned portfolio model on JSS to fix the SA-bias issue

---

### EXP-008: SI2 Policy Validation Cell (CVRP A-n32-k5, hybrid, seed 42)

**Date:** 2026-07-11  
**Problem:** CVRP  
**Instance:** A-n32-k5 (31 customers, optimal 784)  
**Algorithm:** Simulated Annealing  
**Policy mode:** `hybrid` (rules + learned checkpoints)  
**Parameters:**
- Iterations: 500,000
- Seed: 42
- Policy dir: `platform/ml/policies`
- Storage: S3 (`pfrs-research-lab-data`)

**Label:** `val-cvrp-a32k5-sa-hybrid-s42`  
**Matrix cell:** `fast-cvrp-sa` × policy `hybrid` × seed `42` (1 of 30 in this config)

**Reproduce:**

```powershell
cd platform/go
go run ./cmd/owp solve cvrp `
  --instance ../../examples/cvrp/A-n32-k5.vrp `
  --mode sa --iterations 500000 `
  --policy-mode hybrid --policy-dir ../ml/policies `
  --seed 42 --run-label val-cvrp-a32k5-sa-hybrid-s42 `
  --storage s3
```

**Artifacts (per run directory):**

| File | Purpose |
|------|---------|
| `run.json` | Objective, runtime, config fingerprint |
| `policy_decisions.csv` | Checkpoint/restart decisions |
| `policy_evaluation.csv` | Policy quality metrics |
| `policy_learning_report.json` | Post-run learning recommendation |
| `generic_search_assist.csv` | Search assist telemetry |

**Verify:**

1. Lab: `https://pfrs-lab.com/runs/val-cvrp-a32k5-sa-hybrid-s42/summary`
2. Matrix: `/experiment-matrix` → `fast-cvrp-sa` should count this label toward 30/30
3. CLI: `go run ./cmd/owp validate-si2 analyze --prefix val-cvrp-a32k5-sa --runs-dir ../web/pfrs-lab/data/runs` (after local sync)
4. Gap audit: `npm run audit-val-matrix` — see [val-gap-audit.md](./reports/val-gap-audit.md)

**Full suite:** Run all 288 cells via `validate-si2.ps1` + `validate-si2-deep.ps1`. Production snapshot (2026-07-11): **288/288 complete** on S3.

**Observations:**
- Canonical entry point for R&D reviewers — one command, full artifact chain, matrix traceability
- Hybrid mode is the default production policy path (`--policy-mode hybrid`)
- Identical seeds across `rules` / `hybrid` / `learned` enable paired policy comparisons

**Next experiment:**
- Re-run single cell after policy retrain and diff `policy_evaluation.csv`
- Extend to `val-cvrp-a32k5-portfolio-hybrid-s42` for portfolio budget policy evidence
