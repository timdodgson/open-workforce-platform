# Constraint Explanations

## Purpose

Explain why Work Items could not be assigned.

This improves transparency and makes optimisation decisions easier to understand.

---

# Why

An optimiser should not only produce a plan.

It should explain why parts of the plan could not be achieved.

This is particularly important when planners need to decide whether to:

- add more resources
- relax constraints
- change priorities
- extend shifts

---

# Behaviour

Every unassigned Work Item should include one or more explanation codes.

Examples:

- NoAvailableResource
- SkillMismatch
- CapacityExceeded
- TimeWindowExceeded
- ShiftEnded
- TravelTimeExceeded

Multiple explanations may exist for the same Work Item.

The optimiser should report every applicable reason rather than stopping at the first failure.

---

# CLI Output

Example:

Unassigned:

  WI-EVT-004

    Reasons:
      - SkillMismatch
      - TimeWindowExceeded

---

# Architecture

Constraint evaluation remains inside the optimisation layer.

The CLI displays explanations but does not calculate them.

Explanation generation should reuse existing constraint evaluation where practical.

Avoid duplicating validation logic.

---

# Non-Goals

Do not implement:

- natural language generation
- configurable messages
- localisation
- UI changes

Use simple symbolic explanation codes.

---

# Tests

Add tests covering:

- skill mismatch
- capacity exceeded
- time window exceeded
- multiple simultaneous reasons
- assigned work has no explanations
- CLI output

---

# Constraints

Use only the Go standard library.

Do not modify domain objects.

Do not introduce external dependencies.

Keep the implementation intentionally simple.

---

# Definition of Done

The implementation is complete when:

- unassigned work items expose explanation codes
- CLI displays explanations
- existing optimisation behaviour is unchanged
- tests pass
- no external dependencies are introduced

---

# Open Questions

None.