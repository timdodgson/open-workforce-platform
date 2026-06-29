# CLI Utilisation Output

## Purpose

Improve command-line output so Resource utilisation is displayed in minutes rather than mixing assignment count with capacity minutes.

This is a presentation change only.

---

# Why

The optimiser now uses duration-based capacity and travel-aware scheduling.

The current output shows:

RES-001 (2/480)

This mixes two different units:

- assigned Work Item count
- available minutes

The output should show used minutes against available minutes.

---

# Behaviour

For each Resource, display:

- Resource ID
- used minutes
- available capacity minutes
- assigned Work Items

Example:

RES-001
  Used: 135 / 480 mins
  Work Items:
    - WI-EVT-001
    - WI-EVT-003

---

# Architecture

The CLI should not implement optimisation logic.

If required, used minutes should be calculated from data already available in the plan, application result or optimisation layer.

Avoid duplicating complex scheduling logic in the CLI.

---

# Non-Goals

Do not implement:

- route display
- start/end times
- travel breakdown
- calendar output
- UI changes

---

# Tests

Add or update tests covering:

- CLI output no longer mixes job count and minutes
- used minutes are displayed
- existing optimisation commands still run

---

# Constraints

Use only the Go standard library.

Do not change optimisation behaviour.

Do not modify domain objects.

Keep the change intentionally simple.

---

# Definition of Done

The implementation is complete when:

- CLI output displays Resource utilisation in minutes
- CLI no longer shows assignment count divided by minute capacity
- all tests pass
- existing commands still run successfully
- no external dependencies are introduced

---

# Open Questions

None.