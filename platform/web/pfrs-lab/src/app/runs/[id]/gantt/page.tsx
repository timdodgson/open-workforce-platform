import { loadRunMetadata } from '@/lib/data-loader';
import { getStorageProvider } from '@/lib/storage';
import Card from '@/components/Card';
import GanttChart from './GanttChart';

export const dynamic = 'force-dynamic';

interface ScheduledOp {
  JobID: number;
  OpIndex: number;
  Machine: number;
  Start: number;
  End: number;
  Duration: number;
}

interface JSSolution {
  makespan: number;
  jobs: number;
  machines: number;
  operations: ScheduledOp[];
}

export default async function GanttPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const metadata = await loadRunMetadata(id);
  const storage = getStorageProvider();
  const solContent = await storage.readFile(id, 'solution.json');

  if (!solContent) {
    return (
      <Card title="Gantt Chart">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p>No solution data available for this run.</p>
        </div>
      </Card>
    );
  }

  let solution: JSSolution;
  try {
    solution = JSON.parse(solContent);
  } catch {
    return (
      <Card title="Gantt Chart">
        <div className="text-red-400 text-sm">Failed to parse solution.json</div>
      </Card>
    );
  }

  if (!solution.operations || solution.operations.length === 0) {
    return (
      <Card title="Gantt Chart">
        <div className="text-gray-500 text-sm">No operations in solution (not a JSS run?).</div>
      </Card>
    );
  }

  const meta = metadata as unknown as Record<string, unknown>;
  const instance = String(meta?.instance || 'unknown');

  return (
    <div className="space-y-4">
      <Card title="Job Shop Schedule (Gantt Chart)">
        <div className="flex gap-4 text-xs text-gray-400 mb-4">
          <span><span className="text-gray-500">Instance:</span> {instance}</span>
          <span><span className="text-gray-500">Makespan:</span> <span className="text-emerald-400 font-semibold">{solution.makespan}</span></span>
          <span><span className="text-gray-500">Jobs:</span> {solution.jobs}</span>
          <span><span className="text-gray-500">Machines:</span> {solution.machines}</span>
          <span><span className="text-gray-500">Operations:</span> {solution.operations.length}</span>
        </div>
        <GanttChart
          operations={solution.operations}
          jobs={solution.jobs}
          machines={solution.machines}
          makespan={solution.makespan}
        />
      </Card>
    </div>
  );
}
