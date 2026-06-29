# Incremental Optimisation

## Purpose

Introduce warm-start optimisation.

Algorithms should be able to begin from an existing plan rather than always constructing a new one.

This allows the optimiser to repair plans after small changes instead of rebuilding them from scratch.

---

# Why

Real workforce planning is dynamic.

Examples include:

- engineer unavailable
- new urgent job
- cancelled appointment
- extended job duration
- customer reschedule

A complete rebuild may unnecessarily disturb the existing schedule.

Warm-start optimisation improves stability and performance.

---

# Behaviour

OptimisationContext may optionally contain an existing OptimisedPlan.

If present:

- search algorithms should begin from that plan
- constructive remains unchanged

The optimiser should repair and improve the supplied plan.

---

# Architecture

OptimisationContext gains:

ExistingPlan() (optional)

Algorithms choose whether to use it.

Constructive ignores it.

Hill Climbing starts from it.

Simulated Annealing starts from it.

Future algorithms may also use it.

---

# Plan Stability

Introduce a new objective.

Minimise unnecessary changes.

Example:

Existing assignment:

WI-001 -> RES-A

New plan:

WI-001 -> RES-A

Reward.

Changing assignments unnecessarily should incur a small penalty.

This objective must remain much smaller than assignment completion.

---

# Objective Breakdown

Add:

Plan Stability

Example

Assignment: 3000
Travel: -50
Preferred Resource: 25
Plan Stability: 18

---

# Tests

Add tests covering:

- optimisation without existing plan
- optimisation with existing plan
- hill climbing repairs an existing plan
- constructive ignores existing plans
- stability objective included
- deterministic behaviour

---

# Non-Goals

Do not implement:

- partial optimisation
- locking assignments
- frozen resources
- user editing

---

# Constraints

Use only the Go standard library.

Do not modify domain objects.

Keep implementation intentionally simple.

---

# Definition of Done

The implementation is complete when:

- OptimisationContext optionally contains an existing plan
- search algorithms can start from that plan
- constructive behaviour unchanged
- stability objective included
- benchmark runner still succeeds
- tests pass

---

# Open Questions

None.