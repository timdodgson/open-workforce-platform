import { listRunsAsync, loadRunSummary, loadTree, loadDiversity, loadDiscoveries, loadWorkerLifecycles, loadPlateaus } from '@/lib/data-loader';
import Card from '@/components/Card';
import HeadToHead from './HeadToHead';
import { RunSummary, TreeNode, DiversityRecord, DiscoveryRecord, WorkerLifecycle, PlateauEvent } from '@/lib/types';

export const dynamic = 'force-dynamic';

export interface RunData {
  id: string;
  summary: RunSummary;
  nodes: TreeNode[];
  diversity: DiversityRecord[];
  discoveries: DiscoveryRecord[];
  workers: WorkerLifecycle[];
  plateaus: PlateauEvent[];
}

export default async function ComparePage() {
  const runs = await listRunsAsync();

  if (runs.length < 2) {
    return (
      <Card title="Head-to-Head">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p>Need at least 2 runs to compare.</p>
        </div>
      </Card>
    );
  }

  // Load full data for all runs.
  const runData: RunData[] = [];
  for (const run of runs) {
    const [summary, nodes, diversity, discoveries, workers, plateaus] = await Promise.all([
      loadRunSummary(run.id),
      loadTree(run.id),
      loadDiversity(run.id),
      loadDiscoveries(run.id),
      loadWorkerLifecycles(run.id),
      loadPlateaus(run.id),
    ]);
    runData.push({ id: run.id, summary, nodes, diversity, discoveries, workers, plateaus });
  }

  return <HeadToHead runs={runData} />;
}
