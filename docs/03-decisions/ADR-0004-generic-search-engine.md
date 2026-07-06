# ADR-0004

## Title

Single generic search engine serves all domains

## Status

Accepted

## Context

Multiple metaheuristic algorithms (SA, LAHC, Tabu, Portfolio, Adaptive) need to work across all problem domains. Each algorithm follows the same pattern: generate move, evaluate, accept/reject, record discoveries.

## Decision

Implement all algorithms in a single `internal/optimisation/` package. `RunSearch(problem, config)` dispatches to the correct algorithm based on `config.Mode`. All algorithms produce a uniform `SearchResult` with discoveries.

## Alternatives

- **Per-domain algorithm implementations.** Maximises domain-specific tuning but duplicates core logic.
- **Algorithm as interface with domain-provided hooks.** Adds unnecessary indirection.
- **External solver library.** Introduces dependency, reduces control, limits telemetry.

## Consequences

- One implementation of each algorithm for the entire platform.
- Bug fixes to an algorithm benefit all domains immediately.
- New algorithms (e.g. Genetic, Ant Colony) can be added once and work everywhere.
- Algorithm parameters are generic (temperature, iterations) not domain-specific.
- Algorithms cannot exploit domain structure directly — but the Problem interface allows domains to encode that knowledge in their TryMove/Evaluate implementations.
