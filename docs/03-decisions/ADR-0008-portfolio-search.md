# ADR-0008

## Title

Portfolio mode runs algorithms in parallel and keeps the best

## Status

Accepted

## Context

Different algorithms perform differently on different instances. SA may beat LAHC on one instance and lose on another. Rather than manually selecting the best algorithm per instance, run them all and take the winner.

## Decision

Portfolio mode launches SA, LAHC, and Tabu as parallel goroutines. Each gets its own Problem instance (via `CreateInitialSolution`). Results are collected via channels. The best objective value wins. Portfolio always performs at least as well as the best individual algorithm.

## Alternatives

- **Sequential trials.** Run SA then LAHC then Tabu. 3× slower than parallel.
- **Algorithm selection via machine learning.** Requires training data and adds complexity without guaranteed improvement.
- **Ensemble with solution sharing.** More complex coordination. Potential for interference between strategies.

## Consequences

- Portfolio is a safe default — it never underperforms the best individual algorithm.
- Uses multiple CPU cores effectively (each strategy in its own goroutine).
- Wall-clock time equals the slowest strategy, not the sum.
- No communication between strategies during search (independent runs).
- The winning strategy is recorded in metadata (`winnerStrategy` field) for analysis.
