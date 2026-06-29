# Optimisation Input Rename

## Purpose

Rename optimisation input types so their names match their current responsibilities.

This is a refactor-only task.

No behaviour should change.

---

# Why

The optimiser now consumes richer structured inputs.

ResourceCapacity now contains:

- resource identity
- capacity
- availability
- skills

WorkItemPriority now contains:

- work item identity
- priority
- required skill
- duration

The old names are now misleading.

The new names should describe their role as optimisation inputs.

---

# Required Renames

Rename:

ResourceCapacity

to:

ResourceInput

Rename:

WorkItemPriority

to:

WorkItemInput

---

# Behaviour

No behaviour should change.

The following must remain true:

- capacity is respected
- availability is respected
- skills are respected
- priority is respected
- duration is respected
- all algorithms continue to work
- CLI output remains unchanged

---

# Constraints

Use only the Go standard library.

Do not introduce new abstractions.

Do not change algorithm behaviour.

Do not change domain objects.

Do not change CLI output.

---

# Tests

Existing tests should continue to pass.

Update test names only where helpful.

Do not rewrite tests unnecessarily.

---

# Definition of Done

The refactor is complete when:

- ResourceCapacity has been renamed to ResourceInput
- WorkItemPriority has been renamed to WorkItemInput
- all references are updated
- all tests pass
- CLI commands still run successfully
- behaviour is unchanged

---

# Open Questions

None.