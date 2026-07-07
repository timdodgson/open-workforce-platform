import Card from '@/components/Card';
import ExperimentManager from './ExperimentManager';
import { listRunsAsync } from '@/lib/data-loader';
import { RunMetadata } from '@/lib/types';

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

  const runInfos: RunInfo[] = runs.map(r => {
    const meta = r.metadata as unknown as Record<string, unknown>;
    return {
      id: r.id,
      metadata: r.metadata,
      totalPenalty: Number(meta?.bestObjective || meta?.bestDistance || meta?.bestMakespan || meta?.totalPenalty || 0),
      numWeeks: 0,
      totalWorkers: 0,
      totalDurationMs: Number(meta?.runtimeMs || 0),
    };
  });

  return (
    <div className="space-y-4">
      <Card title="Experiments (Experimental)">
        <p className="text-xs text-amber-400 mb-2">
          This page is a local notebook for tracking hypotheses. Data is stored in your browser only.
        </p>
        <p className="text-xs text-gray-500">
          In a future version, experiments will become the organising layer of the platform —
          runs will belong to experiments, statistics will operate on experiments, and conclusions
          will be generated automatically from experiment data.
        </p>
      </Card>
      <ExperimentManager runs={runInfos} />
    </div>
  );
}
