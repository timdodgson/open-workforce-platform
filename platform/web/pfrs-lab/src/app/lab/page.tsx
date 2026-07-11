import type { Metadata } from 'next';
import { listRunsAsync } from '@/lib/data-loader';
import LabHub, { type LabStats } from '@/features/lab/LabHub';

export const metadata: Metadata = {
  title: 'Research Lab',
  description: 'Live metrics, benchmarks, runs, and statistical validation for PFRS Lab.',
};

export const dynamic = 'force-dynamic';

function buildStats(runs: Awaited<ReturnType<typeof listRunsAsync>>): LabStats {
  const domains: Record<string, number> = {};
  let valRuns = 0;

  for (const run of runs) {
    const pt = (run.metadata?.problemType || 'nrp').toLowerCase();
    domains[pt] = (domains[pt] ?? 0) + 1;
    if (run.id.startsWith('val-')) valRuns++;
  }

  return { totalRuns: runs.length, valRuns, domains };
}

export default async function LabPage() {
  const runs = await listRunsAsync();
  const stats = buildStats(runs);
  const recent = runs.slice(0, 50);

  return <LabHub runs={recent} stats={stats} />;
}
