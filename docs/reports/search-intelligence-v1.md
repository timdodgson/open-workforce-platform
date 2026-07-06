# Search Intelligence v1 — Validation Report

## Summary

Search Intelligence v1 is the first online AI advisory system for the Open Workforce Platform's optimisation solvers. It has been validated safe on two distinct problem domains (NRP and CVRP), three algorithms (SA, LAHC, Tabu), and three solver architectures (beam search, single-search, portfolio).

**Status:** Validated on tested configurations. Not yet proven universal.

---

## Experiments Conducted

### Experiment 1: NRP SA (Beam Search)

| Metric | Off | Shadow | Assist |
|--------|-----|--------|--------|
| Mean objective | 1731 | 1789 | **1716** |
| Best objective | 1595 | 1725 | **1635** |
| Global bests missed | 0 | 0 | **0** |
| Safety overrides | — | — | Active |

**Conclusion:** SAFE. Assist mean is 1% better than off.

### Experiment 2: NRP Portfolio (Beam Search)

| Metric | Off | Shadow | Assist |
|--------|-----|--------|--------|
| Mean objective | 1805 | 1756 | **1771** |
| Std deviation | 67 | 63 | **38** |
| Global bests missed | 0 | 0 | **0** |

**Conclusion:** SAFE. Lower variance indicates more consistent results.

### Experiment 3: CVRP A-n32-k5 (Generic Search Engine)

#### SA (500K iterations)

| Metric | Off | Assist | Improvement |
|--------|-----|--------|-------------|
| Mean distance | 801.2 | 801.2 | Same |
| Mean runtime | 79ms | 30ms | **62% faster** |
| Mean candidates | 500K | 198K | **60% fewer** |

#### LAHC (500K iterations)

| Metric | Off | Assist | Improvement |
|--------|-----|--------|-------------|
| Mean distance | 810.4 | 810.4 | Same |
| Mean runtime | 71ms | 21ms | **70% faster** |
| Mean candidates | 500K | 140K | **72% fewer** |

#### Portfolio (100K per strategy)

| Metric | Off | Assist |
|--------|-----|--------|
| Mean distance | 786.4 | **784.0** |
| Best-known hit rate | 4/5 | **5/5** |

**Conclusion:** SAFE. Identical quality with 62–72% compute reduction. Portfolio assist improves results.

---

## Safety Rules

### WorkerAssist (NRP Beam Search)
- Never skip global-best lineage workers
- Never skip with confidence < 0.65
- Never skip workers at distance = 0 from global best
- Skip requires 3+ independent negative signals

### SearchAssist (Generic Search)
- Never early-stop before 20% of budget consumed
- Never stop within 5K candidates of a recent improvement
- Never reduce budget below minimum fraction

### PortfolioAssist (Portfolio Mode)
- Never skip all strategies
- At least 2 strategies must run in 3+ portfolio
- No strategy gets more than 2× budget boost
- No strategy gets less than 0.25× budget

---

## Telemetry Files

| File | Source | Content |
|------|--------|---------|
| `worker_learning.csv` | All solvers | Per-worker/run training data |
| `worker_decisions.csv` | NRP shadow/assist | Rule engine predictions vs outcomes |
| `worker_assist.csv` | NRP assist | Accepted/rejected worker decisions |
| `generic_search_assist.csv` | CVRP/JSS/VRPTW shadow/assist | Checkpoint recommendations |
| `portfolio_assist.csv` | Portfolio shadow/assist | Budget allocation decisions |
| `worker_model.json` | ML pipeline | Trained decision tree model |
| `worker_predictions.json` | ML pipeline | Per-worker predictions with explanations |

---

## Limitations

- Validated on n012w8 (NRP) and A-n32-k5 (CVRP) only
- JSS and VRPTW not yet experimentally validated (architecture tested, not quality-tested)
- Worker Value Model trained on limited data (805 records, mostly CVRP)
- PortfolioAssist uses simple heuristics, not learned model
- SearchAssist early-stop threshold is fixed, not adaptive
- No counterfactual simulation yet

---

## Cross-Domain Evidence

| Domain | Beam Search | Single-Search | Portfolio | Validated |
|--------|-------------|---------------|-----------|-----------|
| NRP | ✅ SA, Portfolio | — | — | **SAFE** |
| CVRP | — | ✅ SA, LAHC | ✅ | **SAFE** |
| JSS | — | Architecture ready | Architecture ready | Not tested |
| VRPTW | — | Architecture ready | Architecture ready | Not tested |

---

## Recommendation

Search Intelligence v1 is ready for production use on NRP and CVRP with the following caveats:
- Default mode remains `off`
- `shadow` mode is safe for all workloads (zero behaviour change)
- `assist` mode is validated safe on tested configurations
- JSS and VRPTW should be validated before enabling assist on those domains
