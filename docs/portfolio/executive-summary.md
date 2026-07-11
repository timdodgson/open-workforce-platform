# PFRS Lab — Executive Summary

**Tim Dodgson** · Principal Software Engineer · [pfrs-lab.com](https://pfrs-lab.com) · [LinkedIn](https://www.linkedin.com/in/tim-dodgson/) · [GitHub](https://github.com/timdodgson/open-workforce-platform)

---

## Problem

NP-hard optimisation (rostering, routing, scheduling) is usually tackled as isolated solvers or slide-deck ML. Rarely do you get **multi-domain search**, **statistical validation**, **production telemetry**, and a **public audit trail** in one system built by the same engineer who runs it.

## Approach

PFRS Lab combines:

- **Go metaheuristic engine** — SA, LAHC, Tabu, portfolio, multi-week NRP beam search  
- **Search Intelligence** — checkpoint policies trained on real search telemetry; 40–73% compute reduction on tested configs  
- **Statistical rigour** — 320+ runs, multi-seed validation, Welch / Mann–Whitney / effect sizes in the lab UI  
- **Extensibility** — `owp-sdk` v0.1.0 BYOD registry; working TSP demo + custom **greedy** search mode  

## Evidence (measured, not claimed)

| Domain | Result |
|--------|--------|
| CVRP | Identical quality, 60–73% less compute |
| JSS | Optimal on la01 reference, ~40% compute saved |
| VRPTW | ~19% better quality (p &lt; 0.001) |
| NRP | No degradation vs baseline; 0 feasibility regressions across validation |

**Audit path:** [Open Lab](https://pfrs-lab.com/lab) → Benchmarks → Statistics → Experiment Matrix → individual runs.

## Extensibility

Third-party domains implement `searchdef.Problem`, register via `owp-sdk`, and run through `owp solve` without forking the engine. Published module: `github.com/timdodgson/open-workforce-platform/owp-sdk@v0.1.0`.

## Builder

**Tim Dodgson** — Principal Software Engineer at CDL Software (10+ years, junior → principal). Serverless architecture, IaC, CI/CD, AWS-certified. Earlier: Java/Android production apps for banking clients; diagnostic platform credited with **£1M+** business impact. BSc (First), Manchester Metropolitan University. Dissertation roots in nurse rostering (2014); PFRS Lab reunites that research with two decades of production engineering and applied ML.

---

*This document is the hiring-manager view. Technical depth: [/research](https://pfrs-lab.com/research). Live metrics: [/lab](https://pfrs-lab.com/lab).*
