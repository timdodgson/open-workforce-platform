# Benchmark Suite

## Purpose

Reproducible benchmark commands for all domains, instances, and algorithm variations.
Run from `platform/go`. All commands upload to S3 for dashboard visibility.

---

## Algorithms

| Mode | Description |
|------|-------------|
| sa | Simulated Annealing (adaptive cooling) |
| lahc | Late Acceptance Hill Climbing |
| tabu | Tabu Search (best-move neighbourhood) |
| portfolio | Parallel multi-strategy (SA + LAHC + Tabu) |
| adaptive | SA with LAHC escape bursts on stagnation |

---

## CVRP — Capacitated Vehicle Routing

**Instances:** A-n32-k5 (31 customers), A-n45-k6 (44), A-n60-k9 (59), A-n80-k10 (79)

```bash
# A-n32-k5 (optimal: 784)
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n32-k5.vrp --run-label cvrp-a32k5-sa --mode sa --iterations 500000 --storage s3
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n32-k5.vrp --run-label cvrp-a32k5-lahc --mode lahc --iterations 500000 --storage s3
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n32-k5.vrp --run-label cvrp-a32k5-tabu --mode tabu --iterations 500000 --storage s3
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n32-k5.vrp --run-label cvrp-a32k5-portfolio --mode portfolio --iterations 500000 --storage s3
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n32-k5.vrp --run-label cvrp-a32k5-adaptive --mode adaptive --iterations 500000 --storage s3

# A-n45-k6 (optimal: 944)
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n45-k6.vrp --run-label cvrp-a45k6-sa --mode sa --iterations 500000 --storage s3
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n45-k6.vrp --run-label cvrp-a45k6-lahc --mode lahc --iterations 500000 --storage s3
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n45-k6.vrp --run-label cvrp-a45k6-tabu --mode tabu --iterations 500000 --storage s3
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n45-k6.vrp --run-label cvrp-a45k6-portfolio --mode portfolio --iterations 500000 --storage s3
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n45-k6.vrp --run-label cvrp-a45k6-adaptive --mode adaptive --iterations 500000 --storage s3

# A-n60-k9 (optimal: 1354)
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n60-k9.vrp --run-label cvrp-a60k9-sa --mode sa --iterations 500000 --storage s3
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n60-k9.vrp --run-label cvrp-a60k9-lahc --mode lahc --iterations 500000 --storage s3
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n60-k9.vrp --run-label cvrp-a60k9-tabu --mode tabu --iterations 500000 --storage s3
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n60-k9.vrp --run-label cvrp-a60k9-portfolio --mode portfolio --iterations 500000 --storage s3
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n60-k9.vrp --run-label cvrp-a60k9-adaptive --mode adaptive --iterations 500000 --storage s3

# A-n80-k10 (optimal: 1763)
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n80-k10.vrp --run-label cvrp-a80k10-sa --mode sa --iterations 500000 --storage s3
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n80-k10.vrp --run-label cvrp-a80k10-lahc --mode lahc --iterations 500000 --storage s3
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n80-k10.vrp --run-label cvrp-a80k10-tabu --mode tabu --iterations 500000 --storage s3
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n80-k10.vrp --run-label cvrp-a80k10-portfolio --mode portfolio --iterations 500000 --storage s3
go run ./cmd/owp solve-cvrp --instance ../../examples/cvrp/A-n80-k10.vrp --run-label cvrp-a80k10-adaptive --mode adaptive --iterations 500000 --storage s3
```

**Total: 20 runs**

---

## JSS — Job Shop Scheduling

**Instances:** ft06 (6×6, optimal 55), ft10 (10×10, optimal 930), la01 (10×5, optimal 666)

