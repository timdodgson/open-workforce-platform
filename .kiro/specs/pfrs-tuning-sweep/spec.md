# PFRS Tuning Sweep

## Purpose

Add a tuning command for Parallel Feasible Roster Search.

The goal is to compare PFRS parameter combinations on the standard INRC-II benchmark and identify strong default settings.

---

## Behaviour

Add:

go run ./cmd/owp tune-pfrs

Default target:

- n012w8
- SA mode
- official-penalty scoring

The command should run a small deterministic grid of parameter combinations.

Initial grid:

- iterations per worker: 30000, 60000, 100000
- max total workers: 16, 32
- initial temperature: 1.0, 2.0, 5.0
- cooling rate: 0.0009, 0.0005, 0.0001

Keep max concurrent workers default at 4 unless overridden.

---

## Output

Display a ranked table:

Rank
Iterations
Workers
Initial Temp
Cooling Rate
Penalty
Soft Violations
Runtime
Candidates

Best result first.

Only hard-valid results should be ranked.

Invalid results should be hidden unless --show-invalid is supplied.

---

## CLI

Support:

--instance <path-or-name>
--max-concurrent-workers <int>
--show-invalid

---

## Architecture

This is a benchmarking/tuning utility only.

Do not change PFRS behaviour.

Do not change scoring.

Do not change validator parity.

---

## Tests

Add tests for:

- grid generation
- deterministic ordering
- ranking by hard-valid then penalty
- command runs on a small fixture

---

## Definition of Done

- tune-pfrs command exists
- runs deterministic parameter grid
- ranks valid results by penalty
- shows best settings clearly
- all tests pass