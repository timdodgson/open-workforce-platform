import type { RunMetadata } from '@/lib/types';
import { resolveProblemType } from '@/lib/resolve-problem-type';

export type RunMode = 'pfrs' | 'ilp' | 'cvrp' | 'jss' | 'vrptw';

/** Derive sidebar/dashboard mode from run.json metadata and run id. */
export function deriveRunMode(
  meta: RunMetadata | Record<string, unknown> | null,
  runId?: string,
): RunMode {
  if (!meta && !runId) return 'pfrs';
  const m = meta ?? {};
  const problemType = resolveProblemType(runId ?? String(m.runLabel || ''), m);
  const mode = String(m.mode || m.algorithm || '').toLowerCase();
  if (problemType === 'cvrp' || mode === 'cvrp') return 'cvrp';
  if (problemType === 'jss' || mode === 'jss' || mode === 'jobshop') return 'jss';
  if (problemType === 'vrptw' || mode === 'vrptw') return 'vrptw';
  if (mode === 'ilp') return 'ilp';
  return 'pfrs';
}
