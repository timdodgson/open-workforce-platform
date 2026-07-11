# Open Workforce Platform — Multi-Domain Optimisation Research Lab

A research platform for combinatorial optimisation, supporting multiple problem domains through a unified search engine and shared analytics dashboard.

## Supported Domains

| Domain | Problem | Benchmark | Status |
|--------|---------|-----------|--------|
| **NRP** | Nurse Rostering (INRC-II) | n012w8 (12 nurses, 8 weeks) | Production — best result 3,465 |
| **CVRP** | Capacitated Vehicle Routing | CVRPLIB (EUC_2D instances) | Active development |
| **JSS** | Job Shop Scheduling | Taillard / OR-Library | Active development |
| **VRPTW** | Vehicle Routing with Time Windows | Solomon C101 (100 customers) | Active development |
| **ILP** | Integer Linear Programming baseline | HiGHS solver | Benchmarking only |

## Architecture

```
Problems          NRP · CVRP · JSS · VRPTW
                           ↓
Interface         CreateInitialSolution · TryMove · Evaluate · Undo · Serialize
                           ↓
Algorithms        SA · LAHC · Tabu · Portfolio · Adaptive
                           ↓
Search Intelligence   Off · Shadow · Assist · Adaptive
                      WorkerAssist · SearchAssist · PortfolioAssist
                           ↓
Telemetry         run.json · discoveries.csv · worker_learning.csv · portfolio_assist.csv
                           ↓
Learning          worker_model.json · portfolio_budget_model.json · Feature Importance
                           ↓
Storage           Local Filesystem · S3 (versioned) · Manifest Index
                           ↓
Dashboard         Benchmarks · Statistics · Search Intelligence · Route Viewer · Gantt
                           ↓
Research Outputs  Validation Reports · Gap Analysis · Statistical Evidence
```

The optimiser knows nothing about nurses, vehicles, or any specific domain. It operates entirely through the `Problem` interface. Each domain provides:
- Solution representation
- Move generation and validation
- Objective evaluation
- Hard constraint checking
- Serialisation for the dashboard

## Search Algorithms

All algorithms operate through the same generic interface. They work identically on NRP and CVRP.

| Algorithm | Acceptance Criterion | Key Parameter |
|-----------|---------------------|---------------|
| **SA** | Metropolis: P(accept worse) = e^(-Δ/T) | Initial temperature |
| **LAHC** | Accept if ≤ current OR ≤ fitness[v] | Buffer length |
| **Tabu** | Accept non-tabu moves; aspiration for global best | Tenure |
| **Portfolio** | Run all strategies, keep the best | Strategy list |
| **Adaptive** | SA primary + LAHC escape bursts on stagnation | Stagnation window |

### NRP-Specific Extensions

NRP adds on top of the generic engine:
- **Beam Search** — multi-history planning across 8 weeks
- **Portfolio Branching** — spawn workers per strategy on global best
- **Look-ahead** — amortized global constraint bias for beam ranking
- **Diversity Slots** — preserve underrepresented beam families
- **Refinement** — violation-count post-processing pass

## Search Intelligence

Search Intelligence has **two layers** that stack — not an old/new version pair.

| Layer | CLI flag | What it does | Status |
|-------|----------|--------------|--------|
| **Assist** | `--worker-decision-mode off\|shadow\|assist\|adaptive` | Rule-based WorkerAssist / SearchAssist / PortfolioAssist checkpoints | **Production** — 320-run statistical validation (40–73% compute saved on CVRP/JSS) |
| **Policies** | `--policy-mode rules\|hybrid\|learned` | Learned stagnation/restart/budget JSON policies distilled to Go trees | **Production** — 12/12 active policies, val-* harness + counterfactual gates |

Use **both** on NRP: `--worker-decision-mode assist --policy-mode hybrid`. Assist handles worker spawn safety; Policies handle stagnation/restart timing.

**Assist modes:** `--worker-decision-mode off|shadow|assist|adaptive`

| Mode | Behaviour | Use Case |
|------|-----------|----------|
| `off` | No AI, existing behaviour unchanged | Default |
| `shadow` | Records recommendations, no behaviour change | Data collection |
| `assist` | Applies safe recommendations (static checkpoints) | Production |
| `adaptive` | Live-updating decisions based on search progress | Advanced |

