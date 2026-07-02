import { listRunsAsync, loadRunSummary, loadRunMetadata } from '@/lib/data-loader';
import Card from '@/components/Card';
import StatisticalAnalysis from './StatisticalAnalysis';
import { RunMetadata, RunSummary } from '@/lib/types';

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

  const entries: RunEntry[] = [];
  for (const run of runs) {
    const [metadata, summary] = await Promise.all([
      loadRunMetadata(run.id),
      loadRunSummary(run.id),
    ]);
    entries.push({ id: run.id, metadata, summary });
  }

  return <StatisticalAnalysis runs={entries} />;
}
