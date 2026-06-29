# Benchmark Delta Percentage

## Purpose

Improve benchmark output by showing percentage improvement over the constructive baseline.

---

# Why

Absolute delta is useful, but it does not show the relative size of an improvement.

A +100 improvement means different things on a score of 2,000 versus a score of 100,000.

Percentage delta makes benchmark comparison clearer.

---

# Behaviour

Add a benchmark column:

Delta %

For each dataset:

Delta % = ((algorithm objective - constructive objective) / constructive objective) * 100

Constructive baseline should display:

0.0%

Positive values should display with a plus sign.

Example:

+50.0%

---

# Edge Cases

If constructive objective is zero, display:

n/a

---

# Architecture

This is benchmark display only.

Do not change optimisation behaviour.

Do not change objective scoring.

---

# Tests

Add tests covering:

- constructive delta percentage is 0.0%
- positive percentage shows plus sign
- negative percentage displays correctly
- zero baseline displays n/a
- benchmark output remains deterministic

---

# Definition of Done

- benchmark output includes Delta %
- values are relative to constructive per dataset
- tests pass
- no optimisation behaviour changes