**Policy layer:** `--policy-mode rules|hybrid|learned` with `--policy-dir ../ml/policies` (defaults when `--policy-mode` is set)

| Policy mode | Behaviour |
|-------------|-----------|
| `rules` | Rule-based checkpoints only (assist layer when worker mode set) |
| `hybrid` | Learned stagnation/restart when confident; rules fallback |
| `learned` | Learned policies for all search decisions |

`--policy-mode` and `--worker-decision-mode` work together: policy mode controls learned JSON policies; worker-decision mode controls assist recording and safety behaviour.

**Integration styles:**

| Style | Solver | Actions |
|-------|--------|---------|
| WorkerAssist | NRP beam search | Skip/reduce/increase workers |
| SearchAssist | SA/LAHC/Tabu (single) | Early stop, budget extend |
| PortfolioAssist | All portfolio modes | Learned budget allocation |

**Validated results (320 runs, 10 seeds, Welch t-test):**

| Domain | Adaptive vs Off | Compute Saved | Verdict |
|--------|----------------|---------------|---------|
| CVRP | Identical quality | **60-73%** | ✅ SAFE |
| JSS | Identical quality | **41%** | ✅ SAFE |
| NRP | Within variance | — | ✅ SAFE |
| VRPTW | **19% better** (p<0.001) | Trades for quality | ✅✅ BETTER |

Zero feasibility regressions. Zero missed bests. All safety invariants hold.

**Policy layer guide:** [docs/SEARCH_INTELLIGENCE_V2.md](docs/SEARCH_INTELLIGENCE_V2.md)

### ILP Baseline

The ILP solver (HiGHS) provides optimal/near-optimal solutions for small instances. This establishes the optimality gap for heuristic methods. It is not a scalable solver — it's a benchmark reference.

- ILP baseline (n012w8, 5hr): **3,020**
- Best PFRS (portfolio+lookahead+fw2, beam 12, 1.5M iter): **3,465**
- Gap to optimality: ~14.7%

## Quick Start

### Run NRP (Nurse Rostering)

```bash
cd platform/go

# Single run with SA
go run ./cmd/owp tune-pfrs --pfrs-mode sa --pfrs-beam-width 5 \
  --pfrs-beam-seeds 42,101,202 --pfrs-run-label my-nrp-run

# Portfolio mode (SA + LAHC + Tabu)
go run ./cmd/owp tune-pfrs --pfrs-mode portfolio \
  --pfrs-portfolio sa,lahc,tabu --pfrs-beam-width 12 \
  --pfrs-beam-seeds 42,101,202,303,404 --iterations 1500000 \
  --pfrs-beam-strategy budget --pfrs-lookahead-weight 4.0 \
  --pfrs-final-window-weeks 2 --pfrs-run-label portfolio-full

# Upload to S3 for deployed dashboard
go run ./cmd/owp tune-pfrs --pfrs-storage s3 --pfrs-run-label my-run ...

# Policy layer — PFRS worker policy (learned worker_policy.json)
go run ./cmd/owp tune-pfrs --instance n012w8 \
  --worker-decision-mode assist --policy-mode hybrid \
  --pfrs-iterations-per-worker 30000 --pfrs-max-total-workers 8 \
  --pfrs-run-label si2-pfrs --pfrs-storage local
```

### Run CVRP (Vehicle Routing)

