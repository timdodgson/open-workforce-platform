# Availability Constraint

## Purpose

Introduce resource availability into the optimiser.

The optimiser should only assign Work Items to Resources that are available.

This is the platform's second optimisation constraint after capacity.

---

# Why

Capacity determines how much work a Resource can take.

Availability determines whether a Resource can take work at all.

A Resource that is unavailable should not receive assignments, even if it has capacity.

---

# Problem Statement

Given:

- a collection of Work Items
- a collection of Resources
- resource capacity
- resource availability

The optimiser should assign Work Items only to Resources that are available and have remaining capacity.

---

# Availability

Availability remains business information.

The Resource domain object remains generic.

Availability is supplied within the Resource details JSON.

Example:

{
  "capacity": 2,
  "available": true
}

The application layer is responsible for interpreting availability.

The optimisation layer receives availability as structured optimisation input.

If a Resource has no availability value, treat it as unavailable.

---

# Behaviour

The optimiser should:

- ignore unavailable Resources
- assign Work Items only to available Resources
- continue to respect capacity
- continue to prioritise higher-priority Work Items
- remain deterministic

If no Resources are available:

- no Work Items should be assigned
- all Work Items should be reported as unassigned
- score should reflect the number of assigned Work Items

---

# Non-Goals

Do not implement:

- calendars
- working hours
- shifts
- time windows
- skills
- routing
- weighted objectives
- generic constraint framework

---

# Optimised Plan

The Optimised Plan should continue to report:

- assignments
- assigned work item count
- unassigned work item count
- total resource capacity
- utilisation percentage
- optimisation score

Unavailable Resources should not receive assignments.

---

# Tests

Add tests covering:

- available Resources can receive assignments
- unavailable Resources do not receive assignments
- missing availability defaults to unavailable
- capacity is still respected
- priority ordering is still respected
- no available Resources results in all Work Items unassigned
- optimiser remains deterministic

---

# Constraints

Use only the Go standard library.

Do not introduce external dependencies.

Do not introduce interfaces or a generic constraint framework.

Do not add availability fields to the Resource domain object.

Keep the implementation intentionally simple.

---

# Definition of Done

The implementation is complete when:

- unavailable Resources are never assigned Work Items
- capacity constraints are still respected
- priority ordering is still respected
- all tests pass
- the command-line example still runs
- no external dependencies are introduced
- architecture remains consistent with steering documents

---

# Open Questions

None.