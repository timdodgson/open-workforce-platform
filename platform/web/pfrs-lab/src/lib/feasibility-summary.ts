import type { JSSSolution, RoutingProblemType, RoutingSolution } from './solution-types';

export interface CapacityFeasibility {
  kind: 'capacity';
  routes: number;
  totalLoad: number;
  avgUtilisation: number;
  maxUtilisation: number;
  overloadedRoutes: number;
  feasible: boolean;
}

export interface VRPTWFeasibility extends CapacityFeasibility {
  kind: 'vrptw';
  timeWindowViolations: number;
  infeasibleRoutes: number;
}

export interface JSSFeasibility {
  kind: 'jss';
  makespan: number;
  jobs: number;
  machines: number;
  operations: number;
  avgMachineUtilisation: number;
  maxMachineUtilisation: number;
  bottleneckMachine: number;
}

export type FeasibilitySummary = CapacityFeasibility | VRPTWFeasibility | JSSFeasibility;

function capacityStats(solution: RoutingSolution, capacity: number): CapacityFeasibility {
  const totalLoad = solution.routes.reduce((s, r) => s + r.load, 0);
  const utils = solution.routes.map((r) => (capacity > 0 ? (r.load / capacity) * 100 : 0));
  const overloadedRoutes = capacity > 0
    ? solution.routes.filter((r) => r.load > capacity).length
    : 0;
  const maxUtilisation = utils.length > 0 ? Math.max(...utils) : 0;
  const avgUtilisation = utils.length > 0
    ? utils.reduce((s, u) => s + u, 0) / utils.length
    : 0;

  return {
    kind: 'capacity',
    routes: solution.routes.length,
    totalLoad,
    avgUtilisation,
    maxUtilisation,
    overloadedRoutes,
    feasible: solution.feasible && overloadedRoutes === 0,
  };
}

export function buildRoutingFeasibility(
  solution: RoutingSolution,
  problemType: RoutingProblemType,
  capacity: number,
): FeasibilitySummary {
  const base = capacityStats(solution, capacity);
  if (problemType !== 'vrptw') return base;

  const twViolations = solution.timeWindowViolations ?? 0;
  const infeasibleRoutes = solution.routes.filter((r) => r.feasible === false).length;

  return {
    ...base,
    kind: 'vrptw',
    timeWindowViolations: twViolations,
    infeasibleRoutes,
    feasible: base.feasible && twViolations === 0 && infeasibleRoutes === 0,
  };
}

export function buildJSSFeasibility(solution: JSSSolution): JSSFeasibility {
  const makespan = solution.makespan || 1;
  const machineBusy = new Map<number, number>();

  for (const op of solution.operations) {
    const busy = (op.End - op.Start);
    machineBusy.set(op.Machine, (machineBusy.get(op.Machine) ?? 0) + busy);
  }

  const utils = [...machineBusy.entries()].map(([machine, busy]) => ({
    machine,
    util: (busy / makespan) * 100,
  }));

  const maxEntry = utils.reduce(
    (best, cur) => (cur.util > best.util ? cur : best),
    { machine: 0, util: 0 },
  );
  const avgMachineUtilisation = utils.length > 0
    ? utils.reduce((s, u) => s + u.util, 0) / utils.length
    : 0;

  return {
    kind: 'jss',
    makespan: solution.makespan,
    jobs: solution.jobs,
    machines: solution.machines,
    operations: solution.operations.length,
    avgMachineUtilisation,
    maxMachineUtilisation: maxEntry.util,
    bottleneckMachine: maxEntry.machine,
  };
}