```bash
cd platform/go

# SA on a CVRPLIB instance
go run ./cmd/owp solve cvrp --instance ../../examples/cvrp/A-n32-k5.vrp \
  --mode sa --iterations 500000 --seed 42

# LAHC
go run ./cmd/owp solve cvrp --instance ../../examples/cvrp/A-n32-k5.vrp \
  --mode lahc --iterations 500000 --late-acceptance-length 1000

# Tabu
go run ./cmd/owp solve cvrp --instance ../../examples/cvrp/A-n32-k5.vrp \
  --mode tabu --iterations 500000 --tabu-tenure 7

# Portfolio (compare all algorithms)
go run ./cmd/owp solve cvrp --instance ../../examples/cvrp/A-n32-k5.vrp \
  --mode portfolio --portfolio sa,lahc,tabu --iterations 500000

# Adaptive (SA with LAHC escape on stagnation)
go run ./cmd/owp solve cvrp --instance ../../examples/cvrp/A-n32-k5.vrp \
  --mode adaptive --iterations 500000

# Save telemetry for dashboard
go run ./cmd/owp solve cvrp --instance ../../examples/cvrp/A-n32-k5.vrp \
  --mode sa --iterations 500000 --run-label cvrp-a32k5-sa

# Policy layer hybrid (learned stagnation + restart; writes policy_decisions.csv)
go run ./cmd/owp solve cvrp --instance ../../examples/cvrp/A-n32-k5.vrp \
  --mode sa --iterations 500000 --policy-mode hybrid --seed 42 \
  --run-label si2-cvrp-hybrid --storage local

# Policy layer portfolio (learned budget allocation + per-strategy search policies)
go run ./cmd/owp solve cvrp --instance ../../examples/cvrp/A-n32-k5.vrp \
  --mode portfolio --iterations 500000 --policy-mode hybrid --seed 42 \
  --run-label si2-cvrp-portfolio --storage local
```

### Run JSS (Job Shop Scheduling)

```bash
cd platform/go

# SA on Fisher & Thompson ft06 (6 jobs × 6 machines, optimal = 55)
go run ./cmd/owp solve jss --instance internal/infrastructure/jobshop/testdata/ft06.txt \
  --mode sa --iterations 500000 --seed 42

# Save to dashboard
go run ./cmd/owp solve jss --instance internal/infrastructure/jobshop/testdata/ft06.txt \
  --mode sa --iterations 500000 --run-label jss-ft06-sa --storage s3
```

### Run VRPTW (Vehicle Routing with Time Windows)

```bash
cd platform/go

go run ./cmd/owp solve vrptw --instance ../../examples/vrptw/C101.txt \
  --mode portfolio --iterations 500000 --run-label vrptw-c101-portfolio --storage s3
```

### Run NRP (single week via SDK)

Multi-week beam search still uses `tune-pfrs`. For a single INRC-II week:

```bash
cd platform/go

go run ./cmd/owp solve nrp --instance n012w8 \
  --mode sa --iterations 500000 --run-label nrp-n012w8-sa --storage local
```

### Run ILP Benchmark

```bash
cd platform/go

# 5-hour ILP solve with S3 upload
go run ./cmd/owp benchmark-ilp --instance n012w8 --weeks 8 \
  --time-limit 18000 --storage s3 --run-label ilp-n012w8-8w
```

### Train ML Model

```bash
cd platform/ml

# Train worker value model from local run data
python -m worker_model.train --data-dir ../web/pfrs-lab/data --output worker_model.json

# Train and upload to S3 (deployed dashboard picks it up automatically)
python -m worker_model.train --data-dir ../web/pfrs-lab/data --output worker_model.json --storage s3

# Generate per-worker predictions
python -m worker_model.predict --data-dir ../web/pfrs-lab/data --output worker_predictions.json

# Generate and upload to S3
python -m worker_model.predict --data-dir ../web/pfrs-lab/data --output worker_predictions.json --storage s3
```

Requires: `pip install -e .` (for scikit-learn, pandas). For S3 upload: `pip install -e ".[s3]"` (adds boto3).

### View Dashboard (Local)

```bash
cd platform/web/pfrs-lab
npm install
npm run dev
# Open http://localhost:3000
```

### Deploy Infrastructure

```bash
cd platform/infra
npm install
npx cdk deploy PfrsResearchLabStack   # S3 bucket
npx cdk deploy DashboardStack         # ECS Fargate + ALB
```

## Telemetry Files

Each run with `--run-label` produces:

