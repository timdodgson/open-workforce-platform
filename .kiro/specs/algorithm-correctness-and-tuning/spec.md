# Algorithm Correctness And Tuning

## Purpose

Fix the current optimisation algorithms so their implementations match their names, expose their real tuning parameters, and allow each algorithm to be run independently with explicit values.

This is a full correction spec.

No half measures.

---

## Why

The current audit showed several problems:

- Simulated Annealing is not true simulated annealing.
- Algorithm tuning is incomplete.
- Some behaviour is hidden behind profiles.
- The previous `--time` concept was misleading.
- Algorithms cannot be cleanly run one at a time with their own parameters.
- Research benchmarking requires explicit, reproducible configuration.

Before implementing Tim's Algorithm, the baseline algorithms must be correct, tunable and reproducible.

---

## Required Outcomes

The platform must support:

- correct algorithm implementations
- explicit per-algorithm configuration
- running one algorithm at a time
- benchmark runs with either all algorithms or one selected algorithm
- effective configuration shown in output
- deterministic reproducible results
- no misleading time-based tuning

---

## Global Rules

Do not change scoring.

Do not change validator parity.

Do not change INRC-II official scoring.

Do not change domain objects.

Do not put INRC-II business rules inside algorithms.

Algorithms must remain generic.

---

# Algorithm Requirements

## Constructive

Constructive remains a single-pass deterministic greedy algorithm.

It has no tuning parameters.

It should be clearly reported as having no tunables.

---

## Hill Climbing

Hill Climbing must expose:

- `HCMaxIterations`

Behaviour:

- Start from constructive or warm-start plan.
- Evaluate neighbourhood moves.
- Accept improving moves only.
- Stop when no improvement is found or max iterations reached.

Optional behaviour:

- keep first-improvement strategy for now.

Do not pretend Hill Climbing has temperature, tabu or LNS settings.

---

## Simulated Annealing

Simulated Annealing must be corrected.

It must implement real simulated annealing behaviour.

Required tunables:

- `SAMaxIterations`
- `SAInitialTemperature`
- `SACoolingRate`
- `SAMinTemperature`

Required behaviour:

- Start from constructive or warm-start plan.
- Generate candidate moves.
- Evaluate candidate score delta.
- Always accept improving moves.
- Accept worse moves according to temperature-based acceptance probability.
- Cool temperature using cooling rate.
- Stop when max iterations reached or temperature falls below minimum temperature.

Acceptance rule:

Use the standard form:

`acceptanceProbability = exp(delta / temperature)`

where:

- delta is candidateScore - currentScore
- improving delta is positive
- worse delta is negative

Because this platform values deterministic reproducibility, do not use true randomness.

Instead, use a deterministic pseudo-random source seeded from stable input:

- algorithm name
- dataset or context identity where available
- profile values

If no stable identity exists, use a fixed seed.

The same input and same parameters must produce the same result.

---

## Tabu Search

Tabu Search must expose:

- `TabuMaxIterations`
- `TabuListSize`
- `TabuAspirationEnabled`

Required behaviour:

- Start from constructive or warm-start plan.
- Generate candidate moves.
- Exclude tabu moves.
- If aspiration is enabled, allow a tabu move if it improves the best score found so far.
- Select best admissible move.
- Add accepted move to tabu list.
- Stop when max iterations reached or no admissible moves exist.

Tabu Search must remain deterministic.

---

## Large Neighbourhood Search

Large Neighbourhood Search must expose:

- `LNSIterations`
- `LNSDestroySize`
- `LNSRepairStrategy`

Required behaviour:

- Start from constructive or warm-start plan.
- Destroy part of the solution.
- Repair using the selected repair strategy.
- Accept repaired solution if it improves the current best score.
- Stop when iteration limit reached.

Supported repair strategies:

- `greedy`
- `priority`
- `best-fit`

Definitions:

`greedy`:
Use existing constructive-style repair.

`priority`:
Repair highest-priority unassigned work first.

`best-fit`:
For each unassigned item, choose the feasible resource that gives the best resulting score.

Do not add random destroy strategies in this spec.

---

# Algorithm Profile

Update `AlgorithmProfile` to include all algorithm-specific fields:

```go
type AlgorithmProfile struct {
    HCMaxIterations int

    SAMaxIterations int
    SAInitialTemperature float64
    SACoolingRate float64
    SAMinTemperature float64

    TabuMaxIterations int
    TabuListSize int
    TabuAspirationEnabled bool

    LNSIterations int
    LNSDestroySize int
    LNSRepairStrategy string
}
```

