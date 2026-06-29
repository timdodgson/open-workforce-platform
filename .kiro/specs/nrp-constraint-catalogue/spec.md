# NRP Constraint Catalogue

## Purpose

Document the hard and soft constraints the platform should support for Nurse Rostering Problem datasets.

This is a documentation and design task only.

No optimiser behaviour should change.

---

# Why

The platform now supports a simplified NRP adapter.

Before adding more NRP behaviour or custom algorithms, the project needs a clear constraint catalogue.

This ensures future work aligns with recognised nurse rostering requirements rather than ad-hoc scheduling rules.

---

# Hard Constraints

Document the following hard constraints:

- Single assignment per nurse per day
- Nurse must have required skill
- Required staffing demand should be satisfied where possible
- No overlapping shifts
- Resource availability
- Shift time windows
- Maximum working minutes per day
- Legal shift succession / rest period rules

---

# Soft Constraints

Document the following soft constraints:

- Preferred day on / day off
- Preferred shift on / shift off
- Minimum assigned shifts
- Maximum assigned shifts
- Minimum consecutive working days
- Maximum consecutive working days
- Minimum consecutive days off
- Maximum consecutive days off
- Weekend fairness
- Night shift fairness
- Workload balance
- Plan stability
- Preference satisfaction

---

# Output

Create:

docs/architecture/nrp-constraints.md

The document should explain:

- what each constraint means
- whether it is currently implemented
- whether it should be hard or soft initially
- what platform component should own it
- suggested implementation priority

---

# Constraints

Do not change code.

Do not change datasets.

Do not change algorithms.

This is a design catalogue only.