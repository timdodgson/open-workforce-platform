import { loadRunSummary, loadRunMetadata } from '@/lib/data-loader';
import { getStorageProvider } from '@/lib/storage';
import Card from '@/components/Card';
import MetricCard from '@/components/MetricCard';

export const dynamic = 'force-dynamic';

export default async function RunSummaryPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const [d, rawMeta] = await Promise.all([
    loadRunSummary(id),
    getStorageProvider().readFile(id, 'run.json'),
  ]);

  // Parse raw metadata for ILP/CVRP fields not in RunMetadata type.
  const meta = rawMeta ? JSON.parse(rawMeta) : {};
  const mode = meta.mode || d.metadata?.mode || '—';
  const problemType = meta.problemType || 'nrp';
  const isILP = mode === 'ilp';
  const isCVRP = problemType === 'cvrp';

  // ILP-specific display.
  if (isILP) {
    return (
      <div>
        <Card title={`Run: ${id}`}>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            <MetricCard label="Solver" value="ILP (HiGHS)" color="blue" />
            <MetricCard label="Instance" value={meta.instance || '—'} color="default" />
            <MetricCard label="Status" value={meta.status || '—'} color={meta.status === 'FEASIBLE' || meta.status === 'OPTIMAL' ? 'green' : 'red'} />
            <MetricCard label="Problem" value={problemType.toUpperCase()} color="default" />
          </div>
        </Card>

        <Card title="Results">
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            <MetricCard label="Objective" value={(meta.objective || 0).toLocaleString()} color="green" />
            <MetricCard label="Lower Bound" value={(meta.bound || 0).toLocaleString()} color="blue" />
            <MetricCard label="Gap" value={`${(meta.gap || 0).toFixed(2)}%`} color="amber" />
            <MetricCard label="Runtime" value={`${((meta.runtime || 0) / 60).toFixed(1)} min`} color="default" />
            {meta.timeLimit && <MetricCard label="Time Limit" value={`${(meta.timeLimit / 60).toFixed(0)} min`} color="default" />}
            {meta.vehicles && <MetricCard label="Vehicles" value={String(meta.vehicles)} color="default" />}
            {meta.customers && <MetricCard label="Customers" value={String(meta.customers)} color="default" />}
            {meta.capacity && <MetricCard label="Capacity" value={String(meta.capacity)} color="default" />}
          </div>
        </Card>
      </div>
    );
  }

  // CVRP-specific display.
  if (isCVRP && !isILP) {
    const objective = meta.bestDistance || d.totalPenalty || 0;
    const initial = meta.initialDistance || d.weeks?.[0]?.startPenalty || 0;
    const improvement = initial > 0 ? ((initial - objective) / initial * 100).toFixed(1) : '0';

    return (
      <div>
        <Card title={`Run: ${id}`}>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            <MetricCard label="Algorithm" value={mode.toUpperCase()} color="blue" />
            <MetricCard label="Instance" value={meta.instance || '—'} color="default" />
            <MetricCard label="Customers" value={String(meta.customers || '—')} color="default" />
            <MetricCard label="Capacity" value={String(meta.capacity || '—')} color="default" />
          </div>
        </Card>

        <Card title="Results">
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            <MetricCard label="Best Distance" value={objective.toLocaleString()} color="green" />
            <MetricCard label="Initial Distance" value={initial.toLocaleString()} color="default" />
            <MetricCard label="Improvement" value={`${improvement}%`} color="green" />
            <MetricCard label="Feasible" value={meta.feasible === false ? '✗' : '✓'} color={meta.feasible === false ? 'red' : 'green'} />
            <MetricCard label="Runtime" value={`${((meta.runtimeMs || d.totalDurationMs || 0) / 1000).toFixed(1)}s`} color="default" />
            <MetricCard label="Candidates" value={d.totalCandidates > 0 ? `${(d.totalCandidates / 1000).toFixed(0)}K` : '—'} color="default" />
            <MetricCard label="Iterations" value={meta.iterations ? `${(meta.iterations / 1000).toFixed(0)}K` : '—'} color="default" />
            <MetricCard label="Seed" value={String(meta.seed || '—')} color="default" />
          </div>
        </Card>
      </div>
    );
  }

  // JSS-specific display.
  if (problemType === 'jss' && !isILP) {
    const makespan = meta.bestMakespan || d.totalPenalty || 0;
    const initial = meta.initialMakespan || 0;
    const improvement = initial > 0 ? ((initial - makespan) / initial * 100).toFixed(1) : '0';

    return (
      <div>
        <Card title={`Run: ${id}`}>
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            <MetricCard label="Algorithm" value={mode.toUpperCase()} color="blue" />
            <MetricCard label="Instance" value={String(meta.instance || '—')} color="default" />
            <MetricCard label="Jobs" value={String(meta.jobs || '—')} color="default" />
            <MetricCard label="Machines" value={String(meta.machines || '—')} color="default" />
          </div>
        </Card>

        <Card title="Results">
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            <MetricCard label="Best Makespan" value={makespan.toLocaleString()} color="green" />
            <MetricCard label="Initial Makespan" value={initial.toLocaleString()} color="default" />
            <MetricCard label="Improvement" value={`${improvement}%`} color="green" />
            <MetricCard label="Runtime" value={`${((meta.runtimeMs || d.totalDurationMs || 0) / 1000).toFixed(1)}s`} color="default" />
            <MetricCard label="Iterations" value={meta.iterations ? `${(meta.iterations / 1000).toFixed(0)}K` : '—'} color="default" />
            <MetricCard label="Seed" value={String(meta.seed || '—')} color="default" />
          </div>
        </Card>
      </div>
    );
  }

  // NRP (default) display.
  return (
    <div>
      <Card title={`Run: ${id}`}>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <MetricCard label="Algorithm" value={d.metadata?.mode?.toUpperCase() || '—'} color="blue" />
          <MetricCard label="Instance" value={d.metadata?.instance || '—'} color="default" />
          <MetricCard label="Beam Width" value={String(d.metadata?.beamWidth || 0)} color="default" />
          <MetricCard label="Iterations" value={`${((d.metadata?.iterationsPerWorker || 0) / 1000).toFixed(0)}K`} color="default" />
        </div>
      </Card>

      <Card title="Results">
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <MetricCard label="Total Penalty" value={d.totalPenalty.toLocaleString()} color="green" />
          <MetricCard label="Weeks" value={String(d.numWeeks)} color="default" />
          <MetricCard label="Total Workers" value={d.totalWorkers.toLocaleString()} color="blue" />
          <MetricCard label="Total Candidates" value={`${(d.totalCandidates / 1_000_000).toFixed(1)}M`} color="default" />
          <MetricCard label="Worst Week" value={`W${d.maxWeekNum}: ${d.maxWeekPenalty.toLocaleString()}`} color="red" />
          <MetricCard label="Runtime" value={`${(d.totalDurationMs / 1000).toFixed(1)}s`} color="default" />
          <MetricCard label="Hard Reject %" value={`${d.hardRejectRate.toFixed(1)}%`} color="amber" />
          {d.metadata?.mode === 'lahc' ? (
            <MetricCard label="Accept by Late %" value={`${d.lahcAcceptByLateRate.toFixed(2)}%`} color="amber" />
          ) : (
            <MetricCard label="Accept Worse %" value={`${d.acceptWorseRate.toFixed(2)}%`} color="amber" />
          )}
        </div>
      </Card>

      {d.weeks.length > 0 && (
        <Card title="Per-Week Breakdown">
          <table className="w-full text-xs">
            <thead>
              <tr className="text-gray-500 uppercase">
                <th className="text-left p-2">Week</th>
                <th className="text-right p-2">Penalty</th>
                <th className="text-right p-2">Cumulative</th>
                <th className="text-right p-2">Workers</th>
                <th className="text-right p-2">Candidates</th>
              </tr>
            </thead>
            <tbody>
              {d.weeks.map((w, i) => (
                <tr key={w.week} className="border-t border-gray-800">
                  <td className="p-2">{w.week}</td>
                  <td className="text-right p-2">{w.finalPenalty.toLocaleString()}</td>
                  <td className="text-right p-2">{d.cumulativePenalties[i].toLocaleString()}</td>
                  <td className="text-right p-2">{w.workersStarted}</td>
                  <td className="text-right p-2">{w.candidates.toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </Card>
      )}
    </div>
  );
}
