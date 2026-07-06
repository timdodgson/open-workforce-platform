'use client';

import { useMemo } from 'react';
import {
  LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer,
  BarChart, Bar, PieChart, Pie, Cell, Legend,
} from 'recharts';
import Card from '@/components/Card';

interface SimResult {
  totalWorkers: number;
  workersRun: number;
  workersSkipped: number;
  workersReduced: number;
  workersIncreased: number;
  cpuSavedPct: number;
  globalBestsMissed: number;
  totalGlobalBests: number;
  workers: { action: string }[];
}

interface ThresholdPoint {
  threshold: number;
  cpuSaved: number;
  globalBestsMissed: number;
  improvement: number;
  skipped: number;
}

interface Props {
  result: SimResult;
  thresholdCurve: ThresholdPoint[];
}

const COLOURS = ['#6b7280', '#ef4444', '#f59e0b', '#34d399'];

export default function SimulationCharts({ result, thresholdCurve }: Props) {
  // Pie chart data.
  const pieData = useMemo(() => [
    { name: 'Run', value: result.workersRun },
    { name: 'Skip', value: result.workersSkipped },
    { name: 'Reduce', value: result.workersReduced },
    { name: 'Increase', value: result.workersIncreased },
  ].filter(d => d.value > 0), [result]);

  // Recommendation breakdown bar.
  const breakdownData = useMemo(() => [
    { name: 'Run', count: result.workersRun, pct: (result.workersRun / result.totalWorkers) * 100 },
    { name: 'Skip', count: result.workersSkipped, pct: (result.workersSkipped / result.totalWorkers) * 100 },
    { name: 'Reduce', count: result.workersReduced, pct: (result.workersReduced / result.totalWorkers) * 100 },
    { name: 'Increase', count: result.workersIncreased, pct: (result.workersIncreased / result.totalWorkers) * 100 },
  ], [result]);

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
      {/* CPU Saved vs Risk (Threshold Curve) */}
      <Card title="CPU Saved vs Risk">
        <p className="text-[10px] text-gray-500 mb-2">
          Lower thresholds save more CPU but risk missing global bests.
        </p>
        <div className="h-48">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={thresholdCurve} margin={{ top: 5, right: 20, bottom: 20, left: 10 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
              <XAxis
                dataKey="threshold" tick={{ fontSize: 9, fill: '#9ca3af' }}
                label={{ value: 'Confidence Threshold', position: 'bottom', fontSize: 9, fill: '#6b7280' }}
              />
              <YAxis tick={{ fontSize: 9, fill: '#9ca3af' }} />
              <Tooltip
                contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '6px' }}
                itemStyle={{ fontSize: 10 }}
              />
              <Line type="monotone" dataKey="cpuSaved" name="CPU Saved %" stroke="#34d399" strokeWidth={2} dot={{ r: 3 }} />
              <Line type="monotone" dataKey="globalBestsMissed" name="GB Missed" stroke="#ef4444" strokeWidth={2} dot={{ r: 3 }} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </Card>

      {/* Expected Improvement Retention */}
      <Card title="Improvement Retention vs Threshold">
        <p className="text-[10px] text-gray-500 mb-2">
          Percentage of total improvement retained at each threshold.
        </p>
        <div className="h-48">
          <ResponsiveContainer width="100%" height="100%">
            <LineChart data={thresholdCurve} margin={{ top: 5, right: 20, bottom: 20, left: 10 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
              <XAxis
                dataKey="threshold" tick={{ fontSize: 9, fill: '#9ca3af' }}
                label={{ value: 'Confidence Threshold', position: 'bottom', fontSize: 9, fill: '#6b7280' }}
              />
              <YAxis tick={{ fontSize: 9, fill: '#9ca3af' }} domain={[0, 100]} />
              <Tooltip
                contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '6px' }}
                itemStyle={{ fontSize: 10 }}
                formatter={(value) => [`${value}%`, 'Retained']}
              />
              <Line type="monotone" dataKey="improvement" name="Improvement %" stroke="#fbbf24" strokeWidth={2} dot={{ r: 3 }} />
            </LineChart>
          </ResponsiveContainer>
        </div>
      </Card>

      {/* Budget Allocation Pie */}
      <Card title="Budget Allocation">
        <p className="text-[10px] text-gray-500 mb-2">
          How workers would be allocated under current settings.
        </p>
        <div className="h-48">
          <ResponsiveContainer width="100%" height="100%">
            <PieChart>
              <Pie
                data={pieData}
                cx="50%"
                cy="50%"
                outerRadius={70}
                dataKey="value"
                label={({ name, percent }) => `${name} ${((percent ?? 0) * 100).toFixed(0)}%`}
                labelLine={false}
              >
                {pieData.map((_, i) => (
                  <Cell key={i} fill={COLOURS[i % COLOURS.length]} />
                ))}
              </Pie>
              <Tooltip
                contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '6px' }}
                itemStyle={{ fontSize: 10 }}
              />
              <Legend wrapperStyle={{ fontSize: 10 }} />
            </PieChart>
          </ResponsiveContainer>
        </div>
      </Card>

      {/* Recommendation Breakdown */}
      <Card title="Recommendation Breakdown">
        <p className="text-[10px] text-gray-500 mb-2">
          Count and percentage of each action.
        </p>
        <div className="h-48">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={breakdownData} margin={{ top: 5, right: 10, bottom: 20, left: 10 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
              <XAxis dataKey="name" tick={{ fontSize: 10, fill: '#9ca3af' }} />
              <YAxis tick={{ fontSize: 9, fill: '#9ca3af' }} />
              <Tooltip
                contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '6px' }}
                itemStyle={{ fontSize: 10 }}
                formatter={(value, name) => {
                  if (name === 'count') return [value, 'Workers'];
                  return [`${(value as number).toFixed(1)}%`, 'Percentage'];
                }}
              />
              <Bar dataKey="count" fill="#60a5fa" radius={[2, 2, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </Card>
    </div>
  );
}
