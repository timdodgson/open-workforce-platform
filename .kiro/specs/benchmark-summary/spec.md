# Benchmark Summary

## Purpose

Add an aggregate summary to the benchmark runner.

The benchmark table shows per-dataset results. The summary should show how each algorithm performs overall.

---

# Why

As benchmark datasets grow, individual rows become harder to interpret.

A summary helps identify which algorithms are consistently improving over the constructive baseline.

---

# Behaviour

After the benchmark table, display a summary by algorithm.

For each algorithm show:

- number of datasets run
- average objective score
- average delta
- average delta percentage
- total candidates evaluated

Example:

Benchmark Summary:

Algorithm                  Datasets   Avg Objective   Avg Delta   Avg Delta %   Candidates
constructive              8          3200            0           0.0%          0
tabu-search               8          3475            +275        +8.6%        2311

---

# Architecture

This is benchmark reporting only.

Do not change optimisation behaviour.

Do not change objective scoring.

---

# Tests

Add tests covering:

- summary includes all algorithms
- averages are calculated correctly
- constructive average delta is zero
- total candidates are summed
- output remains deterministic

---

# Definition of Done

- benchmark output includes summary section
- summary groups by algorithm
- tests pass
- no optimisation behaviour changes