Existing profiles must be updated:

- `default`
- `fast`
- `quality`
- `research`

Profiles must use sensible values for every field.

---

# CLI Requirements

## Existing Commands

Update:

```bash
owp optimise
owp benchmark
owp solve-inrc2
owp benchmark-inrc2
```

## Algorithm Selection

All relevant commands must support:

```bash
--algorithm <name>
```

If omitted in benchmark commands, run all algorithms.

If supplied in benchmark commands, run only that algorithm.

---

## Explicit Tuning Flags

Support explicit algorithm tuning flags:

```bash
--hc-max-iterations <int>

--sa-max-iterations <int>
--sa-initial-temperature <float>
--sa-cooling-rate <float>
--sa-min-temperature <float>

--tabu-max-iterations <int>
--tabu-list-size <int>
--tabu-aspiration <true|false>

--lns-iterations <int>
--lns-destroy-size <int>
--lns-repair-strategy <greedy|priority|best-fit>
```

Explicit CLI flags override profile values.

Profiles provide defaults only.

---

## Remove Misleading Time Flag

Remove or deprecate:

```bash
--time
```

Do not use `time` unless it means real wall-clock time.

If the flag already exists, make it fail clearly with:

```text
--time is not supported. Use explicit algorithm tuning flags such as --tabu-max-iterations or --sa-max-iterations.
```

---

## Effective Configuration Output

Every optimisation and benchmark run must display the effective algorithm configuration.

Example:

```text
Algorithm: tabu-search
Profile: research

Effective Configuration:
  TabuMaxIterations: 500
  TabuListSize: 50
  TabuAspirationEnabled: true
```

For benchmark runs, include configuration either:

- once at the top when running one algorithm
- per algorithm in a concise section when running all algorithms

---

# INRC-II Benchmark Behaviour

`benchmark-inrc2` must support:

Run all algorithms:

```bash
go run ./cmd/owp benchmark-inrc2
```

Run one algorithm:

```bash
go run ./cmd/owp benchmark-inrc2 --algorithm tabu-search --tabu-max-iterations 5000 --tabu-list-size 100
```

Run SA with explicit real SA parameters:

```bash
go run ./cmd/owp benchmark-inrc2 --algorithm simulated-annealing --sa-max-iterations 10000 --sa-initial-temperature 1000 --sa-cooling-rate 0.995 --sa-min-temperature 0.01
```

---

# Validity Handling

Do not show invalid hard-constraint solutions in the main benchmark ranking.

Default benchmark output must show valid solutions only.

Invalid solutions may be shown only with:

```bash
--show-invalid
```

No hard-constraint-violating solution may be ranked as a candidate solution.

---

# Tests

Add tests covering:

## Profile Tests

- default profile contains all fields
- fast/quality/research profiles contain all fields
- explicit CLI flags override profile values
- invalid profile values fail clearly

## Simulated Annealing Tests

- uses temperature values
- cooling rate reduces temperature
- improving moves are accepted
- worse moves may be accepted according to deterministic probability
- deterministic output with same seed/config
- different SA parameters can change behaviour

## Tabu Tests

- tabu list size is applied
- aspiration allows best-ever tabu move when enabled
- aspiration disabled blocks tabu moves
- deterministic output

## LNS Tests

- destroy size is applied
- repair strategy is applied
- invalid repair strategy fails clearly
- deterministic output

## CLI Tests

- one algorithm can be run alone
- benchmark can run one algorithm only
- effective config is displayed
- `--time` fails clearly
- explicit flags override profile values

## Regression

- official INRC-II validator parity remains unchanged
- existing OWP benchmarks still pass
- invalid hard-constraint solutions remain hidden by default

---

# Non-Goals

Do not implement real wall-clock time limits.

Do not implement parallel execution.

Do not improve INRC-II-specific heuristics in this spec.

Do not implement Tim's Algorithm in this spec.

---

# Definition Of Done

The implementation is complete when:

- Simulated Annealing is a real temperature-based algorithm.
- All algorithms expose their relevant tunables.
- All tunables can be set through profiles.
- All tunables can be overridden through CLI flags.
- One algorithm can be run independently with explicit values.
- Benchmark commands can run either all algorithms or one selected algorithm.
- Effective configuration is displayed.
- `--time` is removed or rejected clearly.
- Hard-invalid solutions are hidden by default.
- Official INRC-II validator parity still passes.
- All tests pass.