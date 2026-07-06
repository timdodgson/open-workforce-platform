# Principles

Non-negotiable design rules for the PFRS Research Lab platform. These guardrails keep the platform coherent as it grows from four domains to ten or more.

---

## 1. Every optimisation problem implements the same Problem interface

No exceptions. Seven methods. Opaque Solution and Move types. If a domain cannot express itself through this interface, the interface is reconsidered — not bypassed.

This is the foundation everything else depends on.

---

## 2. Search algorithms must remain domain-agnostic

SA, LAHC, Tabu, Portfolio, and Adaptive know nothing about nurses, vehicles, machines, or time windows. They call `TryMove`, `Evaluate`, and `UndoMove`. That's it.

A new algorithm benefits every domain. A new domain benefits from every algorithm. This multiplication is the platform's core value proposition.

---

## 3. Telemetry is a first-class feature, not an afterthought

Every search run produces discovery events. Every discovery has a timestamp, candidate count, and objective improvement. This data powers convergence analysis, statistical comparison, and the Benchmark Ladder.

If a feature cannot produce telemetry, it is incomplete.

---

## 4. New dashboard pages should work across domains where possible

The Summary page, Search Progress, Statistics, and Benchmark Ladder work for NRP, CVRP, JSS, and VRPTW without domain-specific code. Domain-specific visualisations (Gantt, Route Viewer, Schedule) are additions, not replacements.

When adding a page, ask: "Does this work for all domains?" If yes, build it generically. If no, build it as a domain-specific opt-in.

---

## 5. Exact solvers are benchmarks, not replacements for heuristics

ILP provides reference bounds. It answers "how close are we to optimal?" — not "how do we solve this at scale?" Heuristics are the production solvers. ILP is the calibration tool.

Do not invest in scaling ILP formulations. Invest in better heuristics and better measurement of heuristic quality.

---

## 6. Every architectural decision should make adding the next domain easier than the previous one

CVRP was harder to add than JSS. JSS was harder to add than VRPTW. VRPTW reused CVRP's neighbourhood operators. Each domain addition should validate and strengthen the architecture.

If a new domain requires changes to existing algorithms, storage, or telemetry — that's a design smell. Fix the architecture, don't work around it.

---

## 7. The contract between producer and consumer is the schema

The Go CLI produces `run.json`. The dashboard consumes it. The `bestObjective` field is the universal truth. If both sides agree on the schema, they can evolve independently.

Document the contract. Test the contract. Don't couple the implementations.

---

## 8. Reproducibility is non-negotiable

Every experiment has a seed. Same seed, same result. The benchmark suite documents exact commands. Anyone can reproduce any result on any machine.

If a result cannot be reproduced, it is not a result.

---

## 9. Measure before optimising

The platform exists to measure algorithm performance rigorously. Statistical significance, effect sizes, gap-to-optimal percentages. Gut feelings are hypotheses — the platform proves or disproves them.

If you think Algorithm A is better than Algorithm B, run the experiment. Report the p-value.

---

## 10. Quality over quantity

Four domains done well is worth more than eight domains done poorly. Each domain should have:

- Complete Problem interface implementation.
- Full test coverage.
- Benchmark results with known-optimal references.
- Dashboard support (summary + search progress at minimum).
- Documentation in DOMAINS.md.

If a domain cannot meet this bar, it is not ready for the platform.

---

## How to Use This Document

When making a design decision, check it against these principles. If a proposed change violates one, either:

1. Reconsider the change.
2. Or propose an amendment to the principle with justification.

Principles are not immutable. But changing one requires the same rigour as an Architecture Decision Record.
