# Architecture Steering

## Purpose

This document summarises the architectural boundaries of the Open Workforce Platform.

All implementation should respect these boundaries.

If a proposed implementation conflicts with the architecture, raise the concern rather than working around it.

---

# Core Domain

The platform is built around five core concepts:

- Business Event
- Work Item
- Resource
- Constraint
- Objective

The optimiser combines these concepts to produce an Optimised Plan.

---

# Domain Flow

Business Event

↓

Work Item

↓

Optimisation

↓

Optimised Plan

Resources, Constraints and Objectives influence optimisation but do not own it.

---

# Responsibilities

Business Events

- represent something that happened
- create or modify Work Items
- never perform optimisation

Work Items

- represent work to be completed
- contain business information
- are independent of optimisation algorithms

Resources

- represent people, vehicles or assets capable of completing work

Constraints

- define what is allowed
- never decide what is optimal

Objectives

- describe what "better" means
- influence optimisation
- never contain business rules

Optimisation

- consumes Work Items, Resources, Constraints and Objectives
- produces one or more Optimised Plans
- does not own business knowledge

---

# Architectural Boundaries

Business knowledge belongs in the domain.

Optimisation consumes business knowledge.

Infrastructure supports the domain.

The CLI orchestrates behaviour but does not contain business logic.

---

# Package Responsibilities

Keep packages focused on a single responsibility.

Avoid packages that become general utilities or catch-all locations.

If functionality does not clearly belong in an existing package, reconsider the design before creating a new package.

---

# Architectural Review

Before introducing new code, confirm:

- Does this belong in this package?
- Does it introduce coupling between bounded contexts?
- Is business knowledge being placed in the correct location?
- Does this strengthen or weaken the architecture?