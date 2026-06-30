'use client';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid, Cell } from 'recharts';
import Card from '@/components/Card';
import { WorkerLifecycle } from '@/lib/types';

interface Props {
  workers: WorkerLifecycle[];
}

function formatMs(ms: number): string {
  if (ms >= 1000) return `${(ms / 1000).toFixed(1)}s`;
  return `${ms}ms`;
}

export default function WorkerLifecycles({ workers }: Props) {
  if (workers.length === 0) {
    return (
      <Card title="Worker Cooling Lifecycles">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p>No worker lifecycle data available.</p>
          <p className="text-xs mt-2">Run with latest code to generate workers.csv.</p>
        </div>
      </Card>
    );
  }

  // Summary stats.
  const total = workers.length;
  const avgDuration = workers.reduce((s, w) => s + w.finishTimeMs, 0) / total;
  const avgFinishCand = workers.reduce((s, w) => s + w.finishCandidate, 0) / total;
  const avgTempAtBest = workers.reduce((s, w) => s + w.temperatureAtBest, 0) / total;
  const plateauPct = (workers.filter(w => w.plateauCount > 0).length / total * 100);
  const globalBestPct = (workers.filter(w => w.producedGlobalBest).length / total * 100);
  const avgBranches = workers.reduce((s, w) => s + w.branchCount, 0) / total;

  // Gantt data: worker runtime bars (sorted by start time then worker ID).
  const ganttData = workers
    .sort((a, b) => a.startTimeMs - b.startTimeMs || a.workerID - b.workerID)
    .map(w => ({
      name: `W${w.workerID}`,
      start: w.startTimeMs,
      duration: w.finishTimeMs - w.startTimeMs,
      finish: w.finishTimeMs,
      producedBest: w.producedGlobalBest,
      plateaus: w.plateauCount,
      depth: w.depth,
      week: w.week,
    }));

  return (
    <>
      <Card title="Worker Lifecycle Summary">
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          <Metric label="Total Workers" value={String(total)} />
          <Metric label="Avg Duration" value={formatMs(avgDuration)} />
          <Metric label="Avg Finish Candidate" value={`${(avgFinishCand / 1000).toFixed(0)}K`} />
          <Metric label="Avg Temp at Best" value={avgTempAtBest.toFixed(4)} />
          <Metric label="% Plateaued" value={`${plateauPct.toFixed(0)}%`} />
          <Metric label="% Produced Global Best" value={`${globalBestPct.toFixed(1)}%`} />
          <Metric label="Avg Branches/Worker" value={avgBranches.toFixed(1)} />
          <Metric label="Max Depth" value={String(Math.max(...workers.map(w => w.depth)))} />
        </div>
      </Card>

      <Card title="Worker Runtime Gantt (first 50 workers)">
        <ResponsiveContainer width="100%" height={Math.min(400, ganttData.slice(0, 50).length * 14 + 40)}>
          <BarChart data={ganttData.slice(0, 50)} layout="vertical" margin={{ left: 40, right: 10, top: 5, bottom: 5 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" horizontal={false} />
            <XAxis type="number" stroke="#9ca3af" fontSize={9} tickFormatter={(v: number) => formatMs(v)} />
            <YAxis type="category" dataKey="name" stroke="#9ca3af" fontSize={8} width={35} />
            <Tooltip
              contentStyle={{ background: '#1f2937', border: '1px solid #374151', fontSize: 11 }}
              content={({ payload }) => {
                if (!payload || payload.length === 0) return null;
                const d = payload[0].payload;
                return (
                  <div className="bg-gray-800 border border-gray-600 rounded p-2 text-xs">
                    <div><span className="text-gray-400">Worker:</span> {d.name}</div>
                    <div><span className="text-gray-400">Week:</span> {d.week}</div>
                    <div><span className="text-gray-400">Duration:</span> {formatMs(d.duration)}</div>
                    <div><span className="text-gray-400">Depth:</span> {d.depth}</div>
                    <div><span className="text-gray-400">Plateaus:</span> {d.plateaus}</div>
                    <div><span className="text-gray-400">Global Best:</span> {d.producedBest ? 'Yes' : 'No'}</div>
                  </div>
                );
              }}
            />
            <Bar dataKey="duration" radius={[0, 3, 3, 0]}>
              {ganttData.slice(0, 50).map((entry, index) => (
                <Cell key={index} fill={entry.producedBest ? '#34d399' : entry.plateaus > 0 ? '#fbbf24' : '#60a5fa'} />
              ))}
            </Bar>
          </BarChart>
        </ResponsiveContainer>
        <div className="flex gap-4 text-[10px] text-gray-500 mt-2">
          <span><span className="inline-block w-3 h-3 rounded bg-emerald-400 mr-1"></span>Produced global best</span>
          <span><span className="inline-block w-3 h-3 rounded bg-amber-400 mr-1"></span>Plateaued</span>
          <span><span className="inline-block w-3 h-3 rounded bg-blue-400 mr-1"></span>Normal</span>
        </div>
      </Card>
    </>
  );
}

function Metric({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-gray-800 rounded px-3 py-2">
      <div className="text-[9px] text-gray-500 uppercase">{label}</div>
      <div className="text-blue-400 font-semibold text-sm">{value}</div>
    </div>
  );
}
