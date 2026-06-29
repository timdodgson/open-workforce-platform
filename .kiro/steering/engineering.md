# Engineering Steering

## Purpose

This document defines the engineering principles that should guide all AI-assisted implementation within the Open Workforce Platform.

The objective is not to maximise generated code.

The objective is to maximise engineering quality.

Where trade-offs exist, prefer the solution that best aligns with the project's engineering principles rather than the solution that produces the least code.

---

# Core Principles

Always:

- Add value.
- Prefer simplicity.
- Respect the architecture.
- Preserve the domain model.
- Optimise for long-term maintainability.
- Produce readable code.
- Explain non-obvious decisions.

Never:

- Introduce unnecessary complexity.
- Add dependencies without justification.
- Mix business logic with infrastructure.
- Move business rules into the optimisation engine.
- Ignore existing architectural decisions.

---

# Dependencies

Before introducing a dependency, always ask:

- What value does it provide?
- Can the standard library solve this problem?
- Is the dependency actively maintained?
- Does it introduce unnecessary transitive dependencies?
- Will it increase long-term maintenance cost?

Every dependency must earn its place.

---

# Code Quality

Prefer:

- Small functions.
- Small packages.
- Explicit behaviour.
- Clear naming.
- Composition over unnecessary abstraction.
- Readability over cleverness.

Avoid:

- Hidden behaviour.
- Magic values.
- Premature abstraction.
- Over-engineering.

---

# Architecture

Respect documented architecture.

Do not:

- Move business rules into optimisation.
- Duplicate domain concepts.
- Cross bounded-context boundaries without justification.
- Introduce shortcuts that weaken the architecture.

If the architecture appears incorrect, raise the concern rather than working around it.

---

# Documentation

Documentation is part of the product.

If implementation changes:

- architecture
- behaviour
- workflows
- public interfaces

recommend updating the appropriate documentation.

---

# Testing

Test behaviour, not implementation.

Every new feature should include appropriate tests.

Every production defect should result in a regression test.

---

# Function Design

Design functions with a single clear responsibility.

Prefer passing domain objects over long parameter lists.

Avoid functions with more than four parameters.

If more than four parameters are required, stop and reconsider the design.

Consider whether:

- a domain object should be introduced
- a value object should be introduced
- the responsibility should be split
- the function is doing too much

Functions should be easy to understand from their signature alone.

---

# Domain Objects

Prefer rich domain objects over primitive values.

Avoid passing multiple related primitive values where a single domain concept exists.

For example, prefer:

Assignment

over:

resourceId, workItemId, startTime, endTime

---

# Constructors

Constructors should establish a valid object.

Avoid constructors that require excessive parameters.

Where construction becomes complex, prefer a builder or factory over a large constructor.

Objects should not exist in an invalid state.

---

# Final Review

Before completing any implementation, confirm:

- Does this add value?
- Is this the simplest reasonable solution?
- Does it respect the domain model?
- Does it respect architectural boundaries?
- Does every dependency earn its place?
- Would another engineer easily understand this code?
- Does the implementation align with the Engineering Principles?

If the answer to any question is "No", reconsider the implementation before completing the task.