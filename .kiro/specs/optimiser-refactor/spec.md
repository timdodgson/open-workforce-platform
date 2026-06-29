# Optimiser Refactor

## Purpose

Refactor the current optimiser implementation to improve readability and maintainability before adding further optimisation algorithms.

This is a refactor only.

The optimiser behaviour must remain unchanged.

---

# Why

The optimiser now supports:

- capacity
- availability
- skills
- priority
- scoring

The current implementation has grown from a trivial walking skeleton into the first real optimisation capability.

Before adding more behaviour, the code should be made easier to understand and extend.

---

# Scope

Refactor the optimiser implementation without changing behaviour.

The optimiser should remain simple.

Do not introduce a generic constraint framework.

Do not introduce interfaces.

Do not introduce external dependencies.

---

# Expected Behaviour

After the refactor:

- capacity is still respected
- unavailable Resources are skipped
- skill requirements are respected
- higher-priority Work Items are assigned first
- scoring remains unchanged
- utilisation remains unchanged
- command-line output remains unchanged
- all existing tests still pass

---

# Suggested Refactor Direction

Split the optimiser into small private functions where useful.

Possible responsibilities include:

- preparing ordered Work Items
- checking whether a Resource can accept a Work Item
- assigning Work Items to Resources
- tracking used capacity
- calculating score
- calculating utilisation
- building the Optimised Plan

These are suggestions, not mandatory names.

Prefer clear code over excessive decomposition.

---

# Constraints

Use only the Go standard library.

Do not introduce:

- external dependencies
- interfaces
- generic constraint engines
- solver frameworks
- unnecessary abstractions

Do not rename public types unless the benefit clearly outweighs the cost.

Do not change public behaviour.

---

# Tests

Existing tests should continue to pass.

Add tests only if the refactor reveals behaviour that is currently untested.

Do not rewrite tests unnecessarily.

---

# Definition of Done

The refactor is complete when:

- go test ./... passes
- the optimisation command still runs successfully
- behaviour is unchanged
- readability is improved
- no unnecessary abstractions are introduced
- no external dependencies are introduced

---

# Open Questions

None.