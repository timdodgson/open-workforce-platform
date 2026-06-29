# Resource Assignment Walking Skeleton

## Purpose

Extend the walking skeleton so the optimiser assigns Work Items to Resources.

The goal is not to produce a real optimisation algorithm yet.

The goal is to prove that the platform can load Resources, pass them to the optimiser, and produce an Optimised Plan that shows which Resource is assigned to which Work Item.

---

# Why

The first walking skeleton proved the pipeline from Business Events to Work Items to Optimised Plan.

The next step is to introduce supply.

Work Items represent demand.

Resources represent supply.

An optimisation platform must eventually match demand to supply.

---

# Target Command

From `platform/go`:

go run ./cmd/owp optimise ../../examples/datasets/simple-events.json

---

# Expected Behaviour

The command should:

1. Load Business Events from the dataset.
2. Load Resources from the dataset.
3. Convert Business Events into Work Items.
4. Pass Work Items and Resources into the optimiser.
5. Produce an Optimised Plan containing assignments.
6. Print a readable summary showing Resources and assigned Work Items.

---

# Dataset

Update:

examples/datasets/simple-events.json

The dataset should contain:

- businessEvents
- resources

Resources should use the existing Resource domain shape:

- id
- type
- details

---

# Assignment

Introduce the smallest useful assignment model.

An assignment connects:

- one Resource
- one Work Item

The initial optimiser may assign all Work Items to the first available Resource.

This is intentionally simple.

The purpose is to prove assignment flow, not optimisation quality.

---

# OptimisedPlan

Update OptimisedPlan so it contains assignments rather than only selected Work Items.

The plan should remain immutable.

Accessors should return defensive copies where needed.

---

# Responsibilities

The infrastructure layer should:

- load Business Events and Resources from JSON
- validate both using domain constructors

The application layer should:

- orchestrate loading inputs into Work Items
- pass Work Items and Resources to the optimiser

The optimisation layer should:

- accept Work Items and Resources
- return an Optimised Plan containing assignments

The CLI should:

- print the resulting assignments clearly

---

# Non-Goals

Do not implement:

- real optimisation
- resource skills
- availability
- constraints
- objectives
- scoring
- routing
- scheduling
- external integrations

---

# Tests

Add or update tests for:

- loading Resources from the dataset
- creating assignments
- optimiser assignment behaviour
- OptimisedPlan immutability
- application workflow
- readable CLI output where practical

---

# Constraints

Use only the Go standard library.

Do not introduce external dependencies.

Do not introduce unnecessary interfaces or abstractions.

If an abstraction is introduced, explain why it has earned its place.

---

# Definition of Done

The walking skeleton is complete when:

- go test ./... passes
- the target command runs successfully
- the command prints Resources with assigned Work Items
- no external dependencies are introduced
- implementation respects the steering documents

---

# Open Questions

None.