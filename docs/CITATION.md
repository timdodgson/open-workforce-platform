# Citation & Reproduction — PFRS Lab

How to cite PFRS Lab in academic work and reproduce published results on your own machine.

**Live entry point:** [pfrs-lab.com/reproduce](https://pfrs-lab.com/reproduce)  
**R&D operations:** [RUNBOOK.md](./RUNBOOK.md) · **Experiment notebook:** [EXPERIMENTS.md](./EXPERIMENTS.md)

---

## How to cite

### Software / platform (recommended)

```bibtex
@software{dodgson2026pfrs,
  author       = {Dodgson, Tim},
  title        = {PFRS Lab: Multi-Domain Optimisation Research Platform},
  year         = {2026},
  url          = {https://pfrs-lab.com},
  repository   = {https://github.com/timdodgson/open-workforce-platform},
  note         = {Open-source platform spanning NRP, CVRP, JSS, and VRPTW with Search Intelligence validation}
}
```

Plain text:

> Dodgson, T. (2026). *PFRS Lab: Multi-Domain Optimisation Research Platform*. https://pfrs-lab.com. Source: https://github.com/timdodgson/open-workforce-platform

### Search Intelligence validation (320-run study)

When citing the assist/adaptive statistical validation (EXP-007):

```bibtex
@misc{dodgson2026si,
  author       = {Dodgson, Tim},
  title        = {Search Intelligence Statistical Validation — 320 Multi-Seed Experiments},
  year         = {2026},
  howpublished = {Technical report, PFRS Lab},
  url          = {https://pfrs-lab.com/statistics},
  note         = {Welch t-tests across NRP, CVRP, JSS, VRPTW; full commands in docs/reports/search-intelligence-statistical-validation.md}
}
```

### SI2 policy matrix (288 val-* runs)

When citing the rules/hybrid/learned policy sweep:

```bibtex
@misc{dodgson2026si2matrix,
  author       = {Dodgson, Tim},
  title        = {Search Intelligence 2.0 Policy Validation Matrix},
  year         = {2026},
  howpublished = {Experiment catalog, PFRS Lab},
  url          = {https://pfrs-lab.com/experiment-matrix},
  note         = {288 labelled runs (val-* / val-deep-*); coverage audit in docs/reports/val-gap-audit.md}
}
```

### Foundational dissertation

The nurse rostering lineage:

```bibtex
@mastersthesis{dodgson2014nrp,
  author       = {Dodgson, Tim},
  title        = {Final Project: Nurse Rostering Optimisation},
  school       = {Manchester Metropolitan University},
  year         = {2014},
  url          = {https://github.com/timdodgson/open-workforce-platform/blob/main/inspiration/Final_Project_Tim_Dodgson.pdf}
}
```

---

## Benchmark instances (cite separately)

PFRS Lab uses standard OR literature instances. Cite the original sources when comparing gaps:

| Domain | Instance(s) in lab | Primary reference |
|--------|-------------------|-------------------|
| NRP | n012w8 (INRC-II) | Ceschia, S., Di Gaspero, L., & Schaerf, A. (2013). The INRC-II nurse rostering competition. |
| CVRP | A-n32-k5, A-n80-k10 (CVRPLIB) | Augerat et al. (1995); http://vrp.atd-lab.inf.puc-rio.br |
| JSS | la01, ft10 (OR-Library) | Taillard, E. (1993). Benchmarks for basic scheduling problems. |
| VRPTW | C101 (Solomon) | Solomon, M. M. (1987). Algorithms for the vehicle routing and scheduling problems with time window constraints. |

---

## Student reproduction path

Designed for coursework, dissertations, and seminar projects. No AWS account required for the first two levels.

### Level 1 — Browse (30 minutes)

No install. Use the public lab to understand domains and evidence.

1. Open [pfrs-lab.com/lab](https://pfrs-lab.com/lab) — read domain metrics.
2. Visit [Benchmark ladder](https://pfrs-lab.com/benchmarks) — compare SA, LAHC, Tabu, Portfolio gaps to known optima.
3. Open [Statistics](https://pfrs-lab.com/statistics) — Welch t-tests and effect sizes for Search Intelligence.
4. Read [Experiment matrix](https://pfrs-lab.com/experiment-matrix) — which flags are on/off per run variation.
5. Pick one run (e.g. `val-cvrp-a32k5-sa-hybrid-s42`) and explore `/runs/<label>/summary`.

**Deliverable idea:** One-page comparison of algorithm gaps on a single instance with screenshots from the lab.

### Level 2 — One command (2 hours)

Reproduce a single published cell locally.

**Prerequisites:** Go 1.22+, `git clone` of this repository.

```powershell
cd platform/go
go test ./... -short

go run ./cmd/owp solve cvrp `
  --instance ../../examples/cvrp/A-n32-k5.vrp `
  --mode sa --iterations 500000 `
  --policy-mode hybrid --policy-dir ../ml/policies `
  --seed 42 --run-label my-cvrp-sa-hybrid-s42 `
  --storage local
```

```powershell
cd ../web/pfrs-lab
npm install
npm run dev
```

Open `http://localhost:3000/runs/my-cvrp-sa-hybrid-s42/summary`.

Compare your `bestObjective` to the live run `val-cvrp-a32k5-sa-hybrid-s42` on pfrs-lab.com. Same seed → same result (see [PRINCIPLES.md](./PRINCIPLES.md)).

**Deliverable idea:** Reproduction report: command, hardware, objective, runtime, diff vs published run.

Full write-up: [EXPERIMENTS.md § EXP-008](./EXPERIMENTS.md).

### Level 3 — Algorithm comparison (half day)

Reproduce the benchmark ladder for one domain.

```powershell
cd platform/go
go run ./cmd/owp solve cvrp `
  --instance ../../examples/cvrp/A-n32-k5.vrp `
  --mode sa --iterations 500000 --seed 42 `
  --worker-decision-mode off `
  --run-label student-cvrp-a32k5-sa-s42 --storage local
```

Repeat for `lahc`, `tabu`, `portfolio`. Commands for all domains: [benchmark-suite.md](./06-engineering/benchmark-suite.md).

**Deliverable idea:** Table of modes × objective × gap-to-optimal with brief interpretation.

### Level 4 — Statistical replication (multi-day)

Replicate EXP-007 (assist modes) or the SI2 matrix subset.

| Study | Runs | Script / doc |
|-------|------|--------------|
| Assist validation (320) | 10 seeds × 4 modes × 8 configs | [search-intelligence-statistical-validation.md](./reports/search-intelligence-statistical-validation.md) |
| SI2 policy matrix (288) | 3 policies × seeds × 16 configs | [RUNBOOK.md](./RUNBOOK.md), `validate-si2.ps1` |

Start with smoke: `platform/go/scripts/validate-si2-quick.ps1` (6 runs).

**Deliverable idea:** Replicate one domain's Welch comparison; discuss compute savings vs quality trade-offs.

---

## Reproducing specific published claims

| Claim | Evidence URL | Reproduce from |
|-------|--------------|----------------|
| CVRP assist saves 60–73% compute, identical quality | `/statistics` | EXP-007 commands in statistical validation report |
| JSS Tabu reaches la01 optimal (666) | `/benchmarks` | `solve-jobshop` la01 tabu, 100K iters |
| VRPTW adaptive improves quality (p < 0.001) | `/statistics` | VRPTW SA adaptive, seed sweep |
| SI2 matrix complete on production | `/experiment-matrix` | `npm run audit-val-matrix` (S3) |
| Portfolio beam NRP best known | `/runs` (EXP-001 labels) | `tune-pfrs` portfolio on n012w8 |

---

## Data availability

| Artifact | Location |
|----------|----------|
| Run telemetry | `platform/web/pfrs-lab/data/runs/<label>/` (local) or S3 bucket `pfrs-research-lab-data` |
| Policy JSON | `platform/ml/policies/` |
| Precomputed stats | Live `/statistics`, `/intelligence` |
| Source code | https://github.com/timdodgson/open-workforce-platform |

Run metadata schema: [run-metadata-schema.md](./05-design/run-metadata-schema.md).

---

## Contact & attribution

- **Author:** Tim Dodgson — [LinkedIn](https://www.linkedin.com/in/tim-dodgson/), [GitHub](https://github.com/timdodgson)
- **License:** See repository `LICENSE`
- **Issues / questions:** GitHub Issues on the repository

If you use PFRS Lab in teaching or publication, a link to https://pfrs-lab.com helps others find the live evidence chain.
