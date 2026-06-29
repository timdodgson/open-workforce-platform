# INRC-II Compliance

## Purpose

Implement full INRC-II-style nurse rostering constraint support for the Open Workforce Platform.

This is not a partial NRP feature.

The platform should support the recognised INRC-II hard and soft constraint model so nurse rostering datasets can be evaluated meaningfully and consistently.

---

## Why

The platform now has:

- NRP adapter
- optimisation context
- hard constraint validation
- objective scoring
- algorithm benchmarking
- multiple search algorithms

Before adding custom algorithms, the platform must ensure it is solving the correct nurse rostering problem.

The goal is to align the optimiser with INRC-II-style nurse rostering constraints rather than continuing with simplified synthetic constraints.

---

## Target Model

Implement INRC-II-style constraints.

The optimiser should support:

- hard constraints that invalidate plans
- soft constraints that contribute penalties or rewards to objective scoring
- constraint explanations for invalid or unassigned work
- objective breakdown for soft-constraint penalties

The optimiser must remain algorithm-independent.

Algorithms should continue to ask:

1. Is this plan valid?
2. What is this plan's objective score?

Algorithms must not contain nurse rostering business rules.

---

# Hard Constraints

## H1 - Single Assignment Per Day

A nurse may be assigned to at most one shift on the same day.

If two Work Items have the same roster day, they must not be assigned to the same Resource.

Violation code:

SameDayAssignment

---

## H2 - Under-Staffing / Minimum Coverage

For each day, shift type and skill, the number of assigned nurses must meet the minimum required demand.

In the current platform model, each generated demand Work Item represents one required coverage unit.

A plan is incomplete if one of these Work Items remains unassigned.

For INRC-II compliance, unassigned mandatory demand should be treated as a hard coverage violation.

Violation code:

UnderStaffed

Note:

The platform may still return a partial plan for explainability, but the plan should be marked as having hard violations.

---

## H3 - Legal Shift Type Successions

A nurse may only work shift sequences that are legal according to the scenario.

Example:

- Late followed by Early may be illegal
- Night followed by Early may be illegal

The NRP adapter must expose shift type on WorkItemInput.

The optimisation layer must validate consecutive assigned shifts for each nurse.

Violation code:

IllegalShiftSuccession

---

# Soft Constraints

## S1 - Optimal Coverage

For each day, shift type and skill, coverage above the minimum but below the optimal staffing level should incur a penalty.

This is separate from H2.

Minimum coverage is hard.

Optimal coverage is soft.

Objective contribution:

Optimal Coverage

---

## S2 - Consecutive Working Days

Each nurse contract may define:

- minimum consecutive working days
- maximum consecutive working days

Violations should be penalised.

Objective contribution:

Consecutive Working Days

---

## S3 - Consecutive Days Off

Each nurse contract may define:

- minimum consecutive days off
- maximum consecutive days off

Violations should be penalised.

Objective contribution:

Consecutive Days Off

---

## S4 - Consecutive Assignments To Same Shift Type

Each shift type may define:

- minimum consecutive assignments
- maximum consecutive assignments

Example:

A nurse may be allowed to work Night shifts, but too many consecutive Night shifts should be penalised.

Objective contribution:

Consecutive Shift Type

---

## S5 - Working Weekends

Each nurse contract may define a maximum number of working weekends.

A weekend is considered worked if the nurse works Saturday and/or Sunday.

Objective contribution:

Working Weekends

---

## S6 - Complete Weekend

For nurses whose contract requires complete weekends:

- working both Saturday and Sunday is valid
- working neither is valid
- working only one of Saturday or Sunday is penalised

Objective contribution:

Complete Weekend

---

## S7 - Total Assignments

Each nurse contract may define:

- minimum total assignments
- maximum total assignments

Violations are penalised at the end of the planning horizon.

Objective contribution:

Total Assignments

---

## S8 - Shift Requests

Nurses may request:

- shift on
- shift off

A violated request should produce a soft penalty.

Objective contribution:

Shift Requests

---

## S9 - Day Requests

Nurses may request:

- day on
- day off

A violated request should produce a soft penalty.

Objective contribution:

Day Requests

---

# Input Model Extensions

The simplified NRP JSON format must be extended to support INRC-II-style data.

## Scenario

Add:

- skills
- shiftTypes
- forbiddenShiftSuccessions
- contracts

## Nurses

Each nurse should have:

- id
- skills
- contractId
- availability where applicable

## Contracts

Each contract should support:

- minAssignments
- maxAssignments
- minConsecutiveWorkingDays
- maxConsecutiveWorkingDays
- minConsecutiveDaysOff
- maxConsecutiveDaysOff
- maxWorkingWeekends
- completeWeekend

## Shift Types

Each shift type should support:

- id
- startMinute
- endMinute
- duration
- minConsecutiveAssignments
- maxConsecutiveAssignments

## Coverage Requirements

Each requirement should support:

- day
- shiftType
- skill
- minimum
- optimal

## Requests

Requests should support:

