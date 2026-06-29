# Algorithm Architecture

## Purpose

Refactor the optimisation package so multiple optimisation algorithms can be hosted cleanly.

The platform currently supports:

- constructive
- hill-climbing

The goal is to make adding future algorithms easier without changing existing algorithm implementations.

---

# Why

The project now has more than one optimisation strategy.

An abstraction has earned its place.

The platform should allow new algorithms to be added as separate implementations while preserving a stable application and CLI layer.

This supports future algorithms such as:

- simulated annealing
- tabu search
- genetic algorithms
- CP-SAT adapters

---

# Scope

Refactor the optimisation layer so algorithms are selected through a common contract.

The refactor should preserve existing behaviour.

This is an architecture refactor, not a new optimisation algorithm.

---

# Required Behaviour

The following commands must continue to work:

go run ./cmd/owp optimise ../../examples/datasets/simple-events.json

go run ./cmd/owp optimise ../../examples/datasets/simple-events.json --algorithm hill-climbing

The default algorithm remains constructive.

Output should still show the algorithm name.

---

# Algorithm Contract

Introduce a small algorithm contract that represents an optimisation strategy.

The contract should allow the application layer to run an algorithm without knowing its implementation details.

Each algorithm should expose:

- its name
- a solve operation

The solve operation should use the existing optimisation inputs and return an Optimised Plan.

---

# Package Structure

Refactor towards separate algorithm packages.

Suggested structure:

optimisation/
  algorithm.go
  constructive/
    solver.go
  hillclimbing/
    solver.go

This is a suggested structure.

If a better structure is identified, explain it before implementing.

---

# Algorithm Selection

The application layer should select the requested algorithm by name.

Supported names:

- constructive
- hill-climbing

Unknown algorithm names should return a clear error.

---

# Non-Goals

Do not implement:

- new optimisation algorithms
- generic plugin loading
- dynamic runtime discovery
- reflection-based registration
- external dependencies
- configuration files
- dependency injection framework

---

# Tests

Add or update tests covering:

- default algorithm remains constructive
- constructive algorithm can be selected explicitly
- hill-climbing algorithm can be selected explicitly
- unknown algorithm returns a clear error
- both algorithms produce valid plans
- existing optimisation behaviour remains unchanged

---

# Constraints

Use only the Go standard library.

Do not introduce external dependencies.

Do not over-engineer the algorithm registry.

Keep the abstraction small.

The abstraction has earned its place because multiple algorithm implementations now exist.

---

# Definition of Done

The implementation is complete when:

- constructive and hill-climbing are separate algorithm implementations
- the application layer selects algorithms through the common contract
- existing CLI commands still work
- all tests pass
- no external dependencies are introduced
- adding a future algorithm would not require modifying existing algorithm implementations
- architecture remains consistent with steering documents

---

# Open Questions

None.