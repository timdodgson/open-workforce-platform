# Objective Score Output

## Purpose

Expose objective scoring in the command-line output.

The optimiser now uses objective scoring internally, but the CLI still only displays the percentage assignment score.

The goal is to make optimisation quality more explainable.

---

# Why

Users need to understand why a plan is considered better.

The percentage score explains how many Work Items were assigned.

The objective score explains optimisation preference, including assignment count and workload balance.

Both are useful, but they answer different questions.

---

# Behaviour

The CLI should display:

- assignment score percentage
- objective score
- assignments
- unassigned Work Items

Do not remove the existing percentage score.

Rename the displayed percentage score to make it clear.

Example:

Assignment Score: 100
Objective Score: 3001

---

# Architecture

Objective scoring remains in the optimisation layer.

The Optimised Plan may expose the objective score if required.

The CLI should display values provided by the plan or application layer.

The CLI must not calculate objective scores itself.

---

# Non-Goals

Do not implement:

- weighted objective configuration
- objective breakdown by component
- external reporting
- UI changes
- new optimisation algorithms

---

# Tests

Add or update tests covering:

- objective score is available from the Optimised Plan or application result
- CLI output includes assignment score
- CLI output includes objective score
- existing algorithms still work

---

# Constraints

Use only the Go standard library.

Do not introduce external dependencies.

Do not duplicate objective scoring logic in the CLI.

Keep the implementation simple.

---

# Definition of Done

The implementation is complete when:

- CLI output clearly distinguishes assignment score from objective score
- objective score is calculated by the optimisation layer
- all tests pass
- existing commands still run successfully
- no external dependencies are introduced

---

# Open Questions

None.