# Parallel Feasible Roster Search

## Purpose

Implement Tim's optimisation algorithm for INRC-II nurse rostering.

The key principle is:

Hard constraints must be satisfied first.

Once a feasible roster exists, the algorithm should never search invalid solutions. It should only explore hard-feasible rosters and reduce soft constraint violations / official penalty.

This algorithm is based on the user's dissertation approach:

- start from a complete feasible roster
- preserve feasibility during search
- use nurse/day shift swaps as the core move
- optimise soft constraints only
- use annealing or late-acceptance-style exploration
- whenever a new global best is found, start a new parallel search branch from that solution with reset search state

---

## Why

Current generic algorithms work on work-item/resource assignment moves.

That is useful for the general Open Workforce Platform, but it is not the right search representation for INRC-II nurse rostering.

For INRC-II, the natural representation is a roster:

```text
nurse x day -> shift/off
```

A good roster search should preserve hard feasibility and only optimise soft penalties.

Hard constraint violations are not candidate solutions.

They are invalid and must be discarded.

---

## Algorithm Name

Register a new algorithm:

```text
parallel-feasible-roster-search
```

Short display name:

```text
PFRS
```

---

## Core Principle

The search must operate in this order:

1. Build or receive a hard-feasible roster.
2. If a hard-feasible roster cannot be built, fail clearly.
3. Search only from hard-feasible rosters.
4. Generate nurse/day shift swaps.
5. Reject any swap that violates hard constraints.
6. Score accepted candidates by soft penalty / soft violation count.
7. Track global best.
8. When a new global best is found, spawn a new parallel branch from that solution.
9. The spawned branch receives a full fresh iteration budget and reset temperature / late acceptance state.
10. Return the best hard-feasible roster found.

---

## Initial Feasible Roster

The algorithm must start from a complete hard-feasible roster.

Supported start modes:

### Existing Solution

If the context contains an existing feasible INRC-II solution, use it.

### Feasible Builder

If no existing solution is supplied, build a feasible roster from coverage requirements.

The builder must satisfy hard constraints:

- one shift per nurse per day
- required skill
- minimum coverage
- forbidden shift successions
- history-to-week forbidden succession

If no hard-feasible roster can be built, the algorithm must return a clear error:

```text
No hard-feasible initial roster found
```

Do not continue with an invalid roster.

---

## Representation

Use an INRC-II-specific internal representation.

Example:

```go
type Roster struct {
    Assignment map[NurseDay]ShiftAssignment
}
```

or an equivalent matrix/slice representation.

The representation must make it cheap to:

- get nurse assignment on a day
- swap two nurses on a day
- validate hard constraints
- calculate soft score
- convert to official INRC-II solution output

Do not force this algorithm through the generic work-item CandidateMove model.

---

## Move Operator

The primary move is:

```text
swap two nurses on the same day
```

For a selected day:

- choose nurse A
- choose nurse B
- swap their shift/off assignments for that day

This preserves:

- day coverage count
- number of shifts assigned that day
- one row per nurse/day

But the move may still violate:

- required skill
- forbidden shift succession
- availability/history-related hard constraints

Therefore every proposed swap must be hard-validated.

Invalid swaps are discarded, never accepted.

---

## Search Modes

Support two worker search modes:

```text
sa
lahc
```

### SA Mode

Use real simulated annealing.

Worker parameters:

- iterations per worker
- initial temperature
- cooling rate
- minimum temperature

Acceptance:

- accept improving soft score
- accept worse soft score using temperature probability
- reject hard-invalid candidates always

### LAHC Mode

Use Late Acceptance Hill Climbing.

Worker parameters:

- iterations per worker
- late acceptance length

Acceptance:

- accept candidate if better than current
- or candidate score is no worse than score from N iterations ago
- reject hard-invalid candidates always

---

## Parallel Branching

The algorithm must run multiple workers.

A worker starts from a hard-feasible roster and runs a full local search.

When a worker finds a solution better than the current global best:

1. update global best
2. create a new branch from that solution
3. reset the branch's SA temperature or LAHC history
4. give the branch a full fresh iteration budget
5. start the branch if capacity allows

Important:

- A branch is created only for new global best solutions.
- Do not branch on local improvements.
- Do not spawn unbounded goroutines.

---

## Parallelism Controls

Expose configuration:

```go
PFRSMode string // sa or lahc
PFRSIterationsPerWorker int
PFRSMaxConcurrentWorkers int
PFRSMaxTotalWorkers int
PFRSBranchOnGlobalBest bool
PFRSInitialTemperature float64
PFRSCoolingRate float64
PFRSMinTemperature float64
PFRSLateAcceptanceLength int
PFRSSeed int64
PFRSScoringMode string // official-penalty or soft-violation-count
```

