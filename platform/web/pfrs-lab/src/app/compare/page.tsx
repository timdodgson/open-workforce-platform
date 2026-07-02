import { listRuns, loadRunSummary, loadTree, loadDiversity } from '@/lib/data-loader';
import Card from '@/components/Card';
import ComparePanel from './ComparePanel';
import { RunSummary, TreeNode, DiversityRecord } from '@/lib/types';

export const dynamic = 'force-dynamic';

interface RunData {
  id: string;
  summary: RunSummary;
  nodes: TreeNode[];
  diversity: DiversityRecord[];
}

export default async function ComparePage() {
  const runs = listRuns();

  if (runs.length < 2) {
    return (
      <Card title="Compare Runs">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p>Need at least 2 runs to compare.</p>
        </div>
      </Card>
    );
  }

  // Load data for all runs.
  const runData: RunData[] = [];
  for (const run of runs) {
    const [summary, nodes, diversity] = await Promise.all([
      loadRunSummary(run.id),
      loadTree(run.id),
      loadDiversity(run.id),
    ]);
    runData.push({ id: run.id, summary, nodes, diversity });
  }

  return <ComparePanel runs={runData} />;
}
