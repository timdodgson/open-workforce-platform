# NRP Benchmark Integration

## Purpose

Allow converted NRP datasets to be included in benchmark runs.

---

# Why

The NRP adapter proves the platform can convert nurse rostering data into OWP datasets.

The next step is to benchmark algorithms against the converted NRP dataset alongside the existing synthetic benchmark datasets.

---

# Behaviour

Create a generated OWP-compatible NRP dataset file:

examples/datasets/nrp-simple.json

This file should be produced from:

examples/nrp/simple-nrp.json

The benchmark runner should pick it up automatically because it is a normal OWP dataset JSON file.

---

# Tests

Add tests verifying:

- the NRP sample can be converted
- the converted NRP dataset can be loaded
- all algorithms can run against the converted NRP dataset
- benchmark runner includes the converted NRP dataset

---

# Non-Goals

Do not modify optimisation algorithms.

Do not add official NRP XML support yet.

Do not change the benchmark runner architecture.

---

# Definition of Done

- nrp-simple.json exists under examples/datasets
- benchmark command includes nrp-simple
- all tests pass
- no optimiser changes