import { listRunsAsync, loadRunSummary, objectiveFromMetadata, emptyRunSummary } from '@/lib/data-loader';
import Card from '@/components/Card';
import StatisticalAnalysis from './StatisticalAnalysis';
import { RunMetadata, RunSummary } from '@/lib/types';
import type { Metadata } from 'next';

export const metadata: Metadata = {
  title: 'Statistics',
  description: 'Statistical analysis of algorithm performance. Welch t-test, box plots, and effect sizes across domains.',
};

export const dynamic = 'force-dynamic';

export interface RunEntry {
  id: string;
  metadata: RunMetadata | null;
  summary: RunSummary;
}

export default async function StatisticsPage() {
  const runs = await listRunsAsync();

  if (runs.length < 2) {
    return (
      <Card title="Statistical Analysis">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p className="mb-2">Need at least 2 runs for statistical analysis.</p>
          <p className="text-xs">Run multiple experiments with different configurations.</p>
        </div>
      </Card>
    );
  }

  const entries: RunEntry[] = await Promise.all(
    runs.map(async (run) => {
      const metadata = run.metadata;
      const meta = metadata as unknown as Record<string, unknown> | undefined;
      const mode = meta ? String(meta.mode || metadata?.mode || '') : '';
      const needsSummary = !metadata || objectiveFromMetadata(meta, mode) <= 0;
      const summary = needsSummary
        ? await loadRunSummary(run.id)
        : emptyRunSummary(metadata);
      return { id: run.id, metadata, summary };
    })
  );

  return <StatisticalAnalysis runs={entries} />;
}
