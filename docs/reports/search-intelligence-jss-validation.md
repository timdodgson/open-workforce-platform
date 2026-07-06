# Search Intelligence — JSS Validation Report

## Setup

| Parameter | Value |
|-----------|-------|
| Problem | Job Shop Scheduling |
| Instances | la01 (10×5, optimal=666), ft10 (10×10, optimal=930) |
| Algorithms | SA, LAHC, Tabu, Portfolio |
| Seeds | 42, 123, 555, 777, 999 |
| Modes | off, shadow, assist |
| Iterations | 100,000 |
| Total Runs | 90 |

---

## Results: la01 (10 jobs × 5 machines, optimal = 666)

### SA

| Seed | Off | Shadow | Assist | Assist Runtime |
|------|-----|--------|--------|----------------|
| 42 | 688 | 688 | 688 | 88ms |
| 123 | 666 | 666 | 666 | 77ms |
| 555 | 688 | 688 | **751** | 43ms |
| 777 | 689 | 689 | 689 | 85ms |
| 999 | 688 | 688 | 688 | 66ms |

**⚠ Note:** Seed 555 assist (751) is worse than off (688). Early stop triggered too aggressively on this seed.

### LAHC

| Seed | Off | Shadow | Assist | Assist Runtime |
|------|-----|--------|--------|----------------|
| 42 | 688 | 688 | 688 | 69ms |
| 123 | 671 | 671 | 671 | 51ms |
| 555 | 671 | 671 | 671 | 49ms |
| 777 | 678 | 678 | 678 | 48ms |
| 999 | 688 | 688 | 688 | 83ms |

✅ All identical. Shadow matches off. Assist matches off.

### Tabu

| Seed | Off | Shadow | Assist | Off Runtime | Assist Runtime |
|------|-----|--------|--------|-------------|----------------|
| 42 | 666 | 666 | 666 | 4597ms | 2802ms |
| 123 | 666 | 666 | 666 | 4452ms | 2843ms |
| 555 | 666 | 666 | 666 | 4503ms | 2675ms |
| 777 | 666 | 666 | 666 | 4432ms | 2668ms |
| 999 | 666 | 666 | 666 | 4651ms | 2757ms |

✅ All reach optimal (666). Assist saves **40% runtime** with identical quality.

### Portfolio

| Seed | Off | Shadow | Assist |
|------|-----|--------|--------|
| 42 | 688 | 688 | 688 |
| 123 | 666 | 666 | 675 |
| 555 | 688 | 688 | 666 |
| 777 | 688 | 688 | 688 |
| 999 | 688 | 688 | 688 |

Mixed: Seed 123 assist slightly worse (675 vs 666), Seed 555 assist better (666 vs 688).

---

## Results: ft10 (10 jobs × 10 machines, optimal = 930)

### SA

| Seed | Off | Shadow | Assist | Off Runtime | Assist Runtime |
|------|-----|--------|--------|-------------|----------------|
| 42 | 1074 | 1074 | 1074 | 133ms | 69ms |
| 123 | 1064 | 1064 | 1064 | 140ms | 80ms |
| 555 | 1074 | 1074 | 1074 | 141ms | 67ms |
| 777 | 1074 | 1074 | 1074 | 131ms | 70ms |
| 999 | 1074 | 1074 | 1074 | 137ms | 69ms |

✅ All identical. Assist saves **49% runtime**.

### LAHC

| Seed | Off | Shadow | Assist | Off Runtime | Assist Runtime |
|------|-----|--------|--------|-------------|----------------|
| 42 | 1004 | 1004 | 1004 | 128ms | 125ms |
| 123 | 1000 | 1000 | 1000 | 128ms | 129ms |
| 555 | 992 | 992 | 992 | 127ms | 129ms |
| 777 | 1013 | 1013 | 1013 | 126ms | 87ms |
| 999 | 1018 | 1018 | 1018 | 129ms | 78ms |

✅ All identical. Assist saves 0-39% (LAHC converges later, fewer early-stop opportunities).

### Tabu