- nurseId
- day
- shiftType optional
- type: shiftOn, shiftOff, dayOn, dayOff
- weight

---

# OWP Mapping

The NRP adapter should convert NRP input into OWP-compatible data.

## Nurses

Map to Resources.

Resource details should include:

- skills
- contractId
- available
- shiftStart
- shiftEnd
- NRP metadata where needed

## Coverage Demand

Each minimum required coverage unit becomes a mandatory Work Item.

Each optimal-but-not-minimum coverage unit may be represented as optional demand or tracked separately for soft coverage scoring.

## Work Item Details

Work Item details should include:

- day
- shiftType
- requiredSkill
- duration
- earliestStart
- latestFinish
- mandatory true/false
- demandGroup
- coverageMinimum
- coverageOptimal

---

# Optimisation Input Extensions

ResourceInput should carry:

- ContractID
- Skills
- Availability
- Day-level availability where available

WorkItemInput should carry:

- Day
- ShiftType
- RequiredSkill
- Duration
- EarliestStart
- LatestFinish
- Mandatory
- DemandGroup

OptimisationContext should carry:

- Contracts
- ShiftTypes
- ForbiddenShiftSuccessions
- Requests
- Coverage requirements
- Existing plan/history where available

---

# Plan Validity

OptimisedPlan should distinguish:

- assignment score
- objective score
- hard constraint violations
- soft constraint penalties
- unassigned mandatory demand
- unassigned optional demand

Plans may still be returned when hard violations exist, but they must be clearly marked.

The CLI must display hard violations separately from soft objective penalties.

---

# Objective Breakdown

Extend objective breakdown to include INRC-II soft constraints:

- Assignment
- Workload Balance
- Travel Time
- Preferred Resource
- Plan Stability
- Optimal Coverage
- Consecutive Working Days
- Consecutive Days Off
- Consecutive Shift Type
- Working Weekends
- Complete Weekend
- Total Assignments
- Shift Requests
- Day Requests

---

# Constraint Explanations

Add explanation codes:

- SameDayAssignment
- UnderStaffed
- IllegalShiftSuccession
- BelowOptimalCoverage
- TooFewAssignments
- TooManyAssignments
- TooFewConsecutiveWorkingDays
- TooManyConsecutiveWorkingDays
- TooFewConsecutiveDaysOff
- TooManyConsecutiveDaysOff
- TooFewConsecutiveShiftType
- TooManyConsecutiveShiftType
- TooManyWorkingWeekends
- IncompleteWeekend
- ShiftRequestViolated
- DayRequestViolated

---

# CLI

Existing commands must continue to work.

The converted NRP dataset should be runnable with:

go run ./cmd/owp optimise ../../examples/datasets/nrp-simple.json --algorithm tabu-search

Benchmark runner should continue to include NRP datasets:

go run ./cmd/owp benchmark ../../examples/datasets

The CLI should display:

- hard violations
- objective breakdown
- INRC-II soft penalties
- unassigned mandatory demand
- benchmark delta
- benchmark delta percentage

---

# Tests

Add tests covering:

## Hard Constraints

- nurse cannot work two shifts on same day
- minimum coverage violation is detected
- illegal shift succession is rejected
- skill mismatch remains enforced

## Soft Constraints

- below optimal coverage penalty
- minimum total assignments penalty
- maximum total assignments penalty
- minimum consecutive working days penalty
- maximum consecutive working days penalty
- minimum consecutive days off penalty
- maximum consecutive days off penalty
- consecutive same-shift penalty
- working weekend penalty
- complete weekend penalty
- shift-off request penalty
- shift-on request penalty
- day-off request penalty
- day-on request penalty

## Integration

- simple NRP JSON converts successfully
- converted NRP dataset optimises successfully
- all algorithms run against the NRP dataset
- benchmark runner includes NRP dataset
- hard violations are reported
- soft penalties appear in objective breakdown
- deterministic output

---

# Non-Goals

Do not implement official INRC-II XML parsing in this spec.

Do not implement multi-stage history files yet.

Do not implement the official validator file format yet.

Do not introduce external dependencies.

Do not change domain objects unnecessarily.

Do not hard-code nurse rostering logic inside algorithms.

---

# Architecture Rules

NRP business rules belong in:

- NRP adapter
- optimisation input extraction
- hard constraint validation
- objective scoring

NRP business rules must not be placed in:

- domain entities
- individual algorithms
- CLI scoring logic

Algorithms must remain generic.

---

# Definition of Done

The implementation is complete when:

- INRC-II hard constraints are represented and enforced
- INRC-II soft constraints are represented and scored
- NRP adapter supports contracts, shift types, coverage, requests and forbidden successions
- OptimisationContext carries all required NRP data
- OptimisedPlan reports hard violations and soft penalties
- CLI displays hard violations and objective breakdown
- benchmark runner still works
- NRP benchmark dataset demonstrates INRC-II-style constraints
- all tests pass
- no external dependencies are introduced

---

# Open Questions

None.