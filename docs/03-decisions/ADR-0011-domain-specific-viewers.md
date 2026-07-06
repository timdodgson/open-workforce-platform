# ADR-0011

## Title

Domain-specific visualisations are opt-in additions, not requirements

## Status

Accepted

## Context

Different domains produce different outputs: NRP produces schedules, CVRP produces routes, JSS produces Gantt charts. A generic summary page cannot meaningfully visualise all of these. But requiring a custom viewer before a domain can be used would slow down new domain adoption.

## Decision

Every domain gets basic support for free (summary metrics, search progress, discovery timeline) via the shared telemetry format. Domain-specific viewers (Gantt chart, route viewer, schedule grid) are added as optional pages. The sidebar navigation adapts per domain to show only relevant pages.

## Alternatives

- **Force all domains through generic visualisation only.** Loses the ability to inspect solutions meaningfully.
- **Require a custom viewer before a domain is considered complete.** Increases barrier to adding domains.
- **Single unified viewer that adapts.** Impractical given how different schedules, routes, and Gantt charts are.

## Consequences

- New domains work immediately on the dashboard (summary + search progress).
- Rich visualisation is added incrementally when the domain matures.
- Sidebar navigation detects `problemType` and shows appropriate pages.
- Each viewer is a self-contained page component with no coupling to other domains.
