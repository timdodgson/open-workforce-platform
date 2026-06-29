# Open Workforce Platform

> **Building the future of workforce optimisation through open engineering, modern algorithms and AI-assisted software engineering.**

---

![Status](https://img.shields.io/badge/status-Early%20Development-brightgreen)

![AI Assisted](https://img.shields.io/badge/AI-Assisted%20Engineering-purple)

![License](https://img.shields.io/badge/license-MIT-green)

---

## Overview

Open Workforce Platform is an open-source engineering platform for solving complex workforce optimisation problems.

The platform is designed to optimise workforce scheduling, task allocation, routing and operational planning across multiple industries including healthcare, utilities, logistics, manufacturing and field services.

Rather than focusing solely on optimisation algorithms, the project demonstrates how modern software engineering, artificial intelligence and cloud-native architecture can be combined to build production-quality optimisation software.

---

## Why this project exists

Ten years ago this project started life as a university dissertation exploring nurse rostering and patient allocation.

Today it has become something much bigger.

Open Workforce Platform is an opportunity to revisit that research using modern optimisation techniques, artificial intelligence and over two decades of professional software engineering experience.

The goal is not simply to rebuild the dissertation.

The goal is to demonstrate what exceptional software engineering looks like in the age of AI.

---

## Engineering Philosophy

Artificial Intelligence is not replacing software engineers.

Artificial Intelligence is changing software engineering.

This project is built on the belief that the future belongs to engineers who know how to collaborate with AI while applying experience, judgement and sound engineering principles.

AI is treated as an engineering collaborator throughout this project.

AI assists with:

- Research
- Architecture
- Documentation
- Design
- Implementation
- Testing
- Engineering review

Final engineering decisions always remain the responsibility of the project maintainers.

> **AI lowers the barrier to software creation. Engineering experience determines the quality of what is created.**

---

## Vision

Create the world's leading open-source workforce optimisation platform while documenting every architectural decision, engineering trade-off and lesson learned.

This repository is intentionally developed in the open to demonstrate modern AI-assisted software engineering.

---

## Current Status

The project has completed its initial architectural foundation and now has a fully runnable optimisation pipeline.

Current capabilities include:

- Business Event domain model
- Work Item domain model
- Optimised Plan domain model
- JSON dataset loading
- Application orchestration layer
- Optimisation pipeline
- Command-line interface
- End-to-end execution
- Comprehensive unit tests

The optimiser is intentionally simple at this stage.

Future iterations will introduce Resources, Constraints, Objectives and advanced optimisation algorithms while preserving the existing architecture.

---

## Running the Project

### Prerequisites

Install Go 1.24 or later.

#### Windows

```powershell
winget install GoLang.Go
```

If `go` is not recognised after installation, restart your terminal or IDE. If necessary, sign out of Windows and back in to refresh your system `PATH`.

Verify:

```powershell
go version
```

#### macOS

```bash
brew install go
```

Verify:

```bash
go version
```

---

### Run the Optimisation Pipeline

From the repository root:

```bash
cd platform/go
```

Run the example optimisation:

```bash
go run ./cmd/owp optimise ../../examples/datasets/simple-events.json
```

Expected output:

```text
=== Optimised Plan ===
Work items planned: 3

  1. [patient.referred] WI-EVT-001
  2. [maintenance.requested] WI-EVT-002
  3. [delivery.scheduled] WI-EVT-003
```

---

### Run the Tests

From `platform/go`:

```bash
go test ./...
```

---

## Planned Capabilities

- Workforce Scheduling
- Constraint Optimisation
- Route Optimisation
- Workforce Planning
- AI Assisted Scheduling
- Explainable Optimisation
- Simulation
- Cloud Deployment
- Plugin Architecture

---

## Documentation

| Document | Purpose |
|----------|---------|
| Project Charter | Why the project exists |
| Roadmap | Long-term direction |
| Domain Model | Core business concepts |
| Architecture | System design |
| ADRs | Engineering decisions |
| Engineering Principles | How engineering decisions are made |
| AI-Assisted Development | Human and AI collaboration methodology |
| Research | Investigation and benchmarking |

---

## Engineering Principles

The platform is built around several key principles.

- Engineering before implementation.
- Domain Driven Design.
- Open architecture.
- Explain every important decision.
- Benchmark before optimising.
- Build for extension.
- Document the journey.
- Every dependency must earn its place.

---

## Contributing

Feedback, discussion and constructive engineering challenges are always welcome.

The project is being developed in the open so that both the software and the engineering methodology can be shared.

---

## About

This project is engineered by **Tim Dodgson** in collaboration with Artificial Intelligence.

It exists to demonstrate what becomes possible when modern AI is combined with decades of software engineering experience, optimisation research and disciplined engineering.

The software is important.

The engineering journey is equally important.