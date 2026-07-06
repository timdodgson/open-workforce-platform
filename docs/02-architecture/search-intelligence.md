# Search Intelligence Architecture

## Purpose

Search Intelligence is a universal AI advisory system for optimisation solvers.
It allows AI to advise any solver architecture without changing core behaviour.

The system supports three integration styles, each suited to a different solver:

| Style | Solver Architecture | Domain | Status |
|-------|-------------------|--------|--------|
| WorkerAssist | Beam search (parallel workers) | NRP | ✅ Implemented, validated safe |
| PortfolioAssist | Multi-strategy portfolio | CVRP, JSS, VRPTW | 📐 Interface defined |
| SearchAssist | Single-algorithm search | All | 📐 Interface defined |

---

## Design Principles

1. **Universal** — one concept works across all domains and architectures
2. **Optional** — default is off, existing behaviour unchanged
3. **Safe** — hard safety rules override AI recommendations
4. **Observable** — every recommendation is logged for analysis
5. **Gradual** — shadow mode records without acting, assist mode acts with overrides

---

## Modes

| Mode | AI Active | Changes Behaviour | Records |
|------|-----------|-------------------|---------|
| `off` | No | No | Nothing |
| `shadow` | Yes | No | Predictions + outcomes |
| `assist` | Yes | Yes (with safety) | Predictions + outcomes + actions |

---

## Integration Styles

### 1. WorkerAssist (NRP Beam Search)

**Decision point:** Each time a worker is about to be submitted to the work queue.

**Actions available:**
- `run` — submit worker as normal
- `skip` — do not submit (save CPU)
- `reduce_budget` — run with fewer iterations
- `increase_budget` — run with more iterations
- `change_algorithm` — run with a different algorithm

**Safety rules:**
- Never skip global-best lineage workers
- Never skip with low confidence (< 0.65)
- Never skip workers near the global best

**Integration point:** `submitWork()` in `pfrs_search.go`

**Output:** `worker_assist.csv`

---

### 2. PortfolioAssist (CVRP, JSS, VRPTW)

**Decision points:**
- Before portfolio run: allocate iteration budgets
- During run (optional): terminate/extend strategies
- After run: record outcomes for learning

**Actions available:**
- `allocate` — set iteration budgets per strategy
- `skip_strategy` — don't run one algorithm
- `terminate` — stop a running strategy early
- `extend` — give more iterations to a promising strategy
- `restart` — restart a strategy with a different seed
- `adjust_params` — change temperature/LAHC/tabu parameters

**Safety rules (proposed):**
- Never skip all strategies (at least 2 must run)
- Never allocate less than minimum viable budget
- Never terminate the currently-best strategy

**Integration point:** `RunPortfolio()` in `optimisation/portfolio.go`

**Output:** `portfolio_assist.csv`

---

### 3. SearchAssist (Single Algorithm)

**Decision points:**
- Periodically during search (every N iterations)
- When stagnation is detected
- When improvement rate changes significantly

**Actions available:**
- `continue` — no change
- `early_stop` — terminate search
- `restart` — restart from best known solution
- `adjust_temp` — change SA temperature
- `adjust_lahc` — change LAHC buffer length
- `adjust_tabu` — change tabu tenure
- `adjust_budget` — change remaining iterations

**Safety rules (proposed):**
- Never early-stop before minimum budget (20% of allocated)
- Never restart if already improving
- Parameter adjustments bounded to ±50% of current value

**Integration point:** search loop in `RunSearch()`

**Output:** `search_assist.csv`

---

## Data Flow

```
┌─────────────────────────────────────────────────┐
│              Search Intelligence                 │
│                                                  │
│  ┌──────────────┐  ┌───────────────┐  ┌───────┐ │
│  │ WorkerAssist │  │PortfolioAssist│  │Search │ │
│  │   (NRP)      │  │ (CVRP/JSS)   │  │Assist │ │
│  └──────┬───────┘  └──────┬────────┘  └───┬───┘ │
└─────────┼──────────────────┼───────────────┼─────┘
          │                  │               │
          ▼                  ▼               ▼
   ┌──────────┐      ┌──────────┐    ┌──────────┐
   │ PFRS     │      │ Portfolio│    │  Search  │
   │ Beam     │      │  Runner  │    │  Loop    │
   │ Search   │      │          │    │          │
   └──────────┘      └──────────┘    └──────────┘
          │                  │               │
          ▼                  ▼               ▼
   worker_assist.csv  portfolio_assist.csv  search_assist.csv
```

---

## Implementation Status

### WorkerAssist (Complete)

- `platform/go/internal/infrastructure/inrc2/worker_decision.go` — rule engine
- `platform/go/internal/infrastructure/inrc2/worker_assist.go` — assist recorder + safety
- `platform/go/internal/infrastructure/inrc2/pfrs_search.go` — integration in submitWork()
- CLI: `--worker-decision-mode assist`
- Dashboard: `/assist`
- Validated: NRP SA (SAFE), NRP Portfolio (SAFE)

### PortfolioAssist (Interface Only)

- `platform/go/internal/optimisation/search_intelligence.go` — interface defined
- Integration point identified: `RunPortfolio()`
- Not yet connected to any solver

### SearchAssist (Interface Only)

- `platform/go/internal/optimisation/search_intelligence.go` — interface defined
- Integration point identified: `RunSearch()` loop
- Not yet connected to any solver

---

## Next Steps

1. Implement `PortfolioAssist` for CVRP (budget allocation based on historical performance)
2. Wire `PortfolioAssist` into `RunPortfolio()` with shadow mode first
3. Validate on CVRP A-n32-k5 (shadow → assist progression)
4. Implement `SearchAssist` for SA (early-stop and restart detection)
5. Add `--search-intelligence` flag as universal replacement for `--worker-decision-mode`
