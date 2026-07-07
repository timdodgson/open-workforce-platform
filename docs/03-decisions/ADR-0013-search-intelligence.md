# ADR-0013: Search Intelligence

## Status

Accepted — implemented and validated on all four domains.

## Context

The optimisation platform runs parallel beam search (NRP) and portfolio-based metaheuristic search (CVRP, JSS, VRPTW). Significant compute is spent on workers/iterations that do not improve the solution. We need a way to reduce wasted compute without harming solution quality.

## Decision

Introduce Search Intelligence — a universal AI advisory system that can advise any solver architecture on compute allocation decisions.

### Modes

- **off** (default): no AI, existing behaviour unchanged
- **shadow**: AI observes and records predictions, no behaviour change
- **assist**: AI recommendations are acted upon with hard safety overrides (static checkpoints)
- **adaptive**: live-updating decisions based on observed search progress and learned models

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

`--worker-decision-mode off|shadow|assist|adaptive` on all solver commands.

## Consequences

- Compute savings of 40–73% demonstrated on CVRP and JSS with zero quality loss
- VRPTW adaptive improves quality by 19% (p<0.001)
- NRP assist produces equal or better objectives within natural beam search variance
- Additional telemetry files generated (worker_assist.csv, generic_search_assist.csv, portfolio_assist.csv)
- Default remains off — no risk to existing users
- Shadow mode provides safe data collection for model training
- Learned portfolio model replaces fixed heuristics when confidence is sufficient
- Validated on tested configurations across all four domains (320 runs, 10 seeds, Welch t-test)

## Validation Evidence

### Initial Validation (5 seeds)
- NRP SA: assist mean 1% better, zero global bests missed
- NRP Portfolio: assist lower variance, zero global bests missed
- CVRP SA: identical distance, 62% fewer candidates
- CVRP LAHC: identical distance, 72% fewer candidates
- CVRP Portfolio: assist hit best-known on all seeds

### Statistical Validation (10 seeds, Welch t-test)
- CVRP: identical quality, 60–73% compute saved (p=1.0, all modes equal)
- JSS Tabu: identical quality (optimal on all seeds), 40% compute saved
- VRPTW SA adaptive: 19% better quality (p<0.001, d=-2.91)
- VRPTW Portfolio adaptive: 11% better quality (p<0.001, d=-3.45)
- NRP: within natural beam search variance (p>0.05)
- Zero feasibility regressions across 320 runs
- Zero missed best-known discoveries

Validated on tested configurations. Not claimed universal.

## Amendment: Learned Portfolio Budget Allocation (2026-07-07)

### Context

The rule-based portfolio advisor uses fixed heuristics (e.g., "SA is generally strong" → +10% budget). Validation showed this causes 2/10 degraded runs on JSS because the heuristic is factually wrong for that domain.

### Decision

Add a learned budget allocation model that replaces fixed rules with data-driven recommendations trained on historical telemetry.

### Behaviour

- Model stored as `portfolio_budget_model.json` (JSON format)
- Loaded via `--portfolio-model <path>` CLI flag
- Falls back to rule-based allocation when:
  - Model file is missing or malformed
  - Model confidence is below 0.60 threshold
  - Fewer than 3 historical samples for the domain/strategy
- Instance-specific entries override domain-wide entries
- Off mode unchanged, shadow mode records learned recommendations without applying

### Safety Gates

- Never allocate zero budget to all strategies
- Never remove all budget from historically strongest strategy
- Cap max boost at 2× original budget
- Floor min budget at 0.25× original budget
- Require confidence threshold (0.60) before applying learned recommendation
- Fall back to rule-based allocation if confidence is low

### CLI

`--portfolio-model <path>` on solve-cvrp, solve-jobshop, solve-vrptw.

### Consequences

- Eliminates the "SA is generally strong" false prior that caused JSS degradations
- Provides per-domain, per-instance budget allocation based on observed performance
- Graceful degradation: no model → rule-based → same behaviour as before
- Dashboard shows source of each recommendation (ML vs Rule)
