# Default INRC-II Benchmark

## Purpose

Establish a single, repeatable benchmark for algorithm comparison.

The project has now reached a stable research-platform state.

Going forward, every optimisation algorithm should be compared against the same official INRC-II benchmark unless explicitly overridden.

The chosen benchmark is:

n010w4

using the

research

algorithm profile.

---

## Goals

Provide a single command that executes the standard benchmark.

Running:

go run ./cmd/owp benchmark-inrc2

should immediately benchmark every registered algorithm using the standard research configuration.

---

## Default Behaviour

If no arguments are supplied:

benchmark-inrc2

shall:

- use the official n010w4 instance
- use the research profile
- execute every registered algorithm
- produce the benchmark summary

---

## Explicit Override

Users must still be able to benchmark any instance.

Examples:

benchmark-inrc2 n005w4

benchmark-inrc2 n020w4

benchmark-inrc2 n050w4

benchmark-inrc2 n120w8

Explicit command-line arguments always override the defaults.

---

## Research Profile

benchmark-inrc2 shall default to:

research

unless another profile is supplied.

Examples:

--profile fast

--profile quality

--profile research

---

## Output

Display the benchmark instance being used.

Example:

=================================================

Open Workforce Platform Benchmark

Instance: n010w4

Profile: research

=================================================

Then execute all algorithms.

---

## League Table

Display results ordered by score.

Columns should include:

Rank

Algorithm

Penalty

Soft Violations

Hard Violations

Objective

Runtime

Candidates

The best algorithm should appear first.

---

## Summary

Display:

Best algorithm

Lowest penalty

Fastest algorithm

Average runtime

Average penalty

Average soft violations

Average hard violations

---

## Architecture

No optimisation algorithms shall change.

No scoring behaviour shall change.

This specification only improves benchmark usability.

---

## Tests

Add tests verifying:

default instance selection

default profile selection

explicit instance override

explicit profile override

stable ordering

league table sorting

---

## Definition of Done

The implementation is complete when:

✓ benchmark-inrc2 requires no parameters.

✓ n010w4 is used automatically.

✓ research profile is selected automatically.

✓ explicit arguments override defaults.

✓ benchmark output is deterministic.

✓ algorithms are ranked by score.

✓ all existing tests continue to pass.