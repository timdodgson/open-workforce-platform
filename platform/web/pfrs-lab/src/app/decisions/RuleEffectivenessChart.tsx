'use client';

import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import Card from '@/components/Card';

interface RuleStats {
  rule: string;
  total: number;
  correct: number;
  accuracy: number;
  improvedPct: number;
  avgRuntimeMs: number;
}

interface Props {
  rules: RuleStats[];
}

export default function RuleEffectivenessChart({ rules }: Props) {
  const chartData = rules.map(r => ({
    name: r.rule.length > 20 ? r.rule.slice(0, 20) + '…' : r.rule,
    fullName: r.rule,
    accuracy: parseFloat(r.accuracy.toFixed(1)),
    improvedPct: parseFloat(r.improvedPct.toFixed(1)),
    count: r.total,
  }));

  return (
    <Card title="Rule Effectiveness">
      <p className="text-xs text-gray-500 mb-4">
        Accuracy per reason code. High accuracy = rule is reliable. High improved% on a skip rule = rule is wrong.
      </p>

      {/* Chart */}
      <div className="h-56 mb-4">
        <ResponsiveContainer width="100%" height="100%">
          <BarChart data={chartData} margin={{ top: 5, right: 20, bottom: 30, left: 20 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis
              dataKey="name"
              tick={{ fontSize: 9, fill: '#9ca3af' }}
              angle={-20}
              textAnchor="end"
              height={50}
            />
            <YAxis
              tick={{ fontSize: 10, fill: '#9ca3af' }}
              domain={[0, 100]}
              label={{ value: '%', position: 'top', fontSize: 10, fill: '#6b7280' }}
            />
            <Tooltip
              contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '6px' }}
              labelStyle={{ color: '#9ca3af', fontSize: 10 }}
              itemStyle={{ fontSize: 11 }}
              formatter={(value) => [`${value}%`]}
              labelFormatter={(label) => {
                const item = chartData.find(d => d.name === String(label));
                return item ? `${item.fullName} (n=${item.count})` : String(label);
              }}
            />
            <Legend wrapperStyle={{ fontSize: 11 }} />
            <Bar dataKey="accuracy" name="Accuracy" fill="#34d399" radius={[2, 2, 0, 0]} />
            <Bar dataKey="improvedPct" name="Actually Improved %" fill="#fbbf24" radius={[2, 2, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </div>

      {/* Table */}
      <table className="w-full text-xs">
        <thead>
          <tr className="text-gray-500 uppercase">
            <th className="text-left p-2">Rule / Reason Code</th>
            <th className="text-right p-2">Triggered</th>
            <th className="text-right p-2">Correct</th>
            <th className="text-right p-2">Accuracy</th>
            <th className="text-right p-2">Improved %</th>
            <th className="text-right p-2">Avg Runtime</th>
          </tr>
        </thead>
        <tbody>
          {rules.map((r, i) => (
            <tr key={r.rule} className={`border-t border-gray-800 ${i === 0 ? 'bg-emerald-900/10' : ''}`}>
              <td className="p-2 text-blue-400 font-mono text-[10px]">{r.rule}</td>
              <td className="text-right p-2">{r.total}</td>
              <td className="text-right p-2">{r.correct}</td>
              <td className="text-right p-2">
                <span className={r.accuracy >= 70 ? 'text-emerald-400' : r.accuracy >= 50 ? 'text-amber-400' : 'text-red-400'}>
                  {r.accuracy.toFixed(1)}%
                </span>
              </td>
              <td className="text-right p-2 text-amber-400">{r.improvedPct.toFixed(1)}%</td>
              <td className="text-right p-2">{r.avgRuntimeMs.toFixed(0)}ms</td>
            </tr>
          ))}
        </tbody>
      </table>
    </Card>
  );
}
