# INRC-II Validator Parity

## Purpose

Achieve exact scoring parity with the official INRC-II validator.

The current implementation supports the official file formats, multi-stage history, solution generation and constraint model.

However, the platform does not yet guarantee identical scoring to the official validator in every case.

This specification completes that work.

The official validator becomes the single source of truth.

---

## Goal

Given any official INRC-II competition instance and solution:

- the platform must report exactly the same hard constraint violations
- the platform must report exactly the same soft penalties
- the platform must report exactly the same total score

No approximations are acceptable.

---

## Source of Truth

The official validator supplied with the competition.

The implementation must never intentionally differ from the validator.

If documentation and validator disagree, the validator wins.

---

## Scope

This specification covers scoring only.

It does not change:

- optimisation algorithms
- neighbourhood generation
- search strategies
- constructive assignment
- CLI behaviour (other than reporting)

---

## Validation Process

For every official sample instance:

1. Run the official validator.

2. Capture:

- hard violations
- every soft constraint contribution
- total score

3. Run the platform validator.

4. Compare every value.

5. Eliminate every difference.

Repeat until the platform produces identical output.

---

## Hard Constraint Validation

Verify complete parity for:

### H1

Single assignment per nurse per day.

### H2

Required skill.

### H3

Minimum coverage.

### H4

Forbidden shift succession.

Including:

- previous history
- first day
- last day
- week boundaries

---

## Soft Constraint Validation

Verify complete parity for:

### S1

Optimal coverage.

### S2

Consecutive working days.

### S3

Consecutive days off.

### S4

Consecutive shift assignments.

### S5

Shift requests.

### S6

Complete weekends.

### S7

Total assignments.

### S8

Working weekends.

Every penalty must match exactly.

---

## Multi-Stage Validation

Verify every week transition.

Including:

- previous shift type
- previous working streak
- previous days off
- previous shift streak
- accumulated weekends
- accumulated assignments

History after week N must produce identical scoring for week N+1.

---

## History Update

The generated history after solving a week must be identical to the history expected by the official validator.

Verify:

- last shift
- consecutive working days
- consecutive days off
- consecutive shift type
- working weekends
- assignment totals

---

## Edge Cases

Validate:

- empty weekends
- partial weekends
- nurses with no assignments
- maximum streak boundaries
- minimum streak boundaries
- shift transitions across weeks
- zero coverage
- excess coverage
- requests on free days
- requests on working days

---

## Scorer Refactoring

If required, refactor the scorer.

Correctness is more important than implementation simplicity.

Do not duplicate logic.

The official scorer should remain deterministic and side-effect free.

---

## Regression Tests

Add regression tests for every discrepancy discovered.

Every bug fixed during parity work must become a permanent automated test.

---

## Benchmark Validation

Run validator parity on:

- n005w4
- every official sample instance included in the repository

The benchmark should report:

- validator score
- platform score
- score difference

Definition of done is:

Difference = 0

for every official sample.

---

## Reporting

validate-inrc2 should display:

Official Validator Score

Platform Score

Difference

If Difference is non-zero:

display every differing constraint contribution.

---

## Architecture Rules

Do not change optimisation algorithms.

Do not embed scoring logic inside algorithms.

Do not change OptimisationContext.

All parity work belongs inside the INRC-II infrastructure package.

---

## Files Expected

Likely files:

internal/infrastructure/inrc2/scorer.go

internal/infrastructure/inrc2/history.go

internal/infrastructure/inrc2/validator.go

parser_test.go

scorer_test.go

history_test.go

new parity regression tests

---

## Non Goals

Do not improve optimisation quality.

Do not introduce new objectives.

Do not add new move operators.

Do not optimise runtime.

This specification is purely about correctness.

---

## Definition of Done

The implementation is complete only when:

✓ Official validator and platform report identical hard violations.

✓ Official validator and platform report identical soft penalties.

✓ Official validator and platform report identical total score.

✓ Generated history produces identical results in following weeks.

✓ All official sample instances pass.

✓ Regression tests cover every discovered discrepancy.

✓ Existing benchmark suite continues to pass.

Only when every official sample produces a score difference of zero may the platform be described as fully INRC-II competition compliant.