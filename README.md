# Open Workforce Platform

> **Building the future of workforce optimisation through open engineering, modern optimisation algorithms and AI-assisted software engineering.**

---

![Status](https://img.shields.io/badge/status-Active%20Development-brightgreen)

![Optimisation](https://img.shields.io/badge/Optimisation-Platform-blue)

![AI Assisted](https://img.shields.io/badge/AI-Assisted%20Engineering-purple)

![License](https://img.shields.io/badge/license-MIT-green)

---

## Overview

Open Workforce Platform is an open-source workforce optimisation platform built to explore how modern optimisation techniques, cloud-native architecture and artificial intelligence can be combined to solve complex operational planning problems.

Rather than focusing solely on optimisation algorithms, this project demonstrates how production-quality optimisation software can be engineered using clean architecture, Domain-Driven Design, explainable optimisation and AI-assisted software engineering.

The long-term vision is a platform capable of solving workforce scheduling, field service optimisation, routing and operational planning problems across industries including healthcare, utilities, logistics, manufacturing and field services.

---

## Why this project exists

Ten years ago this project began as a university dissertation exploring nurse rostering and patient allocation.

Today it has become something much larger.

Open Workforce Platform is an opportunity to revisit those ideas using modern optimisation techniques, artificial intelligence and over two decades of professional software engineering experience.

The objective is no longer simply to build an optimiser.

The objective is to demonstrate what exceptional software engineering looks like in the age of AI.

---

## Engineering Philosophy

Artificial Intelligence is not replacing software engineers.

Artificial Intelligence is changing software engineering.

This project is built on the belief that experienced engineers who understand architecture, design and engineering judgement will produce significantly better software by collaborating with AI.

AI is treated as a first-class engineering tool throughout this project.

It assists with:

- Research
- Architecture
- Documentation
- Design
- Implementation
- Testing
- Code Review

Final engineering decisions always remain the responsibility of the project maintainers.

> **AI lowers the barrier to software creation. Engineering experience determines the quality of what is created.**

---

## Vision

Create the world's leading open-source workforce optimisation platform while documenting every architectural decision, engineering trade-off and lesson learned along the way.

The project is intentionally developed in public so that others can follow both the engineering journey and the evolution of the optimisation platform.

---

## Current Status

The platform is under active development.

A complete optimisation pipeline now exists from business data through to an optimised plan.

Current focus is expanding optimisation capabilities while keeping the architecture clean, testable and explainable.

---

## Current Capabilities

### Domain Model

- Business Events
- Resources
- Work Items
- Assignments
- Optimised Plans

### Optimisation Engine

- Constructive algorithm
- Hill Climbing
- Simulated Annealing
- Shared neighbourhood model
- Explainable objective engine

### Constraints

- Resource availability
- Skills matching
- Priority ordering
- Duration-based capacity
- Sequential scheduling
- Time windows
- Travel-aware scheduling

### Objectives

- Maximise completed work
- Balance workload
- Minimise travel time

### Explainability

- Objective score
- Objective breakdown
- Travel explanation
- Resource utilisation
- Assignment reporting

---

## Current Architecture

```text
Business Events
        │
        ▼
Application Layer
        │
        ▼
Optimisation Input Builder
        │
        ▼
Optimisation Context
        │
        ▼
Algorithms
 ├─ Constructive
 ├─ Hill Climbing
 └─ Simulated Annealing
        │
        ▼
Neighbourhood
        │
        ▼
Constraint Engine
        │
        ▼
Objective Engine
        │
        ▼
Optimised Plan
        │
        ▼
CLI Output
```

The optimisation engine deliberately operates on optimisation inputs rather than business domain objects.

This separation allows new algorithms and constraints to be introduced without leaking business concerns into optimisation code.

---

## Running

From the repository root:

```bash
cd platform/go

go run ./cmd/owp optimise ../../examples/datasets/simple-events.json
```

Run a specific optimisation algorithm:

```bash
go run ./cmd/owp optimise ../../examples/datasets/simple-events.json --algorithm hill-climbing

go run ./cmd/owp optimise ../../examples/datasets/simple-events.json --algorithm simulated-annealing
```

Run the test suite:

```bash
go test ./...
```

---

## Example Output

```text
=== Optimised Plan ===

Algorithm: hill-climbing
Assignment Score: 100
Objective Score:  2951

Objective Breakdown:
  Assignment: 3000
  Workload Balance: 1
  Travel Time: -50

Resources: 2
Capacity:  960

Assignments:

  RES-002
    Used: 120 / 480 mins
    Work Items:
      - WI-EVT-002

  RES-001
    Used: 135 / 480 mins
    Work Items:
      - WI-EVT-001
      - WI-EVT-003

Unassigned: None

Travel:

  RES-002
    BASE-SOUTH -> LOC-B: 25 mins
    Total: 25 mins

  RES-001
    BASE-NORTH -> LOC-A: 10 mins
    LOC-A -> LOC-C: 15 mins
    Total: 25 mins

Done.
```

---

## Engineering Principles

The platform is built around several key principles.

- Domain Driven Design
- Clean Architecture
- Engineering before implementation
- Explain every important decision
- Optimisation separated from business rules
- Small, incremental architectural evolution
- Features earn their place
- Test-driven development
- AI-assisted engineering
- Explainable optimisation

---

## Planned Capabilities

The architecture is intentionally designed to support future optimisation techniques without significant architectural change.

Planned work includes:

- Tabu Search
- Large Neighbourhood Search
- Genetic Algorithms
- OR-Tools / CP-SAT integration
- Vehicle routing
- Multi-day planning
- Calendar integration
- Overtime and break constraints
- Benchmark datasets
- Visual optimisation explorer

---

## Documentation

| Document | Purpose |
|----------|---------|
| Project Charter | Project vision and goals |
| Roadmap | Long-term direction |
| Domain Model | Core business concepts |
| Architecture | Platform architecture |
| ADRs | Architectural Decision Records |
| Research | Optimisation research and benchmarking |
| Kiro Specs | Incremental feature specifications |

---

## Engineering Journey

Every architectural decision in this repository is introduced incrementally.

Features are added only when they have earned their place within the architecture.

The repository intentionally documents:

- architectural evolution
- engineering trade-offs
- optimisation research
- AI collaboration
- implementation decisions
- lessons learned

The goal is to demonstrate not only what was built, but why it was built that way.

---

## Contributing

Feedback, discussion, ideas and constructive challenges are always welcome.

The project values thoughtful engineering discussion as highly as code contributions.

---

## About

Open Workforce Platform is engineered by **Tim Dodgson** in collaboration with Artificial Intelligence.

The project combines over two decades of software engineering experience with modern optimisation techniques and AI-assisted development to explore what the next generation of engineering can look like.

The software is important.

The engineering journey is equally important.
