# Soft Constraints

## Purpose

Introduce soft constraints through objective scoring.

Soft constraints represent preferences.

They should influence plan quality but must never make a plan invalid.

---

# Why

Hard constraints answer:

Can this assignment happen?

Soft constraints answer:

How good is this assignment?

Real workforce optimisation is usually driven by many preferences that should be rewarded or penalised without making a plan impossible.

---

# Behaviour

Soft constraints should affect the objective score only.

They must not affect:

- assignment validity
- capacity checks
- availability checks
- skills checks
- time-window feasibility

If a soft constraint is not satisfied, the plan may still be valid.

---

# Initial Soft Constraint

Introduce a preferred resource objective.

A Work Item may specify a preferred Resource.

If the Work Item is assigned to its preferred Resource, the plan receives a small positive score contribution.

Example Work Item details:

{
  "preferredResource": "RES-001"
}

The contribution should be smaller than the assignment objective.

The optimiser must never leave work unassigned purely to satisfy a preferred resource preference.

---

# Architecture

Preferred resource is business information.

The WorkItem domain object remains unchanged.

The application layer extracts preferredResource from Work Item details.

The optimisation layer receives preferred resource information through WorkItemInput.

Objective scoring evaluates the preference.

Algorithms compare objective scores only.

---

# Objective Breakdown

Add a new objective contribution:

Preferred Resource

Example:

Objective Breakdown:
  Assignment: 3000
  Workload Balance: 1
  Travel Time: -50
  Preferred Resource: 25

The total objective score must equal the sum of objective contributions.

---

# Non-Goals

Do not implement:

- configurable weights
- runtime objective selection
- multiple preferred resources
- preference hierarchies
- hard preferred-resource constraints
- UI changes

---

# Tests

Add tests covering:

- preferred resource contribution is applied when matched
- no contribution is applied when not matched
- missing preferred resource gives no contribution
- assignment objective remains dominant
- objective breakdown includes preferred resource
- soft constraint does not invalidate a plan
- algorithms still work

---

# Constraints

Use only the Go standard library.

Do not modify domain objects.

Do not introduce external dependencies.

Do not introduce a generic soft-constraint framework.

Keep the implementation intentionally simple.

---

# Definition of Done

The implementation is complete when:

- preferred resource is extracted into WorkItemInput
- objective scoring includes preferred resource contribution
- objective breakdown includes preferred resource contribution
- soft constraints do not affect validity
- all tests pass
- CLI command still runs successfully
- no external dependencies are introduced

---

# Open Questions

None.