'use client';
import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';

interface ProgressPoint {
  elapsed: number;
  incumbent: number | null;
  bound: number | null;
  gap: number;
  nodes: number;
  lpIters: number;
}

export default function ILPProgress({ data }: { data: ProgressPoint[] }) {
  // Format elapsed time for X axis.
  const formatTime = (seconds: number) => {
    if (seconds < 60) return `${seconds.toFixed(0)}s`;
    if (seconds < 3600) return `${(seconds / 60).toFixed(1)}m`;
    return `${(seconds / 3600).toFixed(1)}h`;
  };

  return (
    <div className="space-y-6">
      {/* Incumbent vs Bound over time */}
      <div>
        <h4 className="text-xs text-gray-400 mb-2">Incumbent & Bound Convergence</h4>
        <ResponsiveContainer width="100%" height={300}>
          <LineChart data={data} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis
              dataKey="elapsed"
              tickFormatter={formatTime}
              stroke="#6B7280"
              fontSize={10}
            />
            <YAxis stroke="#6B7280" fontSize={10} />
            <Tooltip
              contentStyle={{ backgroundColor: '#1F2937', border: '1px solid #374151' }}
              labelFormatter={(v) => `Time: ${formatTime(v as number)}`}
            />
            <Legend />
            <Line
              type="stepAfter"
              dataKey="incumbent"
              stroke="#EF4444"
              strokeWidth={2}
              dot={false}
              name="Incumbent (Upper)"
              connectNulls
            />
            <Line
              type="stepAfter"
              dataKey="bound"
              stroke="#3B82F6"
              strokeWidth={2}
              dot={false}
              name="Bound (Lower)"
              connectNulls
            />
          </LineChart>
        </ResponsiveContainer>
      </div>

      {/* Gap over time */}
      <div>
        <h4 className="text-xs text-gray-400 mb-2">Optimality Gap</h4>
        <ResponsiveContainer width="100%" height={200}>
          <LineChart data={data} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis
              dataKey="elapsed"
              tickFormatter={formatTime}
              stroke="#6B7280"
              fontSize={10}
            />
            <YAxis
              stroke="#6B7280"
              fontSize={10}
              unit="%"
            />
            <Tooltip
              contentStyle={{ backgroundColor: '#1F2937', border: '1px solid #374151' }}
              labelFormatter={(v) => `Time: ${formatTime(v as number)}`}
              formatter={(v: number) => [`${v.toFixed(2)}%`, 'Gap']}
            />
            <Line
              type="monotone"
              dataKey="gap"
              stroke="#10B981"
              strokeWidth={2}
              dot={false}
              name="Gap %"
            />
          </LineChart>
        </ResponsiveContainer>
      </div>

      {/* Nodes explored over time */}
      <div>
        <h4 className="text-xs text-gray-400 mb-2">Nodes Explored</h4>
        <ResponsiveContainer width="100%" height={200}>
          <LineChart data={data} margin={{ top: 5, right: 30, left: 20, bottom: 5 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis
              dataKey="elapsed"
              tickFormatter={formatTime}
              stroke="#6B7280"
              fontSize={10}
            />
            <YAxis stroke="#6B7280" fontSize={10} />
            <Tooltip
              contentStyle={{ backgroundColor: '#1F2937', border: '1px solid #374151' }}
              labelFormatter={(v) => `Time: ${formatTime(v as number)}`}
            />
            <Line
              type="monotone"
              dataKey="nodes"
              stroke="#F59E0B"
              strokeWidth={1.5}
              dot={false}
              name="Nodes"
            />
          </LineChart>
        </ResponsiveContainer>
      </div>

      {/* Summary stats */}
      <div className="grid grid-cols-3 gap-4 text-center">
        <div>
          <p className="text-[10px] uppercase text-gray-500">Data Points</p>
          <p className="text-sm font-bold text-white">{data.length}</p>
        </div>
        <div>
          <p className="text-[10px] uppercase text-gray-500">Final Nodes</p>
          <p className="text-sm font-bold text-white">{data[data.length - 1]?.nodes.toLocaleString() ?? '—'}</p>
        </div>
        <div>
          <p className="text-[10px] uppercase text-gray-500">Final LP Iterations</p>
          <p className="text-sm font-bold text-white">{data[data.length - 1]?.lpIters.toLocaleString() ?? '—'}</p>
        </div>
      </div>
    </div>
  );
}
