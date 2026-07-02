'use client';
import {
  LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer,
  CartesianGrid, ScatterChart, Scatter, BarChart, Bar, Legend,
} from 'recharts';
import Card from '@/components/Card';
import MetricCard from '@/components/MetricCard';
import { DiscoveryRecord } from '@/lib/types';

interface Props {
  records: DiscoveryRecord[];
}

export default function DiscoveryTimeline({ records }: Props) {
  if (records.length === 0) return null;

  // Summary metrics.
  const totalDiscoveries = records.length;
  const localBests = records.filter(r => r.eventType === 'LOCAL_BEST').length;
  const globalBests = records.filter(r => r.eventType === 'GLOBAL_BEST').length;
  const reheats = records.filter(r => r.eventType === 'REHEAT');
  const reheatCount = reheats.length;
  const avgImprovement = totalDiscoveries > 0
    ? Math.round(records.reduce((s, r) => s + r.improvement, 0) / totalDiscoveries)
    : 0;
  const largestImprovement = Math.max(...records.map(r => r.improvement));
  const avgCandsBetween = totalDiscoveries > 0
    ? Math.round(records.reduce((s, r) => s + r.candsSincePrevious, 0) / totalDiscoveries)
    : 0;
  const avgYield = totalDiscoveries > 0
    ? parseFloat((records.reduce((s, r) => s + r.improvementPer10K, 0) / totalDiscoveries).toFixed(2))
    : 0;

  // Reheat effectiveness metrics.
  const reheatsImproved = reheats.filter(r => r.postReheatImproved).length;
  const reheatsBranched = reheats.filter(r => r.postReheatSpawnedBranch).length;
  const reheatsOnWinning = reheats.filter(r => r.postReheatOnWinningLineage).length;
  const reheatImprovedPct = reheatCount > 0 ? ((reheatsImproved / reheatCount) * 100).toFixed(1) : '0';
  const reheatBranchedPct = reheatCount > 0 ? ((reheatsBranched / reheatCount) * 100).toFixed(1) : '0';
  const reheatWinningPct = reheatCount > 0 ? ((reheatsOnWinning / reheatCount) * 100).toFixed(1) : '0';

  // Chart 1: Penalty vs Candidate (stepping down line).
  // Build a running best penalty series.
  const penaltyData = (() => {
    let runningBest = records[0]?.previousBest ?? 0;
    return records.map(r => {
      runningBest = Math.min(runningBest, r.newBest);
      return {
        candidate: r.candidate,
        bestPenalty: runningBest,
        week: r.week,
        workerID: r.workerID,
        eventType: r.eventType,
        improvement: r.improvement,
        temperature: r.temperatureAtEvent,
        elapsedMs: r.elapsedMs,
        previousBest: r.previousBest,
        newBest: r.newBest,
      };
    });
  })();

  // Chart 2: Improvement size at each discovery.
  const improvementData = records.map(r => ({
    candidate: r.candidate,
    improvement: r.improvement,
    eventType: r.eventType,
    week: r.week,
    workerID: r.workerID,
  }));

  // Chart 3: Search yield (improvement per 10k candidates).
  const yieldData = records.map(r => ({
    candidate: r.candidate,
    yieldPer10K: r.improvementPer10K,
    week: r.week,
    workerID: r.workerID,
  }));

  // Custom tooltip for Chart 1.
  const PenaltyTooltip = ({ active, payload }: { active?: boolean; payload?: Array<{ payload: typeof penaltyData[0] }> }) => {
    if (!active || !payload || !payload[0]) return null;
    const d = payload[0].payload;
    return (
      <div className="bg-gray-900 border border-gray-700 rounded p-2 text-xs">
        <div className="text-gray-400">Worker {d.workerID} · Week {d.week}</div>
        <div>Beam path: {d.eventType}</div>
        <div>Temperature: {d.temperature.toFixed(2)}</div>
        <div>Old penalty: {d.previousBest.toLocaleString()}</div>
        <div>New penalty: {d.newBest.toLocaleString()}</div>
        <div className="text-emerald-400">Improvement: {d.improvement}</div>
        <div>Candidate: {d.candidate.toLocaleString()}</div>
        <div>Elapsed: {(d.elapsedMs / 1000).toFixed(1)}s</div>
      </div>
    );
  };

  return (
    <>
      {/* Summary Cards */}
      <div className="grid grid-cols-2 sm:grid-cols-4 lg:grid-cols-7 gap-3 mb-4">
        <MetricCard label="Total Discoveries" value={totalDiscoveries.toLocaleString()} color="blue" />
        <MetricCard label="Local Bests" value={localBests.toLocaleString()} color="amber" />
        <MetricCard label="Global Bests" value={globalBests.toLocaleString()} color="green" />
        <MetricCard label="Avg Improvement" value={avgImprovement.toLocaleString()} color="default" />
        <MetricCard label="Largest" value={largestImprovement.toLocaleString()} color="green" />
        <MetricCard label="Avg Cands Between" value={avgCandsBetween.toLocaleString()} color="default" />
        <MetricCard label="Avg Yield / 10K" value={avgYield.toFixed(2)} color="amber" />
      </div>

      {/* Reheat Effectiveness — only shown if reheats occurred */}
      {reheatCount > 0 && (
        <Card title="Reheat Effectiveness">
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-3">
            <MetricCard label="Reheats" value={reheatCount.toLocaleString()} color="blue" />
            <MetricCard label="Produced Improvement" value={`${reheatImprovedPct}%`} color={parseFloat(reheatImprovedPct) > 10 ? 'green' : 'red'} />
            <MetricCard label="Beat Pre-Reheat Best" value={`${reheatsImproved}/${reheatCount}`} color="amber" />
            <MetricCard label="Produced Branch" value={`${reheatBranchedPct}%`} color={parseFloat(reheatBranchedPct) > 0 ? 'green' : 'red'} />
            <MetricCard label="Beat Global" value={`${reheatsBranched}/${reheatCount}`} color="default" />
            <MetricCard label="On Winning Lineage" value={`${reheatWinningPct}%`} color={parseFloat(reheatWinningPct) > 0 ? 'green' : 'red'} />
          </div>
        </Card>
      )}

      {/* Chart 1: Penalty vs Candidate */}
      <Card title="Discovery Timeline — Penalty vs Candidate">
        <p className="text-xs text-gray-500 mb-3">
          Best penalty stepping down over search effort. Each point is a discovery.
        </p>
        <ResponsiveContainer width="100%" height={240}>
          <LineChart data={penaltyData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="candidate" stroke="#9ca3af" fontSize={11} tickFormatter={v => `${(v / 1000).toFixed(0)}k`} />
            <YAxis stroke="#9ca3af" fontSize={11} />
            <Tooltip content={<PenaltyTooltip />} />
            <Line type="stepAfter" dataKey="bestPenalty" stroke="#34d399" strokeWidth={2} dot={{ r: 3, fill: '#34d399' }} />
          </LineChart>
        </ResponsiveContainer>
      </Card>

      {/* Chart 2: Improvement Size */}
      <Card title="Improvement Size">
        <p className="text-xs text-gray-500 mb-3">
          Size of each discovery. Large jumps indicate high-value moves.
        </p>
        <ResponsiveContainer width="100%" height={200}>
          <ScatterChart>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis type="number" dataKey="candidate" stroke="#9ca3af" fontSize={11} tickFormatter={v => `${(v / 1000).toFixed(0)}k`} />
            <YAxis type="number" dataKey="improvement" stroke="#9ca3af" fontSize={11} />
            <Tooltip
              contentStyle={{ background: '#1f2937', border: '1px solid #374151' }}
              formatter={(value: number, name: string) => [value.toLocaleString(), name]}
            />
            <Legend wrapperStyle={{ fontSize: 11 }} />
            <Scatter
              data={improvementData.filter(d => d.eventType === 'LOCAL_BEST')}
              fill="#fbbf24"
              name="Local Best"
              opacity={0.7}
            />
            <Scatter
              data={improvementData.filter(d => d.eventType === 'GLOBAL_BEST')}
              fill="#34d399"
              name="Global Best"
              opacity={1}
            />
          </ScatterChart>
        </ResponsiveContainer>
      </Card>

      {/* Chart 3: Search Yield */}
      <Card title="Search Yield — Improvement per 10K Candidates">
        <p className="text-xs text-gray-500 mb-3">
          When this approaches zero, search effort stops producing value.
        </p>
        <ResponsiveContainer width="100%" height={200}>
          <LineChart data={yieldData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="candidate" stroke="#9ca3af" fontSize={11} tickFormatter={v => `${(v / 1000).toFixed(0)}k`} />
            <YAxis stroke="#9ca3af" fontSize={11} />
            <Tooltip contentStyle={{ background: '#1f2937', border: '1px solid #374151' }} />
            <Line type="monotone" dataKey="yieldPer10K" stroke="#60a5fa" strokeWidth={1.5} dot={{ r: 2, fill: '#60a5fa' }} />
          </LineChart>
        </ResponsiveContainer>
      </Card>
    </>
  );
}
