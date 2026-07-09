export interface RouteStop {
  customers: number[];
  load: number;
  distance: number;
  feasible?: boolean;
}

export interface RoutingSolution {
  routes: RouteStop[];
  totalCost: number;
  vehicles: number;
  feasible: boolean;
  timeWindowViolations?: number;
}

export interface JSSScheduledOp {
  JobID: number;
  OpIndex: number;
  Machine: number;
  Start: number;
  End: number;
  Duration: number;
}

export interface JSSSolution {
  makespan: number;
  jobs: number;
  machines: number;
  operations: JSSScheduledOp[];
}

export type RoutingProblemType = 'cvrp' | 'vrptw';

export function isRoutingSolution(value: unknown): value is RoutingSolution {
  if (!value || typeof value !== 'object') return false;
  const v = value as Record<string, unknown>;
  return Array.isArray(v.routes) && typeof v.totalCost === 'number';
}

export function isJSSSolution(value: unknown): value is JSSSolution {
  if (!value || typeof value !== 'object') return false;
  const v = value as Record<string, unknown>;
  return Array.isArray(v.operations) && typeof v.makespan === 'number';
}
