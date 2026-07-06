# ADR-0005

## Title

Shared telemetry format across all algorithms and domains

## Status

Accepted

## Context

The dashboard needs to display search progress, convergence, and discovery timelines for any run regardless of domain or algorithm. If each domain produces different telemetry formats, the dashboard becomes a collection of domain-specific parsers.

## Decision

All algorithms emit `[]Discovery` — a slice of timestamped global best improvements. Every `SearchResult` includes candidates, accepted, rejected, improved counts, and duration. This uniform telemetry powers all dashboard charts.

## Alternatives

- **Domain-specific telemetry structs.** Rich but requires dashboard changes per domain.
- **Structured event log.** More flexible but higher storage cost and parsing complexity.
- **No telemetry, just final result.** Insufficient for research analysis.

## Consequences

- Dashboard search progress, timeline, and convergence pages work for any domain without code changes.
- Adding a new domain gets full telemetry visualisation for free.
- Domain-specific telemetry (e.g. NRP per-week breakdown) is layered on top, not a replacement.
- Discovery events are lightweight (5 fields) — negligible storage overhead.
