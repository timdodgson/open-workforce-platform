# ADR-0012

## Title

Statistical comparison is domain-agnostic using objective values

## Status

Accepted

## Context

The platform needs to compare algorithm performance across multiple runs. Each domain has a different objective (penalty, distance, makespan) but they all share one property: lower is better. Statistical tests need numbers, not domain knowledge.

## Decision

The Statistics page operates on objective values only. It groups runs by configuration, computes descriptive statistics (mean, median, std dev, CI), performs pairwise Welch's t-tests, and reports Cohen's d effect sizes. It does not need to know what the numbers represent. A domain filter ensures only same-domain runs are compared.

## Alternatives

- **Domain-specific statistical pages.** Duplicates analysis logic per domain.
- **Normalised scores.** Transform all objectives to a common scale. Adds complexity and risks information loss.
- **No statistical tools.** Rely on eyeballing benchmark ladder numbers.

## Consequences

- Adding a new domain gets statistical comparison for free.
- Domain filter prevents meaningless cross-domain comparisons (can't compare a penalty of 3000 with a distance of 800).
- The objective label adapts per domain (Penalty / Distance / Makespan) for display purposes.
- Statistical rigour (p-values, effect sizes) supports publication-quality claims about algorithm superiority.
- The dashboard requires at least 2 runs per group for t-tests — single runs show descriptive stats only.
