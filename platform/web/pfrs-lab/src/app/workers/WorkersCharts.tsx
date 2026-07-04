'use client';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid, Legend } from 'recharts';
import Card from '@/components/Card';

interface WorkerData { week: string; workers: number; branches: number; dropped: number; depth: number; }

export default function WorkersCharts({ data }: { data: WorkerData[] }) {
  return (
    <>
      <Card title="Workers & Branches per Week">
        <p className="text-xs text-gray-500 mb-3">
          Workers deployed per week and how many spawned branches (portfolio diversification). High branch counts indicate the algorithm found global bests frequently. Dropped workers were pruned by beam selection.
        </p>
        <ResponsiveContainer width="100%" height={200}>
          <BarChart data={data}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="week" stroke="#9ca3af" fontSize={11} />
            <YAxis stroke="#9ca3af" fontSize={11} />
            <Tooltip contentStyle={{ background: '#1f2937', border: '1px solid #374151' }} />
            <Legend wrapperStyle={{ fontSize: 11 }} />
            <Bar dataKey="workers" fill="#a78bfa" name="Workers" />
            <Bar dataKey="branches" fill="#60a5fa" name="Branches" />
            <Bar dataKey="dropped" fill="#f87171" name="Dropped" />
          </BarChart>
        </ResponsiveContainer>
      </Card>

      <Card title="Winning Branch Depth">
        <p className="text-xs text-gray-500 mb-3">
          How many branch generations deep the final winning solution comes from. Higher depth means the portfolio kept finding improvements through successive diversification — a sign the strategy is working well.
        </p>
        <ResponsiveContainer width="100%" height={160}>
          <BarChart data={data}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="week" stroke="#9ca3af" fontSize={11} />
            <YAxis stroke="#9ca3af" fontSize={11} />
            <Tooltip contentStyle={{ background: '#1f2937', border: '1px solid #374151' }} />
            <Bar dataKey="depth" fill="#34d399" name="Depth" radius={[3, 3, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </Card>
    </>
  );
}
