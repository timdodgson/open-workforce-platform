# Objective Scoring

## Purpose

Separate optimisation objectives from optimisation constraints.

The optimiser should first determine whether a candidate plan is valid.

Only then should it evaluate how good that valid plan is.

This prepares the platform for richer optimisation algorithms and multiple business objectives.

---

# Why

Today the optimisation score is primarily based on the number of assigned Work Items.

That is a useful feasibility metric but it is not a true optimisation objective.

Future optimisation should be able to express preferences such as:

- balanced workloads
- reduced travel
- preferred engineers
- continuity of care
- SLA compliance

without changing constraint logic.

---

# Behaviour

Validation remains unchanged.

Capacity, availability and skills continue to determine whether a candidate assignment is valid.

Objective scoring determines which valid plan is preferred.

---

# Initial Objectives

Introduce a simple additive scoring model.

The overall score should be the sum of objective contributions.

Initial objectives:

- assigned work items
- workload balance

---

# Assigned Work

Continue rewarding assignment of Work Items.

This remains the dominant objective.

---

# Workload Balance

Introduce a small reward for evenly distributing work across available Resources.

Example:

2 + 2 assignments

should score higher than

4 + 0 assignments

when both plans assign the same number of Work Items.

The balance objective should be deliberately weaker than assignment count.

Never sacrifice assigning work purely to improve balance.

---

# Architecture

Introduce an Objective component.

Objectives should evaluate a complete candidate plan.

Algorithms should not know how individual objectives are calculated.

Algorithms simply compare total scores.

---

# Non-Goals

Do not implement:

- weighted configuration
- runtime objective selection
- external configuration
- generic plugin frameworks
- external dependencies

---

# Tests

Add tests covering:

- assignment objective
- workload balance objective
- assignment remains dominant
- deterministic scoring
- algorithms still return valid plans

---

# Constraints

Use only the Go standard library.

Do not introduce external dependencies.

Keep the objective model intentionally simple.

---

# Definition of Done

The implementation is complete when:

- objectives are evaluated independently of constraints
- assignment remains the dominant objective
- workload balance influences scoring
- all algorithms continue to work
- tests pass
- CLI output remains unchanged except for improved scoring behaviour

---

# Open Questions

None.