# NRP Constraint Catalogue

## Purpose

This document catalogues the hard and soft constraints the Open Workforce Platform should support for Nurse Rostering Problem datasets.

It serves as a design reference for future implementation work.

---

## Hard Constraints

Hard constraints determine whether a plan is valid. A plan that violates a hard constraint is invalid and must be rejected.

| Constraint | Description | Implemented | Owner | Priority |
|------------|-------------|-------------|-------|----------|
| Single assignment per shift | A nurse can only work one shift per day | Partially (via time windows) | Optimisation | High |
| Required skill | A nurse must have the skill required by the shift | Yes | Optimisation (canAccept) | Done |
| Staffing demand | Required number of nurses per shift should be satisfied | Yes (via work item generation) | NRP Adapter | Done |
| No overlapping shifts | A nurse cannot be assigned two shifts that overlap in time | Yes (via sequential scheduling) | Optimisation (scheduleFeasible) | Done |
| Resource availability | Only available nurses can be assigned | Yes | Optimisation (canAccept) | Done |
| Shift time windows | Work must be completed within shift boundaries | Yes | Optimisation (assignItems) | Done |
| Maximum working minutes | A nurse cannot exceed their daily capacity | Yes (via shiftEnd) | Optimisation | Done |
| Rest period rules | Minimum hours between consecutive shifts | No | Optimisation | Medium |

---

## Soft Constraints

Soft constraints represent preferences. Violating a soft constraint does not invalidate a plan but reduces its objective score.

| Constraint | Description | Implemented | Owner | Priority |
|------------|-------------|-------------|-------|----------|
| Preferred day on / day off | Nurse requests specific days on or off | No | Objective scoring | Medium |
| Preferred shift | Nurse prefers specific shift types | Partially (preferredResource maps to this) | Objective scoring | Medium |
| Minimum assigned shifts | Each nurse should work at least N shifts per period | No | Objective scoring | Medium |
| Maximum assigned shifts | Each nurse should not exceed N shifts per period | No | Objective scoring | Medium |
| Minimum consecutive working days | Avoid isolated single working days | No | Objective scoring | Low |
| Maximum consecutive working days | Limit how many days in a row a nurse works | No | Objective scoring | High |
| Minimum consecutive days off | Ensure meaningful rest periods | No | Objective scoring | Medium |
| Maximum consecutive days off | Prevent excessive absence | No | Objective scoring | Low |
| Weekend fairness | Distribute weekend shifts fairly across nurses | No | Objective scoring | Medium |
| Night shift fairness | Distribute night shifts fairly | No | Objective scoring | Medium |
| Workload balance | Even distribution of total work across nurses | Yes | Objective scoring (balanceObjective) | Done |
| Plan stability | Minimise unnecessary changes from existing plans | Yes | Objective scoring (stabilityObjective) | Done |
| Preference satisfaction | Honour nurse shift preferences | Yes | Objective scoring (preferredResourceObjective) | Done |

---

## Current Platform Coverage

The platform currently implements:

**Hard constraints (fully supported):**
- Skills
- Availability
- Capacity / duration
- Time windows
- Travel-aware scheduling
- Sequential shift enforcement

**Soft constraints (fully supported):**
- Workload balance
- Plan stability
- Preferred resource / shift preference
- Travel minimisation

---

## Implementation Approach

### Hard constraints

Hard constraints belong in `canAccept()` and `scheduleFeasible()` within the optimisation layer.

New hard constraints should:
- Be validated during assignment
- Be checked by `scheduleFeasible` for search algorithm validation
- Generate explanation codes when violated

### Soft constraints

Soft constraints belong in `objective.go` as additional objective functions.

New soft constraints should:
- Return a raw value (count or measure)
- Be multiplied by their weight from `ObjectiveWeights`
- Appear in the objective breakdown
- Never invalidate a plan

### Multi-day extension

Many NRP constraints (consecutive days, rest periods, fairness) require multi-day planning. This requires:

1. Extending the NRP adapter to generate multi-day datasets
2. Extending time representation beyond single-day minutes
3. Adding nurse-level tracking across multiple days

This is the next major capability after the current single-day model.

---

## Suggested Implementation Priority

### Phase 1 (Current)
- ✅ Skills
- ✅ Availability
- ✅ Time windows
- ✅ Workload balance
- ✅ Preference satisfaction
- ✅ Plan stability

### Phase 2 (Next)
- Rest period rules (hard)
- Maximum consecutive working days (soft)
- Multi-day NRP adapter
- Preferred day on / day off (soft)

### Phase 3 (Future)
- Weekend fairness (soft)
- Night shift fairness (soft)
- Minimum/maximum assigned shifts (soft)
- Consecutive days off rules (soft)

---

## References

- INRC-I (International Nurse Rostering Competition 2010)
- INRC-II (International Nurse Rostering Competition 2014)
- Burke et al., "The State of the Art of Nurse Rostering" (2004)
