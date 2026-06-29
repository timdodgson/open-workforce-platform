# Priority Optimisation

## Purpose

Introduce priority-aware assignment.

The optimiser should prefer higher-priority Work Items when resource capacity is limited.

This is the platform's first optimisation objective.

---

# Why

Capacity prevents Resources from being overloaded.

Priority determines which Work Items should be assigned first when not all Work Items can be completed.

This moves the optimiser from simple constraint satisfaction toward making better business decisions.

---

# Problem Statement

Given:

- a collection of Work Items
- a collection of Resources
- resource capacity

Each Work Item may have a priority value in its details JSON.

The optimiser should assign higher-priority Work Items before lower-priority Work Items when capacity is limited.

---

# Priority

Priority remains business information.

The WorkItem domain object remains generic.

Priority is supplied within the Work Item details JSON.

Example:

{
  "priority": 100
}

The application layer is responsible for interpreting priority.

The optimisation layer receives priority as structured optimisation input.

---

# Behaviour

The optimiser should:

- read priority values prepared by the application layer
- sort Work Items by priority before assignment
- assign higher-priority Work Items first
- remain deterministic when priorities are equal

If a Work Item has no priority, treat it as priority 0.

Higher numbers mean higher priority.

---

# Non-Goals

Do not implement:

- availability
- skills
- routing
- calendars
- working hours
- weighted objectives
- generic objective framework

---

# Optimised Plan

The Optimised Plan should continue to report:

- assignments
- assigned work item count
- unassigned work item count
- total resource capacity
- utilisation percentage
- optimisation score

If low-priority Work Items are left unassigned because capacity is limited, that is acceptable.

---

# Tests

Add tests covering:

- higher-priority Work Items are assigned first
- lower-priority Work Items are left unassigned when capacity is limited
- missing priority defaults to 0
- equal priorities are handled deterministically
- existing capacity behaviour still works

---

# Constraints

Use only the Go standard library.

Do not introduce external dependencies.

Do not introduce a generic objective framework.

Do not add priority fields to the WorkItem domain object.

Keep the implementation intentionally simple.

---

# Definition of Done

The implementation is complete when:

- the optimiser assigns higher-priority Work Items first
- capacity constraints are still respected
- all tests pass
- the command-line example still runs
- no external dependencies are introduced
- architecture remains consistent with steering documents

---

# Open Questions

None.