'use client';

import { useMemo } from 'react';
import { ScatterChart, Scatter, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, Legend } from 'recharts';
import Card from '@/components/Card';

interface DataPoint {
  confidence: number;
  correct: boolean;
  recommendation: string;
  improved: boolean;
  improvementAmount: number;
}

interface Props {
  data: DataPoint[];
}

export default function ConfidenceScatter({ data }: Props) {
  const { correctPoints, incorrectPoints } = useMemo(() => {
    const correct: { confidence: number; improvement: number }[] = [];
    const incorrect: { confidence: number; improvement: number }[] = [];

    for (const d of data) {
      const point = { confidence: d.confidence, improvement: d.improvementAmount };
      if (d.correct) {
        correct.push(point);
      } else {
        incorrect.push(point);
      }
    }
    return { correctPoints: correct, incorrectPoints: incorrect };
  }, [data]);

  // Confidence bucket accuracy for the bar-style view.
  const bucketAccuracy = useMemo(() => {
    const buckets = [
      { label: '0.0–0.3', min: 0, max: 0.3 },
      { label: '0.3–0.5', min: 0.3, max: 0.5 },
      { label: '0.5–0.7', min: 0.5, max: 0.7 },
      { label: '0.7–0.9', min: 0.7, max: 0.9 },
      { label: '0.9–1.0', min: 0.9, max: 1.01 },
    ];

    return buckets.map(b => {
      const inBucket = data.filter(d => d.confidence >= b.min && d.confidence < b.max);
      const correct = inBucket.filter(d => d.correct).length;
      const accuracy = inBucket.length > 0 ? (correct / inBucket.length) * 100 : 0;
      return { ...b, total: inBucket.length, correct, accuracy };
    });
  }, [data]);

  return (
    <Card title="Confidence vs Correctness">
      <p className="text-xs text-gray-500 mb-4">
        Higher confidence should correlate with higher accuracy. If not, the confidence scoring needs recalibration.
      </p>

      {/* Scatter plot */}
      <div className="h-64 mb-6">
        <ResponsiveContainer width="100%" height="100%">
          <ScatterChart margin={{ top: 10, right: 20, bottom: 20, left: 20 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis
              dataKey="confidence"
              type="number"
              domain={[0, 1]}
              name="Confidence"
              tick={{ fontSize: 10, fill: '#9ca3af' }}
              label={{ value: 'Confidence', position: 'bottom', fontSize: 10, fill: '#6b7280' }}
            />
            <YAxis
              dataKey="improvement"
              type="number"
              name="Improvement"
              tick={{ fontSize: 10, fill: '#9ca3af' }}
              label={{ value: 'Improvement Δ', angle: -90, position: 'left', fontSize: 10, fill: '#6b7280' }}
            />
            <Tooltip
              contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '6px' }}
              labelStyle={{ color: '#9ca3af', fontSize: 10 }}
              itemStyle={{ fontSize: 11 }}
            />
            <Legend wrapperStyle={{ fontSize: 11 }} />
            <Scatter name="Correct" data={correctPoints} fill="#34d399" opacity={0.6} />
            <Scatter name="Incorrect" data={incorrectPoints} fill="#f87171" opacity={0.6} />
          </ScatterChart>
        </ResponsiveContainer>
      </div>

      {/* Accuracy by confidence bucket */}
      <p className="text-xs text-gray-500 mb-2">Accuracy by confidence bucket:</p>
      <div className="grid grid-cols-5 gap-2">
        {bucketAccuracy.map(b => (
          <div key={b.label} className="bg-gray-800 rounded p-2 text-center">
            <div className="text-[9px] text-gray-500">{b.label}</div>
            <div className={`text-sm font-bold ${b.accuracy >= 70 ? 'text-emerald-400' : b.accuracy >= 50 ? 'text-amber-400' : 'text-red-400'}`}>
              {b.accuracy.toFixed(0)}%
            </div>
            <div className="text-[9px] text-gray-500">n={b.total}</div>
          </div>
        ))}
      </div>
    </Card>
  );
}
