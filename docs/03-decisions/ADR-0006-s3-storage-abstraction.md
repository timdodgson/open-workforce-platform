# ADR-0006

## Title

Storage abstraction with S3 as production backend

## Status

Accepted

## Context

The dashboard needs to read run data. In development, data lives on the local filesystem. In production (ECS Fargate), there is no persistent filesystem — data must come from S3. The dashboard code should not know which backend it uses.

## Decision

Implement a `StorageProvider` interface with `listRuns()`, `readFile()`, and `exists()`. Two implementations: `LocalStorageProvider` (reads from `data/runs/`) and `S3StorageProvider` (reads from S3 bucket). Selected via `STORAGE_PROVIDER` environment variable.

## Alternatives

- **S3 only.** Requires AWS credentials for local development.
- **Database (DynamoDB/Postgres).** Heavier infrastructure for what is essentially file storage.
- **Git-based storage.** Limits run size and requires push/pull workflow.

## Consequences

- Development works with local files, no AWS setup needed.
- Production reads from S3 with no code changes.
- The Go CLI writes locally AND uploads to S3 — both backends always have data.
- Manifest pattern (`manifest.json`) gives O(1) run discovery without bucket listing.
- Storage interface is simple enough that a third backend (e.g. database) could be added without dashboard changes.
