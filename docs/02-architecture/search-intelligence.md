# Search Intelligence Architecture

## Purpose

Search Intelligence is a universal AI advisory system for optimisation solvers.
It allows AI to advise any solver architecture without changing core behaviour.

The system supports three integration styles, each suited to a different solver:

| Style | Solver Architecture | Domain | Status |
|-------|-------------------|--------|--------|
| WorkerAssist | Beam search (parallel workers) | NRP | ✅ Implemented, validated on tested configurations |
| PortfolioAssist | Multi-strategy portfolio | CVRP, JSS, VRPTW, NRP | ✅ Implemented, learned model + rule-based fallback |
| SearchAssist | Single-algorithm search | CVRP, JSS, VRPTW | ✅ Implemented, validated on tested configurations |

---

## Design Principles

1. **Universal** — one concept works across all domains and architectures
2. **Optional** — default is off, existing behaviour unchanged
3. **Safe** — hard safety rules override AI recommendations
4. **Observable** — every recommendation is logged for analysis
5. **Gradual** — shadow mode records without acting, assist mode acts with overrides

---

## Modes

| Mode | Active | Changes Behaviour | Records |
|------|--------|-------------------|---------|
| `off` | No | No | Nothing |
| `shadow` | Yes | No | Predictions + outcomes |
| `assist` | Yes | Yes (static checkpoints, safety overrides) | Predictions + outcomes + actions |
| `adaptive` | Yes | Yes (live-updating, learned models) | Predictions + outcomes + actions |

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

**Safety rules:**
- Never skip all strategies (at least 2 must run in 3+ portfolio)
- Never allocate below 0.25× or above 2× base budget
- Require confidence threshold (0.60) for learned model recommendations
- Fall back to rule-based if model confidence is low

**Integration point:** `RunPortfolioWithAssist()` in `optimisation/portfolio_assist.go`

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

**Safety rules:**
- Never early-stop before minimum budget (20% of allocated)
- Never stop immediately after an improvement
- Budget adjustments bounded by safety floor/ceiling

**Integration point:** `SearchHookRunner` in `optimisation/search_assist_hooks.go`

**Output:** `generic_search_assist.csv`

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

All three integration styles are implemented and validated on tested configurations.

### WorkerAssist (Complete)

- `platform/go/internal/infrastructure/inrc2/worker_decision.go` — rule engine
- `platform/go/internal/infrastructure/inrc2/worker_assist.go` — assist recorder + safety
- `platform/go/internal/infrastructure/inrc2/pfrs_search.go` — integration in submitWork()
- CLI: `--worker-decision-mode off|shadow|assist|adaptive`
- Dashboard: `/intelligence` (Assist Validation tab)
- Validated on tested configurations: NRP SA, NRP Portfolio

### SearchAssist (Complete)

- `platform/go/internal/optimisation/search_assist_hooks.go` — rule-based engine + hook runner
- `platform/go/internal/optimisation/adaptive_search_assist.go` — adaptive mode (live-updating)
- Integrated in `RunSearch()` via `SearchHookRunner`
- CLI: `--worker-decision-mode off|shadow|assist|adaptive`
- Validated on tested configurations: CVRP SA/LAHC, JSS Tabu, VRPTW SA

### PortfolioAssist (Complete)

- `platform/go/internal/optimisation/portfolio_assist.go` — rule-based + learned allocation
- `platform/go/internal/optimisation/portfolio_budget_model.go` — learned model with fallback
- Integrated via `RunPortfolioWithAssist()`
- CLI: `--worker-decision-mode off|shadow|assist|adaptive`, `--portfolio-model <path>`
- Validated on tested configurations: CVRP Portfolio, JSS Portfolio, VRPTW Portfolio, NRP Portfolio

---

## Validation

320 statistical validation runs. 10 seeds. 4 domains. Welch t-test at 95% confidence.

Validated on tested configurations. Not claimed universal.

See `docs/reports/search-intelligence-statistical-validation.md` for full evidence.
