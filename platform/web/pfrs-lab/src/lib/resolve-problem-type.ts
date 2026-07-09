/** Infer problem domain from run.json and run id (ilp-jss-ft06 → jss). */
export function resolveProblemType(
  runId: string,
  meta: { problemType?: unknown; runLabel?: unknown } | Record<string, unknown> | null | undefined,
): string {
  const fromMeta = String(meta?.problemType || '').toLowerCase();
  if (fromMeta && fromMeta !== 'nrp') return fromMeta;

  const lower = runId.toLowerCase();
  if (lower.startsWith('vrptw') || lower.includes('ilp-vrptw') || lower.includes('-vrptw-')) return 'vrptw';
  if (lower.startsWith('cvrp') || lower.includes('ilp-cvrp') || lower.includes('-cvrp-')) return 'cvrp';
  if (lower.startsWith('jss') || lower.includes('ilp-jss') || lower.includes('-jss-') || lower.includes('jobshop')) {
    return 'jss';
  }
  return fromMeta || 'nrp';
}
