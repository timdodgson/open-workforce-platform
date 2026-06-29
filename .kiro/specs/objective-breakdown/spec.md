# Objective Breakdown

## Purpose

Expose the individual objective contributions that make up the overall objective score.

The optimiser currently reports a single objective score.

The goal is to make the score explainable by showing which objectives contributed to it.

---

# Why

Objective scoring is only useful if users can understand it.

A total score explains which plan is better.

A breakdown explains why.

This improves explainability and prepares the platform for future objectives such as travel distance, SLA compliance, preferred resources and continuity.

---

# Behaviour

The CLI should display an objective breakdown.

Example:

Objective Score: 3001

Objective Breakdown:
  Assignment: 3000
  Workload Balance: 1

The total objective score should equal the sum of the breakdown values.

---

# Architecture

Objective scoring remains in the optimisation layer.

The CLI must not calculate objective scores or breakdowns.

The Optimised Plan may expose the objective breakdown if required.

---

# Non-Goals

Do not implement:

- configurable objective weights
- runtime objective selection
- external reporting
- UI changes
- new objectives beyond the existing assignment and workload balance objectives

---

# Tests

Add or update tests covering:

- objective breakdown is available
- breakdown total equals objective score
- assignment objective contribution
- workload balance objective contribution
- CLI output includes the breakdown
- algorithms still work

---

# Constraints

Use only the Go standard library.

Do not introduce external dependencies.

Do not duplicate objective scoring logic in the CLI.

Keep the implementation simple.

---

# Definition of Done

The implementation is complete when:

- objective breakdown is calculated in the optimisation layer
- CLI output displays the objective breakdown
- total objective score equals the sum of objective contributions
- all tests pass
- existing commands still run successfully
- no external dependencies are introduced

---

# Open Questions

None.