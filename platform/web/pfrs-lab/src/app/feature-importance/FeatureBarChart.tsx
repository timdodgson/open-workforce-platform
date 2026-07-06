'use client';

import { BarChart, Bar, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

interface DataPoint {
  feature: string;
  importance: number;
  label: string;
}

interface Props {
  data: DataPoint[];
  colour: string;
}

export default function FeatureBarChart({ data, colour }: Props) {
  const chartData = data.map(d => ({
    name: d.label,
    feature: d.feature,
    importance: parseFloat((d.importance * 100).toFixed(1)),
  }));

  if (chartData.length === 0) {
    return (
      <div className="text-center text-gray-500 text-xs py-8">
        No feature importance data available.
      </div>
    );
  }

  return (
    <div className="h-64">
      <ResponsiveContainer width="100%" height="100%">
        <BarChart
          data={chartData}
          layout="vertical"
          margin={{ top: 5, right: 30, bottom: 5, left: 120 }}
        >
          <CartesianGrid strokeDasharray="3 3" stroke="#374151" horizontal={false} />
          <XAxis
            type="number"
            tick={{ fontSize: 10, fill: '#9ca3af' }}
            domain={[0, 'auto']}
            tickFormatter={(v) => `${v}%`}
          />
          <YAxis
            type="category"
            dataKey="name"
            tick={{ fontSize: 10, fill: '#9ca3af' }}
            width={110}
          />
          <Tooltip
            contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '6px' }}
            labelStyle={{ color: '#9ca3af', fontSize: 10 }}
            itemStyle={{ fontSize: 11 }}
            formatter={(value) => [`${value}%`, 'Importance']}
          />
          <Bar
            dataKey="importance"
            fill={colour}
            radius={[0, 4, 4, 0]}
            maxBarSize={24}
          />
        </BarChart>
      </ResponsiveContainer>
    </div>
  );
}
