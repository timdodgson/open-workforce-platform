# Optimisation Context

## Purpose

Introduce an OptimisationContext to represent the complete optimisation problem passed to algorithms.

This is a refactor-only task.

No optimisation behaviour should change.

---

# Why

Algorithms now consume multiple structured inputs.

Current inputs include:

- WorkItemInput
- ResourceInput

Future inputs may include:

- travel matrices
- time windows
- objectives
- constraint settings
- algorithm configuration

Passing each input separately will make algorithm signatures grow over time.

An OptimisationContext provides a stable contract for algorithms.

---

# Behaviour

No behaviour should change.

Existing algorithms must continue to produce the same results.

CLI output must remain unchanged.

---

# OptimisationContext

Introduce:

OptimisationContext

containing:

- WorkItems
- Resources

The context should be immutable where practical.

Accessors should return defensive copies where needed.

---

# Algorithm Contract

Update the algorithm contract so algorithms receive OptimisationContext rather than separate input slices.

Algorithms should still return an OptimisedPlan.

---

# Non-Goals

Do not implement:

- travel matrices
- time windows
- algorithm configuration
- runtime constraint configuration
- external dependencies
- generic solver frameworks

---

# Tests

Update tests to use OptimisationContext.

Existing behaviour tests should continue to pass.

Add tests for context construction and defensive copying where useful.

---

# Constraints

Use only the Go standard library.

Do not introduce external dependencies.

Do not change domain objects.

Do not change CLI output.

Do not change optimisation behaviour.

---

# Definition of Done

The implementation is complete when:

- OptimisationContext exists
- algorithms consume OptimisationContext
- existing algorithms still work
- all tests pass
- CLI commands still run successfully
- behaviour is unchanged
- no external dependencies are introduced

---

# Open Questions

None.