```bash
# ft06 (optimal: 55)
go run ./cmd/owp solve-jobshop --instance internal/infrastructure/jobshop/testdata/ft06.txt --run-label jss-ft06-sa --mode sa --iterations 500000 --storage s3
go run ./cmd/owp solve-jobshop --instance internal/infrastructure/jobshop/testdata/ft06.txt --run-label jss-ft06-lahc --mode lahc --iterations 500000 --storage s3
go run ./cmd/owp solve-jobshop --instance internal/infrastructure/jobshop/testdata/ft06.txt --run-label jss-ft06-tabu --mode tabu --iterations 500000 --storage s3
go run ./cmd/owp solve-jobshop --instance internal/infrastructure/jobshop/testdata/ft06.txt --run-label jss-ft06-portfolio --mode portfolio --iterations 500000 --storage s3
go run ./cmd/owp solve-jobshop --instance internal/infrastructure/jobshop/testdata/ft06.txt --run-label jss-ft06-adaptive --mode adaptive --iterations 500000 --storage s3

# ft10 (optimal: 930)
go run ./cmd/owp solve-jobshop --instance internal/infrastructure/jobshop/testdata/ft10.txt --run-label jss-ft10-sa --mode sa --iterations 500000 --storage s3
go run ./cmd/owp solve-jobshop --instance internal/infrastructure/jobshop/testdata/ft10.txt --run-label jss-ft10-lahc --mode lahc --iterations 500000 --storage s3
go run ./cmd/owp solve-jobshop --instance internal/infrastructure/jobshop/testdata/ft10.txt --run-label jss-ft10-tabu --mode tabu --iterations 500000 --storage s3
go run ./cmd/owp solve-jobshop --instance internal/infrastructure/jobshop/testdata/ft10.txt --run-label jss-ft10-portfolio --mode portfolio --iterations 500000 --storage s3
go run ./cmd/owp solve-jobshop --instance internal/infrastructure/jobshop/testdata/ft10.txt --run-label jss-ft10-adaptive --mode adaptive --iterations 500000 --storage s3

# la01 (optimal: 666)
go run ./cmd/owp solve-jobshop --instance internal/infrastructure/jobshop/testdata/la01.txt --run-label jss-la01-sa --mode sa --iterations 500000 --storage s3
go run ./cmd/owp solve-jobshop --instance internal/infrastructure/jobshop/testdata/la01.txt --run-label jss-la01-lahc --mode lahc --iterations 500000 --storage s3
go run ./cmd/owp solve-jobshop --instance internal/infrastructure/jobshop/testdata/la01.txt --run-label jss-la01-tabu --mode tabu --iterations 500000 --storage s3
go run ./cmd/owp solve-jobshop --instance internal/infrastructure/jobshop/testdata/la01.txt --run-label jss-la01-portfolio --mode portfolio --iterations 500000 --storage s3
go run ./cmd/owp solve-jobshop --instance internal/infrastructure/jobshop/testdata/la01.txt --run-label jss-la01-adaptive --mode adaptive --iterations 500000 --storage s3
```

**Total: 15 runs**

---

## VRPTW — Vehicle Routing with Time Windows

**Instances:** C101 (100 customers, BKS: 828)

```bash
# C101 (best known: 828)
go run ./cmd/owp solve-vrptw --instance ../../examples/vrptw/C101.txt --run-label vrptw-c101-sa --mode sa --iterations 500000 --storage s3
go run ./cmd/owp solve-vrptw --instance ../../examples/vrptw/C101.txt --run-label vrptw-c101-lahc --mode lahc --iterations 500000 --storage s3
go run ./cmd/owp solve-vrptw --instance ../../examples/vrptw/C101.txt --run-label vrptw-c101-tabu --mode tabu --iterations 500000 --storage s3
go run ./cmd/owp solve-vrptw --instance ../../examples/vrptw/C101.txt --run-label vrptw-c101-portfolio --mode portfolio --iterations 500000 --storage s3
go run ./cmd/owp solve-vrptw --instance ../../examples/vrptw/C101.txt --run-label vrptw-c101-adaptive --mode adaptive --iterations 500000 --storage s3
```

**Total: 5 runs**

---

## NRP — Nurse Rostering

**Instances:** n012w8 (12 nurses, 8 weeks), n030w4 (30 nurses, 4 weeks)

NRP uses the tuning grid by default. For benchmark purposes, use single-config mode with best known parameters.

