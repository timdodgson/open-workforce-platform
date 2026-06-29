# NRP Single Assignment Per Day

## Purpose

Implement the NRP hard constraint that a nurse may work at most one shift per day.

---

## Why

In Nurse Rostering Problems, a nurse should not normally be assigned multiple shifts on the same day.

The current simplified NRP adapter can produce multiple shift demands on the same day, and the optimiser may assign more than one of those shifts to the same nurse if the time windows fit.

This is not valid for standard nurse rostering.

---

## Behaviour

A Resource must not be assigned more than one Work Item on the same roster day.

For NRP-generated Work Items, the day should be extracted from the Work Item details.

Example:

```json
{
  "day": 1,
  "shift": "early",
  "duration": 480,
  "requiredSkill": "general"
}
```

If two Work Items have the same day value, they must not be assigned to the same Resource.

---

## Architecture

This is a hard constraint.

It belongs in optimisation validation.

Domain objects must remain unchanged.

The NRP adapter should include day information in Work Item details.

The application layer should extract day into WorkItemInput.

The optimisation layer should enforce the constraint.

---

## Existing Behaviour

Existing non-NRP datasets should continue to work.

If a Work Item has no day value, treat it as day 0.

Day 0 means no day-level constraint applies.

---

## Constraint Explanations

If a Work Item cannot be assigned because the Resource is already assigned another Work Item on the same day, include explanation code:

SameDayAssignment

---

## Tests

Add tests covering:

- same nurse cannot receive two shifts on the same day
- different nurses can receive shifts on the same day
- same nurse can receive shifts on different days
- missing day defaults to no day-level constraint
- constraint explanation includes SameDayAssignment
- all algorithms still respect the constraint
- NRP benchmark still runs

---

## Non-Goals

Do not implement:

- rest periods
- maximum consecutive days
- weekend constraints
- contract constraints
- fairness objectives

---

## Constraints

Use only the Go standard library.

Do not modify domain objects.

Do not change optimisation algorithms unnecessarily.

Keep the implementation intentionally simple.

---

## Definition of Done

The implementation is complete when:

- WorkItemInput carries day information
- NRP adapter populates day
- same-resource same-day assignments are rejected
- explanation code is shown where relevant
- all tests pass
- benchmark runner still succeeds

---

## Open Questions

None.