# Full INRC-II Competition Compliance

## Purpose

Make Open Workforce Platform capable of participating in the INRC-II competition format.

This means supporting the official INRC-II input/output model, multi-stage history, official constraints, official scoring behaviour, and validator-compatible output.

This is not INRC-inspired support.

This is competition compliance.

---

## Required Outcome

The platform must be able to:

1. Load official INRC-II competition files.
2. Load scenario, week/data and history/state files.
3. Optimise the required planning week.
4. Produce an INRC-II compatible solution file.
5. Update history/state for the next stage.
6. Score the solution using the same hard and soft constraint model as INRC-II.
7. Match the official validator score for supported instances.

If any official format detail is unclear, stop and ask for the official file/schema/example before implementing.

---

## Scope

Implement full INRC-II competition compatibility.

This includes:

- official input parsing
- official output writing
- multi-stage history support
- official hard constraints
- official soft constraints
- official penalty weights
- validator-compatible scoring
- benchmark execution over INRC-II instances

---

## Official File Support

Add support for the official INRC-II file structure.

The implementation must support:

- scenario file
- week/data file
- history/state file
- solution output file
- updated history/state output where required

Do not rely only on the simplified JSON NRP format.

The simplified JSON adapter may remain, but it must not be considered competition compliant.

---

## Multi-Stage Scheduling

INRC-II is multi-stage.

The solver must support solving one week at a time while carrying state across weeks.

History/state must include at minimum:

- last assigned shift type before the current week
- consecutive working days before the current week
- consecutive days off before the current week
- consecutive same-shift assignments before the current week
- total assignments so far
- total working weekends so far

The optimiser must use this history when evaluating constraints for the current week.

---

## Hard Constraints

Implement official INRC-II hard constraints.

At minimum:

### H1 - Single Assignment Per Day

A nurse may work at most one shift per day.

### H2 - Required Skill

A nurse assigned to a shift requiring a skill must possess that skill.

### H3 - Minimum Coverage

Minimum demand for each day, shift type and skill must be satisfied.

### H4 - Shift Succession

Forbidden shift type successions must not be violated, including successions between the previous history state and the first day of the current week.

Hard violations must be reported separately from soft penalties.

---

## Soft Constraints

Implement official INRC-II soft constraints and scoring.

At minimum:

### S1 - Insufficient Staffing For Optimal Coverage

Penalise coverage below optimal demand but above or equal to minimum demand.

### S2 - Consecutive Assignments

Penalise violations of minimum and maximum consecutive working days.

### S3 - Consecutive Days Off

Penalise violations of minimum and maximum consecutive days off.

### S4 - Consecutive Assignments Of Same Shift Type

Penalise violations of minimum and maximum consecutive assignments to the same shift type.

### S5 - Preferences / Requests

Penalise violated shift-on, shift-off, day-on and day-off requests.

### S6 - Complete Weekend

Penalise incomplete weekends where the nurse contract requires complete weekends.

### S7 - Total Assignments

Penalise total assignment count below minimum or above maximum over the complete planning horizon.

### S8 - Total Working Weekends

Penalise working weekends above the contract maximum over the complete planning horizon.

All penalty weights must match the official INRC-II specification.

---

## Output Compatibility

Add a writer for official INRC-II solution output.

The output must include the assignment of nurses to shifts in the format expected by the official validator.

The CLI must support:

```bash
go run ./cmd/owp solve-inrc2 <scenario-file> <week-file> <history-file> <solution-output-file>
```

If the official format requires additional output files, support them.

---

## Validation Compatibility

Add a validation mode:

```bash
go run ./cmd/owp validate-inrc2 <scenario-file> <week-file> <history-file> <solution-file>
```

This should calculate the platform score.

Where the official validator executable or reference output is available, add tests that compare:

- platform hard violations
- platform soft penalty
- total score

against the official validator result.

Definition of done requires score agreement on included official sample instances.

---

## Benchmark Compatibility

Add benchmark support for INRC-II instances:

```bash
go run ./cmd/owp benchmark-inrc2 <instance-directory> --profile research
```

The benchmark output should include:

- instance name
- week/stage
- algorithm
- hard violations
- soft penalty
- total objective
- delta versus constructive
- runtime
- candidates evaluated

---

## Architecture Rules

The optimiser must remain generic.

INRC-II-specific logic belongs in:

- INRC-II adapter/parser
- INRC-II history/state model
- optimisation input extraction
- shared hard constraint validation
- shared objective scoring
- INRC-II solution writer

INRC-II logic must not be placed inside individual algorithms.

Algorithms must continue to operate through OptimisationContext.

---

## Data Model Extensions

OptimisationContext must carry all INRC-II data required for compliance:

- nurses
- contracts
- skills
- shift types
- forbidden shift successions
- coverage requirements
- preferences/requests
- planning week information
- historical state
- total-horizon counters
- scoring weights

WorkItemInput and ResourceInput may be extended, but domain entities must not be modified unless there is a clear architectural reason.

---

## Tests

Add tests covering:

### Parser Tests

- official scenario file parsing
- official week/data file parsing
- official history/state file parsing

### Writer Tests

- solution output formatting
- deterministic solution output

### Hard Constraint Tests

- single shift per nurse per day
- required skill violation
- minimum coverage violation
- forbidden shift succession violation
- forbidden succession from previous history into current week

### Soft Constraint Tests

- optimal coverage penalty
- min/max consecutive working days
- min/max consecutive days off
- min/max consecutive same shift type
- shift-on request
- shift-off request
- day-on request
- day-off request
- complete weekend
- total assignments over horizon
- working weekends over horizon

### History Tests

- previous working streak affects current week
- previous days-off streak affects current week
- previous shift type affects first-day succession
- accumulated assignments affect total assignment penalty
- accumulated weekends affect weekend penalty

### Integration Tests

- solve official sample instance
- write official solution format
- validate platform score
- compare against official validator or known expected score
- run all algorithms on at least one official-style instance

---

## Non-Goals

Do not implement unrelated NRP variants.

Do not change algorithms unless needed to consume OptimisationContext correctly.

Do not introduce external dependencies unless official XML parsing cannot reasonably be done with Go standard library.

Do not claim compliance until official sample files round-trip and scoring matches expected validator output.

---

## Definition of Done

The implementation is complete only when:

- official INRC-II scenario/week/history files can be parsed
- official INRC-II solution files can be written
- multi-stage history is loaded and applied
- all official INRC-II hard constraints are enforced
- all official INRC-II soft constraints are scored
- official penalty weights are used
- platform scoring matches official validator or known expected outputs for included sample instances
- benchmark-inrc2 runs across official-style instances
- existing synthetic and simplified NRP benchmarks still pass
- all tests pass
- algorithms remain generic
- no nurse rostering rules are embedded inside algorithm implementations

---

## Open Questions

None. If official file format details are unavailable in the repository, stop and request official sample files before implementation.