```bash
# n012w8 — SA (best config from tuning: iter=60K, workers=32, temp=5.0, rate=0.0009)
go run ./cmd/owp tune-pfrs --instance n012w8 --pfrs-run-label nrp-n012w8-sa --pfrs-mode sa --pfrs-iterations-per-worker 60000 --pfrs-max-total-workers 32 --pfrs-initial-temperature 5.0 --pfrs-cooling-rate 0.0009 --pfrs-storage s3

# n012w8 — LAHC
go run ./cmd/owp tune-pfrs --instance n012w8 --pfrs-run-label nrp-n012w8-lahc --pfrs-mode lahc --pfrs-iterations-per-worker 60000 --pfrs-max-total-workers 32 --pfrs-late-acceptance-length 1000 --pfrs-storage s3

# n012w8 — Portfolio (SA+LAHC+Tabu)
go run ./cmd/owp tune-pfrs --instance n012w8 --pfrs-run-label nrp-n012w8-portfolio --pfrs-mode portfolio --pfrs-portfolio sa,lahc,tabu --pfrs-iterations-per-worker 60000 --pfrs-max-total-workers 32 --pfrs-storage s3

# n012w8 — Portfolio with look-ahead
go run ./cmd/owp tune-pfrs --instance n012w8 --pfrs-run-label nrp-n012w8-portfolio-look --pfrs-mode portfolio --pfrs-portfolio sa,lahc,tabu --pfrs-iterations-per-worker 60000 --pfrs-max-total-workers 32 --pfrs-lookahead-weight 0.3 --pfrs-storage s3

# n012w8 — Portfolio with final window coupling (2 weeks)
go run ./cmd/owp tune-pfrs --instance n012w8 --pfrs-run-label nrp-n012w8-portfolio-fw2 --pfrs-mode portfolio --pfrs-portfolio sa,lahc,tabu --pfrs-iterations-per-worker 60000 --pfrs-max-total-workers 32 --pfrs-final-window-weeks 2 --pfrs-storage s3

# n012w8 — Portfolio with look-ahead + final window
go run ./cmd/owp tune-pfrs --instance n012w8 --pfrs-run-label nrp-n012w8-portfolio-look-fw2 --pfrs-mode portfolio --pfrs-portfolio sa,lahc,tabu --pfrs-iterations-per-worker 60000 --pfrs-max-total-workers 32 --pfrs-lookahead-weight 0.3 --pfrs-final-window-weeks 2 --pfrs-storage s3

# n030w4 — SA
go run ./cmd/owp tune-pfrs --instance n030w4 --pfrs-run-label nrp-n030w4-sa --pfrs-mode sa --pfrs-iterations-per-worker 100000 --pfrs-max-total-workers 32 --pfrs-initial-temperature 5.0 --pfrs-cooling-rate 0.0009 --pfrs-storage s3

# n030w4 — LAHC
go run ./cmd/owp tune-pfrs --instance n030w4 --pfrs-run-label nrp-n030w4-lahc --pfrs-mode lahc --pfrs-iterations-per-worker 100000 --pfrs-max-total-workers 32 --pfrs-late-acceptance-length 1000 --pfrs-storage s3

# n030w4 — Portfolio
go run ./cmd/owp tune-pfrs --instance n030w4 --pfrs-run-label nrp-n030w4-portfolio --pfrs-mode portfolio --pfrs-portfolio sa,lahc,tabu --pfrs-iterations-per-worker 100000 --pfrs-max-total-workers 32 --pfrs-storage s3

# n030w4 — Portfolio with look-ahead
go run ./cmd/owp tune-pfrs --instance n030w4 --pfrs-run-label nrp-n030w4-portfolio-look --pfrs-mode portfolio --pfrs-portfolio sa,lahc,tabu --pfrs-iterations-per-worker 100000 --pfrs-max-total-workers 32 --pfrs-lookahead-weight 0.3 --pfrs-storage s3
```

**Total: 10 runs**

---

## Summary

| Domain | Instances | Modes | Runs |
|--------|-----------|-------|------|
| CVRP | 4 | 5 | 20 |
| JSS | 3 | 5 | 15 |
| VRPTW | 1 | 5 | 5 |
| NRP | 2 | 5 variations | 10 |
| **Total** | **10** | — | **50** |

---

## Notes

- All CVRP/JSS/VRPTW runs use 500K iterations, seed 42, adaptive cooling.
- NRP runs use single-config mode (not tuning grid) for reproducible, comparable results.
- Portfolio runs SA+LAHC+Tabu in parallel, returns the best.
- Adaptive uses SA primary with LAHC escape bursts on stagnation detection.
- Re-running with the same label overwrites the S3 data (latest result wins).
- Results appear on the Benchmark Ladder automatically after upload.

---

## Reproduction

To reproduce the full benchmark suite from scratch:

1. `cd platform/go`
2. Run all commands above in sequence (or parallel where independent)
3. Verify on dashboard: `/benchmarks` page shows all instances × modes
4. Expected runtime: ~5 minutes total (all are sub-second per run at 500K iterations)
