# Open Workforce Platform — Multi-Domain Optimisation Research Lab

A research platform for combinatorial optimisation: unified Go search engine, Search Intelligence policies, S3 telemetry, and a public Next.js lab ([pfrs-lab.com](https://pfrs-lab.com)).

## Supported Domains

| Domain | Problem | Benchmark | Status |
|--------|---------|-----------|--------|
| **NRP** | Nurse Rostering (INRC-II) | n012w8 (12 nurses, 8 weeks) | Flagship — best published PFRS **3,440** (~13.9% vs ILP feasible 3,020) |
| **CVRP** | Capacitated Vehicle Routing | CVRPLIB (EUC_2D) | Production — typically within ~0–4% of BKS |
| **JSS** | Job Shop Scheduling | Taillard / OR-Library | Production — often at optimal on small instances |
| **VRPTW** | Vehicle Routing with Time Windows | Solomon C101 | Production — ~0.1% of BKS on C101 |
| **ILP** | Integer Linear Programming baseline | HiGHS | Benchmarking / bounds only |
| **BYOD** | Bring-your-own domain | `owp-sdk` + examples | Extensible — see `/lab/byod` and `examples/byod-*` |

## Architecture

```
Problems          NRP · CVRP · JSS · VRPTW
                           ↓
Interface         CreateInitialSolution · TryMove · Evaluate · Undo · Serialize
                           ↓
Search             Metaheuristics (SA · LAHC · Tabu) · Population-based (GA)
                   · Portfolio · Adaptive · NRP beam
                           ↓
Search Intelligence   Assist (off/shadow/assist/adaptive)
                      · Policies (rules/hybrid/learned)
                           ↓
Telemetry         run.json · discoveries.csv · worker_learning.csv · policy_decisions.csv
                           ↓
Learning          worker / budget / restart policies · Feature Importance
                           ↓
Storage           Local filesystem · S3 (versioned) · Manifest index
                           ↓
Lab (OpenNext)    Getting Started · Benchmarks · Statistics · SI · BYOD · Assistant
                           ↓
Research Outputs  Validation reports · Gap analysis · Statistical evidence
```

The optimiser knows nothing about nurses, vehicles, or any specific domain. It operates entirely through the `Problem` interface. Each domain provides:
- Solution representation
- Move generation and validation
- Objective evaluation
- Hard constraint checking
- Serialisation for the dashboard

## Search Algorithms

All algorithms share the same generic `Problem` interface. **SA / LAHC / Tabu** are single-trajectory metaheuristics; **GA** is population-based (elite + crossover + mutation). Portfolio runs them together (default `sa,lahc,tabu,ga`).

| Algorithm | Kind | Acceptance / mechanism | Key parameter |
|-----------|------|------------------------|---------------|
| **SA** | Metaheuristic | Metropolis: P(accept worse) = e^(-Δ/T) | Initial temperature |
| **LAHC** | Metaheuristic | Accept if ≤ current OR ≤ fitness[v] | Buffer length |
| **Tabu** | Metaheuristic | Non-tabu moves; aspiration for global best | Tenure |
| **GA** | Population-based | Elite ∪ crossover ∪ mutate | Population size |
| **Portfolio** | Composite | Parallel strategies, keep the best | Strategy list |
| **Adaptive** | Composite | SA primary + LAHC escape on stagnation | Stagnation window |

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
| SearchAssist | SA/LAHC/Tabu/GA (single) | Early stop, budget extend |
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

- ILP best feasible (n012w8, published reference): **3,020**
- Best PFRS (SI hybrid + diversity 30% + fw 6M, beam 12, 3M/worker): **3,440**
- Gap to ILP feasible: ~13.9% (do not confuse with the dual/MIP lower bound ~1,845)
- Prior best (portfolio+lookahead+fw2, 1.5M): 3,465

## Quick Start

### Run NRP (Nurse Rostering)

```bash
cd platform/go

# Single run with SA
go run ./cmd/owp tune-pfrs --pfrs-mode sa --pfrs-beam-width 5 \
  --pfrs-beam-seeds 42,101,202 --pfrs-run-label my-nrp-run

# Portfolio mode (SA + LAHC + Tabu + GA)
go run ./cmd/owp tune-pfrs --pfrs-mode portfolio \
  --pfrs-portfolio sa,lahc,tabu,ga --pfrs-beam-width 12 \
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

# Portfolio (compare all algorithms; default includes GA)
go run ./cmd/owp solve cvrp --instance ../../examples/cvrp/A-n32-k5.vrp \
  --mode portfolio --portfolio sa,lahc,tabu,ga --iterations 500000

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
go run ./cmd/owp solve jobshop --instance internal/infrastructure/jobshop/testdata/ft06.txt \
  --mode sa --iterations 500000 --seed 42

# Save to dashboard
go run ./cmd/owp solve jobshop --instance internal/infrastructure/jobshop/testdata/ft06.txt \
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

### View Lab (Local)

```bash
cd platform/web/pfrs-lab
npm install
npm run dev
# Open http://localhost:3000
```

**Live site:** [pfrs-lab.com](https://pfrs-lab.com)

| Path | What |
|------|------|
| `/` | Public marketing home |
| [`/getting-started`](https://pfrs-lab.com/getting-started) | 5-minute CLI Quick start |
| [`/lab`](https://pfrs-lab.com/lab) | Live lab hub (runs, benchmarks, SI, …) |
| [`/lab/byod`](https://pfrs-lab.com/lab/byod) | Bring-your-own domain / algorithm |
| [`/reproduce`](https://pfrs-lab.com/reproduce) | Cite + learning path |
| [`/experiments/chat`](https://pfrs-lab.com/experiments/chat) | Research assistant (Cognito sign-in) |

Production dashboard is **OpenNext on CloudFront + Lambda** (SST). Cognito still gates Admin and the assistant. Deploy path: GitHub Actions → semantic-release → `npx sst deploy --stage production` (see [docs/OPENNEXT_MIGRATION.md](docs/OPENNEXT_MIGRATION.md)).

### Deploy Infrastructure

```bash
# Primary (production lab): from CI or locally
cd platform/web/pfrs-lab
npx sst deploy --stage production

# Supporting CDK stacks (S3 research bucket, Cognito / legacy ECS)
cd platform/infra
npm install
npx cdk deploy PfrsResearchLabStack   # S3 bucket
npx cdk deploy DashboardStack         # Cognito (+ optional legacy ECS)
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

## Lab / Dashboard Pages

### Site & lab hub
- **Getting Started** — Quick start + essential/advanced CLI reference
- **BYOD / BYOA** — Copy-paste extend examples
- **Capabilities / Experiment Matrix** — What runs where, and which knobs matter
- **Assistant** — Cognito-gated experiment planner (Anthropic API; key in SSM)
- **Admin** — Release metadata, assistant config/prompt, token usage (authenticated)

### Universal (all problem types)
- **Summary** — Key metrics and objective value
- **Search Progress** — Discovery timeline, improvement rate
- **Statistics** — Cross-run comparison with t-tests and box plots
- **Compare** — Head-to-head A vs B analysis
- **Trends** — Regression analysis across experiments
- **Search Intelligence** — Assist / policy dashboards

### NRP-Specific
- **Schedule** — Full nurse roster grid
- **Timeline** — Event-by-event visualisation
- **Beam Tree** — Ancestry and pruning history
- **Diversity** — Hamming distance and beam health
- **Landscape** — Fitness landscape scatter
- **Workers** — Parallel worker lifecycle

### CVRP / VRPTW / JSS
- **Route Viewer** — Vehicle routes with capacity utilisation (CVRP/VRPTW)
- **Gantt** — Machine schedules (JSS)
- **Timeline / Search Map** — Discovery views where applicable

## Storage

| Backend | Config | Use Case |
|---------|--------|----------|
| Local | `STORAGE_PROVIDER=local` | Development |
| S3 | `STORAGE_PROVIDER=s3` | Production (OpenNext Lambda + CLI uploads) |

S3 bucket: `pfrs-research-lab-data` (eu-west-1, versioned, intelligent tiering).

Manifest-based listing — no full bucket scan needed. Delete removes from manifest only; files preserved in versioned bucket.

## Project Structure

```
platform/
├── go/
│   ├── cmd/owp/                         # CLI (solve, tune-pfrs, benchmark-ilp, …)
│   └── internal/
│       ├── optimisation/                # Problem interface + SA/LAHC/Tabu/GA/Portfolio + SI
│       ├── sdk/                         # Built-in + BYOD domain registration
│       └── infrastructure/
│           ├── inrc2/                   # NRP (beam, portfolio workers, official scorer)
│           ├── cvrp/ · vrptw/ · jobshop/
│           └── ilp/                     # HiGHS baseline
├── ml/                                  # Policy training, validation harness
├── owp-sdk/                             # External BYOD Go module
├── web/pfrs-lab/                        # Next.js lab (OpenNext / SST)
│   ├── sst.config.ts                    # CloudFront + Lambda production
│   ├── src/app/                         # Site, lab, admin, assistant API
│   └── src/lib/assistant-prompt.md      # Research assistant system prompt
└── infra/                               # CDK: S3 research bucket, Cognito, legacy ECS
examples/
├── byod-tsp/ · byod-byoa/               # Extend-the-platform demos
└── cvrp/ · vrptw/ · inrc2/              # Benchmark instances
docs/                                    # Architecture, ADRs, benchmarks, SI guides
```

## Dependencies

| Tool | Version | Purpose |
|------|---------|---------|
| Go | 1.22+ | Algorithms + CLI |
| Node.js | 22+ | Lab (OpenNext / SST) |
| HiGHS | 1.15+ | ILP benchmark (optional) |
| AWS CDK / SST | current | Infra + OpenNext deploy |
| AWS CLI | 2.x | S3 / SSM |

## Key Results (NRP, n012w8)

| Configuration | Mean / note | Best | Notes |
|--------------|-------------|------|-------|
| **SI + div30 + fw 6M (3M/worker)** | single seed | **3,440** | **Current platform best** (−25 vs prior) |
| Portfolio+lookahead+fw2 (published ladder) | mean 3,567 (6 runs) | 3,465 | Prior best |
| Portfolio+GA 3M (single seed) | — | 3,515 | No SI |
| Same + SI hybrid (single seed) | — | 3,485 | −30 vs 3M baseline |
| SA / LAHC / Tabu (earlier ladder) | — | 3,565 / 3,630 / 5,395 | Tabu weak alone |
| ILP (5hr HiGHS) | — | **3,020** feasible | Dual/MIP bound much lower (~1.8k–1.9k) |

Current best: portfolio with GA, SI hybrid assist, diversity slots 30%, and 6M iterations on the final 2-week window (`portfolio-ga-3m-si-div30-fw6m`). Still a single-seed result — multi-seed confirmation is next. Week 8 remains ~43% of total penalty. Tabu stays a diversifier inside portfolio, not a standalone winner.


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
| PFRS (SI+div30+fw6m, 3M) | 3,440 | +13.9% | ~hours (beam) | n120+ |
| PFRS (portfolio 1.5M prior) | 3,465 | +14.7% | 38 seconds | n120+ |
| PFRS (SA baseline) | 3,583 | +18.6% | 5 seconds | n120+ |
| Tabu standalone | 5,395 | +78.6% | 5 seconds | n120+ |

The heuristic scales far beyond ILP. The ~13.9% gap is the remaining flagship challenge.

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
  --time-limit 18000 --compare-pfrs 3440 --compare-pfrs-runtime 38.5
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
