# PFRS Lab v3.0 — Search Intelligence

A multi-domain optimisation research platform that benchmarks algorithms and uses Search Intelligence to learn how to search better.

---

## Architecture

PFRS Lab is built on a generic Problem interface that separates domain knowledge from search logic. Any domain that implements seven methods gains immediate access to the full algorithm portfolio, telemetry pipeline, statistical analysis, and Search Intelligence advisory system.

```
Problems → Generic Interface → Algorithms → Search Intelligence → Telemetry → Learning → Storage → Dashboard → Research Outputs
```

---

## Domains

| Domain | Problem | Benchmark | Best Result |
|--------|---------|-----------|-------------|
| NRP | Nurse Rostering (INRC-II) | n012w8, n030w4, n030w8 | 3,440 (13.9% gap to ILP) |
| CVRP | Capacitated Vehicle Routing | A-n32-k5 through A-n80-k10 | 784 (optimal on A-n32-k5) |
| JSS | Job Shop Scheduling | la01, ft10 | 666 (optimal on la01) |
| VRPTW | Vehicle Routing + Time Windows | Solomon C101 | 829 (0.1% gap to BKS) |

---

## Algorithms

Five metaheuristic algorithms operate through the same interface on all domains:

- **Simulated Annealing** — Metropolis acceptance with adaptive cooling
- **Late Acceptance Hill Climbing** — accepts if better than L iterations ago
- **Tabu Search** — best-move neighbourhood with recency prohibition
- **Portfolio** — runs all strategies in parallel, keeps the best
- **Adaptive** — SA primary with LAHC escape bursts on stagnation

Additionally, ILP benchmarks (HiGHS) provide proven optimality bounds for calibration.

---

## Search Intelligence

A universal advisory system that monitors search progress and makes safe compute allocation decisions. Four modes, three integration styles.

**Modes:**

| Mode | Behaviour |
|------|-----------|
| `off` | No intelligence. Existing behaviour. Zero overhead. |
| `shadow` | Records predictions. No behaviour change. Safe data collection. |
| `assist` | Applies safe recommendations at static checkpoints. |
| `adaptive` | Live-updating decisions based on observed search progress. |

**Integration styles:**

| Style | Solver | Domains |
|-------|--------|---------|
| WorkerAssist | Beam search | NRP |
| SearchAssist | Single-search (SA/LAHC/Tabu) | CVRP, JSS, VRPTW |
| PortfolioAssist | Multi-strategy portfolio | All |

**Safety rules (non-negotiable):**
- Never stop before 20% of budget consumed
- Never stop immediately after an improvement
- Never skip all portfolio strategies
- Never allocate below 0.25× or above 2× base budget
- Never skip global-best lineage workers

**CLI:** `--worker-decision-mode off|shadow|assist|adaptive`

---

## Validation

320 experiment runs. 10 seeds per configuration. Welch t-test at 95% confidence.

| Domain | Adaptive vs Off | Compute Saved | Verdict |
|--------|----------------|---------------|---------|
| CVRP | Identical quality | 60–73% | Safe |
| JSS | Identical quality (optimal) | 40% | Safe |
| VRPTW | 19% better (p<0.001) | Trades for quality | Improved |
| NRP | Within variance | — | Safe |

Zero feasibility regressions. Zero missed best-known discoveries. Shadow mode behaviourally neutral on all deterministic solvers.

Validated on tested configurations. Not claimed universal.

---

## Dashboard

Next.js application deployed on AWS ECS Fargate with S3-backed telemetry storage.

- Redesigned homepage (research platform narrative)
- Inline SVG architecture diagram
- Unified Search Intelligence section with 7 tabs (Overview, Learning, Model, Predictions, Decision Analysis, What-If Lab, Assist Validation)
- Benchmark Ladder with gap-to-optimal tracking
- Statistical comparison (Welch t-test, box plots, Cohen's d)
- Per-domain visualisations (Gantt, Route Viewer, Schedule, Beam Tree)

---

## ML Pipeline

Offline worker value model trained on search telemetry.

- Decision tree classifier/regressor (scikit-learn)
- Portfolio budget model (learned allocation with rule-based fallback)
- `--storage s3` flag for automatic artefact upload
- Feature importance, predictions, and What-If simulation in dashboard

```bash
python -m worker_model.train --data-dir data --output worker_model.json --storage s3
python -m worker_model.predict --data-dir data --output worker_predictions.json --storage s3
```

---

## Deployment

- CI/CD: GitHub Actions (Go tests → Jest → Cypress → semantic-release → Docker build → ECR push → ECS deploy)
- Infrastructure: AWS CDK (S3, ECS Fargate, ALB, Cognito, ECR)
- Cost-optimised: Fargate Spot, no NAT gateway, 0.25 vCPU / 512MB

---

## Known Limitations

- JSS Portfolio has a known SA-bias in rule-based allocation (2/10 seeds marginally degraded). Learned model addresses this when loaded.
- Search Intelligence not yet validated on instances larger than 100 customers / 30 nurses.
- main.go is 131KB (single CLI file). Functional but maintenance cost grows. Splitting planned for v4.
- Adaptive stagnation thresholds are heuristic, not proven optimal.
- ML pipeline not integrated into CI (manual training step required).

---

## Future Direction

- Split main.go into per-command files
- Add CLI integration tests
- Include ML pipeline in CI (pytest validation)
- Expand to larger benchmark instances (Christofides, Taillard ta01-ta80)
- Adaptive portfolio with runtime reallocation based on live strategy progress
- Counterfactual simulation for post-hoc decision analysis
- Per-algorithm, per-domain stagnation profiles (learned convergence curves)
- Publication: multi-domain metaheuristic comparison with Search Intelligence

---

## Reproducibility

All results are reproducible from documented commands with explicit seeds.

```bash
owp solve-cvrp --instance A-n32-k5.vrp --mode portfolio --iterations 500000 --seed 42 --worker-decision-mode adaptive --run-label my-run
owp solve-jobshop --instance la01.txt --mode tabu --iterations 100000 --seed 42 --worker-decision-mode adaptive --run-label my-run
owp solve-vrptw --instance C101.txt --mode sa --iterations 100000 --seed 42 --worker-decision-mode adaptive --run-label my-run
owp tune-pfrs --instance n012w8 --pfrs-mode sa --pfrs-iterations-per-worker 100000 --seeds 42 --worker-decision-mode assist --pfrs-run-label my-run
```

---

## Summary

PFRS Lab v3.0 delivers a validated Search Intelligence system that observes, learns, and adapts search behaviour across four optimisation domains — while preserving zero-overhead backwards compatibility for users who don't enable it.

Everything measurable. Everything reproducible. Everything benchmarked. Everything explainable.