Default mode:

```text
sa
```

Default scoring:

```text
official-penalty
```

The dissertation-style scoring mode should be available:

```text
soft-violation-count
```

---

## Determinism

Parallel execution can make results non-deterministic.

For research repeatability, the algorithm must use deterministic seeds per worker:

```text
seed = baseSeed + workerID
```

The final result must be deterministic when:

- same dataset
- same profile
- same algorithm parameters
- same seed
- same max workers

If full determinism is not possible with goroutine scheduling, document it clearly and provide a deterministic mode:

```text
PFRSDeterministic bool
```

When deterministic mode is enabled, workers may be processed through a controlled queue while still using the same branching logic.

---

## Integration

The algorithm is INRC-II-specific but must not pollute generic algorithms.

Allowed locations:

```text
internal/infrastructure/inrc2/
internal/optimisation/
```

Preferred approach:

- INRC-II roster search implementation in `internal/infrastructure/inrc2`
- algorithm registration bridge in `internal/optimisation` if needed

The algorithm may be available only for INRC-II solving / benchmark commands.

If run on a non-INRC dataset, it should fail clearly:

```text
parallel-feasible-roster-search requires INRC-II context
```

---

## CLI

Support:

```bash
go run ./cmd/owp benchmark-inrc2 --algorithm parallel-feasible-roster-search
```

Support tunables:

```bash
--pfrs-mode sa
--pfrs-iterations-per-worker 30000
--pfrs-max-concurrent-workers 8
--pfrs-max-total-workers 64
--pfrs-initial-temperature 1
--pfrs-cooling-rate 0.0009
--pfrs-min-temperature 0.0001
--pfrs-late-acceptance-length 1000
--pfrs-seed 42
--pfrs-scoring-mode soft-violation-count
--pfrs-deterministic false
```

Also support LAHC:

```bash
go run ./cmd/owp benchmark-inrc2 \
  --algorithm parallel-feasible-roster-search \
  --pfrs-mode lahc \
  --pfrs-iterations-per-worker 30000 \
  --pfrs-late-acceptance-length 1000
```

---

## Output

Benchmark output must show:

- algorithm name
- hard violations
- soft violations
- official penalty
- scoring mode
- workers started
- branches created
- iterations completed
- candidates evaluated
- improvements accepted
- best updates
- runtime

The main ranked benchmark table must still include only hard-valid solutions.

---

## Statistics

Extend algorithm statistics for PFRS:

```go
WorkersStarted int
BranchesCreated int
BestUpdates int
InvalidMovesRejected int
```

Do not remove existing statistics.

---

## Tests

Add tests covering:

### Initial Feasible Builder

- builds hard-feasible roster for small official instance
- fails clearly when impossible
- satisfies minimum coverage
- satisfies required skill
- satisfies one shift per nurse per day
- respects forbidden succession from history

### Swap Operator

- swap preserves coverage
- swap preserves one assignment per nurse/day
- invalid skill swap rejected
- invalid succession swap rejected
- valid swap accepted

### Search

- SA mode improves or preserves soft score
- LAHC mode improves or preserves soft score
- new global best creates a new branch
- branch receives full iteration budget
- branch resets temperature / LAHC history
- max concurrent workers respected
- max total workers respected
- deterministic mode is reproducible

### CLI

- algorithm can run alone
- CLI parameters override profile
- non-INRC run fails clearly
- output includes PFRS statistics

### Regression

- official INRC-II validator parity remains unchanged
- existing algorithms still run
- synthetic benchmarks unaffected

---

## Non-Goals

Do not change official INRC-II scoring.

Do not change validator parity.

Do not replace the generic optimisation algorithms.

Do not add random non-deterministic benchmarking by default.

Do not hide hard-invalid solutions as valid.

Do not implement UI.

---

## Definition Of Done

The implementation is complete when:

- `parallel-feasible-roster-search` is registered.
- It starts only from hard-feasible rosters.
- It never accepts hard-invalid moves.
- It uses nurse/day swap moves.
- It supports SA mode.
- It supports LAHC mode.
- It branches on new global best solutions.
- Each branch receives a full fresh iteration budget.
- Parallel execution is bounded.
- CLI exposes all PFRS tunables.
- Benchmark can run PFRS alone.
- Hard-valid PFRS results appear in the ranked table.
- Invalid results are never ranked.
- Official validator parity remains unchanged.
- All tests pass.