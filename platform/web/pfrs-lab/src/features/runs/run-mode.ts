import type { RunMetadata } from '@/lib/types';

export type RunMode = 'pfrs' | 'ilp' | 'cvrp' | 'jss' | 'vrptw';

/** Derive sidebar/dashboard mode from run.json metadata. */
export function deriveRunMode(meta: RunMetadata | Record<string, unknown> | null): RunMode {
  if (!meta) return 'pfrs';
  const problemType = String(meta.problemType || '').toLowerCase();
  const mode = String(meta.mode || meta.algorithm || '').toLowerCase();
  if (problemType === 'cvrp' || mode === 'cvrp') return 'cvrp';
  if (problemType === 'jss' || mode === 'jss' || mode === 'jobshop') return 'jss';
  if (problemType === 'vrptw' || mode === 'vrptw') return 'vrptw';
  if (mode === 'ilp') return 'ilp';
  return 'pfrs';
}
