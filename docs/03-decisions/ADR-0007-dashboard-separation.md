# ADR-0007

## Title

Dashboard is a separate application from the optimisation engine

## Status

Accepted

## Context

The optimisation engine is a Go CLI that runs experiments. The dashboard is a web application that visualises results. These have different deployment models, technology stacks, and development cadences.

## Decision

The dashboard is a standalone Next.js application. It reads data from S3 (or local files). It has no Go dependency at runtime. The Go CLI writes JSON/CSV outputs that the dashboard consumes. The contract between them is the `run.json` schema.

## Alternatives

- **Monolithic Go web server.** Simpler deployment but limits UI development velocity and forces Go templating.
- **Dashboard embedded in CLI (TUI).** No remote access, limited visualisation.
- **Notebook-based analysis (Jupyter).** Good for exploration but not a persistent shared dashboard.

## Consequences

- Go and TypeScript codebases evolve independently.
- Dashboard deploys via Docker/ECS. CLI runs on developer machines.
- The schema contract (`run.json`) must be maintained by both sides.
- Dashboard can be developed and tested without running any experiments.
- Multiple people can view results simultaneously via the web interface.
