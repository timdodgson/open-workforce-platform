# Candidate Move Review

## Purpose

Review and simplify the CandidateMove model after adding placement, displacement and swap moves.

This is a refactor-only task.

No optimisation behaviour should change.

---

# Why

CandidateMove now supports multiple move shapes.

Before adding more optimisation behaviour, the model should be reviewed for clarity.

The goal is to ensure CandidateMove remains understandable and has not become an accidental catch-all structure.

---

# Scope

Review the current CandidateMove implementation.

Refactor only if the current structure can be made clearer without introducing unnecessary abstraction.

---

# Behaviour

The following behaviour must remain unchanged:

- placement moves still work
- displacement moves still work
- swap moves still work
- hill climbing still works
- simulated annealing still works
- CLI output remains unchanged

---

# Review Questions

Answer these before changing code:

- Is CandidateMove still easy to understand?
- Are optional fields clear?
- Are move types mutually exclusive?
- Would a future engineer understand how to add another move type?
- Would splitting into multiple structs improve clarity or introduce unnecessary complexity?

---

# Constraints

Use only the Go standard library.

Do not introduce:

- interfaces
- class hierarchies
- generic move frameworks
- external dependencies

Do not change public optimiser behaviour.

---

# Definition of Done

The task is complete when:

- CandidateMove has been reviewed
- any justified refactor is implemented
- existing behaviour is unchanged
- all tests pass
- CLI commands still run

---

# Open Questions

None.