| File | Purpose |
|------|---------|
| `run.json` | Metadata (problemType, mode, instance, parameters) |
| `results.csv` | Per-week/per-run summary metrics |
| `discoveries.csv` | Every global best improvement with timing |
| `solution.json` | Final solution (roster or routes) |
| `tree.csv` | Beam search ancestry (NRP only) |
| `diversity.csv` | Hamming distance / fingerprints (NRP only) |
| `workers.csv` | Worker lifecycle data |
| `improvements.csv` | Global best update events |
| `plateaus.csv` | Stagnation detection events |
| `roster.json` | NRP schedule viewer data |

CVRP runs emit `run.json`, `results.csv`, `discoveries.csv`, and `solution.json`. The dashboard detects `problemType` and shows appropriate pages.

## Dashboard Pages

### Universal (all problem types)
- **Summary** — Key metrics and objective value
- **Search Progress** — Discovery timeline, improvement rate
- **Statistics** — Cross-run comparison with t-tests and box plots
- **Compare** — Head-to-head A vs B analysis
- **Trends** — Regression analysis across experiments

### NRP-Specific
- **Schedule** — Full nurse roster grid
- **Timeline** — Event-by-event visualisation
- **Beam Tree** — Ancestry and pruning history
- **Diversity** — Hamming distance and beam health
- **Landscape** — Fitness landscape scatter
- **Workers** — Parallel worker lifecycle
- **Insights** — AI-generated analysis

### CVRP-Specific
- **Route Viewer** — Vehicle routes with capacity utilisation
- **Timeline** — Search event timeline
- **Search Map** — Discovery geography

## Storage

| Backend | Config | Use Case |
|---------|--------|----------|
| Local | `STORAGE_PROVIDER=local` | Development |
| S3 | `STORAGE_PROVIDER=s3` | Production (ECS dashboard) |

S3 bucket: `pfrs-research-lab-data` (eu-west-1, versioned, intelligent tiering).

Manifest-based listing — no full bucket scan needed. Delete removes from manifest only; files preserved in versioned bucket.

## Project Structure

```
platform/
├── go/
│   ├── cmd/owp/                         # CLI (owp solve, tune-pfrs, benchmark-ilp)
│   └── internal/
│       ├── optimisation/
│       │   ├── problem.go               # Generic Problem interface
│       │   └── search.go                # SA / LAHC / Tabu / Portfolio engine
│       └── infrastructure/
│           ├── inrc2/                   # NRP domain (INRC-II)
│           │   ├── nrp_problem.go       # Problem interface implementation
│           │   ├── pfrs_search.go       # NRP-specific workers + branching
│           │   ├── pfrs_beam.go         # Multi-week beam search
│           │   └── scorer.go            # Official INRC-II validation
│           ├── cvrp/                    # CVRP domain
│           │   ├── problem.go           # Problem interface implementation
│           │   ├── neighbourhood.go     # Relocate / Swap / IntraSwap / 2-Opt / Or-opt
│           │   ├── constructive.go      # Nearest-neighbour initial solution
│           │   ├── scorer.go            # Distance + capacity validation
│           │   └── loader/              # CVRPLIB parser (TSPLIB format)
│           ├── jobshop/                 # JSS domain
│           │   ├── problem.go           # Problem interface implementation
│           │   ├── constructive.go      # SPT dispatch rule
│           │   ├── scorer.go            # Makespan + precedence/overlap validation
│           │   └── loader.go            # Taillard/OR-Library format parser
│           └── ilp/                     # ILP baseline (HiGHS)
├── web/pfrs-lab/                        # Next.js dashboard
│   ├── src/app/runs/[id]/              # Per-run pages
│   ├── src/app/statistics/             # Cross-run analysis
│   └── src/lib/                        # Data loading, CSV parsing
└── infra/                              # AWS CDK (S3, ECS, Cognito)
```

## Dependencies

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.22+ | Algorithm + CLI |
| Node.js | 20+ | Dashboard |
| HiGHS | 1.15+ | ILP benchmark (optional) |
| AWS CDK | 2.x | Infrastructure |
| AWS CLI | 2.x | S3 upload |

## Key Results (NRP, n012w8)

