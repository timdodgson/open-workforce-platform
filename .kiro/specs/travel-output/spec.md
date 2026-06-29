# Travel Output

## Purpose

Improve command-line output so the travel penalty shown in the objective breakdown is explainable.

This is a presentation change only.

---

# Why

The CLI now shows:

Travel Time: -50

This explains the penalty value but not where it came from.

Users should be able to see the travel legs that contributed to the travel objective.

---

# Behaviour

For each Resource, display the travel route used for its assigned Work Items.

Example:

Travel:

  RES-001
    BASE-NORTH -> LOC-A: 10 mins
    LOC-A -> LOC-C: 15 mins
    Total: 25 mins

  RES-002
    BASE-SOUTH -> LOC-B: 25 mins
    Total: 25 mins

The total travel minutes should match the absolute value of the travel objective contribution.

---

# Architecture

The CLI should not calculate optimisation scores.

It may format travel information using:

- the assignments in the Optimised Plan
- Resource locations
- Work Item locations
- the travel matrix already loaded from the dataset

Do not duplicate objective scoring logic.

---

# Non-Goals

Do not implement:

- route optimisation
- map display
- geocoding
- external APIs
- travel affecting objectives differently
- new algorithms

---

# Tests

Add or update tests covering:

- travel output is displayed
- travel legs are shown in assignment order
- resource travel totals are shown
- existing commands still run

---

# Constraints

Use only the Go standard library.

Do not change optimisation behaviour.

Do not modify domain objects.

Do not introduce external dependencies.

Keep the change intentionally simple.

---

# Definition of Done

The implementation is complete when:

- CLI output explains travel time by Resource
- travel totals align with the travel objective
- all tests pass
- existing commands still run successfully
- no external dependencies are introduced

---

# Open Questions

None.