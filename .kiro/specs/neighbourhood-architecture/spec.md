# Neighbourhood Architecture

## Purpose

Extract neighbourhood generation from optimisation algorithms.

Neighbourhood generation defines how candidate optimisation plans are produced.

Multiple optimisation algorithms should be able to reuse the same neighbourhood generation logic.

---

# Why

The hill climbing optimiser currently generates neighbouring solutions internally.

Future algorithms such as:

- simulated annealing
- tabu search
- genetic algorithms

will require the same capability.

The neighbourhood has therefore earned its place as a reusable optimisation component.

---

# Scope

Refactor only.

Do not change optimisation behaviour.

Do not introduce new optimisation algorithms.

---

# Behaviour

The constructive algorithm remains unchanged.

The hill climbing algorithm should delegate neighbour generation to the neighbourhood component.

The generated neighbours should be identical to the current implementation.

---

# Neighbourhood

The neighbourhood should be responsible only for generating valid candidate moves.

It should not:

- score candidates
- choose candidates
- accept candidates

Those remain algorithm responsibilities.

---

# Candidate Move

The initial neighbourhood supports:

- moving a single Work Item from one Resource to another

No other move types are introduced.

---

# Future Direction

The neighbourhood should be capable of supporting future move types such as:

- swapping two Work Items
- moving multiple Work Items
- balancing workload
- route improvements

These are future capabilities only.

---

# Tests

Existing behaviour must remain unchanged.

Neighbour generation should be independently testable.

---

# Constraints

Use only the Go standard library.

Do not introduce:

- generic optimisation frameworks
- plugin systems
- reflection
- external dependencies

Keep the abstraction focused.

---

# Definition of Done

- behaviour unchanged
- hill climbing delegates neighbour generation
- constructive unchanged
- tests pass
- CLI unchanged

---

# Open Questions

None.