# Roadmap

---

## Completed

| Item | Priority | Complexity | Dependencies |
|------|----------|------------|--------------|
| NRP domain (INRC-II format, beam search, 8-week planning) | — | High | — |
| CVRP domain (CVRPLIB format, 5 neighbourhood operators) | — | Medium | Problem interface |
| JSS domain (Taillard format, permutation encoding) | — | Medium | Problem interface |
| VRPTW domain (Solomon format, time-window penalties) | — | Medium | CVRP as reference |
| Generic search engine (SA, LAHC, Tabu) | — | High | Problem interface |
| Portfolio mode (parallel multi-strategy) | — | Medium | SA, LAHC, Tabu |
| Adaptive mode (SA + LAHC escape bursts) | — | Medium | SA, LAHC |
| ILP benchmarks (HiGHS, NRP + CVRP formulations) | — | High | HiGHS binary |
| Dashboard (Next.js, ECS deployment) | — | High | S3 storage |
| Benchmark Ladder (all domains, gap-to-optimal) | — | Medium | Dashboard, S3 |
| Statistical comparison (Welch's t-test, box plots) | — | Medium | Dashboard |
| JSS Gantt chart visualisation | — | Low | JSS domain |
| Shared S3 upload function | — | Low | S3 storage |
| run.json schema standardisation (bestObjective) | — | Low | All domains |
| Admin page (schema reference, system info) | — | Low | Dashboard |
| Cypress E2E test suite | — | Medium | Dashboard |
| CI/CD pipeline (GitHub Actions, semantic-release, ECS) | — | Medium | AWS infra |
| Architecture documentation (VISION, ARCHITECTURE, ADRs) | — | Low | — |
| Benchmark suite documentation (50 reproducible commands) | — | Low | All domains |

---

## Current

| Item | Priority | Complexity | Dependencies |
|------|----------|------------|--------------|
| Full benchmark run (all 50 commands across 4 domains) | High | Low | All domains working |
| CVRP portfolio re-run (fix mode field in S3 data) | High | Low | CVRP portfolio fix deployed |
| NRP additional instances (n030w4 LAHC/Tabu) | Medium | Low | NRP S3 upload fix |
| Cypress CI stabilisation (verify green pipeline) | High | Low | Cypress installed |

---

## Next

| Item | Priority | Complexity | Dependencies |
|------|----------|------------|--------------|
| VRPTW timed route viewer (show arrival times, time windows) | Medium | Medium | VRPTW domain |
| CVRP route visualisation improvement (map-style, coordinates) | Medium | Medium | CVRP domain |
| Multi-seed benchmark runs (5 seeds per config) | High | Low | Benchmark suite |
| NRP beam search telemetry (diversity, lineage, families) for new instances | Medium | Low | NRP domain |
| Additional Solomon instances (R101, RC101) | Medium | Low | VRPTW domain |
| Additional Taillard instances (ta01-ta10) | Medium | Low | JSS domain |
| Normalise NRP CLI flags (accept --storage alongside --pfrs-storage) | Low | Low | CLI |

---

## Future

| Item | Priority | Complexity | Dependencies |
|------|----------|------------|--------------|
| Adaptive hyper-heuristic with learning (reward-based strategy selection) | High | High | Adaptive mode |
| VRPTW ILP formulation (time-indexed model) | Medium | High | HiGHS, VRPTW domain |
| Pickup & Delivery Problem (PDPTW) | Medium | High | VRPTW as base |
| Multi-Depot VRP | Medium | High | CVRP as base |
| Workforce Scheduling (generalised NRP) | Medium | High | NRP as base |
| Publication: multi-domain metaheuristic comparison | High | Medium | Full benchmarks |
| Convergence analysis tooling (iteration vs quality curves) | Medium | Medium | Telemetry |
| Algorithm parameter tuning automation (irace-style) | Medium | High | Search engine |
| Dashboard authentication (Cognito login for admin pages) | Low | Medium | AWS Cognito |
| Export to standard solution formats (CVRPLIB .sol, Solomon .sol) | Low | Medium | All domains |

---

## Far Future

| Item | Priority | Complexity | Dependencies |
|------|----------|------------|--------------|
| Commercial workforce scheduling product | Low | Very High | Platform maturity |
| Real-time re-optimisation (dynamic events during execution) | Low | Very High | Event-driven architecture |
| Cloud-hosted solver API (submit instance, get solution) | Low | High | AWS infrastructure |
| Constraint learning from historical data | Low | Very High | ML pipeline |
| Research paper: generic vs domain-specific solver comparison | Medium | Medium | Publication benchmarks |
| Research paper: adaptive portfolio learning | Medium | Medium | Adaptive with learning |
| Integration with OR-Tools / CPLEX for comparison | Low | Medium | External solvers |
| Mobile-friendly dashboard | Low | Medium | Dashboard |
| Multi-tenant SaaS deployment | Low | Very High | Authentication, isolation |

---

## Notes

- Priority is relative within each phase.
- Complexity: Low (< 1 day), Medium (1-3 days), High (1-2 weeks), Very High (> 2 weeks).
- Dependencies list what must be completed before the item can start.
- Items may move between phases as priorities shift.
