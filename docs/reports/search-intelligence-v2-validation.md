# Search Intelligence 2.0 — Validation Report

## Status: Not Yet Evaluated

This report will be populated with real experimental data once the validation suite is executed.
No fabricated results exist in this document.

---

## Experiment Design

| Parameter | Value |
|-----------|-------|
| Total experiments | 240 |
| Domains | 4 (NRP, CVRP, JSS, VRPTW) |
| Configurations | 8 (domain × algorithm) |
| Policy modes | 3 (rules, hybrid, learned) |
| Seeds | 10 (42, 123, 555, 777, 999, 1001, 2022, 3033, 4044, 5055) |
| Statistical tests | Welch t-test, Mann-Whitney U, Cohen's d |
| Significance level | α = 0.05 |

---

## Configurations

| Domain | Instance | Algorithm | Iterations |
|--------|----------|-----------|-----------|
| CVRP | A-n32-k5 | SA | 500,000 |
| CVRP | A-n32-k5 | Portfolio | 500,000 |
| JSS | la01 | Tabu | 100,000 |
| JSS | la01 | Portfolio | 100,000 |
| VRPTW | C101 | SA | 100,000 |
| VRPTW | C101 | Portfolio | 100,000 |
| NRP | n012w8 | SA | 100,000 |
| NRP | n012w8 | Portfolio | 100,000 |

---

## Results: CVRP A-n32-k5

### SA (500K iterations)

| Mode | N | Mean | Median | StdDev | 95% CI | Min | Max | Runtime |
|------|---|------|--------|--------|--------|-----|-----|---------|
| rules | — | — | — | — | — | — | — | — |
| hybrid | — | — | — | — | — | — | — | — |
| learned | — | — | — | — | — | — | — | — |

**Not Yet Evaluated**

### Portfolio

| Mode | N | Mean | Median | StdDev | 95% CI | Min | Max | Runtime |
|------|---|------|--------|--------|--------|-----|-----|---------|
| rules | — | — | — | — | — | — | — | — |
| hybrid | — | — | — | — | — | — | — | — |
| learned | — | — | — | — | — | — | — | — |

**Not Yet Evaluated**

---

## Results: JSS la01

### Tabu (100K iterations)

| Mode | N | Mean | Median | StdDev | 95% CI | Min | Max | Runtime |
|------|---|------|--------|--------|--------|-----|-----|---------|
| rules | — | — | — | — | — | — | — | — |
| hybrid | — | — | — | — | — | — | — | — |
| learned | — | — | — | — | — | — | — | — |

**Not Yet Evaluated**

### Portfolio

**Not Yet Evaluated**

---

## Results: VRPTW C101

### SA (100K iterations)

**Not Yet Evaluated**

### Portfolio

**Not Yet Evaluated**

---

## Results: NRP n012w8

### SA (100K iterations)

**Not Yet Evaluated**

### Portfolio

**Not Yet Evaluated**

---

## Statistical Comparisons

### Rules vs Hybrid

| Domain | Algorithm | Mean(Rules) | Mean(Hybrid) | t-stat | p-value | Cohen's d | W/L/T | Verdict |
|--------|-----------|-------------|--------------|--------|---------|-----------|-------|---------|
| CVRP SA | — | — | — | — | — | — | — | Not Yet Evaluated |
| JSS Tabu | — | — | — | — | — | — | — | Not Yet Evaluated |
| VRPTW SA | — | — | — | — | — | — | — | Not Yet Evaluated |
| NRP SA | — | — | — | — | — | — | — | Not Yet Evaluated |

### Rules vs Learned

| Domain | Algorithm | Mean(Rules) | Mean(Learned) | t-stat | p-value | Cohen's d | W/L/T | Verdict |
|--------|-----------|-------------|---------------|--------|---------|-----------|-------|---------|
| — | — | — | — | — | — | — | — | Not Yet Evaluated |

### Hybrid vs Learned

| Domain | Algorithm | Mean(Hybrid) | Mean(Learned) | t-stat | p-value | Cohen's d | W/L/T | Verdict |
|--------|-----------|--------------|---------------|--------|---------|-----------|-------|---------|
| — | — | — | — | — | — | — | — | Not Yet Evaluated |

---

## Policy Metrics

| Domain | Mode | Mean Confidence | Mean Regret | Fallback Rate | Safety Overrides |
|--------|------|-----------------|-------------|---------------|------------------|
| — | — | — | — | — | — |

**Not Yet Evaluated**

---

## Acceptance Criteria

| Criterion | Result |
|-----------|--------|
| Hybrid not statistically worse than rules on any domain | **Not Yet Evaluated** |
| Learned not statistically worse than rules on any domain | **Not Yet Evaluated** |
| Zero feasibility regressions | **Not Yet Evaluated** |
| Confidence calibration error < 15% | **Not Yet Evaluated** |
| Fallback rate < 50% (hybrid should use learned majority of time) | **Not Yet Evaluated** |
| Safety override rate < 5% | **Not Yet Evaluated** |

---

## Reproducibility

```bash
# Execute the full validation suite (from platform/go):
go run ./cmd/owp validate-si2 --output validation/si2

# Or run individual configurations:
owp solve-cvrp --instance examples/cvrp/A-n32-k5.vrp --mode sa --iterations 500000 \
  --policy-mode rules --seed 42 --run-label si2-cvrp-A-n32-k5-sa-rules-s42

owp solve-cvrp --instance examples/cvrp/A-n32-k5.vrp --mode sa --iterations 500000 \
  --policy-mode hybrid --policy-dir policies/ --seed 42 --run-label si2-cvrp-A-n32-k5-sa-hybrid-s42

owp solve-cvrp --instance examples/cvrp/A-n32-k5.vrp --mode sa --iterations 500000 \
  --policy-mode learned --policy-dir policies/ --seed 42 --run-label si2-cvrp-A-n32-k5-sa-learned-s42
```

Seeds: 42, 123, 555, 777, 999, 1001, 2022, 3033, 4044, 5055

---

## Final Verdict

**Not Yet Evaluated**

This report will be updated with real experimental results once the validation suite has been executed across all configurations.
