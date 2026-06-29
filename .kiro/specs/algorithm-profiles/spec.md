# Algorithm Profiles

## Purpose

Allow optimisation algorithms to be configured without changing source code.

Profiles make benchmarking repeatable and allow different search strategies to be compared fairly.

---

# Why

Currently several algorithms contain hard-coded values such as:

- iteration limits
- tabu list size
- simulated annealing temperature
- cooling rate
- destroy size
- neighbourhood limits

These should become configurable.

---

# Behaviour

Introduce an AlgorithmProfile.

Each algorithm reads its configuration from the profile.

Example:

Constructive

(no configuration)

Hill Climbing

- MaxIterations

Simulated Annealing

- MaxIterations
- InitialTemperature
- CoolingRate

Tabu Search

- MaxIterations
- TabuListSize

Large Neighbourhood Search

- MaxIterations
- DestroySize

---

# Profiles

Provide built-in profiles.

default

Current behaviour.

fast

Fewer iterations.

quality

More iterations.

research

Larger search limits.

---

# Architecture

Profiles belong to OptimisationContext.

Algorithms read configuration from the context.

Algorithms must not know which named profile was selected.

---

# CLI

Support:

--profile default

--profile fast

--profile quality

--profile research

---

# Benchmark

Benchmark summary should include the selected profile.

---

# Tests

Verify:

- default reproduces existing behaviour
- fast uses fewer iterations
- research uses more iterations
- algorithms receive profile values correctly

---

# Non Goals

Do not load configuration files.

Do not introduce YAML.

Do not introduce JSON configuration.

Profiles are hard-coded Go values.

---

# Definition of Done

- algorithms use AlgorithmProfile
- CLI supports --profile
- benchmark displays profile
- tests pass