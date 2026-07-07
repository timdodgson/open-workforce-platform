# Dashboard

Documentation of every page in the PFRS Research Lab dashboard.

**URL:** http://pfrsda-albae-jbxmpxaiudkb-144941416.eu-west-1.elb.amazonaws.com/

---

## Home (`/`)

**Purpose:** Platform overview and run list.

**Data source:** S3 manifest (run list), static content (domain cards, features).

**Metrics:** None — informational page.

**How to interpret:** Domain cards explain what each problem type is. Run list shows all experiments with domain pills, algorithm labels, and instance names.

**When to use:** Starting point. Navigate to specific runs or platform pages.

<!-- Screenshot: Home page with domain cards and run list -->

---

## Benchmark Ladder (`/benchmarks`)

**Purpose:** Compare algorithm performance across all instances and domains. The primary research output page.

**Data source:** S3 manifest + `run.json` metadata for each run.

**Metrics:**
- Experimental setup (iterations, seeds, algorithms, runtime).
- Algorithm leaderboard (wins per algorithm).
- Average gap to reference.
- Per-instance results: SA, LAHC, Tabu, Portfolio, Adaptive values.
- Gap% to known optimal / best-known.
- Winner per instance.

**How to interpret:**
- Lower objective is better.
- Green sparkline bars show quality relative to reference.
- "✓ optimal" means the algorithm found the proven best solution.
- Gap% shows how far the best heuristic is from the reference.
- The leaderboard shows which algorithm wins most often across all instances.

**When to use:** Evaluating which algorithm to use. Comparing platform performance to published benchmarks. Deciding whether iteration budget should be increased.

<!-- Screenshot: Benchmark Ladder with leaderboard and CVRP table -->

---

## Statistics (`/statistics`)

**Purpose:** Rigorous statistical comparison between algorithm configurations.

**Data source:** S3 — all run metadata and summary data.

**Metrics:**
- Box plots (95% CI, mean, median, data points).
- Group statistics table (N, mean, median, best, worst, std dev, CI).
- Pairwise Welch's t-test (t-stat, p-value, significance).
- Cohen's d effect size.
- Histogram overlay.
- Observations (auto-generated from statistical results).

**How to interpret:**
- Domain filter: select which problem type to analyse (prevents cross-domain comparison).
- Group by: choose what defines a "configuration" (mode, instance, iterations, etc.).
- p < 0.05: statistically significant difference between groups.
- |d| > 0.8: large practical effect size.
- Tight box plots: consistent algorithm. Wide: high variance.

**When to use:** Deciding whether one configuration is genuinely better than another. Publishing comparative results. Identifying high-variance configurations that need more seeds.

<!-- Screenshot: Statistics page with box plots and significance table -->

---

## Run Summary (`/runs/[id]/summary`)

**Purpose:** Overview of a single experiment's results.

**Data source:** `run.json` metadata + `results.csv` (NRP only).

**Metrics (vary by domain):**
- **NRP:** Total penalty, weeks, workers, candidates, per-week breakdown.
- **CVRP:** Best distance, initial distance, improvement%, feasibility, runtime.
- **JSS:** Best makespan, initial makespan, improvement%, jobs, machines.
- **VRPTW:** Best distance, vehicles used, max vehicles, feasibility, improvement%.

**How to interpret:** The primary result is the "best" metric (distance/makespan/penalty). Improvement% shows how much the search improved over the constructive heuristic baseline.

**When to use:** Quick assessment of a run's quality. Starting point before drilling into search progress or visualisations.

<!-- Screenshot: CVRP summary showing distance and improvement -->

---

## Search Progress (`/runs/[id]/search`)

**Purpose:** Visualise how the search evolved over time.

**Data source:** `results.csv` (per-week/per-iteration stats), `discoveries.csv`.

**Metrics:**
- Penalty/distance by week (bar chart).
- Cumulative penalty (line chart).
- Candidate efficiency (penalty reduction per million candidates).
- Discovery timeline (penalty vs candidate number).
- Discovery statistics (total, globals, avg improvement, largest).

**How to interpret:**
- Steep drops in the discovery timeline indicate productive search phases.
- Flat sections indicate plateaus.
- High candidate efficiency early, declining later, is normal (diminishing returns).
- Discovery frequency shows how often the search finds improvements.

**When to use:** Understanding search dynamics. Diagnosing stagnation. Deciding whether to increase iteration budget.

<!-- Screenshot: Search progress with discovery timeline -->

---

## Timeline (`/runs/[id]/timeline`)

**Purpose:** Discovery events plotted against elapsed wall-clock time.

**Data source:** `discoveries.csv` (elapsedMs field).

**Metrics:**
- Elapsed time (ms) on x-axis.
- Penalty/objective on y-axis.
- Each point is a global best improvement.

**How to interpret:** Early clusters of discoveries indicate effective initial search. Late discoveries (near the end of runtime) suggest the budget was well-spent. No late discoveries means the search converged early — could try reheating or longer budget.

**When to use:** Evaluating whether the iteration budget was sufficient. Comparing convergence speed between algorithms.

---

## Search DNA (`/runs/[id]/dna`)

**Purpose:** Fingerprint the search behaviour — characterise what "type" of search this run performed.

**Data source:** Per-week audit data (NRP), discoveries.

**Metrics:**
- Acceptance pattern (how often worse moves are accepted).
- Improvement frequency.
- Plateau detection.
- Temperature trajectory (SA).

**How to interpret:** Identifies whether the search was exploratory (high acceptance) or exploitative (low acceptance). Useful for comparing runs with different parameters.

**When to use:** NRP runs with beam search. Understanding why one configuration outperformed another.

---

## Genealogy (`/runs/[id]/genealogy`)

