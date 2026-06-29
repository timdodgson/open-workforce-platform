# Benchmark Delta Output

## Purpose

Improve benchmark output by showing each algorithm's objective score improvement compared with the constructive baseline.

---

# Why

Benchmark output currently shows objective scores, but it does not clearly show whether an algorithm improved over the baseline.

A delta column makes algorithm comparison easier.

---

# Behaviour

For each dataset, use the constructive algorithm as the baseline.

Add a column:

Delta

The delta is:

algorithm objective score - constructive objective score

Example:

Dataset        Algorithm       Objective   Delta
travel-trap    constructive    2890        0
travel-trap    tabu-search     2981        +91

---

# Architecture

The benchmark runner should calculate deltas only for display.

Do not change optimisation behaviour.

Do not change objective scoring.

---

# Tests

Add tests covering:

- constructive delta is zero
- positive delta is displayed with plus sign
- negative delta is displayed correctly
- benchmark output remains deterministic

---

# Constraints

Use only the Go standard library.

Do not introduce external dependencies.

Keep the change intentionally simple.

---

# Definition of Done

- benchmark output includes delta column
- delta is relative to constructive per dataset
- tests pass
- benchmark command still runs