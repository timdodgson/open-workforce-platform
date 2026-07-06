# ADR-0010

## Title

CSV as the primary telemetry interchange format

## Status

Accepted

## Context

The Go CLI produces per-run telemetry (discoveries, audit rows, search progress). The Next.js dashboard consumes this data. The format must be easy to produce in Go, easy to parse in TypeScript, and human-readable for debugging.

## Decision

Use CSV for tabular telemetry data (`discoveries.csv`, `results.csv`). Use JSON for structured metadata (`run.json`, `solution.json`). CSV is chosen for telemetry because it streams well, appends efficiently, and opens directly in spreadsheet tools for manual inspection.

## Alternatives

- **JSON arrays.** Require full file read before parsing. Large for high-frequency telemetry.
- **Parquet/Arrow.** Efficient but adds binary format dependency and complicates debugging.
- **SQLite.** Powerful queries but heavy dependency for simple append-only data.
- **Protocol Buffers.** Fast but not human-readable.

## Consequences

- Telemetry files are human-readable and can be inspected with `head`, `tail`, or Excel.
- Parsing is simple — split lines, split commas.
- No schema validation at the format level (CSV is untyped). The dashboard parsers handle type conversion.
- Large NRP runs (8 weeks × many workers) produce large CSV files, but still manageable at current scale.
- JSON is used where structure matters (metadata, solutions) and CSV where tabular data matters (telemetry).
