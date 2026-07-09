import type { JSSScheduledOp, JSSSolution, RoutingSolution } from './solution-types';

export interface RouteConstraintRow {
  routeIndex: number;
  customers: number;
  load: number;
  capacity: number;
  utilisation: number;
  overloaded: boolean;
  distance: number;
  feasible: boolean;
  twOk?: boolean;
}

export interface JSSViolation {
  code: string;
  detail: string;
}

export interface MachineLoadRow {
  machine: number;
  operations: number;
  busyTime: number;
  utilisation: number;
}

export function buildRouteConstraintRows(
  solution: RoutingSolution,
  capacity: number,
  includeTW = false,
): RouteConstraintRow[] {
  return solution.routes.map((route, i) => {
    const utilisation = capacity > 0 ? (route.load / capacity) * 100 : 0;
    const overloaded = capacity > 0 && route.load > capacity;
    return {
      routeIndex: i + 1,
      customers: route.customers.length,
      load: route.load,
      capacity,
      utilisation,
      overloaded,
      distance: route.distance,
      feasible: !overloaded && (includeTW ? route.feasible !== false : true),
      twOk: includeTW ? route.feasible !== false : undefined,
    };
  });
}

export function validateJSSSolution(solution: JSSSolution): JSSViolation[] {
  const violations: JSSViolation[] = [];
  const lookup = new Map<string, JSSScheduledOp>();
  for (const op of solution.operations) {
    lookup.set(`${op.JobID}:${op.OpIndex}`, op);
  }

  // Precedence within each job.
  const byJob = new Map<number, JSSScheduledOp[]>();
  for (const op of solution.operations) {
    if (!byJob.has(op.JobID)) byJob.set(op.JobID, []);
    byJob.get(op.JobID)!.push(op);
  }
  for (const [jobId, ops] of byJob) {
    ops.sort((a, b) => a.OpIndex - b.OpIndex);
    for (let i = 1; i < ops.length; i++) {
      const prev = ops[i - 1];
      const curr = ops[i];
      if (curr.Start < prev.End) {
        violations.push({
          code: 'PRECEDENCE',
          detail: `Job ${jobId} op ${curr.OpIndex}: starts at ${curr.Start} before op ${prev.OpIndex} ends at ${prev.End}`,
        });
      }
    }
  }

  // Machine overlap.
  const byMachine = new Map<number, JSSScheduledOp[]>();
  for (const op of solution.operations) {
    if (!byMachine.has(op.Machine)) byMachine.set(op.Machine, []);
    byMachine.get(op.Machine)!.push(op);
  }
  for (const [machine, ops] of byMachine) {
    for (let i = 0; i < ops.length; i++) {
      for (let j = i + 1; j < ops.length; j++) {
        const a = ops[i];
        const b = ops[j];
        if (a.Start < b.End && b.Start < a.End) {
          violations.push({
            code: 'OVERLAP',
            detail: `Machine ${machine}: J${a.JobID}O${a.OpIndex} [${a.Start}-${a.End}] overlaps J${b.JobID}O${b.OpIndex} [${b.Start}-${b.End}]`,
          });
        }
      }
    }
  }

  return violations;
}

export function buildMachineLoadRows(solution: JSSSolution): MachineLoadRow[] {
  const makespan = solution.makespan || 1;
  const busy = new Map<number, number>();
  const opCount = new Map<number, number>();

  for (const op of solution.operations) {
    busy.set(op.Machine, (busy.get(op.Machine) ?? 0) + (op.End - op.Start));
    opCount.set(op.Machine, (opCount.get(op.Machine) ?? 0) + 1);
  }

  const machines = Math.max(solution.machines, ...solution.operations.map((o) => o.Machine + 1), 0);
  const rows: MachineLoadRow[] = [];
  for (let m = 0; m < machines; m++) {
    const busyTime = busy.get(m) ?? 0;
    rows.push({
      machine: m,
      operations: opCount.get(m) ?? 0,
      busyTime,
      utilisation: (busyTime / makespan) * 100,
    });
  }
  return rows;
}
