# Repository Layout

## Purpose

This document defines the intended repository structure for the Open
Workforce Platform.

The goal is to keep the codebase modular, understandable and
maintainable as the platform grows.

The repository layout should reflect the architecture of the platform
rather than the technology stack.

------------------------------------------------------------------------

# Design Principles

-   The repository should make domain boundaries visible.
-   Documentation remains a first-class part of the project.
-   The optimisation core should not depend on user interface or
    infrastructure code.
-   Dependencies must earn their place.
-   Simple code is preferred over unnecessary packages.
-   The first implementation should be small, but the structure should
    allow the platform to grow.

------------------------------------------------------------------------

# Proposed Structure

``` text
open-workforce-platform/
├── docs/
│   ├── 00-charter/
│   ├── 01-research/
│   ├── 02-architecture/
│   ├── 03-decisions/
│   ├── 04-roadmap/
│   └── 05-design/
│
├── src/
│   └── open_workforce/
│       ├── domain/
│       ├── work_management/
│       ├── resource_management/
│       ├── constraint_management/
│       ├── objective_management/
│       ├── optimisation/
│       └── cli/
│
├── tests/
│   ├── unit/
│   ├── integration/
│   └── fixtures/
│
├── examples/
│   └── datasets/
│
├── scripts/
│
└── README.md
```

------------------------------------------------------------------------

# Top-Level Folders

## docs

Contains project documentation, architecture, decisions, research and
design notes.

Documentation is treated as part of the product.

## src

Contains production source code.

The structure under `src/open_workforce/` should reflect the platform
bounded contexts.

## tests

Contains automated tests.

Tests should validate the domain model, constraints, optimisation
behaviour and command-line workflows.

## examples

Contains small example datasets and scenarios used to demonstrate the
platform.

The initial MVP dataset will live here.

## scripts

Contains helper scripts for local development and maintenance.

Scripts should remain small and should not become hidden application
logic.

------------------------------------------------------------------------

# Source Layout

## domain

Contains shared domain concepts that are independent of any bounded
context implementation.

Examples:

-   Business Event
-   Work Item
-   Resource
-   Constraint
-   Objective

## work_management

Owns the lifecycle of work.

## resource_management

Owns resources, availability, capacity and capabilities.

## constraint_management

Owns constraint definitions and configuration.

## objective_management

Owns optimisation objectives and their relative importance.

## optimisation

Owns solution generation, solution evaluation and scenario comparison.

## cli

Provides the command-line entry point for the MVP.

The CLI should orchestrate the workflow without owning business logic.

------------------------------------------------------------------------

# Dependency Principle

Dependencies must earn their place.

Every external package should provide clear value and justify its
maintenance cost.

Packages should be used where they reduce risk, provide complex
functionality, or represent well-maintained ecosystem standards.

Packages should be avoided when they replace simple, understandable
code.

A dependency is treated as a long-term maintenance commitment, not a
convenience.
