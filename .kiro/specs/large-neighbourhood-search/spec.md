# Large Neighbourhood Search

## Purpose

Introduce Large Neighbourhood Search as a new optimisation algorithm.

Large Neighbourhood Search repeatedly destroys part of a solution and repairs it, allowing the optimiser to escape local optima more effectively than small local moves.

---

# Why

Hill Climbing and Tabu Search explore relatively small neighbourhood moves.

Large Neighbourhood Search explores larger changes by temporarily removing multiple assignments and rebuilding part of the plan.

This can improve solutions where single moves or swaps are insufficient.

---

# Behaviour

Register a new algorithm:

large-neighbourhood-search

The algorithm should:

1. Start from the constructive solution or an existing plan.
2. Select a small deterministic subset of assignments to remove.
3. Treat removed Work Items as unassigned.
4. Repair the plan using existing placement logic.
5. Score the repaired plan.
6. Keep the repaired plan if it improves the objective score.
7. Repeat for a fixed number of iterations.
8. Return the best plan encountered.

---

# Destroy Strategy

Use a deterministic destroy strategy.

Initial strategy:

- remove up to 2 assignments per iteration
- choose assignments based on iteration index
- do not use randomness

---

# Repair Strategy

Use existing assignment logic where practical.

Repair should respect:

- availability
- skills
- capacity
- duration
- time windows
- travel-aware scheduling

Invalid repaired plans must be rejected.

---

# Architecture

Implement as another Algorithm.

Reuse:

- OptimisationContext
- Objective engine
- Schedule validation
- Candidate move / assignment helpers
- Statistics

Do not introduce a generic LNS framework.

---

# Statistics

Populate:

- iterations
- candidates evaluated
- improvements accepted
- final objective score

Algorithm name:

large-neighbourhood-search

---

# CLI

Support:

--algorithm large-neighbourhood-search

Benchmark runner should automatically include it via the algorithm registry.

---

# Tests

Cover:

- registration
- deterministic behaviour
- destroy removes assignments
- repair produces valid plans
- statistics populated
- benchmark runner succeeds

---

# Non-Goals

Do not implement:

- random destroy
- adaptive destroy size
- multiple destroy strategies
- simulated annealing acceptance
- parallel search

Keep the implementation intentionally simple.

---

# Definition of Done

- large-neighbourhood-search registered
- CLI supports it
- benchmark runner includes it
- statistics populated
- tests pass
- no external dependencies

---

# Open Questions

None.