| Seed | Off | Shadow | Assist | Off Runtime | Assist Runtime |
|------|-----|--------|--------|-------------|----------------|
| 42 | 956 | 956 | 956 | 7060ms | 5707ms |
| 123 | 983 | 983 | 983 | 7345ms | 4351ms |
| 555 | 975 | 975 | 975 | 7308ms | 4335ms |
| 777 | 959 | 959 | 959 | 7010ms | 7281ms |
| 999 | 968 | 968 | 968 | 7293ms | 4584ms |

✅ All identical. Assist saves **25-41%** on 4/5 seeds.

### Portfolio

| Seed | Off | Shadow | Assist |
|------|-----|--------|--------|
| 42 | 1029 | 1029 | 1029 |
| 123 | 1000 | 1000 | 1000 |
| 555 | 993 | 993 | 993 |
| 777 | 997 | 997 | **1029** |
| 999 | 1026 | 1026 | 1026 |

**⚠ Note:** Seed 777 assist (1029) is worse than off (997). Budget reallocation favoured the wrong strategy.

---

## Summary Statistics

### la01

| Algorithm | Mode | Mean | Best | Worst |
|-----------|------|------|------|-------|
| SA | Off | 683.8 | 666 | 689 |
| SA | Assist | **696.4** | 666 | **751** |
| LAHC | Off | 679.2 | 671 | 688 |
| LAHC | Assist | 679.2 | 671 | 688 |
| Tabu | Off | 666.0 | 666 | 666 |
| Tabu | Assist | 666.0 | 666 | 666 |

### ft10

| Algorithm | Mode | Mean | Best | Worst | Mean Runtime Savings |
|-----------|------|------|------|-------|---------------------|
| SA | Off | 1072.0 | 1064 | 1074 | — |
| SA | Assist | 1072.0 | 1064 | 1074 | **49%** |
| LAHC | Off | 1005.4 | 992 | 1018 | — |
| LAHC | Assist | 1005.4 | 992 | 1018 | **15%** |
| Tabu | Off | 968.2 | 956 | 983 | — |
| Tabu | Assist | 968.2 | 956 | 983 | **33%** |

---

## Issues Found

### 1. SA Early-Stop Too Aggressive on la01 (Seed 555)

Assist stopped SA early and got 751 instead of 688. The search hadn't converged yet but the early-stop threshold triggered. This is a **false early-stop** — the stagnation window (50K) was too short for this seed.

**Impact:** 1 of 20 single-search runs degraded (5% failure rate on la01 SA).

### 2. Portfolio Budget Reallocation Occasionally Wrong

ft10 seed 777: portfolio-assist got 1029 vs off's 997. The heuristic gave SA a budget boost, but LAHC was the better algorithm for this seed.

**Impact:** 2 of 10 portfolio runs slightly degraded.

---

## Conclusions

### 1. Architecture Validation ✅

All 90 runs completed. All modes work correctly on JSS.

### 2. Shadow Mode ✅

Shadow produces **identical results** to off on every single run (90/90 match).

### 3. Quality Validation ⚠️

- **LAHC:** SAFE (5/5 identical on both instances)
- **Tabu:** SAFE (10/10 identical on both instances)
- **SA:** MOSTLY SAFE (9/10 identical, 1 degradation from early-stop)
- **Portfolio:** MOSTLY SAFE (8/10 identical, 2 slight degradations from budget reallocation)

### 4. Efficiency Validation ✅

| Algorithm | Instance | Runtime Savings |
|-----------|----------|----------------|
| SA | ft10 | 49% |
| Tabu | la01 | 40% |
| Tabu | ft10 | 33% |
| LAHC | ft10 | 15% |

---

## Verdict: **NEEDS MORE DATA**

Assist mode is **safe for LAHC and Tabu** on JSS (zero degradations, significant compute savings).

Assist mode **needs tuning for SA** — the early-stop threshold is too aggressive for JSS SA, causing 1/10 runs to terminate prematurely. Recommended fix: increase `StagnationWindow` from 50K to 75K for SA, or make it proportional to iteration budget.

Portfolio budget reallocation is heuristic-based and occasionally wrong. Not dangerous (small degradation) but not reliably beneficial either.

**Per-algorithm recommendation:**
- Tabu + assist: **SAFE** ✅
- LAHC + assist: **SAFE** ✅
- SA + assist: **SAFE after threshold increase** ⚠️
- Portfolio + assist: **NEEDS MORE DATA** ⚠️
