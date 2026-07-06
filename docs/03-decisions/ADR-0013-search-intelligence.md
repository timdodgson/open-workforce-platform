# ADR-0013: Search Intelligence

## Status

Accepted

## Context

The optimisation platform runs parallel beam search (NRP) and portfolio-based metaheuristic search (CVRP, JSS, VRPTW). Significant compute is spent on workers/iterations that do not improve the solution. We need a way to reduce wasted compute without harming solution quality.

## Decision

Introduce Search Intelligence — a universal AI advisory system that can advise any solver architecture on compute allocation decisions.

### Modes

- **off** (default): no AI, existing behaviour unchanged
- **shadow**: AI observes and records predictions, no behaviour change
- **assist**: AI recommendations are acted upon with hard safety overrides

### Integration Styles

| Style | Architecture | Actions |
|-------|-------------|---------|
| WorkerAssist | Beam search (NRP) | Skip/reduce/increase/change workers |
| SearchAssist | Single search (SA/LAHC/Tabu) | Early stop, budget adjust |
| PortfolioAssist | Portfolio | Budget allocation across strategies |

### Safety Rules (non-negotiable)

- WorkerAssist: never skip global-best lineage, never skip low-confidence, require 3+ signals for skip
- SearchAssist: never stop before 20% budget, never stop after recent improvement
- PortfolioAssist: never skip all strategies, minimum 2 must run

### CLI

`--worker-decision-mode off|shadow|assist` on all solver commands.

## Consequences

- Compute savings of 62-72% demonstrated on CVRP with zero quality loss
- NRP assist produces equal or better objectives with lower variance
- Additional telemetry files generated (worker_assist.csv, generic_search_assist.csv, portfolio_assist.csv)
- Default remains off — no risk to existing users
- Shadow mode provides safe data collection for future model training

## Validation Evidence

- NRP SA: 5 seeds, assist mean 1% better, zero global bests missed
- NRP Portfolio: 5 seeds, assist lower variance, zero global bests missed
- CVRP SA: 5 seeds, identical distance, 62% fewer candidates
- CVRP LAHC: 5 seeds, identical distance, 72% fewer candidates
- CVRP Portfolio: 5 seeds, assist hit best-known on all seeds (off missed 1)
