# PFRS Research Lab — Vision

## What This Project Is

PFRS Research Lab is a multi-domain optimisation research platform. It provides a unified framework for solving NP-hard combinatorial problems using metaheuristic search algorithms, with full telemetry, statistical comparison, and interactive visualisation.

The platform is not a single-purpose solver. It is an engineering system designed to make it straightforward to add new optimisation domains and immediately benefit from the full algorithm portfolio, dashboard, and analysis infrastructure.

## Why It Exists

Optimisation research is fragmented. Each problem domain typically has its own bespoke solver, its own evaluation framework, and its own reporting tools. Comparing algorithms across domains requires building everything from scratch each time.

This platform exists to eliminate that repetition. A well-defined Problem interface separates domain knowledge from search algorithms. Once a domain implements that interface, it automatically gains access to every algorithm, every telemetry system, and every visualisation tool the platform provides.

## Long-Term Goal

Build a credible, extensible research platform that:

- Solves real-world combinatorial optimisation problems across multiple domains.
- Provides rigorous statistical comparison between algorithms.
- Produces publication-quality results with full reproducibility.
- Demonstrates that a single generic search engine can compete with domain-specific solvers.
- Serves as a reference implementation for metaheuristic algorithm design.

## Design Philosophy

- **Separation of concerns.** Business logic belongs in the domain. Search logic belongs in the engine. Infrastructure supports both.
- **Simplicity over cleverness.** Prefer straightforward implementations that are easy to understand, test, and extend.
- **Everything earns its place.** No abstraction, dependency, or feature exists without clear justification.
- **Quality over quantity.** A small number of well-implemented domains is worth more than many half-finished ones.

## Architecture

The platform is built from five independent layers:

### Optimisation Problems

Each domain implements a generic `Problem` interface:

- `CreateInitialSolution` — construct a feasible starting point.
- `CloneSolution` — deep copy for parallel search.
- `Evaluate` — compute the objective value (lower is better).
- `TryMove` — generate and apply a random neighbourhood move.
- `UndoMove` — revert a rejected move.
- `SolutionFingerprint` — hash for diversity measurement.
- `SerializeSolution` — export for dashboard consumption.

Domains know nothing about which algorithm will search them. Algorithms know nothing about what domain they are solving.

### Search Algorithms

The generic search engine dispatches to:

- **Simulated Annealing (SA)** — Metropolis acceptance with adaptive cooling.
- **Late Acceptance Hill Climbing (LAHC)** — accepts if better than L iterations ago.
- **Tabu Search** — best-move neighbourhood with recency-based prohibition.
- **Portfolio** — runs SA, LAHC, and Tabu in parallel, returns the best.
- **Adaptive** — SA primary with LAHC escape bursts on stagnation detection.

All algorithms operate through the Problem interface. Adding a new algorithm benefits every domain.

### Orchestration

The CLI (`owp`) provides commands for each domain and mode. It handles:

- Instance loading and validation.
- Algorithm configuration and execution.
- Result serialisation and S3 upload.
- Manifest management for dashboard discovery.

### Telemetry

Every search run produces:

- Discovery timeline (every global best improvement with timestamp).
- Search progress metrics (candidates, acceptance rate, temperature).
- Solution serialisation (routes, schedules, assignments).
- Run metadata (configuration, instance, objective, runtime).

### Dashboard

A Next.js application deployed on AWS ECS that provides:

- Benchmark Ladder with algorithm leaderboard and gap-to-optimal tracking.
- Statistical comparison (Welch's t-test, box plots, Cohen's d).
- Per-run analysis (summary, search progress, discovery timeline).
- Domain-specific visualisations (Gantt charts, route viewers, schedules).
- Export tools for publication figures.

## Supported Domains

### Nurse Rostering (NRP)

Based on the INRC-II competition format. Assigns shifts to nurses over a multi-week horizon, minimising soft constraint penalties while satisfying coverage, skills, and contractual rules. Solved using beam search with per-week optimisation.

### Capacitated Vehicle Routing (CVRP)

Based on CVRPLIB benchmark instances. Finds minimum-distance routes for capacity-limited vehicles serving customers from a depot. Five neighbourhood operators (Relocate, Swap, IntraSwap, 2-Opt, Or-opt).

### Job Shop Scheduling (JSS)

Based on Taillard/Fisher & Thompson benchmark instances. Schedules operations across machines to minimise makespan. Permutation-based encoding with greedy decoding.

### Vehicle Routing with Time Windows (VRPTW)

Based on Solomon benchmark instances. Extends CVRP with temporal constraints — each customer has a service time window. Vehicles must arrive within the window or wait; late arrivals are infeasible.

## Future Domains

- **Pickup & Delivery** — extend VRPTW with paired pickup/delivery constraints.
- **Workforce Scheduling** — generalise NRP beyond healthcare to field service, logistics, retail.
- **Multi-Depot VRP** — multiple depots with heterogeneous fleets.

## What Every Domain Gets

When a new domain implements the Problem interface, it automatically gains:

- SA, LAHC, Tabu, Portfolio, and Adaptive search modes.
- S3 telemetry upload with manifest registration.
- Dashboard discovery, summary page, and search progress visualisation.
- Benchmark Ladder integration with gap-to-optimal tracking.
- Statistical comparison against other configurations and algorithms.
- ILP benchmark support (where a mathematical formulation is practical).

No dashboard code changes are required for basic support. Domain-specific visualisations (Gantt charts, route viewers) are added as needed.

## What This Project Is Not

- It is not a commercial solver. It is a research platform.
- It is not a university dissertation. It is production-quality engineering.
- It is not a wrapper around existing solvers. All algorithms are implemented from scratch.
- It is not limited to a single domain. The architecture is explicitly multi-domain.
