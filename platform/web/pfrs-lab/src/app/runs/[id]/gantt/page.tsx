import { loadRunMetadata } from '@/lib/data-loader';
import { getStorageProvider } from '@/lib/storage';
import RunPageShell from '@/features/runs/RunPageShell';
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
  try {
    const metadata = await loadRunMetadata(id);
    const storage = getStorageProvider();
    const solContent = await storage.readFile(id, 'solution.json');

    if (!solContent) {
      return (
        <RunPageShell
          title="Gantt Chart"
          empty
          emptyMessage="No solution data available for this run."
        >
          {null}
        </RunPageShell>
      );
    }

    let solution: JSSolution;
    try {
      solution = JSON.parse(solContent);
    } catch {
      return (
        <RunPageShell title="Gantt Chart" error="Failed to parse solution.json">
          {null}
        </RunPageShell>
      );
    }

    if (!solution.operations || solution.operations.length === 0) {
      return (
        <RunPageShell
          title="Gantt Chart"
          empty
          emptyMessage="No operations in solution (not a JSS run?)."
        >
          {null}
        </RunPageShell>
      );
    }

    const meta = metadata as unknown as Record<string, unknown>;
    const instance = String(meta?.instance || 'unknown');

    return (
      <RunPageShell title="Gantt Chart">
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
      </RunPageShell>
    );
  } catch (err) {
    return (
      <RunPageShell title="Gantt Chart" error={String(err)}>
        {null}
      </RunPageShell>
    );
  }
}
