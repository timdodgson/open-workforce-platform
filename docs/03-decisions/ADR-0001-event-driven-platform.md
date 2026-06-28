# ADR-0001

## Title

Adopt an Event-Driven Optimisation Platform Architecture

---

## Status

Accepted

---

## Context

The Open Workforce Platform is intended to support workforce optimisation across multiple industries rather than solve a single scheduling problem.

The platform must remain independent of any specific optimisation engine, programming language, deployment model or industry.

Business events naturally represent the origin of work within most operational environments.

Examples include:

- Customer orders
- Equipment failures
- Planned maintenance
- Patient referrals
- Emergency incidents

These events create demand that must be transformed into an optimised plan.

---

## Decision

The platform will adopt an event-driven architecture.

Business Events create Work Items.

Bounded contexts collaborate through business events rather than direct knowledge of one another.

The Optimisation Context consumes business information and produces an optimised plan.

Execution remains outside the responsibility of the platform.

---

## Consequences

### Positive

- Low coupling between bounded contexts.
- Clear separation of responsibilities.
- Easier integration with external systems.
- Supports asynchronous processing.
- Supports future AI-driven capabilities.
- Suitable for multiple industries.
- Scales naturally as additional capabilities are added.

### Trade-offs

- More architectural concepts than a tightly coupled application.
- Requires careful event modelling.
- Eventual consistency must be considered where appropriate.

---

## Alternatives Considered

### Direct Service Calls

Each bounded context directly invokes the next.

Rejected because this increases coupling and makes future evolution more difficult.

### Monolithic Scheduling Engine

One component owns work, resources, constraints and optimisation.

Rejected because it combines multiple business responsibilities into a single context and reduces extensibility.

---

## Rationale

Open Workforce Platform is designed as a platform rather than a single application.

The event-driven architecture supports this vision by allowing business capabilities to evolve independently while collaborating through well-defined business events.