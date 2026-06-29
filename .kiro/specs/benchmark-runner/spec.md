# Benchmark Runner

## Purpose

Add a command-line benchmark runner that compares optimisation algorithms across benchmark datasets.

The benchmark runner should make it easy to see how each algorithm performs across multiple optimisation scenarios.

---

# Why

The project now has:

- multiple optimisation algorithms
- multiple benchmark datasets
- optimisation statistics
- objective scoring

Running each dataset and algorithm manually is slow and error-prone.

A benchmark runner makes algorithm comparison repeatable.

---

# Behaviour

Add a new CLI command:

go run ./cmd/owp benchmark ../../examples/datasets

The command should:

- discover benchmark dataset JSON files in the supplied directory
- run each supported algorithm against each dataset
- print a readable comparison table

---

# Algorithms

Run all currently supported algorithms:

- constructive
- hill-climbing
- simulated-annealing

Do not hard-code behaviour beyond the currently available algorithms.

Use the existing algorithm registry where possible.

---

# Output

Example:

Dataset                  Algorithm             Score   Objective   Assigned   Duration   Candidates
constructive-baseline    constructive          100     2976        3          1ms        0
constructive-baseline    hill-climbing         100     2976        3          2ms        12
constructive-baseline    simulated-annealing   100     2976        3          3ms        18

The exact formatting may vary, but it must be readable.

---

# Architecture

The CLI should orchestrate benchmark execution.

The benchmark runner should reuse existing:

- dataset loader
- application workflow
- optimisation algorithms
- optimisation statistics

Do not duplicate optimisation logic.

---

# Tests

Add tests covering:

- benchmark command discovers datasets
- benchmark command runs all algorithms
- benchmark results include statistics
- invalid dataset directory returns a clear error

---

# Non-Goals

Do not implement:

- CSV export
- JSON export
- charts
- historical comparisons
- performance profiling
- parallel execution

---

# Constraints

Use only the Go standard library.

Do not introduce external dependencies.

Do not change optimisation behaviour.

Keep the implementation intentionally simple.

---

# Definition of Done

The implementation is complete when:

- benchmark command runs from the CLI
- all benchmark datasets are executed
- all supported algorithms are executed
- output includes objective score and statistics
- tests pass
- no external dependencies are introduced

---

# Open Questions

None.
