'use client';
import {
  ScatterChart, Scatter, XAxis, YAxis, Tooltip, ResponsiveContainer,
  CartesianGrid, BarChart, Bar, Legend, LineChart, Line,
} from 'recharts';
import Card from '@/components/Card';
import MetricCard from '@/components/MetricCard';
import { DiversityRecord } from '@/lib/types';

interface Props {
  records: DiversityRecord[];
}

export default function DiversityCharts({ records }: Props) {
  // Group by week for aggregation.
  const byWeek = new Map<number, DiversityRecord[]>();
  for (const r of records) {
    const list = byWeek.get(r.week) || [];
    list.push(r);
    byWeek.set(r.week, list);
  }
  const weeks = Array.from(byWeek.keys()).sort((a, b) => a - b);

  // Beam Spread data: one entry per week.
  const beamSpreadData = weeks.map(w => {
    const weekRecords = byWeek.get(w)!;
    return {
      week: `W${w}`,
      beamSpread: weekRecords[0]?.beamSpread ?? 0,
      retained: weekRecords.filter(r => r.retained).length,
      total: weekRecords.length,
    };
  });

  // Hamming Distance data: all paths per week.
  const hammingData = weeks.map(w => {
    const weekRecords = byWeek.get(w)!;
    const retainedRecords = weekRecords.filter(r => r.retained);
    const avgHammingBest = retainedRecords.length > 0
      ? retainedRecords.reduce((sum, r) => sum + r.hammingToBest, 0) / retainedRecords.length
      : 0;
    const avgHammingParent = retainedRecords.length > 0
      ? retainedRecords.reduce((sum, r) => sum + r.hammingToParent, 0) / retainedRecords.length
      : 0;
    const maxHammingBest = retainedRecords.reduce((max, r) => Math.max(max, r.hammingToBest), 0);
    return {
      week: `W${w}`,
      avgHammingBest: parseFloat((avgHammingBest * 100).toFixed(2)),
      avgHammingParent: parseFloat((avgHammingParent * 100).toFixed(2)),
      maxHammingBest: parseFloat((maxHammingBest * 100).toFixed(2)),
    };
  });

  // Structural Diversity: scatter of each path's Hamming distance vs penalty.
  const scatterData = records.filter(r => r.retained).map(r => ({
    hammingToBest: parseFloat((r.hammingToBest * 100).toFixed(2)),
    weekPenalty: r.weekPenalty,
    week: r.week,
    pathID: r.pathID,
    winning: r.winning,
  }));

  // Near-Duplicate Detection: count per week.
  const nearDupData = weeks.map(w => {
    const weekRecords = byWeek.get(w)!;
    const nearDups = weekRecords.filter(r => r.nearDuplicate).length;
    const total = weekRecords.length;
    return {
      week: `W${w}`,
      nearDuplicates: nearDups,
      unique: total - nearDups,
      pct: total > 0 ? parseFloat(((nearDups / total) * 100).toFixed(1)) : 0,
    };
  });

  // Summary metrics.
  const totalPaths = records.length;
  const totalNearDups = records.filter(r => r.nearDuplicate).length;
  const avgBeamSpread = beamSpreadData.length > 0
    ? Math.round(beamSpreadData.reduce((s, d) => s + d.beamSpread, 0) / beamSpreadData.length)
    : 0;
  const avgHammingAll = records.length > 0
    ? parseFloat(((records.reduce((s, r) => s + r.hammingToBest, 0) / records.length) * 100).toFixed(1))
    : 0;

  return (
    <>
      {/* Summary metrics */}
      <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-4">
        <MetricCard label="Total Paths" value={totalPaths.toLocaleString()} color="blue" />
        <MetricCard label="Avg Beam Spread" value={avgBeamSpread.toLocaleString()} color="green" />
        <MetricCard label="Avg Hamming %" value={`${avgHammingAll}%`} color="amber" />
        <MetricCard label="Near Duplicates" value={`${totalNearDups} (${totalPaths > 0 ? ((totalNearDups / totalPaths) * 100).toFixed(1) : 0}%)`} color="red" />
      </div>

      {/* Beam Spread */}
      <Card title="Beam Spread">
        <p className="text-xs text-gray-500 mb-3">
          Penalty gap between best and worst retained paths per week. Higher spread indicates more exploration diversity.
        </p>
        <ResponsiveContainer width="100%" height={200}>
          <BarChart data={beamSpreadData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="week" stroke="#9ca3af" fontSize={11} />
            <YAxis stroke="#9ca3af" fontSize={11} />
            <Tooltip contentStyle={{ background: '#1f2937', border: '1px solid #374151' }} />
            <Bar dataKey="beamSpread" fill="#34d399" name="Beam Spread" radius={[3, 3, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </Card>

      {/* Hamming Distance */}
      <Card title="Hamming Distance">
        <p className="text-xs text-gray-500 mb-3">
          Average roster distance between retained paths. Higher values mean more structurally diverse solutions.
        </p>
        <ResponsiveContainer width="100%" height={200}>
          <LineChart data={hammingData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="week" stroke="#9ca3af" fontSize={11} />
            <YAxis stroke="#9ca3af" fontSize={11} unit="%" />
            <Tooltip contentStyle={{ background: '#1f2937', border: '1px solid #374151' }} />
            <Legend wrapperStyle={{ fontSize: 11 }} />
            <Line type="monotone" dataKey="avgHammingBest" stroke="#60a5fa" name="Avg to Best" strokeWidth={2} />
            <Line type="monotone" dataKey="avgHammingParent" stroke="#fbbf24" name="Avg to Parent" strokeWidth={2} />
            <Line type="monotone" dataKey="maxHammingBest" stroke="#f87171" name="Max to Best" strokeWidth={1} strokeDasharray="4 4" />
          </LineChart>
        </ResponsiveContainer>
      </Card>

      {/* Structural Diversity */}
      <Card title="Structural Diversity">
        <p className="text-xs text-gray-500 mb-3">
          Hamming distance vs week penalty for retained paths. Clusters indicate convergence; spread indicates exploration.
        </p>
        <ResponsiveContainer width="100%" height={240}>
          <ScatterChart>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis type="number" dataKey="hammingToBest" name="Hamming %" stroke="#9ca3af" fontSize={11} unit="%" />
            <YAxis type="number" dataKey="weekPenalty" name="Week Penalty" stroke="#9ca3af" fontSize={11} />
            <Tooltip
              contentStyle={{ background: '#1f2937', border: '1px solid #374151' }}
              formatter={(value: number, name: string) => [name === 'Hamming %' ? `${value}%` : value.toLocaleString(), name]}
            />
            <Scatter
              data={scatterData.filter(d => !d.winning)}
              fill="#60a5fa"
              name="Retained"
              opacity={0.6}
            />
            <Scatter
              data={scatterData.filter(d => d.winning)}
              fill="#34d399"
              name="Winning"
              opacity={1}
            />
          </ScatterChart>
        </ResponsiveContainer>
      </Card>

      {/* Near-Duplicate Detection */}
      <Card title="Near-Duplicate Detection">
        <p className="text-xs text-gray-500 mb-3">
          Paths with Hamming distance &lt; 5% from the best path. High near-duplicate counts suggest the beam is converging too quickly.
        </p>
        <ResponsiveContainer width="100%" height={200}>
          <BarChart data={nearDupData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="week" stroke="#9ca3af" fontSize={11} />
            <YAxis stroke="#9ca3af" fontSize={11} />
            <Tooltip contentStyle={{ background: '#1f2937', border: '1px solid #374151' }} />
            <Legend wrapperStyle={{ fontSize: 11 }} />
            <Bar dataKey="nearDuplicates" stackId="a" fill="#f87171" name="Near Duplicates" />
            <Bar dataKey="unique" stackId="a" fill="#34d399" name="Unique" />
          </BarChart>
        </ResponsiveContainer>
      </Card>
    </>
  );
}
