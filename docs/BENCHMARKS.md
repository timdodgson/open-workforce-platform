# Benchmarks

Reference document for all benchmark datasets supported by the platform.

---

## INRC-II — Nurse Rostering

**Problem:** Assign shifts to nurses over a multi-week horizon. Minimise soft constraint penalties while satisfying coverage, skills, contractual rules, and succession patterns.

**Source:** International Nurse Rostering Competition II (http://mobiz.vives.be/inrc2/)

**Objective:** Minimise total weighted soft constraint penalty. Hard constraints must have zero violations.

**Official best known:** Competition results vary by instance. No single published optimal for most instances.

**Reference solver:** HiGHS ILP (time-limited, provides bound for small instances).

**Instances currently used:**

| Instance | Nurses | Weeks | ILP Reference | Platform Best | Gap |
|----------|--------|-------|---------------|---------------|-----|
| n012w8 | 12 | 8 | 3,020 | 3,465 (Portfolio) | +14.7% |
| n030w4 | 30 | 4 | — | 6,120 (Portfolio) | — |

**Future benchmark plan:**
- Add n030w8, n040w4, n050w4 instances.
- Run ILP on n030w4 (may require long time limits).
- Multi-seed runs (5 seeds per configuration) for statistical rigour.

---

## CVRPLIB — Capacitated Vehicle Routing

**Problem:** Find minimum-distance routes for capacity-limited vehicles serving customers from a depot. Every customer visited exactly once. No route exceeds vehicle capacity.

**Source:** CVRPLIB (http://vrp.atd-lab.inf.puc-rio.br/)

**Objective:** Minimise total Euclidean travel distance (rounded, TSPLIB convention).

**Official best known:** Published optimal solutions for all Augerat Set A instances.

**Reference solver:** Published CVRPLIB optimal values (proven by branch-and-cut).

**Instances currently used:**

| Instance | Customers | Vehicles | Optimal | Platform Best | Gap | Winner |
|----------|-----------|----------|---------|---------------|-----|--------|
| A-n32-k5 | 31 | 5 | 784 | 784 (LAHC) | ✓ optimal | LAHC |
| A-n45-k6 | 44 | 6 | 944 | 968 (LAHC) | +2.5% | LAHC |
| A-n60-k9 | 59 | 9 | 1,354 | 1,358 (Tabu) | +0.3% | Tabu |
| A-n80-k10 | 79 | 10 | 1,763 | 1,826 (Tabu) | +3.6% | Tabu |

**Future benchmark plan:**
- Add Augerat Set B instances (B-n31-k5 through B-n78-k10).
- Add Christofides instances (larger, 100-200 customers).
- Run 1M iteration budget for larger instances.
- Multi-seed statistical analysis.

---

## Solomon — Vehicle Routing with Time Windows

**Problem:** Extend CVRP with time window constraints. Each customer has a service time window [ready, due]. Vehicles must arrive within the window (may wait if early). Late arrivals are infeasible. Each customer has a service duration.

**Source:** Solomon Benchmark Instances (https://www.sintef.no/projectweb/top/vrptw/solomon-benchmark/)

**Objective:** Minimise total travel distance (primary), minimise vehicles used (secondary).

**Official best known:** Published best-known solutions for all Solomon instances.

**Reference solver:** Published Solomon BKS values.

**Instances currently used:**

| Instance | Customers | Type | BKS Distance | BKS Vehicles | Platform Best | Gap | Winner |
|----------|-----------|------|--------------|--------------|---------------|-----|--------|
| C101 | 100 | Clustered, tight TW | 828 | 10 | 829 (LAHC) | +0.1% | LAHC |

**Instance classes:**
- C1/C2: Clustered customers (easier routing, harder time windows).
- R1/R2: Random customers (harder routing, mixed time windows).
- RC1/RC2: Mixed clustered and random.
- Type 1: Tight time windows, small vehicle capacity.
- Type 2: Wide time windows, large vehicle capacity.

**Future benchmark plan:**
- Add R101 (random layout, 100 customers).
- Add RC101 (mixed layout, 100 customers).
- Add C201 (wide time windows, fewer vehicles needed).
- Compare across instance classes to understand algorithm strengths.

---

## Taillard / OR-Library — Job Shop Scheduling

**Problem:** Schedule operations across machines to minimise makespan. Each job has an ordered sequence of operations, each requiring a specific machine for a specific duration. No machine can process two operations simultaneously.

**Source:** OR-Library / Fisher & Thompson / Lawrence instances (http://jobshop.jjvh.nl/)

**Objective:** Minimise makespan (completion time of the last operation).

**Official best known:** Proven optimal solutions for most classic instances.

**Reference solver:** Published optimal values (proven by branch-and-bound or constraint programming).

**Instances currently used:**

| Instance | Jobs | Machines | Optimal | Platform Best | Gap | Winner |
|----------|------|----------|---------|---------------|-----|--------|
| ft06 | 6 | 6 | 55 | 55 (LAHC) | ✓ optimal | LAHC |
| ft10 | 10 | 10 | 930 | 956 (Tabu) | +2.8% | Tabu |
| la01 | 10 | 5 | 666 | 666 (Tabu) | ✓ optimal | Tabu |

**Future benchmark plan:**
- Add ft20 (20 jobs × 5 machines, optimal 1165).
- Add la06-la10 (15 jobs × 5 machines).
- Add ta01-ta10 (Taillard instances, 15×15).
- Increase iteration budget for harder instances.
- Compare with published metaheuristic results from literature.

---

## TSP — Travelling Salesman Problem (Future)

**Problem:** Find the shortest Hamiltonian cycle visiting all cities exactly once.

**Source:** TSPLIB (http://comopt.ifi.uni-heidelberg.de/software/TSPLIB95/)

**Objective:** Minimise total tour length.

**Status:** Not yet implemented. Natural candidate — simpler than CVRP (no capacity, no depot legs). Would validate the generic engine on the most well-studied combinatorial problem.

**Future benchmark plan:**
- Implement TSP as a Problem (trivial — subset of CVRP with single route, no capacity).
- Use TSPLIB instances (att48, eil51, berlin52, st70, pr76).
- Compare with published LKH results.

---

## Platform Performance Summary

| Domain | Instances | At Optimal | Best Gap | Worst Gap | Avg Gap |
|--------|-----------|------------|----------|-----------|---------|
| CVRP | 4 | 1 | +0.3% | +3.6% | +1.6% |
| JSS | 3 | 2 | +0% | +2.8% | +0.9% |
| VRPTW | 1 | 0 | +0.1% | +0.1% | +0.1% |
| NRP | 2 | 0 | +14.7% | — | — |

**Overall average gap to reference: +2.4%** (where reference is available).

---

## Search Intelligence Benchmark Ladder

Search Intelligence v1 validated on tested configurations (267 runs, 5 seeds each).
Modes: `off` (baseline), `shadow` (record-only), `assist` (active advisory).

### CVRP — SA Mode

| Instance | Mode | Mean Objective | Gap to BKS | Mean Runtime | Compute Saved |
|----------|------|----------------|------------|--------------|---------------|
| A-n32-k5 | off | 801.2 | +2.2% | 78ms | — |
| A-n32-k5 | shadow | 801.2 | +2.2% | 78ms | 0% |
| A-n32-k5 | **assist** | **801.2** | **+2.2%** | **30ms** | **62%** |
| A-n60-k9 | off | 1410.8 | +4.2% | 88ms | — |
| A-n60-k9 | shadow | 1410.8 | +4.2% | 88ms | 0% |
| A-n60-k9 | **assist** | **1410.8** | **+4.2%** | **40ms** | **55%** |
| A-n80-k10 | off | 1865.6 | +5.8% | 93ms | — |
| A-n80-k10 | shadow | 1865.6 | +5.8% | 94ms | 0% |
| A-n80-k10 | **assist** | **1865.6** | **+5.8%** | **48ms** | **48%** |

### CVRP — Portfolio Mode

| Instance | Mode | Mean Objective | Gap to BKS | Degradations |
|----------|------|----------------|------------|--------------|
| A-n32-k5 | off | 786.4 | +0.3% | — |
| A-n32-k5 | **assist** | **786.4** | **+0.3%** | **0/5** |
| A-n60-k9 | off | 1387.6 | +2.5% | — |
| A-n60-k9 | **assist** | **1387.6** | **+2.5%** | **0/5** |
| A-n80-k10 | off | 1810.8 | +2.7% | — |
| A-n80-k10 | **assist** | **1808.2** | **+2.6%** | **0/5** |

### JSS — Tabu Mode

| Instance | Mode | Mean Objective | Gap to Optimal | Mean Runtime | Compute Saved |
|----------|------|----------------|----------------|--------------|---------------|
| la01 | off | 666.0 | ✓ optimal | 4501ms | — |
| la01 | shadow | 666.0 | ✓ optimal | 4521ms | 0% |
| la01 | **assist** | **666.0** | **✓ optimal** | **2709ms** | **40%** |
| ft10 | off | 968.2 | +4.1% | 7114ms | — |
| ft10 | shadow | 968.2 | +4.1% | 6988ms | 0% |
| ft10 | **assist** | **968.2** | **+4.1%** | **5105ms** | **28%** |

### JSS — Portfolio Mode

| Instance | Mode | Mean Objective | Gap to Optimal | Degradations |
|----------|------|----------------|----------------|--------------|
| la01 | off | 683.2 | +2.6% | — |
| la01 | **assist** | **681.0** | **+2.3%** | **1/5 (+9)** |
| ft10 | off | 1009.0 | +8.5% | — |
| ft10 | **assist** | **1015.4** | **+9.2%** | **1/5 (+32)** |

### VRPTW — SA Mode

| Instance | Mode | Mean Objective | Gap to BKS | Improvement |
|----------|------|----------------|------------|-------------|
| C101 | off | 1116 | +34.8% | — |
| C101 | shadow | 1116 | +34.8% | 0% |
| C101 | **assist** | **985** | **+19.0%** | **12% better** |

### VRPTW — Portfolio Mode

| Instance | Mode | Mean Objective | Gap to BKS | Improvement |
|----------|------|----------------|------------|-------------|
| C101 | off | 975 | +17.8% | — |
| C101 | shadow | 975 | +17.8% | 0% |
| C101 | **assist** | **915** | **+10.5%** | **6% better** |

### NRP — SA Mode (WorkerAssist)

| Instance | Mode | Mean Penalty | Assist vs Off |
|----------|------|--------------|---------------|
| n012w8 | off | 3882 | — |
| n012w8 | **assist** | **3928** | **+1.2%** (within variance) |
| n030w4 | off | 6045 | — |
| n030w4 | **assist** | **5998** | **-0.8%** (better) |
| n030w8 | off | 11148 | — |
| n030w8 | **assist** | **11076** | **-0.6%** (better) |

### NRP — Portfolio Mode

| Instance | Mode | Mean Penalty | Assist vs Off |
|----------|------|--------------|---------------|
| n012w8 | off | 3912 | — |
| n012w8 | **assist** | **3891** | **-0.5%** (better) |
| n030w4 | off | 6045 | — |
| n030w4 | **assist** | **6075** | **+0.5%** (within variance) |
| n030w8 | off | 11081 | — |
| n030w8 | **assist** | **11223** | **+1.3%** (slightly worse) |

### Ladder Summary

| Domain | Best Mode | Quality Impact | Compute Savings | Verdict |
|--------|-----------|----------------|-----------------|---------|
| CVRP | SA + assist | Equal | **48–62%** | ✅ SAFE |
| CVRP | Portfolio + assist | Equal | Moderate | ✅ SAFE |
| JSS | Tabu + assist | Equal (optimal) | **33–40%** | ✅ SAFE |
| JSS | Portfolio + assist | Mixed (±) | — | ⚠️ Needs tuning |
| VRPTW | SA + assist | **12% better** | Trades compute for quality | ✅✅ BENEFICIAL |
| VRPTW | Portfolio + assist | **6% better** | Trades compute for quality | ✅✅ BENEFICIAL |
| NRP | SA + assist | -0.6% to +1.2% | — | ✅ SAFE (within variance) |
| NRP | Portfolio + assist | -0.5% to +1.3% | — | ✅ SAFE (within variance) |

---

## Statistical Validation (10 seeds, Welch t-test)

Final release evidence. 320 runs, 10 seeds per configuration, all 4 modes.

| Domain | Algorithm | Off Mean | Adaptive Mean | p-value | Cohen's d | Compute Saved | Verdict |
|--------|-----------|----------|---------------|---------|-----------|---------------|---------|
| CVRP | SA | 802.8 | 802.8 | 1.000 | 0.000 | **60%** | ✅ SAFE |
| CVRP | Portfolio | 790.8 | 790.8 | 1.000 | 0.000 | **73%** | ✅ SAFE |
| JSS | Tabu | 666.0 | 666.0 | 1.000 | 0.000 | **41%** | ✅ SAFE |
| JSS | Portfolio | 682.3 | 678.8 | 0.237 | -0.354 | **23%** | ✅ SAFE |
| NRP | SA | 3943.5 | 3922.0 | 0.183 | -0.225 | — | ✅ SAFE |
| NRP | Portfolio | 3964.5 | 3910.0 | 0.215 | -0.632 | — | ✅ SAFE |
| VRPTW | SA | 1137.6 | 923.2 | <0.001 | -2.91 | Quality trade | ✅✅ BETTER |
| VRPTW | Portfolio | 972.5 | 861.3 | <0.001 | -3.45 | Quality trade | ✅✅ BETTER |

**Release verdict: SAFE. No degradation at 95% confidence. Significant improvements on VRPTW.**

---

## Notes

- All platform results use 500K iterations (CVRP a32k5), 200K (CVRP a60k9/a80k10), 100K (JSS/VRPTW/NRP), seed 42, default parameters.
- Search Intelligence benchmark uses 5 seeds (42, 123, 555, 777, 999) per configuration.
- Statistical validation uses 10 seeds (42, 123, 555, 777, 999, 1001, 2022, 3033, 4044, 5055).
- Gaps are calculated as (platform - reference) / reference × 100%.
- "✓ optimal" means the platform found the proven optimal solution.
- NRP gap is higher because the ILP reference itself has a 56% optimality gap (time-limited solve).
- Compute savings measured as 1 - (assist runtime / off runtime).
- VRPTW assist/adaptive extends budget (longer runtime) for better solutions — a deliberate quality-over-speed trade-off.
- Cohen's d: small (0.2), medium (0.5), large (0.8), very large (>1.0).
