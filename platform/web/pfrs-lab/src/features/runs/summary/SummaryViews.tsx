import Card from '@/components/Card';
import MetricCard from '@/components/MetricCard';
import type { FeasibilitySummary } from '@/lib/feasibility-summary';
import type { RunSummary } from '@/lib/types';
import FeasibilitySummaryCard from '@/components/FeasibilitySummaryCard';
import { improvementPct, runtimeSeconds } from './utils';

export function NRPSummaryView({ id, d }: { id: string; d: RunSummary }) {
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
          <MetricCard label="Runtime" value={runtimeSeconds(d.totalDurationMs)} color="default" />
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

type RunMetaRecord = Record<string, unknown>;

export function ILPSummaryView({ id, meta }: { id: string; meta: RunMetaRecord }) {
  const status = String(meta.status || '—');
  const statusColor = status === 'FEASIBLE' || status === 'OPTIMAL' ? 'green' : 'red';
  const problemType = String(meta.problemType || 'nrp').toUpperCase();

  return (
    <div>
      <Card title={`Run: ${id}`}>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <MetricCard label="Solver" value="ILP (HiGHS)" color="blue" />
          <MetricCard label="Instance" value={String(meta.instance || '—')} color="default" />
          <MetricCard label="Status" value={status} color={statusColor} />
          <MetricCard label="Problem" value={problemType} color="default" />
        </div>
      </Card>

      <Card title="Results">
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <MetricCard label="Objective" value={Number(meta.objective || 0).toLocaleString()} color="green" />
          <MetricCard label="Lower Bound" value={Number(meta.bound || 0).toLocaleString()} color="blue" />
          <MetricCard label="Gap" value={`${Number(meta.gap || 0).toFixed(2)}%`} color="amber" />
          <MetricCard label="Runtime" value={`${(Number(meta.runtime || 0) / 60).toFixed(1)} min`} color="default" />
          {meta.timeLimit != null && (
            <MetricCard label="Time Limit" value={`${(Number(meta.timeLimit) / 60).toFixed(0)} min`} color="default" />
          )}
          {meta.vehicles != null && <MetricCard label="Vehicles" value={String(meta.vehicles)} color="default" />}
          {meta.customers != null && <MetricCard label="Customers" value={String(meta.customers)} color="default" />}
          {meta.capacity != null && <MetricCard label="Capacity" value={String(meta.capacity)} color="default" />}
        </div>
      </Card>
    </div>
  );
}

export function CVRPSummaryView({
  id,
  meta,
  totalPenalty,
  totalDurationMs,
  totalCandidates,
  feasibility,
}: {
  id: string;
  meta: RunMetaRecord;
  totalPenalty: number;
  totalDurationMs: number;
  totalCandidates: number;
  feasibility: FeasibilitySummary | null;
}) {
  const mode = String(meta.mode || '—');
  const objective = Number(meta.bestDistance || totalPenalty || 0);
  const initial = Number(meta.initialDistance || 0);

  return (
    <div>
      <Card title={`Run: ${id}`}>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <MetricCard label="Algorithm" value={mode.toUpperCase()} color="blue" />
          <MetricCard label="Instance" value={String(meta.instance || '—')} color="default" />
          <MetricCard label="Customers" value={String(meta.customers || '—')} color="default" />
          <MetricCard label="Capacity" value={String(meta.capacity || '—')} color="default" />
        </div>
      </Card>

      <Card title="Results">
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <MetricCard label="Best Distance" value={objective.toLocaleString()} color="green" />
          <MetricCard label="Initial Distance" value={initial.toLocaleString()} color="default" />
          <MetricCard label="Improvement" value={`${improvementPct(initial, objective)}%`} color="green" />
          <MetricCard label="Feasible" value={meta.feasible === false ? '✗' : '✓'} color={meta.feasible === false ? 'red' : 'green'} />
          <MetricCard label="Runtime" value={runtimeSeconds(Number(meta.runtimeMs || totalDurationMs))} color="default" />
          <MetricCard label="Candidates" value={totalCandidates > 0 ? `${(totalCandidates / 1000).toFixed(0)}K` : '—'} color="default" />
          <MetricCard label="Iterations" value={meta.iterations ? `${(Number(meta.iterations) / 1000).toFixed(0)}K` : '—'} color="default" />
          <MetricCard label="Seed" value={String(meta.seed || '—')} color="default" />
        </div>
      </Card>

      {feasibility && <FeasibilitySummaryCard summary={feasibility} />}
    </div>
  );
}

export function JSSSummaryView({
  id,
  meta,
  totalPenalty,
  totalDurationMs,
  feasibility,
}: {
  id: string;
  meta: RunMetaRecord;
  totalPenalty: number;
  totalDurationMs: number;
  feasibility: FeasibilitySummary | null;
}) {
  const mode = String(meta.mode || '—');
  const makespan = Number(meta.bestMakespan || totalPenalty || 0);
  const initial = Number(meta.initialMakespan || 0);

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
          <MetricCard label="Improvement" value={`${improvementPct(initial, makespan)}%`} color="green" />
          <MetricCard label="Runtime" value={runtimeSeconds(Number(meta.runtimeMs || totalDurationMs))} color="default" />
          <MetricCard label="Iterations" value={meta.iterations ? `${(Number(meta.iterations) / 1000).toFixed(0)}K` : '—'} color="default" />
          <MetricCard label="Seed" value={String(meta.seed || '—')} color="default" />
        </div>
      </Card>

      {feasibility && <FeasibilitySummaryCard summary={feasibility} />}
    </div>
  );
}

export function VRPTWSummaryView({
  id,
  meta,
  feasibility,
}: {
  id: string;
  meta: RunMetaRecord;
  feasibility: FeasibilitySummary | null;
}) {
  const mode = String(meta.mode || '—');
  const distance = Number(meta.bestDistance || meta.bestObjective || 0);
  const initial = Number(meta.initialDistance || 0);

  return (
    <div>
      <Card title={`Run: ${id}`}>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <MetricCard label="Algorithm" value={mode.toUpperCase()} color="blue" />
          <MetricCard label="Instance" value={String(meta.instance || '—')} color="default" />
          <MetricCard label="Customers" value={String(meta.customers || '—')} color="default" />
          <MetricCard label="Capacity" value={String(meta.capacity || '—')} color="default" />
        </div>
      </Card>

      <Card title="Results">
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <MetricCard label="Best Distance" value={distance.toLocaleString()} color="green" />
          <MetricCard label="Initial Distance" value={initial.toLocaleString()} color="default" />
          <MetricCard label="Improvement" value={`${improvementPct(initial, distance)}%`} color="green" />
          <MetricCard label="Feasible" value={meta.feasible === false ? '✗' : '✓'} color={meta.feasible === false ? 'red' : 'green'} />
          <MetricCard label="Vehicles Used" value={String(meta.bestVehicles || '—')} color="default" />
          <MetricCard label="Max Vehicles" value={String(meta.vehicles || '—')} color="default" />
          <MetricCard label="Runtime" value={runtimeSeconds(Number(meta.runtimeMs))} color="default" />
          <MetricCard label="Seed" value={String(meta.seed || '—')} color="default" />
        </div>
      </Card>

      {feasibility && <FeasibilitySummaryCard summary={feasibility} />}
    </div>
  );
}
