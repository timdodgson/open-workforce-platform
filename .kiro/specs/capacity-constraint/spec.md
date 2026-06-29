# Capacity Constraint

## Purpose

Implement the platform's first optimisation constraint.

Resources have a finite capacity.

The optimiser must respect that capacity when assigning Work Items.

This is the first implementation of genuine optimisation behaviour within the platform.

---

# Why

The current optimiser assigns every Work Item to the first available Resource.

While this proves the optimisation pipeline, it does not optimise.

Introducing a simple capacity constraint demonstrates how business rules influence optimisation decisions while preserving the existing architecture.

This establishes the pattern for introducing future constraints such as skills, availability, working hours, travel time and regulatory compliance.

---

# Problem Statement

Given:

- a collection of Resources
- a collection of Work Items

Each Resource has a maximum capacity.

The optimiser should assign Work Items so that no Resource exceeds its capacity whenever sufficient capacity exists.

---

# Scope

This iteration introduces only capacity.

Do not implement:

- skills
- locations
- travel
- working hours
- calendars
- priorities
- objectives
- weighting
- optimisation frameworks

---

# Capacity

Capacity remains business information.

The platform does not own its meaning.

Capacity is supplied within the Resource details JSON.

Example:

```json
{
    "capacity": 2
}
```

The application layer is responsible for interpreting this information.

The Resource domain object remains generic.

---

# Responsibilities

The application layer should:

- read capacity from Resource details
- provide capacity information to the optimiser

The optimisation layer should:

- respect capacity during assignment
- never knowingly exceed capacity when sufficient capacity exists

The domain layer should remain unchanged unless required to support the Optimised Plan.

---

# Optimisation Behaviour

The initial algorithm should remain intentionally simple.

Process Work Items in order.

Assign each Work Item to the first Resource with available capacity.

When a Resource reaches capacity, continue with the next Resource.

If all Resources are at capacity:

- stop assigning additional Work Items
- return the partially completed Optimised Plan

Do not invent overflow behaviour.

---

# Optimised Plan

Extend the Optimised Plan to report:

- assignments
- assigned work item count
- unassigned work item count
- total resource capacity
- utilisation percentage
- optimisation score

---

# Optimisation Score

Introduce the platform's first scoring mechanism.

The initial score is intentionally simple.

If every Work Item is assigned within capacity:

Score = 100

Otherwise:

Score =

(number of assigned work items ÷ total work items) × 100

Rounded to the nearest whole number.

This scoring mechanism exists only to demonstrate optimisation quality.

Future iterations will replace it with a richer objective model.

---

# Console Output

Example:

```text
=== Optimised Plan ===

Score: 100

Resources: 2
Capacity: 4

Assignments:

RES-001 (2/2)

- WI-001
- WI-002

RES-002 (1/2)

- WI-003

Unassigned:

None
```

---

# Tests

Add tests covering:

- assignment within capacity
- exact capacity
- insufficient capacity
- no resources
- zero capacity
- utilisation calculation
- optimisation score

The optimiser should remain deterministic.

---

# Constraints

Use only the Go standard library.

Do not introduce external dependencies.

Do not introduce interfaces or optimisation frameworks.

Keep the implementation intentionally simple.

If architectural ambiguity exists, stop and ask rather than making assumptions.

---

# Definition of Done

The implementation is complete when:

- the optimiser respects resource capacity
- assignments are deterministic
- optimisation score is reported
- utilisation is reported
- all tests pass
- go run executes successfully
- the architecture remains consistent with existing engineering principles

---

# Open Questions

None.