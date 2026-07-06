# Domains

Documentation of every optimisation problem domain implemented in the platform.

---

## Nurse Rostering (NRP)

### Representation

Weekly shift assignments: each nurse is assigned a shift type (Early, Day, Late, Night, Off) for each day of the week. An 8-week planning horizon is solved sequentially using beam search, with rolling history tracking consecutive assignments.

### Constraints

**Hard (must satisfy — zero violations required):**
- Shift coverage requirements (minimum nurses per shift per day).
- Skill qualifications (nurse must hold required skill for the shift).
- Forbidden shift successions (e.g. Night → Early the next day).
- Maximum consecutive working days.

**Soft (minimise weighted penalties):**
- Preferred shifts and days off.
- Balanced workload distribution.
- Consecutive shift pattern preferences.
- Weekend working rules.

### Objective

Minimise total weighted soft constraint penalty across all weeks. Lower is better. Zero would be a perfect schedule (never achieved on realistic instances).

### Moves

Single-assignment swap: change one nurse's shift on one day to a different shift type. Hard constraints are checked immediately — invalid moves are rejected before evaluation.

### Neighbourhoods

Single operator: random nurse × random day × random shift. High rejection rate (60-70%) due to skill and succession constraints. Effective moves are a small subset of the search space.

### Telemetry

- Per-week penalty breakdown.
- Worker lifecycle (start/finish time, plateau count, best candidate).
- Branch genealogy (beam search lineage).
- Discovery timeline per week.
- Diversity metrics (near-duplicate rate, lineage entropy).

### Visualisation

- Weekly schedule grid (nurse × day matrix with shift colours).
- Per-week penalty waterfall.
- Beam search tree and genealogy.
- Worker efficiency and causality analysis.
- Publication export figures.

### Current Maturity

High. Most mature domain. Full beam search with diversity preservation, look-ahead, final-window coupling, and exhaustive telemetry. ILP benchmark available for small instances.

### Future Work

- Larger instances (n050w8, n080w8).
- Multi-objective (penalty + fairness).
- Real-time re-scheduling (day-of changes).
- Generalised workforce scheduling beyond healthcare.

---

## Capacitated Vehicle Routing (CVRP)

### Representation

Set of routes. Each route is an ordered list of customer indices visited by one vehicle. All routes start and end at the depot (implicit). Each route has a tracked load (sum of customer demands).

### Constraints

**Hard:**
- Every customer visited exactly once.
- No route exceeds vehicle capacity.

**Soft (penalised):**
- Capacity violations penalised at 1000× excess demand (allows infeasible intermediate states during search).

### Objective

Minimise total Euclidean travel distance across all routes. Rounded to integer (TSPLIB convention).

### Moves

Five neighbourhood operators with weighted random selection:

| Operator | Weight | Description |
|----------|--------|-------------|
| Relocate | 25% | Move one customer to a different position |
| Inter-swap | 20% | Swap two customers between different routes |
| Intra-swap | 15% | Swap two positions within the same route |
| 2-Opt | 15% | Reverse a segment within a route |
| Or-opt | 25% | Move a chain of 1-3 customers to a new position |

### Neighbourhoods

Capacity is validated before applying inter-route moves. Infeasible moves are rejected immediately (Valid=false). Distance is pre-computed in a matrix for O(1) edge lookups.

### Telemetry

- Discovery timeline (global best improvements over time).
- Search progress (candidates, acceptance rate).
- Solution serialisation (routes with customer IDs, loads, distances).

### Visualisation

- Route viewer (customer coordinates, route paths).
- Summary page (distance, improvement, feasibility, vehicles).
- Search progress charts.

### Current Maturity

High. Five neighbourhood operators, proven results on CVRPLIB instances (0.3-3.6% gap to optimal). ILP formulation available (MTZ subtour elimination).

### Future Work

- Incremental distance evaluation (avoid full route recalculation).
- Route viewer with actual coordinate map.
- Larger instances (Christofides, 100-200 customers).
- Heterogeneous fleet (different vehicle capacities).

---

## Job Shop Scheduling (JSS)

### Representation

Permutation of operation indices. Each position determines scheduling priority. Decoding: process operations in permutation order, scheduling each at the earliest feasible time respecting precedence (job order) and machine availability (no overlap).

### Constraints

**Hard (enforced by decoding):**
- Precedence: operations within a job execute in order.
- Machine capacity: no two operations on the same machine overlap.
- All operations must be scheduled.

### Objective

Minimise makespan — the completion time of the last operation to finish.

### Moves

Swap: exchange two positions in the permutation. Always valid (any permutation produces a feasible schedule via the greedy decoder).

### Neighbourhoods

Single operator: random position swap. 100% valid moves (no rejection). The decoder handles all constraint satisfaction — the search operates in an unconstrained permutation space.

### Telemetry

- Discovery timeline.
- Search progress.
- Solution serialisation (operations with job, machine, start, end, duration).

### Visualisation

- Gantt chart: machines × time, colour-coded by job.
- Summary page (makespan, jobs, machines, improvement).
- Search progress charts.

### Current Maturity

Medium. Functional with good results (optimal on ft06 and la01, +2.8% on ft10). Simple permutation encoding works but could be improved with more sophisticated move operators.

### Future Work

- Critical path moves (swap operations on the critical path only).
- Block moves (move sequences of operations).
- Larger Taillard instances (ta01-ta80, 15×15 to 100×20).
- Gantt chart interaction (click operation for details).

---

## Vehicle Routing with Time Windows (VRPTW)

### Representation

Same as CVRP: set of routes with ordered customer visits per vehicle. Each route additionally must satisfy time window constraints along the path.

### Constraints

**Hard:**
- Every customer visited exactly once.
- No route exceeds vehicle capacity.
- Vehicle must arrive at customer before the due date (time window).
- Vehicle must return to depot before depot closing time.

**Soft (penalised):**
- Time window violations penalised at 500 per violation.
- Capacity violations penalised at 1000× excess.

### Objective

Minimise total travel distance (same as CVRP). Time windows are hard constraints that restrict feasibility rather than contributing to the objective.

### Moves

Same five operators as CVRP (Relocate, Swap, IntraSwap, 2-Opt, Or-opt). Capacity is checked immediately. Time window feasibility is checked via penalty in Evaluate (not rejected outright — allows search through infeasible space).

### Neighbourhoods

Capacity rejection on inter-route moves. Time window violations are penalised but not rejected — this allows the search to traverse infeasible intermediate states, which is important for VRPTW where feasibility is highly constrained.

### Telemetry

- Discovery timeline.
- Solution serialisation (routes with customers, loads, distances, feasibility, time window violations).

### Visualisation

- Route viewer (shared with CVRP infrastructure).
- Summary page (distance, vehicles used, max vehicles, feasibility).
- Search progress charts.

### Current Maturity

Medium. Functional with excellent results on C101 (+0.1% gap). Solomon format parser, time-ordered constructive heuristic, full Problem interface implementation. Needs more instances and a timed route viewer.

### Future Work

- Timed route viewer (show arrival times, waiting, service at each stop).
- Additional Solomon instances (R101, RC101, C201).
- Vehicle minimisation as secondary objective.
- Waiting time minimisation.
- Service time windows visualisation.