**Purpose:** Visualise the lineage of beam search paths (NRP only).

**Data source:** `tree.csv` (beam search tree data).

**Metrics:**
- Path lineage (parent-child relationships).
- Retained vs pruned paths.
- Winning lineage depth.

**How to interpret:** Healthy beam search shows diverse lineages. If one family dominates, diversity is low and the search may converge prematurely.

**When to use:** NRP beam search runs. Diagnosing diversity loss. Evaluating beam width effectiveness.

---

## Search Map (`/runs/[id]/map`)

**Purpose:** Fitness landscape visualisation — show solution quality across the search space.

**Data source:** Discoveries + diversity data.

**Metrics:**
- Hamming distance from best (x-axis).
- Penalty (y-axis).
- Each point is a discovered solution.

**How to interpret:** Funnel shape indicates a smooth landscape (easy for metaheuristics). Scattered points indicate a rugged landscape (harder to optimise).

**When to use:** Understanding problem difficulty. Comparing landscape structure across instances.

---

## Route Viewer (`/runs/[id]/routes`)

**Purpose:** Visualise vehicle routes (CVRP, VRPTW).

**Data source:** `solution.json` (routes with customer IDs).

**Metrics:**
- Routes drawn as ordered sequences.
- Per-route distance and load.
- Total vehicles used.

**How to interpret:** Compact, non-crossing routes indicate good solutions. Long routes or many vehicles suggest room for improvement.

**When to use:** CVRP and VRPTW runs. Visual inspection of solution quality.

---

## Gantt Chart (`/runs/[id]/gantt`)

**Purpose:** Visualise job shop schedules as a Gantt chart (JSS only).

**Data source:** `solution.json` (operations with JobID, Machine, Start, End).

**Metrics:**
- Operations displayed as coloured bars (one colour per job).
- Machines on y-axis, time on x-axis.
- Makespan shown on time axis.

**How to interpret:** Gaps between operations on a machine indicate idle time. Tight packing with no gaps approaches optimal makespan. The critical path is the longest chain of operations.

**When to use:** JSS runs. Verifying schedule validity. Identifying bottleneck machines.

<!-- Screenshot: Gantt chart with 6 machines and colour-coded jobs -->

---

## Admin (`/admin`)

**Purpose:** System configuration reference and data contract documentation.

**Data source:** Environment variables, S3 manifest, schema constants.

**Metrics:**
- Storage type, bucket, region.
- Total runs, runs by domain.
- run.json schema reference (required + domain-specific fields).
- Objective reading priority.
- S3 storage layout.

**How to interpret:** Reference page. Shows what the platform expects from run.json files and how data flows from CLI to dashboard.

**When to use:** Debugging data issues. Understanding the contract between Go CLI and dashboard. Onboarding new contributors.

---

## Experiments (`/experiments`)

**Purpose:** List and manage experiment configurations.

**Data source:** S3 runs filtered by configuration patterns.

**Metrics:** Run count, configurations tested, date range.

**When to use:** Planning new experiments. Reviewing what has been tested.

---

## Assistant (`/experiments/chat`)

**Purpose:** AI-powered research assistant (behind Cognito authentication).

**Data source:** AWS Bedrock (Claude), dashboard data for context.

**When to use:** Asking questions about results, generating hypotheses, exploring parameter space. Requires login.

---

## Knowledge (`/knowledge`)

**Purpose:** Curated research knowledge base.

**Data source:** Static content + linked documentation.

**When to use:** Learning about optimisation concepts, algorithm theory, benchmark history.

---

## Search Intelligence — Assist (`/assist`)

**Purpose:** Analyse all AI advisory decisions across solver architectures.

**Data source:** `worker_assist.csv`, `generic_search_assist.csv`, `portfolio_assist.csv` from all runs.

**Metrics:**
- Total decisions, accepted, rejected, safety overrides.
- Workers skipped, global bests missed, early stops, budget adjustments.
- Per-architecture breakdown (WorkerAssist, SearchAssist, PortfolioAssist).
- Per-domain tables with strategy, budget, confidence, source (ML vs Rule).
- Learned model vs rule-based fallback counts.

**How to interpret:**
- Safety overrides = the system correctly prevented a dangerous recommendation.
- GB Missed should always be zero.
- Accepted/rejected ratio shows how active the AI is.
- Source column (ML/Rule) shows whether learned model was used.

**When to use:** After running experiments with `--worker-decision-mode shadow|assist|adaptive`. Understanding what Search Intelligence decided and why.

---

## Decisions (`/decisions`)

**Purpose:** Analyse per-worker spawn decisions (NRP beam search).

**Data source:** `worker_decisions.csv` from shadow/assist runs.

**When to use:** Evaluating whether the rule engine correctly predicts worker value.

---

## Learning (`/learning`)

**Purpose:** Visualise how worker/strategy performance data accumulates over time.

**Data source:** `worker_learning.csv` aggregated across runs.

**When to use:** Understanding training data quality for the ML model.

---

## Predictions (`/predictions`)

**Purpose:** Show per-worker predictions from the trained ML model.

**Data source:** `worker_predictions.json` from ML pipeline.

**When to use:** Evaluating model accuracy. Identifying which features drive predictions.

---

## Feature Importance (`/feature-importance`)

**Purpose:** Display which features the decision tree model considers most important.

**Data source:** `worker_model.json` (trained model with feature importances).

**When to use:** Understanding what the ML model learned. Guiding future feature engineering.

---

## What-If Lab (`/what-if`)

**Purpose:** Simulate alternative decisions and predict outcomes.

**Data source:** ML model + historical telemetry.

**When to use:** Exploring counterfactual scenarios. Testing hypotheses about parameter changes.
