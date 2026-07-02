import { loadRunSummary } from '@/lib/data-loader';
import Card from '@/components/Card';
import MetricCard from '@/components/MetricCard';

export const dynamic = 'force-dynamic';

export default async function RunSummaryPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  const d = await loadRunSummary(id);

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
    </div>
  );
}
