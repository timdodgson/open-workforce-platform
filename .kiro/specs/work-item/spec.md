# Work Item

## Purpose

A Work Item represents a unit of work that requires planning or optimisation.

Work Items are created from one or more Business Events and become the primary input to the optimisation engine.

A Work Item is independent of any optimisation algorithm.

---

# Why

Business Events describe what has happened.

Work Items describe what needs to be done.

Separating these concepts allows the same business event to generate different work depending on business rules while keeping optimisation independent of the originating business process.

---

# Responsibilities

A Work Item:

- represents work that may need to be completed
- contains the information required for planning
- can be validated
- can be serialised
- is immutable once created

---

# Non-Responsibilities

A Work Item must not:

- perform optimisation
- allocate Resources
- evaluate Constraints
- calculate Objectives
- understand scheduling
- contain infrastructure concerns

---

# Domain Invariants

A Work Item represents work that may be scheduled.

A Work Item must always:

- have a unique identifier
- have a type
- contain planning details
- be valid when constructed

A Work Item can never exist in an invalid state.

---

# Validation

Validation occurs during construction.

Invalid Work Items must never enter the domain model.

Validation should fail early using clear domain errors.

Objects should not require additional validation after successful construction.

---

# Immutability

Work Items are immutable.

If the work changes, a new Work Item should be created rather than modifying an existing one.

---

# Identity

Every Work Item has a unique identifier.

Equality is determined by identity rather than object instance.

---

# Ownership

Work Items are created by the platform.

Business Events are consumed by the platform.

The platform owns the lifecycle of Work Items.

---

# Acceptance Criteria

The implementation shall:

- provide a WorkItem domain object
- be immutable after creation
- reject invalid construction
- expose only behaviour appropriate to the domain
- be easy to serialise
- rely only on the Go standard library
- be fully unit tested

---

# Dependencies

Implementation should rely solely on the Go standard library unless a dependency can be clearly justified.

Any proposed dependency must explain why it earns its place.

---

# Tests

Tests should verify:

- valid construction
- invalid construction
- immutability
- serialisation
- identity behaviour

Tests should validate business behaviour rather than implementation details.

---

# AI Guidance

The initial implementation should remain intentionally simple.

Do not introduce:

- interfaces
- factories
- builders
- dependency injection
- external packages

unless they are explicitly required by the specification.

Prefer clear, explicit code over extensible abstractions.

---

# Open Questions

None.

---

# Engineering Checklist

- Adds measurable value.
- Respects the domain model.
- Preserves architectural boundaries.
- Introduces no unnecessary dependencies.
- Uses the Go standard library.
- Is simple to understand.
- Is fully tested.
- Updates documentation where required.