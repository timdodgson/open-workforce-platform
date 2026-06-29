# Simulated Annealing Optimiser

## Purpose

Introduce simulated annealing as the platform's next search-based optimisation algorithm.

Simulated annealing should use the existing optimisation inputs, scoring model and neighbourhood generation.

This proves that the platform can host algorithms that explore beyond strictly improving moves.

---

# Why

Hill climbing only accepts improving candidate moves.

This can cause the optimiser to get stuck in local optima.

Simulated annealing allows some worse moves early in the search, reducing the chance of getting trapped too soon.

This is the next natural step after hill climbing.

---

# Behaviour

The simulated annealing optimiser should:

- start from the constructive solution
- use the existing neighbourhood generation
- evaluate candidate moves using the existing score
- accept improving moves
- sometimes accept worse moves according to temperature
- reduce temperature over time
- remain deterministic for now

---

# Determinism

This implementation must remain deterministic.

Do not introduce randomness yet.

Use a deterministic acceptance rule for worse moves based on iteration and temperature.

The goal is to prove the architecture, not produce a production-grade annealing algorithm.

---

# CLI

Add support for:

simulated-annealing

Example:

go run ./cmd/owp optimise ../../examples/datasets/simple-events.json --algorithm simulated-annealing

---

# Non-Goals

Do not implement:

- random number generation
- advanced cooling schedules
- tuning parameters
- external dependencies
- generic algorithm framework
- weighted objectives
- parallel search

---

# Tests

Add tests covering:

- simulated annealing can be selected
- starts from constructive solution
- uses neighbourhood generation
- respects capacity
- respects availability
- respects skills
- returns a valid plan
- deterministic behaviour
- unknown algorithms still fail clearly

---

# Constraints

Use only the Go standard library.

Do not change domain objects.

Do not introduce external dependencies.

Do not introduce randomness in this iteration.

Keep the implementation intentionally simple.

---

# Definition of Done

The implementation is complete when:

- simulated-annealing is available as an algorithm option
- existing algorithms still work
- all tests pass
- CLI command runs successfully
- no external dependencies are introduced
- architecture remains consistent with steering documents

---

# Open Questions

None.