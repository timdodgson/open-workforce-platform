# ADR-0009

## Title

ILP is used for benchmarking only, not as a primary solver

## Status

Accepted

## Context

Integer Linear Programming (ILP) via HiGHS can solve small instances to proven optimality. This provides a reference bound for measuring heuristic quality. However, ILP does not scale to large instances and requires a mathematical formulation per domain.

## Decision

ILP serves as a calibration tool. It provides the "reference" column on the Benchmark Ladder — the value against which heuristic gap% is measured. ILP runs are labelled with `mode: "ilp"` and displayed separately from heuristic results.

## Alternatives

- **ILP as primary solver.** Does not scale. NRP n030w4 or CVRP n80 would require impractical solve times.
- **No ILP at all.** Lose the ability to measure absolute quality (gap to optimal).
- **Published optimal values only.** Works for well-studied instances but not for novel problems.

## Consequences

- Heuristic quality is measurable: gap% = (heuristic - ILP) / ILP × 100.
- ILP models need only be written for instances small enough to solve in reasonable time.
- The dashboard distinguishes ILP runs (shown as reference, not competing in leaderboard).
- For instances where no ILP solution exists, gap% shows "—" rather than a misleading value.
