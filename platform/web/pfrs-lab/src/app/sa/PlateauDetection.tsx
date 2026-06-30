'use client';
import { ScatterChart, Scatter, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts';
import Card from '@/components/Card';
import { PlateauEvent } from '@/lib/types';

interface Props {
  events: PlateauEvent[];
}

function formatCandidate(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(0)}K`;
  return String(n);
}

export default function PlateauDetection({ events }: Props) {
  if (events.length === 0) {
    return (
      <Card title="Plateau Detection">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p>No plateau events recorded.</p>
          <p className="text-xs mt-2">Plateaus are detected after 100K candidates without local improvement, once temperature drops below 25% of initial.</p>
          <p className="text-xs mt-1">Run with the latest code to generate plateau data.</p>
        </div>
      </Card>
    );
  }

  // Summary metrics.
  const count = events.length;
  const avgTemp = events.reduce((s, e) => s + e.temperature, 0) / count;
  const avgCandidate = events.reduce((s, e) => s + e.candidate, 0) / count;
  const earliest = Math.min(...events.map(e => e.candidate));
  const latest = Math.max(...events.map(e => e.candidate));

  // Chart data: candidate vs temperature at plateau.
  const chartData = events.map(e => ({
    candidate: e.candidate,
    temperature: e.temperature,
    week: e.week,
    worker: e.workerID,
    penalty: e.currentPenalty,
    localBest: e.localBest,
    sinceImprove: e.candsSinceImprove,
    depth: e.depth,
  }));

  return (
    <Card title="Plateau Detection">
      {/* Summary metrics */}
      <div className="grid grid-cols-2 sm:grid-cols-5 gap-3 mb-4">
        <Metric label="Plateau Count" value={String(count)} />
        <Metric label="Avg Temperature" value={avgTemp.toFixed(4)} />
        <Metric label="Avg Candidate" value={formatCandidate(avgCandidate)} />
        <Metric label="Earliest" value={formatCandidate(earliest)} />
        <Metric label="Latest" value={formatCandidate(latest)} />
      </div>

      {/* Timeline scatter: candidate position vs temperature */}
      <ResponsiveContainer width="100%" height={220}>
        <ScatterChart margin={{ top: 10, right: 20, left: 10, bottom: 10 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
          <XAxis
            dataKey="candidate"
            type="number"
            stroke="#9ca3af"
            fontSize={10}
            tickFormatter={formatCandidate}
            name="Candidate"
          />
          <YAxis
            dataKey="temperature"
            type="number"
            stroke="#9ca3af"
            fontSize={10}
            name="Temperature"
            scale="log"
            domain={['auto', 'auto']}
          />
          <Tooltip
            contentStyle={{ background: '#1f2937', border: '1px solid #374151', fontSize: 11 }}
            formatter={(value: number, name: string) => {
              if (name === 'temperature') return [value.toFixed(6), 'Temperature'];
              if (name === 'candidate') return [formatCandidate(value), 'Candidate'];
              return [value, name];
            }}
            content={({ payload }) => {
              if (!payload || payload.length === 0) return null;
              const d = payload[0].payload;
              return (
                <div className="bg-gray-800 border border-gray-600 rounded p-2 text-xs">
                  <div><span className="text-gray-400">Worker:</span> {d.worker}</div>
                  <div><span className="text-gray-400">Week:</span> {d.week}</div>
                  <div><span className="text-gray-400">Candidate:</span> {formatCandidate(d.candidate)}</div>
                  <div><span className="text-gray-400">Temperature:</span> {d.temperature.toFixed(6)}</div>
                  <div><span className="text-gray-400">Penalty:</span> {d.penalty}</div>
                  <div><span className="text-gray-400">Local Best:</span> {d.localBest}</div>
                  <div><span className="text-gray-400">Since Improve:</span> {formatCandidate(d.sinceImprove)}</div>
                  <div><span className="text-gray-400">Depth:</span> {d.depth}</div>
                </div>
              );
            }}
          />
          <Scatter data={chartData} fill="#f87171" opacity={0.8} />
        </ScatterChart>
      </ResponsiveContainer>

      <p className="text-[10px] text-gray-500 mt-2">
        Each dot = one plateau event (100K candidates without local best improvement, after T ≤ 25% initial).
      </p>
    </Card>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-gray-800 rounded px-3 py-2">
      <div className="text-[9px] text-gray-500 uppercase">{label}</div>
      <div className="text-red-400 font-semibold text-sm">{value}</div>
    </div>
  );
}
