# Hill Climbing Optimiser

## Purpose

Introduce the platform's first search-based optimisation algorithm.

Hill climbing should improve on the existing constructive optimiser by exploring neighbouring assignment plans and keeping better solutions.

This proves that the platform can support multiple optimisation strategies.

---

# Why

The current optimiser builds a valid plan using a simple constructive algorithm.

That is useful, but it only makes one pass through the Work Items and Resources.

A search-based optimiser can start from an initial solution and attempt to improve it.

Hill climbing is the simplest useful local search algorithm and is a suitable first step before more advanced algorithms such as simulated annealing, genetic algorithms or CP-SAT.

---

# Behaviour

The hill climbing optimiser should:

- start from the existing constructive solution
- generate neighbouring solutions by moving assigned Work Items between Resources
- reject moves that violate capacity, availability or skills
- score each candidate plan
- keep a candidate only if it improves the score
- stop when no improving neighbour can be found
- remain deterministic

---

# Inputs

The hill climbing optimiser should use the same optimisation inputs as the existing optimiser:

- Work Items
- Resources
- Resource capacity
- Resource availability
- Resource skills
- Work Item priority
- Work Item required skill

Do not change the domain model.

---

# Scoring

Use the existing score calculation.

Do not introduce a new scoring model in this iteration.

A candidate plan is considered better if its score is higher.

If scores are equal, keep the existing plan to preserve determinism.

---

# Algorithm

The initial implementation should be intentionally simple.

Suggested approach:

1. Generate an initial plan using the existing constructive optimiser.
2. Generate candidate neighbours by attempting to move one Work Item from its current Resource to another valid Resource.
3. Score each candidate.
4. If a candidate improves the score, accept it.
5. Repeat until no improving candidate exists.

Do not introduce randomness.

Do not introduce tuning parameters unless clearly justified.

---

# CLI

Add optional algorithm selection to the existing command.

Supported values:

- constructive
- hill-climbing

Default:

- constructive

Examples:

go run ./cmd/owp optimise ../../examples/datasets/simple-events.json

go run ./cmd/owp optimise ../../examples/datasets/simple-events.json --algorithm hill-climbing

---

# Output

The CLI should continue to print the Optimised Plan.

It should also print which algorithm was used.

Example:

Algorithm: hill-climbing

---

# Non-Goals

Do not implement:

- simulated annealing
- genetic algorithms
- CP-SAT
- random search
- weighted objective models
- generic solver framework
- external dependencies

---

# Tests

Add tests covering:

- hill climbing starts from the constructive solution
- invalid moves are rejected
- improving moves are accepted
- equal-score moves are rejected
- deterministic behaviour
- CLI algorithm selection
- default algorithm remains constructive

---

# Constraints

Use only the Go standard library.

Do not introduce external dependencies.

Do not change existing domain objects.

Do not introduce a generic optimisation framework.

Keep the implementation intentionally simple.

---

# Definition of Done

The implementation is complete when:

- constructive remains the default optimiser
- hill-climbing can be selected from the CLI
- hill-climbing respects capacity, availability and skills
- hill-climbing improves the plan when a better neighbour exists
- all tests pass
- the command-line example runs successfully
- no external dependencies are introduced
- architecture remains consistent with steering documents

---

# Open Questions

None.