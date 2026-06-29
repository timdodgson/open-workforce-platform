# Swap Neighbourhood

## Purpose

Extend neighbourhood generation to support swap candidate moves.

A swap candidate exchanges the Resource assignments of two Work Items.

This gives search-based algorithms a richer neighbourhood to explore.

---

# Why

The current neighbourhood mostly helps place unassigned Work Items.

With only placement moves, hill climbing and simulated annealing often behave similarly.

Swap moves allow algorithms to explore alternative valid assignments even when all Work Items are already assigned.

This is an important step toward meaningful local search.

---

# Behaviour

The neighbourhood should support:

- existing placement candidate moves
- new swap candidate moves

A swap candidate should exchange two assigned Work Items between their assigned Resources.

A swap is valid only if both resulting assignments respect:

- capacity
- availability
- skills

---

# Candidate Move

Extend the existing CandidateMove model only if necessary.

Do not introduce a large move hierarchy.

Keep the representation simple.

The algorithm should remain responsible for deciding whether to accept a candidate move.

---

# Algorithms

Hill climbing and simulated annealing should be able to use the richer neighbourhood.

Do not change the algorithm contract.

Do not introduce new algorithms.

---

# Scoring

Use the existing score calculation.

Do not introduce a new scoring model in this iteration.

If a swap does not improve the current score, hill climbing should reject it.

Simulated annealing may accept non-improving valid swaps during its hot phase according to its existing deterministic acceptance rule.

---

# Non-Goals

Do not implement:

- route moves
- multi-work-item moves
- randomisation
- tabu search
- genetic operators
- weighted scoring
- generic move framework

---

# Tests

Add tests covering:

- valid swap generation
- invalid swap rejected when skills would be violated
- invalid swap rejected when availability would be violated
- algorithms remain deterministic
- hill climbing rejects non-improving swaps
- simulated annealing can accept non-improving swaps during hot phase
- existing placement behaviour remains unchanged

---

# Constraints

Use only the Go standard library.

Do not introduce external dependencies.

Do not introduce a generic move framework.

Keep the implementation intentionally simple.

---

# Definition of Done

The implementation is complete when:

- swap candidate moves can be generated
- invalid swaps are rejected
- existing placement moves still work
- hill climbing and simulated annealing use the richer neighbourhood
- all tests pass
- CLI commands still work
- no external dependencies are introduced

---

# Open Questions

None.