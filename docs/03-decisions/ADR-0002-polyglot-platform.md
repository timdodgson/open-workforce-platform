# ADR-0002

## Title

Go is the platform core and owns the optimisation pipeline. Python may be introduced later for specialist solver experiments where it earns its place

---

## Status

Accepted

---

## Context

The platform requires a primary language for the core runtime — APIs, event processing, domain logic, and infrastructure.

It also requires tooling for solver experimentation, where the operational research ecosystem is strongest in Python.

A decision is needed on which languages are adopted, and equally important, which are explicitly deferred.

---

## Decision

The platform core will be implemented in Go.

Python will be used for optimisation experiments and solver integration.

Node will only be introduced later if a web application or tooling layer requires it.

---

## Consequences

### Positive

- Go is fast, strongly typed, simple to deploy, and has excellent standard library support.
- Go keeps dependency pressure low.
- Python provides access to the strongest optimisation ecosystem (OR-Tools, SciPy, NumPy, etc.).
- Deferring Node removes unnecessary complexity until there is a clear product need.
- Clear language boundaries reinforce bounded context separation.

### Trade-offs

- Two languages means two build pipelines and two sets of tooling.
- Integration between Go and Python requires a well-defined interface (likely gRPC or event-based).
- Contributors need familiarity with at least one of the two languages.

---

## Alternatives Considered

### Python Only

Use Python for everything — core platform and optimisation.

Rejected because Python is not well suited to high-performance, strongly typed platform services at scale.

### Node / TypeScript for Platform Core

Use TypeScript for the platform runtime.

Rejected because Go offers better performance characteristics, simpler deployment, and lower dependency overhead for the type of services this platform requires.

### Introduce All Three Languages Immediately

Start with Go, Python, and Node from day one.

Rejected because introducing a language without a clear product need violates Principle 1 (Everything Must Earn Its Place).

---

## Rationale

Go fits the platform core because it is fast, strongly typed, simple to deploy, has excellent standard library support, and keeps dependency pressure low.

Python remains appropriate for solver experimentation because the optimisation ecosystem is strongest there.

Node is deferred until there is a clear product need.
