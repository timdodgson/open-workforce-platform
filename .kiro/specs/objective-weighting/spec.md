# Configurable Objective Weighting

## Purpose

Make optimisation objective weights configurable rather than hard-coded.

This allows different optimisation strategies without changing the optimisation engine.

---

# Why

Current objective values are constants.

Assignment: 1000
Travel: -1
Preferred Resource: 25
Plan Stability: 10
Balance: 1

Different organisations optimise for different priorities.

Objective weighting allows the same optimisation engine to produce different plans.

---

# Behaviour

Introduce ObjectiveWeights.

Example:

Assignment: 1000
Travel: -5
PreferredResource: 50
PlanStability: 20
Balance: 2

The objective engine multiplies each contribution by its configured weight.

---

# Architecture

ObjectiveWeights belongs to optimisation input.

The application layer provides weights.

OptimisationContext stores them.

Objective evaluation uses them.

Algorithms remain completely unaware of weighting.

---

# Defaults

If no weights are supplied, existing behaviour must remain identical.

Default weights:

Assignment: 1000
Travel: -1
PreferredResource: 25
PlanStability: 10
Balance: 1

---

# CLI

Add an optional argument:

--weights default

Initially only the built-in "default" profile is required.

The architecture should allow future profiles.

---

# Tests

Add tests covering:

- default weights preserve behaviour
- custom weights change objective score
- algorithms remain deterministic
- benchmark runner still works

---

# Non-Goals

Do not implement:

- file loading
- YAML
- JSON config
- runtime editing
- UI

Only support built-in profiles.

---

# Constraints

Use only the Go standard library.

Do not change optimisation algorithms.

Keep implementation intentionally simple.

---

# Definition of Done

- objective weights configurable
- defaults unchanged
- OptimisationContext carries weights
- objective engine consumes weights
- tests pass
- benchmark runner succeeds

---

# Open Questions

None.