| Configuration | Mean Penalty | Best | Significance |
|--------------|-------------|------|--------------|
| Portfolio+lookahead+fw2 (6 runs) | 3,567 | 3,465 | Baseline |
| SA (4 runs) | 3,583 | 3,565 | p=0.60 vs portfolio |
| LAHC (3 runs) | 3,630 | 3,630 | p=0.01 vs portfolio (✓) |
| Tabu standalone (4 runs) | 8,976 | 5,395 | p<0.001 (✗) |
| ILP (5hr, single-threaded) | — | 3,020 | Optimal reference |

Portfolio with lookahead and final-window coupling achieves the lowest mean penalty. Tabu is weak standalone but adds value as a diversifier within portfolio mode.


## Exact Benchmarks / ILP

### Purpose

Integer Linear Programming provides **provably optimal** (or bounded) solutions for small problem instances. ILP is not intended to replace heuristic solvers for production use — it exists purely as a research reference:

1. **Establish the optimality gap** — how far are heuristic solutions from the mathematical optimum?
2. **Validate the constraint model** — does the ILP formulation produce feasible solutions according to the official scorer?
3. **Calibrate parameters** — a heuristic that achieves <15% gap may not benefit from further tuning.

### How It Works

The ILP model encodes all INRC-II hard and soft constraints as linear inequalities. The objective function minimises total weighted soft constraint violations. HiGHS solves the mixed-integer program using branch-and-bound with LP relaxation.

For n012w8 (12 nurses, 8 weeks, all constraints):
- **Variables**: ~6,700 binary decision variables
- **Constraints**: ~15,000 linear constraints
- **Solve time**: 5 hours (single-threaded HiGHS static build)
- **Result**: Objective 3,020 / Lower bound 1,933 / Gap 56.23%

The ILP gap (56%) means HiGHS proved the true optimum is somewhere between 1,933 and 3,020. It could not close this gap within the time limit — the problem is genuinely hard even for exact methods.

### Comparison Framework

| Method | Objective | Gap to ILP | Runtime | Scalability |
|--------|-----------|-----------|---------|-------------|
| ILP (HiGHS) | 3,020 | — | 5 hours | n012 only |
| PFRS (portfolio) | 3,465 | +14.7% | 38 seconds | n120+ |
| PFRS (SA baseline) | 3,583 | +18.6% | 5 seconds | n120+ |
| Tabu standalone | 5,395 | +78.6% | 5 seconds | n120+ |

The heuristic is 500× faster and scales to instances 10× larger. The 14.7% gap represents the price of scalability.

### Running ILP Benchmarks

```bash
cd platform/go

# Quick test (5 minutes, 1 week)
go run ./cmd/owp benchmark-ilp --instance n012w8 --weeks 1 --time-limit 300

# Full benchmark (5 hours, 8 weeks, upload to S3)
go run ./cmd/owp benchmark-ilp --instance n012w8 --weeks 8 \
  --time-limit 18000 --storage s3 --run-label ilp-n012w8-8w

# Compare against heuristic
go run ./cmd/owp benchmark-ilp --instance n012w8 --weeks 8 \
  --time-limit 18000 --compare-pfrs 3465 --compare-pfrs-runtime 38.5
```

### Dashboard

ILP runs appear in the dashboard with their own page set:
- **Summary** — metadata and basic stats
- **Solve Progress** — objective/bound convergence over time (if progress CSV exists)
- **Schedule** — roster grid (if roster.json was generated)
- **Heuristic Gap** — side-by-side comparison table with PFRS best

The Solve Progress page shows the ILP incumbent and lower bound converging over time. Early progress is fast; the final gap closure is exponentially hard.

### HiGHS Installation

1. Download from [github.com/ERGO-Code/HiGHS/releases](https://github.com/ERGO-Code/HiGHS/releases)
2. **Windows**: Use the Apache static build (includes parallel MIP)
3. Add `bin/` to PATH
4. Verify: `highs --version`

The native binary is required. No Python wrapper.

### When to Use ILP

✅ Use ILP when:
- You need a proven lower bound
- The instance is small (≤30 nurses)
- You have hours of compute time
- You're calibrating heuristic parameters

❌ Don't use ILP when:
- The instance is large (>30 nurses)
- You need results in seconds
- You're running parameter sweeps
- You're comparing algorithm designs (use heuristics with fixed iteration budgets)
