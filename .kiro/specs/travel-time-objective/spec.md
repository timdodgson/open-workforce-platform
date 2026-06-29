# Travel Time Objective

## Purpose

Introduce travel time into optimisation scoring.

The optimiser should prefer valid plans with lower travel time.

This is the first operational cost objective.

---

# Why

A valid plan is not always a good plan.

Two plans may both satisfy:

- capacity
- availability
- skills
- time windows

but one may require significantly more travel.

Travel time makes the optimiser choose between valid plans based on operational cost.

---

# Behaviour

Travel time should influence objective scoring.

Lower travel time should produce a better objective score.

Assignment should remain the dominant objective.

The optimiser must never leave work unassigned purely to reduce travel time.

---

# Travel Matrix

Introduce a simple travel matrix.

Travel time is supplied as minutes between locations.

For this iteration, locations may be simple string identifiers.

Example:

{
  "from": "RES-001",
  "to": "LOC-A",
  "minutes": 15
}

Resources provide a starting location in details JSON.

Work Items provide a location in details JSON.

---

# Architecture

Travel data is business input.

Resource and WorkItem domain objects remain unchanged.

The application layer extracts:

- resource start location
- work item location
- travel matrix

The optimisation layer receives travel data through OptimisationContext.

Objective scoring uses the travel matrix.

---

# Objective Scoring

Add a travel objective contribution.

Travel should reduce the objective score.

Example:

Objective Breakdown:
  Assignment: 3000
  Workload Balance: 1
  Travel Time: -45

The total objective score should equal the sum of all objective contributions.

---

# Scheduling

For this iteration:

- travel time affects scoring
- travel time does not affect time window feasibility

Do not add travel time into schedule duration yet.

That will be a future capability.

---

# Dataset

Update the example dataset to include:

- resource start locations
- work item locations
- travel times

---

# Non-Goals

Do not implement:

- routing
- multi-stop route sequencing
- real maps
- geocoding
- external APIs
- travel affecting schedule feasibility
- distance calculations

---

# Tests

Add tests covering:

- lower travel time gives better objective score
- objective breakdown includes travel time
- missing travel time defaults to zero penalty
- assignment remains dominant over travel
- existing algorithms still work
- CLI output includes travel objective contribution

---

# Constraints

Use only the Go standard library.

Do not introduce external dependencies.

Do not call external APIs.

Do not modify domain objects.

Keep the implementation intentionally simple.

---

# Definition of Done

The implementation is complete when:

- travel time is represented in optimisation input
- objective scoring includes travel penalty
- objective breakdown explains travel contribution
- all algorithms still work
- all tests pass
- CLI command runs successfully
- no external dependencies are introduced

---

# Open Questions

None.