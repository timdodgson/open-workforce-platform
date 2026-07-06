# Run Metadata Schema (run.json)

## Purpose

Every experiment run produces a `run.json` file that captures the configuration, results, and metadata needed by the dashboard. This document defines the contract between the Go CLI (producer) and the Next.js dashboard (consumer).

## Schema

### Required Fields (all domains)

| Field | Type | Description |
|-------|------|-------------|
| `problemType` | string | Domain identifier: `"nrp"`, `"cvrp"`, `"vrptw"`, `"jss"`, `"ilp"` |
| `mode` | string | Algorithm used: `"sa"`, `"lahc"`, `"tabu"`, `"portfolio"`, `"adaptive"`, `"ilp"` |
| `instance` | string | Instance name (e.g. `"A-n32-k5"`, `"n030w4"`, `"ft06"`, `"C101"`) |
| `bestObjective` | int | Best objective value achieved (lower is better). Universal field. |
| `runLabel` | string | Unique run identifier used as S3 key prefix |
| `runtimeMs` | int | Total solver runtime in milliseconds |
| `iterations` | int | Iteration budget per strategy |
| `seed` | int | Random seed used |

### Domain-Specific Fields

#### NRP (Nurse Rostering)

| Field | Type | Description |
|-------|------|-------------|
| `totalPenalty` | int | Sum of soft constraint penalties across all weeks (same as `bestObjective`) |

#### CVRP (Capacitated Vehicle Routing)

| Field | Type | Description |
|-------|------|-------------|
| `bestDistance` | int | Total travel distance (same as `bestObjective`) |
| `customers` | int | Number of customers in instance |
| `capacity` | int | Vehicle capacity |
| `feasible` | bool | Whether the best solution satisfies all capacity constraints |
| `initialDistance` | int | Constructive heuristic baseline distance |

#### VRPTW (Vehicle Routing with Time Windows)

| Field | Type | Description |
|-------|------|-------------|
| `bestDistance` | int | Total travel distance (same as `bestObjective`) |
| `customers` | int | Number of customers |
| `capacity` | int | Vehicle capacity |
| `vehicles` | int | Max vehicles available |
| `bestVehicles` | int | Vehicles used in best solution |
| `feasible` | bool | Whether best solution satisfies capacity + time windows |
| `initialDistance` | int | Constructive heuristic baseline |

#### JSS (Job Shop Scheduling)

| Field | Type | Description |
|-------|------|-------------|
| `bestMakespan` | int | Best makespan achieved (same as `bestObjective`) |
| `jobs` | int | Number of jobs |
| `machines` | int | Number of machines |
| `initialMakespan` | int | Constructive heuristic baseline |

#### ILP (Integer Linear Programming)

| Field | Type | Description |
|-------|------|-------------|
| `objective` | int | Proven optimal/best bound (same as `bestObjective`) |
| `weeks` | int | Planning horizon (NRP only) |
| `solver` | string | Solver name (e.g. `"highs"`) |
| `gap` | float | Optimality gap percentage |
| `status` | string | Solve status: `"optimal"`, `"feasible"`, `"infeasible"` |

## Dashboard Reading Priority

The dashboard reads the objective value in this order:

1. `bestObjective` (universal — preferred)
2. `bestDistance` (CVRP/VRPTW)
3. `bestMakespan` (JSS)
4. `totalPenalty` (NRP)
5. `objective` (ILP)
6. Fall back to summary CSV parsing

## S3 Storage Structure

```
pfrs-research-lab-data/
├── manifest.json          # Index of all runs
└── runs/
    └── <runLabel>/
        ├── run.json       # This schema
        ├── solution.json  # Domain-specific solution
        ├── results.csv    # Per-week audit (NRP) or search log
        └── discoveries.csv # Global best improvements timeline
```

## Manifest Entry

Each run is registered in `manifest.json`:

```json
{
  "runId": "cvrp-a32k5-sa",
  "label": "cvrp-a32k5-sa",
  "algorithm": "sa",
  "timestamp": "2026-07-05T14:19:59Z",
  "totalPenalty": 814,
  "storageVersion": "1.0"
}
```

The `totalPenalty` field in the manifest matches `bestObjective` in run.json.
