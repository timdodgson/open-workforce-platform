# Travel-Aware Scheduling

## Purpose

Make travel time affect schedule feasibility.

Travel time currently affects objective scoring only.

This change ensures Work Items are only assigned when the Resource has enough time to travel and complete the work within the relevant time windows.

---

# Why

A plan that ignores travel time may appear valid but be impossible to execute.

If a Resource has a shift from 08:00 to 17:00 and a Work Item takes 90 minutes, the Resource also needs enough time to travel to the Work Item location.

Travel must therefore be considered when checking whether a schedule is feasible.

---

# Behaviour

When assigning a Work Item to a Resource, the optimiser should consider:

- the Resource's current location
- travel time from the current location to the Work Item location
- the Work Item duration
- the Resource shift end
- the Work Item latest finish

A Work Item can be assigned only if:

travel start time + travel time + duration <= Resource shift end

and

work start time + duration <= Work Item latest finish

---

# Sequential Scheduling

Continue using the existing simple sequential scheduling model.

For each Resource:

1. Start at the Resource's start location.
2. Start at the Resource's shift start time.
3. For each assigned Work Item:
   - travel from the current location to the Work Item location
   - wait until the Work Item's earliest start if arriving early
   - perform the Work Item
   - update current time
   - update current location

No routing optimisation is introduced in this iteration.

---

# Travel Matrix

Use the existing travel matrix.

If travel time is missing between two locations, treat it as zero for now.

Do not call external APIs.

---

# Architecture

Travel time remains business input.

Resource and WorkItem domain objects remain unchanged.

The application layer extracts travel-related data.

The optimisation layer uses travel data from OptimisationContext.

---

# Objective Scoring

Travel time should continue to affect objective scoring.

Do not remove the travel time objective.

---

# Non-Goals

Do not implement:

- route optimisation
- real maps
- geocoding
- external APIs
- breaks
- overtime
- multi-day scheduling

---

# Tests

Add tests covering:

- assignment succeeds when travel plus duration fits
- assignment fails when travel plus duration exceeds shift
- assignment fails when travel plus duration exceeds latest finish
- waiting until earliest start works
- sequential location updates are respected
- travel objective still appears in the objective breakdown
- existing algorithms still work

---

# Constraints

Use only the Go standard library.

Do not modify domain objects.

Do not introduce external dependencies.

Keep the implementation intentionally simple.

---

# Definition of Done

The implementation is complete when:

- travel time affects schedule feasibility
- travel objective still affects scoring
- all algorithms still work
- all tests pass
- CLI command runs successfully
- no external dependencies are introduced

---

# Open Questions

None.