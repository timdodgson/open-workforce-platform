# NRP Adapter

## Purpose

Introduce an adapter for Nurse Rostering Problem datasets.

The adapter should convert NRP-style input data into the Open Workforce Platform dataset format so existing algorithms can benchmark against more realistic workforce scheduling problems.

---

# Why

The platform currently uses small synthetic benchmark datasets.

NRP datasets provide a more realistic and recognised scheduling problem shape.

Adding an NRP adapter allows the existing optimiser, algorithms, benchmark runner and statistics to be tested against larger and more meaningful workforce optimisation scenarios.

---

# Scope

Create an adapter that reads an NRP dataset format and produces an Open Workforce Platform dataset.

The adapter should not change optimisation algorithms.

The adapter should not change domain objects.

The adapter should translate input data into:

- Resources
- Business Events
- Work Items
- Travel matrix where applicable

---

# Initial Format

Start with a simple JSON-based NRP input format.

Do not implement XML parsing yet.

The initial format should model:

- nurses
- days
- shifts
- demand
- skills
- preferences where practical

This allows the adapter and optimisation mapping to be developed before supporting external benchmark formats.

---

# Mapping

Map nurses to Resources.

Map required shifts to Business Events.

Map each required shift into a Work Item.

Use:

- nurse ID as Resource ID
- shift demand as generated Work Items
- day and shift as time-window information
- nurse skills as Resource skills
- required skill as Work Item requiredSkill

---

# Scheduling Assumptions

For the first implementation:

- one day may be represented as minutes from midnight
- each shift has a start and end time
- no travel time is required
- no multi-day carry-over is required
- each nurse has one shift capacity per day where applicable

Keep the mapping intentionally simple.

---

# CLI

Add a conversion command:

go run ./cmd/owp convert-nrp <input-file> <output-file>

The output should be a normal Open Workforce Platform dataset JSON file.

The generated dataset should be runnable with:

go run ./cmd/owp optimise <output-file> --algorithm tabu-search

---

# Example Data

Create a small example NRP input file:

examples/nrp/simple-nrp.json

Create the converted output as part of tests or examples where appropriate.

---

# Tests

Add tests covering:

- NRP file can be loaded
- nurses are converted to Resources
- shift demand is converted to Business Events / Work Items
- skills are preserved
- generated dataset can be optimised
- generated dataset is deterministic

---

# Non-Goals

Do not implement:

- official XML NRP parsing
- INRC full format
- multi-week planning
- fairness constraints
- maximum consecutive shifts
- rest-period rules
- contract constraints

These will come later.

---

# Architecture

The NRP adapter belongs in infrastructure or a dedicated adapter package.

It should produce the existing dataset structure consumed by the current loader.

The optimiser should remain unaware of NRP.

---

# Constraints

Use only the Go standard library.

Do not introduce external dependencies.

Do not change optimisation algorithms.

Do not change domain objects.

Keep the first adapter intentionally simple.

---

# Definition of Done

The implementation is complete when:

- simple NRP JSON can be converted to OWP dataset JSON
- generated dataset can be optimised by existing algorithms
- tests pass
- CLI conversion command works
- no external dependencies are introduced

---

# Open Questions

None.