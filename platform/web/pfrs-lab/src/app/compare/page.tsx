import { listRunsAsync } from '@/lib/data-loader';
import Card from '@/components/Card';
import CompareShell from './CompareShell';

export const dynamic = 'force-dynamic';

export interface RunInfo {
  id: string;
  problemType: string;
  mode: string;
  objective: number;
}

export default async function ComparePage() {
  const runs = await listRunsAsync();

  if (runs.length < 2) {
    return (
      <Card title="Compare">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p>Need at least 2 runs to compare.</p>
          <p className="text-xs mt-2">Run multiple experiments with <code className="text-blue-400">--run-label</code> to enable comparison.</p>
        </div>
      </Card>
    );
  }

  // Extract lightweight info for the selector (no heavy data loading).
  const runInfos: RunInfo[] = runs.map(r => {
    const meta = r.metadata as unknown as Record<string, unknown>;
    return {
      id: r.id,
      problemType: String(meta?.problemType || 'nrp'),
      mode: String(meta?.mode || 'unknown'),
      objective: Number(meta?.bestObjective || meta?.bestDistance || meta?.bestMakespan || meta?.totalPenalty || 0),
    };
  }).filter(r => r.objective > 0); // Only show runs with valid objectives

  if (runInfos.length < 2) {
    return (
      <Card title="Compare">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p>Need at least 2 runs with valid objectives to compare.</p>
        </div>
      </Card>
    );
  }

  return <CompareShell runs={runInfos} />;
}
