'use client';
import { LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts';
import Card from '@/components/Card';
import { ImprovementEvent } from '@/lib/types';

interface Props {
  events: ImprovementEvent[];
}

export default function GlobalBestTimeline({ events }: Props) {
  if (events.length === 0) {
    return (
      <Card title="Global Best Timeline">
        <div className="border-2 border-dashed border-gray-700 rounded-lg p-8 text-center text-gray-500">
          <p>No improvement timeline data available.</p>
          <p className="text-xs mt-2">Run with latest code to generate improvements.csv.</p>
        </div>
      </Card>
    );
  }

  // Build penalty-over-time series.
  const data = events.map(e => ({
    elapsedMs: e.elapsedMs,
    penalty: e.newGlobalBest,
    week: e.week,
    worker: e.workerID,
    improvement: e.improvement,
    temperature: e.temperatureAtEvent,
  }));

  const firstPenalty = events[0].oldGlobalBest;
  const finalPenalty = events[events.length - 1].newGlobalBest;
  const totalImprovement = firstPenalty - finalPenalty;
  const totalEvents = events.length;

  return (
    <Card title="Global Best Timeline">
      <ResponsiveContainer width="100%" height={220}>
        <LineChart data={data} margin={{ top: 10, right: 20, left: 10, bottom: 10 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
          <XAxis
            dataKey="elapsedMs"
            stroke="#9ca3af"
            fontSize={10}
            tickFormatter={(v: number) => v >= 1000 ? `${(v/1000).toFixed(0)}s` : `${v}ms`}
          />
          <YAxis stroke="#9ca3af" fontSize={10} domain={['auto', 'auto']} />
          <Tooltip
            contentStyle={{ background: '#1f2937', border: '1px solid #374151', fontSize: 11 }}
            content={({ payload }) => {
              if (!payload || payload.length === 0) return null;
              const d = payload[0].payload;
              return (
                <div className="bg-gray-800 border border-gray-600 rounded p-2 text-xs">
                  <div><span className="text-gray-400">Penalty:</span> {d.penalty}</div>
                  <div><span className="text-gray-400">Improvement:</span> -{d.improvement}</div>
                  <div><span className="text-gray-400">Week:</span> {d.week}</div>
                  <div><span className="text-gray-400">Worker:</span> {d.worker}</div>
                  <div><span className="text-gray-400">Temp:</span> {d.temperature.toFixed(4)}</div>
                  <div><span className="text-gray-400">Elapsed:</span> {d.elapsedMs}ms</div>
                </div>
              );
            }}
          />
          <Line type="stepAfter" dataKey="penalty" stroke="#34d399" strokeWidth={2} dot={{ r: 2, fill: '#34d399' }} />
        </LineChart>
      </ResponsiveContainer>
      <div className="grid grid-cols-4 gap-3 mt-3 text-xs">
        <div className="bg-gray-800 rounded px-3 py-2">
          <div className="text-[9px] text-gray-500 uppercase">Improvements</div>
          <div className="text-emerald-400 font-semibold">{totalEvents}</div>
        </div>
        <div className="bg-gray-800 rounded px-3 py-2">
          <div className="text-[9px] text-gray-500 uppercase">Total Gain</div>
          <div className="text-emerald-400 font-semibold">{totalImprovement}</div>
        </div>
        <div className="bg-gray-800 rounded px-3 py-2">
          <div className="text-[9px] text-gray-500 uppercase">Start</div>
          <div className="text-gray-300 font-semibold">{firstPenalty}</div>
        </div>
        <div className="bg-gray-800 rounded px-3 py-2">
          <div className="text-[9px] text-gray-500 uppercase">Final</div>
          <div className="text-emerald-400 font-semibold">{finalPenalty}</div>
        </div>
      </div>
    </Card>
  );
}
