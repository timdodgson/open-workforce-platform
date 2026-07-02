'use client';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid, Legend } from 'recharts';
import Card from '@/components/Card';

interface SAData {
  week: string;
  better: number;
  worse: number;
  rejectedProb: number;
  worsePct: number;
}

export default function SACharts({ data }: { data: SAData[] }) {
  return (
    <>
      <Card title="SA Acceptance by Week">
        <ResponsiveContainer width="100%" height={220}>
          <BarChart data={data}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="week" stroke="#9ca3af" fontSize={11} />
            <YAxis stroke="#9ca3af" fontSize={11} />
            <Tooltip contentStyle={{ background: '#1f2937', border: '1px solid #374151' }} />
            <Legend wrapperStyle={{ fontSize: 11 }} />
            <Bar dataKey="better" stackId="a" fill="#34d399" name="Improving" />
            <Bar dataKey="worse" stackId="a" fill="#fbbf24" name="Worse (accepted)" />
            <Bar dataKey="rejectedProb" stackId="a" fill="#f87171" name="Rejected by prob" />
          </BarChart>
        </ResponsiveContainer>
      </Card>

      <Card title="Accepted Worse % by Week">
        <ResponsiveContainer width="100%" height={180}>
          <BarChart data={data}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="week" stroke="#9ca3af" fontSize={11} />
            <YAxis stroke="#9ca3af" fontSize={11} unit="%" />
            <Tooltip contentStyle={{ background: '#1f2937', border: '1px solid #374151' }} />
            <Bar dataKey="worsePct" fill="#fbbf24" name="Worse %" radius={[3, 3, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </Card>
    </>
  );
}
