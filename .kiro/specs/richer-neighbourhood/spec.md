# Richer Neighbourhood

## Purpose

Expand the neighbourhood model so search algorithms can explore a larger solution space.

The current neighbourhood primarily performs placement and simple swaps.

Introduce additional move types that preserve validity while allowing algorithms to escape local optima.

---

# Why

Current benchmark datasets show little separation between algorithms.

A richer neighbourhood gives search algorithms more opportunities to improve plans without changing the objective engine.

This benefits every current and future search algorithm.

---

# New Move Types

Introduce the following candidate move types.

## Relocate

Move one assigned Work Item from one Resource to another.

Unlike placement, no unassigned work is involved.

---

## Two-for-One Exchange

Exchange two Work Items on one Resource with one Work Item on another Resource where constraints permit.

Keep the implementation intentionally simple.

---

## Reorder

Reorder Work Items assigned to the same Resource.

This allows different schedule feasibility outcomes where time windows or travel are involved.

Only change execution order.

Do not change assignments.

---

# Architecture

Continue using the existing neighbourhood model.

CandidateMove should continue representing move proposals.

Do not introduce inheritance or complex move hierarchies.

Move generation remains independent from algorithms.

Algorithms simply request available candidate moves.

---

# Constraints

Every generated move must still respect:

- availability
- skills
- capacity
- duration
- time windows
- travel-aware scheduling

Algorithms should reject invalid moves.

---

# Tests

Add tests covering:

- relocate generation
- reorder generation
- exchange generation
- invalid moves rejected
- deterministic generation order
- algorithms remain deterministic

---

# Non-Goals

Do not implement:

- random neighbourhood generation
- adaptive neighbourhoods
- probabilistic moves
- parallel search

---

# Definition of Done

The implementation is complete when:

- neighbourhood supports relocate
- neighbourhood supports reorder
- neighbourhood supports exchange
- existing algorithms use the richer neighbourhood
- benchmark runner still succeeds
- all tests pass
- no external dependencies are introduced

---

# Open Questions

None.