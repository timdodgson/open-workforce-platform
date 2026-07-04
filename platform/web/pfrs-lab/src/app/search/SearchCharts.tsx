'use client';
import { BarChart, Bar, LineChart, Line, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid } from 'recharts';
import Card from '@/components/Card';

interface ChartData {
  week: string;
  penalty: number;
  cumulative: number;
  contribution: number;
  efficiencyPerM: number;
}

export default function SearchCharts({ data }: { data: ChartData[] }) {
  return (
    <>
      <Card title="Penalty by Week">
        <p className="text-xs text-gray-500 mb-3">
          Soft constraint violations per week. Lower bars are better. Spikes indicate weeks the algorithm found difficult — often due to carry-over constraints from previous weeks.
        </p>
        <ResponsiveContainer width="100%" height={200}>
          <BarChart data={data}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="week" stroke="#9ca3af" fontSize={11} />
            <YAxis stroke="#9ca3af" fontSize={11} />
            <Tooltip contentStyle={{ background: '#1f2937', border: '1px solid #374151' }} />
            <Bar dataKey="penalty" fill="#60a5fa" radius={[3, 3, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </Card>

      <Card title="Cumulative Penalty">
        <p className="text-xs text-gray-500 mb-3">
          Running total penalty across weeks. A flatter curve means later weeks contribute less penalty — the algorithm is handling carry-over well. Steep jumps indicate problematic weeks.
        </p>
        <ResponsiveContainer width="100%" height={200}>
          <LineChart data={data}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="week" stroke="#9ca3af" fontSize={11} />
            <YAxis stroke="#9ca3af" fontSize={11} />
            <Tooltip contentStyle={{ background: '#1f2937', border: '1px solid #374151' }} />
            <Line type="monotone" dataKey="cumulative" stroke="#34d399" strokeWidth={2} dot={{ r: 3 }} />
          </LineChart>
        </ResponsiveContainer>
      </Card>

      <Card title="Candidate Efficiency (penalty reduction per million candidates)">
        <p className="text-xs text-gray-500 mb-3">
          How much penalty reduction each million candidates achieved. More negative values mean the search was more productive that week. Near-zero values suggest diminishing returns.
        </p>
        <ResponsiveContainer width="100%" height={200}>
          <BarChart data={data}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="week" stroke="#9ca3af" fontSize={11} />
            <YAxis stroke="#9ca3af" fontSize={11} />
            <Tooltip contentStyle={{ background: '#1f2937', border: '1px solid #374151' }} />
            <Bar dataKey="efficiencyPerM" fill="#fbbf24" radius={[3, 3, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </Card>
    </>
  );
}
