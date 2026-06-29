# Optimisation Statistics

## Purpose

Capture optimisation execution statistics so algorithms can be compared and understood.

The optimiser should report not only the final plan, but also how the algorithm behaved while producing it.

---

# Why

As the platform adds more algorithms, objective functions and neighbourhoods, it becomes important to understand algorithm performance.

A useful optimisation platform should be able to answer:

- which algorithm was used
- how long optimisation took
- how many iterations ran
- how many candidate moves were evaluated
- how many improvements were accepted
- what final objective score was achieved

This supports benchmarking, research and explainability.

---

# Behaviour

Each algorithm should produce optimisation statistics.

Statistics should be included in the Optimised Plan.

The CLI should display the statistics after the objective breakdown.

---

# Statistics

Introduce optimisation statistics containing:

- algorithm name
- duration
- iterations
- candidates evaluated
- improvements accepted
- final objective score

The constructive algorithm may report:

- iterations: 1
- candidates evaluated: 0
- improvements accepted: 0

Search-based algorithms should report meaningful counts where practical.

---

# Architecture

Algorithms are responsible for collecting their own statistics.

The Optimised Plan stores the statistics.

The CLI displays statistics but does not calculate them.

---

# Output

Example:

Optimisation Statistics:
  Algorithm: hill-climbing
  Duration: 4ms
  Iterations: 12
  Candidates Evaluated: 84
  Improvements Accepted: 3
  Final Objective Score: 2976

---

# Non-Goals

Do not implement:

- benchmark runner
- historical reporting
- metrics export
- Prometheus
- tracing
- performance dashboards

---

# Tests

Add tests covering:

- statistics are present on the Optimised Plan
- constructive reports expected baseline statistics
- hill climbing reports iterations and candidates evaluated
- simulated annealing reports iterations and candidates evaluated
- CLI output includes statistics
- statistics do not affect optimisation behaviour

---

# Constraints

Use only the Go standard library.

Do not introduce external dependencies.

Do not modify domain objects.

Do not change optimisation behaviour.

Keep the implementation intentionally simple.

---

# Definition of Done

The implementation is complete when:

- algorithms populate optimisation statistics
- Optimised Plan exposes statistics
- CLI displays statistics
- all tests pass
- existing commands still run successfully
- no external dependencies are introduced

---

# Open Questions

None.