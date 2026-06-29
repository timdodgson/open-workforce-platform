# Business Event

## Purpose

Business Events represent something that has happened in the real world that creates, changes or removes work.

A Business Event is the starting point of the optimisation pipeline.

Business Events are part of the domain model and are independent of any optimisation engine, infrastructure or implementation technology.

---

# Why

Everything that enters the platform begins as a Business Event.

The platform models what has happened before determining what work should be carried out.

This separation allows business processes to evolve independently from optimisation algorithms.

The platform does not understand industry-specific business events. Instead, it provides a generic model that allows each organisation to define the events relevant to its own domain.

---

# Responsibilities

A Business Event:

- represents a business fact
- contains business information describing that fact
- may create one or more Work Items
- may modify existing Work Items
- may cancel existing Work Items
- can be validated
- can be serialised
- is immutable once created

---

# Non-Responsibilities

A Business Event must not:

- perform optimisation
- allocate Resources
- evaluate Constraints
- calculate Objectives
- contain infrastructure concerns
- understand scheduling
- understand optimisation algorithms
- know how Work Items are generated

Business Event processing is the responsibility of the application layer.

---

# Domain Invariants

A Business Event represents a historical fact.

Once created it cannot be changed.

A Business Event must always:

- have a unique identifier
- have a business-defined type
- have an occurrence time
- contain opaque business details supplied by the producing system
- be valid when constructed

A Business Event can never exist in an invalid state.

---

# Validation

Validation occurs during construction.

Invalid Business Events must never enter the domain model.

Validation should fail early using clear domain errors.

Objects should not require additional validation after successful construction.

---

# Immutability

Business Events are immutable.

Once created they represent historical facts.

Changing an existing Business Event would change history.

If business information changes, a new Business Event should be created instead of modifying an existing one.

---

# Identity

Every Business Event has a unique identifier.

Equality is determined by identity rather than object instance.

---

# Acceptance Criteria

The implementation shall:

- provide a BusinessEvent domain object
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

- valid event creation
- invalid event rejection
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
- inheritance
- external packages

unless they are explicitly required by the specification.

Prefer clear, explicit code over extensible abstractions.

---

# Ownership

Business Events originate outside the platform.

The platform consumes Business Events but does not create or modify them.

Identity, type and opaque business details are owned by the producing system.

The platform is responsible for validating that a Business Event is structurally valid before it enters the domain model.

The meaning of the business details remains the responsibility of the producing system or the application layer that understands that domain.

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