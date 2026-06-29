# Resource

## Purpose

A Resource represents something capable of completing Work Items.

Resources may represent people, vehicles, teams, equipment, AI agents, robots or external providers.

Resources are part of the domain model and are independent of any optimisation algorithm, infrastructure or implementation technology.

---

# Why

Optimisation requires both demand and supply.

Work Items represent demand.

Resources represent supply.

Separating Resources from Work Items allows the platform to optimise how available supply is matched to work that needs to be completed.

---

# Responsibilities

A Resource:

- represents something capable of completing work
- contains business information describing the resource
- can be validated
- can be serialised
- is immutable once created

---

# Non-Responsibilities

A Resource must not:

- perform optimisation
- allocate itself to Work Items
- evaluate Constraints
- calculate Objectives
- understand scheduling
- contain infrastructure concerns
- know how Work Items are generated

---

# Domain Invariants

A Resource represents available supply.

A Resource must always:

- have a unique identifier
- have a type
- contain opaque business details
- be valid when constructed

A Resource can never exist in an invalid state.

---

# Validation

Validation occurs during construction.

Invalid Resources must never enter the domain model.

Validation should fail early using clear domain errors.

Objects should not require additional validation after successful construction.

---

# Immutability

Resources are immutable.

If resource information changes, a new Resource should be created rather than modifying an existing one.

---

# Identity

Every Resource has a unique identifier.

Equality is determined by identity rather than object instance.

---

# Ownership

Resources are owned by the platform input model.

The platform consumes Resource definitions when producing an Optimised Plan.

The meaning of opaque business details remains the responsibility of the producing system or application layer that understands that domain.

---

# Acceptance Criteria

The implementation shall:

- provide a Resource domain object
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