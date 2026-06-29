# Benchmark Datasets

## Purpose

Create benchmark datasets that demonstrate the strengths and weaknesses of optimisation algorithms.

These datasets are intentionally designed to expose situations where constructive optimisation is not optimal.

---

# Why

Benchmarking requires meaningful optimisation problems.

Datasets should demonstrate:

- assignment quality
- travel optimisation
- scheduling quality
- neighbourhood improvements
- objective improvements

---

# Required Datasets

Create:

examples/datasets/

    constructive-baseline.json

    skill-trap.json

    travel-trap.json

    time-window-trap.json

    capacity-trap.json

    preference-trap.json

Each dataset should include comments (README if needed) explaining what behaviour is expected.

---

# Behaviour

Each dataset should illustrate one optimisation concept.

Where possible:

Constructive should produce a lower objective score.

Hill Climbing or Simulated Annealing should improve the result.

---

# Tests

Add integration tests that execute every algorithm against every dataset.

The tests should verify:

- algorithms complete successfully
- plans remain valid
- statistics are produced
- objective scores are deterministic

Do not require one algorithm to outperform another unless that dataset was specifically designed to demonstrate it.

---

# Non-Goals

Do not implement new optimisation behaviour.

Do not change algorithms.

---

# Definition of Done

- benchmark datasets created
- integration tests added
- algorithms run successfully against all datasets
- documentation explains each dataset