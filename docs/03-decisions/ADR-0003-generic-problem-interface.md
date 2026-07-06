# ADR-0003

## Title

Generic Problem interface decouples domains from algorithms

## Status

Accepted

## Context

The platform supports multiple optimisation domains (NRP, CVRP, JSS, VRPTW). Each domain has different solution representations, constraint types, and neighbourhood operators. Without abstraction, every algorithm would need domain-specific code paths.

## Decision

Define a `Problem` interface with 7 methods. All algorithms depend only on this interface. Domains implement it independently. `Solution` and `Move` are opaque types — algorithms never inspect their contents.

## Alternatives

- **Domain-specific algorithms.** Each domain gets its own SA/LAHC/Tabu. Duplicates algorithm logic across domains.
- **Shared solution struct with tagged fields.** Couples domains to a common representation. Breaks when domain complexity differs.
- **Plugin architecture with dynamic loading.** Over-engineered for the current scale.

## Consequences

- Adding a new domain requires zero changes to the search engine.
- Algorithms are tested once and work everywhere.
- Each domain owns its representation — no compromise on performance.
- The interface is minimal (7 methods) which keeps implementation effort low.
- Domain-specific optimisations (e.g. incremental evaluation) are still possible within the interface contract.
