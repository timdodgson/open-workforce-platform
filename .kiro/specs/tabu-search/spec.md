# Tabu Search

## Purpose

Introduce Tabu Search as a new optimisation algorithm.

Tabu Search extends local search by remembering recently visited moves, preventing immediate reversal and helping escape local optima.

---

# Why

Hill Climbing stops when no improving neighbour exists.

Tabu Search continues exploring while avoiding cycles.

This often finds better solutions without introducing randomness.

---

# Behaviour

Register a new algorithm:

tabu-search

The algorithm should:

1. Start from the constructive solution or an existing plan.
2. Generate neighbourhood moves.
3. Reject moves currently in the tabu list.
4. Select the highest-scoring admissible move.
5. Add the chosen move to the tabu list.
6. Remove the oldest move when the tabu list reaches its maximum size.
7. Continue until:
   - iteration limit reached
   - or no admissible moves remain.

Return the best plan encountered.

---

# Tabu List

Implement a simple FIFO tabu list.

Store:

- Work Item ID
- Source Resource
- Destination Resource

A fixed size of 10 entries is sufficient.

No aspiration criteria are required.

---

# Architecture

Implement as another Algorithm.

Reuse:

- OptimisationContext
- Objective engine
- CandidateMove generation
- Schedule validation
- Statistics

Algorithms remain independent.

---

# Statistics

Populate:

- iterations
- candidates evaluated
- improvements accepted
- best objective score

Algorithm name:

tabu-search

---

# CLI

Support:

--algorithm tabu-search

Benchmark runner should automatically include it via the algorithm registry.

---

# Tests

Cover:

- registration
- deterministic behaviour
- tabu list prevents immediate reversal
- statistics populated
- benchmark runner succeeds

---

# Non-Goals

Do not implement:

- aspiration
- adaptive tabu tenure
- strategic oscillation
- intensification
- diversification

Keep the implementation intentionally simple.

---

# Definition of Done

- tabu-search registered
- CLI supports it
- benchmark runner includes it
- statistics populated
- tests pass
- no external dependencies