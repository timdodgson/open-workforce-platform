# Constraint Match Reporting

## Purpose

Introduce first-class constraint match reporting into the Open Workforce Platform.

The optimiser currently reports:

- objective score
- objective breakdown
- hard violations

However, it does not report the actual number of soft constraint violations or which individual constraints were violated.

This specification introduces generic constraint reporting.

The optimiser must continue to optimise using weighted penalties, while also exposing the individual constraint matches that produced those penalties.

This capability is intended for:

- users
- UI
- reporting
- benchmarking
- explainability
- future optimisation algorithms

---

## Goals

The platform shall report:

- Hard violation count
- Soft violation count
- Constraint matches
- Constraint summaries
- Weighted penalty

without changing optimisation behaviour.

---

## Architecture

Constraint reporting is a platform capability.

It is **not** an INRC-II feature.

Every optimisation objective may optionally return constraint matches.

Algorithms remain completely unaware of them.

---

## Constraint Match Model

Introduce a generic model.

Example:

```go
type ConstraintMatch struct {
    Constraint string
    Severity string

    ResourceID string
    WorkItemID string

    Day int

    Penalty int

    Description string
}
```

Fields may be extended where useful.

The model must remain generic.

Do not introduce nurse-specific terminology.

---

## Objective Result

Introduce a reusable result model.

Example:

```go
type ObjectiveResult struct {
    Penalty int

    Matches []ConstraintMatch
}
```

Objective functions should return ObjectiveResult instead of only an integer where practical.

ObjectiveScore may continue to aggregate penalties exactly as today.

Search algorithms must not require modification.

---

## Hard Constraints

Hard constraints shall also expose matches.

Each hard violation should become a ConstraintMatch.

The existing HardViolation model may remain if useful, but both models must stay synchronised.

---

## Soft Constraints

Every soft constraint shall produce:

- weighted penalty
- list of individual matches

For example:

Coverage

Penalty = 120

Matches:

Coverage shortfall Monday Early

Coverage shortfall Tuesday Late

Coverage shortfall Friday Night

---

Weekend

Penalty = 60

Matches:

Nurse A incomplete weekend

Nurse C incomplete weekend

---

## Plan Output

OptimisedPlan shall expose:

ConstraintMatches()

SoftConstraintCount()

HardConstraintCount()

ConstraintSummary()

TotalPenalty()

The existing APIs must remain backward compatible.

---

## CLI

Update optimise output.

Display:

Hard Constraints

Total: 0

Soft Constraints

Total: 42

Breakdown

Coverage....................5

Weekend.....................8

Requests...................20

Assignments.................4

Consecutive................5

Penalty

1695

Do not remove the existing objective breakdown.

Both should be shown.

---

## Benchmark

Extend benchmark output.

Per dataset include:

Soft Violations

Hard Violations

Penalty

Summary section should include averages for:

Average Penalty

Average Soft Violations

Average Hard Violations

---

## INRC-II

Map every official constraint to ConstraintMatch entries.

One match should represent one violated constraint.

Penalty remains weighted exactly as required by the official validator.

---

## UI Support

The reporting model must support future UI requirements.

Consumers should be able to:

count violations

group by constraint

group by resource

group by work item

group by day

without recomputing the optimisation.

---

## Explainability

Constraint matches should contain meaningful descriptions.

Example:

Coverage below optimal for Tuesday Early.

Requested Friday Off not honoured.

Maximum consecutive working days exceeded.

Descriptions should be deterministic.

---

## Tests

Add tests verifying:

constraint counts

constraint summaries

penalty totals

objective totals unchanged

hard and soft counts

CLI output

benchmark output

INRC-II constraint match generation

backward compatibility

---

## Non Goals

Do not modify optimisation algorithms.

Do not change scoring behaviour.

Do not alter validator parity.

This specification is purely an explainability and reporting enhancement.

---

## Definition of Done

The implementation is complete only when:

✓ Every hard violation generates a ConstraintMatch.

✓ Every soft violation generates a ConstraintMatch.

✓ SoftConstraintCount() returns the total number of soft violations.

✓ HardConstraintCount() returns the total number of hard violations.

✓ TotalPenalty() equals the existing weighted penalty.

✓ CLI reports both counts and penalties.

✓ Benchmark reports counts and penalties.

✓ Existing objective scores remain unchanged.

✓ Official validator parity remains unchanged.

✓ All tests pass.