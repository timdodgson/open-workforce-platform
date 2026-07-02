import { listRunsAsync, loadRunSummary } from '@/lib/data-loader';
import Card from '@/components/Card';
import ExperimentManager from './ExperimentManager';
import { RunSummary, RunMetadata } from '@/lib/types';

export const dynamic = 'force-dynamic';

export interface RunInfo {
  id: string;
  metadata: RunMetadata | null;
  totalPenalty: number;
  numWeeks: number;
  totalWorkers: number;
  totalDurationMs: number;
}

export default async function ExperimentsPage() {
  const runs = await listRunsAsync();

  const runInfos: RunInfo[] = [];
  for (const run of runs) {
    const summary = await loadRunSummary(run.id);
    runInfos.push({
      id: run.id,
      metadata: summary.metadata,
      totalPenalty: summary.totalPenalty,
      numWeeks: summary.numWeeks,
      totalWorkers: summary.totalWorkers,
      totalDurationMs: summary.totalDurationMs,
    });
  }

  return <ExperimentManager runs={runInfos} />;
}
