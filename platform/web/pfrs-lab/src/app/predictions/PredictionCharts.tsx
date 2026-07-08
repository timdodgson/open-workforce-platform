'use client';

import { useMemo } from 'react';
import { ScatterChart, Scatter, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer, BarChart, Bar } from 'recharts';
import Card from '@/components/Card';
import { WorkerPrediction } from './page.types';

interface Props {
  predictions: WorkerPrediction[];
}

export default function PredictionCharts({ predictions }: Props) {
  // Predicted vs Actual improvement scatter.
  const scatterData = useMemo(() => {
    return predictions.map(p => ({
      actual: p.actual.improvement_amount,
      predicted: p.predicted.expected_improvement,
    }));
  }, [predictions]);

  // Probability buckets vs actual success rate.
  const bucketData = useMemo(() => {
    const buckets = [
      { label: '0–20%', min: 0, max: 0.2 },
      { label: '20–40%', min: 0.2, max: 0.4 },
      { label: '40–60%', min: 0.4, max: 0.6 },
      { label: '60–80%', min: 0.6, max: 0.8 },
      { label: '80–100%', min: 0.8, max: 1.01 },
    ];
    return buckets.map(b => {
      const inBucket = predictions.filter(p => p.predicted.p_improved >= b.min && p.predicted.p_improved < b.max);
      const actuallyImproved = inBucket.filter(p => p.actual.improved).length;
      const rate = inBucket.length > 0 ? (actuallyImproved / inBucket.length) * 100 : 0;
      return { name: b.label, rate: parseFloat(rate.toFixed(1)), count: inBucket.length };
    });
  }, [predictions]);

  // Error histogram.
  const errorHistogram = useMemo(() => {
    const bins = [
      { label: '<-1000', min: -Infinity, max: -1000 },
      { label: '-1000–-500', min: -1000, max: -500 },
      { label: '-500–-100', min: -500, max: -100 },
      { label: '-100–0', min: -100, max: 0 },
      { label: '0–100', min: 0, max: 100 },
      { label: '100–500', min: 100, max: 500 },
      { label: '500–1000', min: 500, max: 1000 },
      { label: '>1000', min: 1000, max: Infinity },
    ];
    return bins.map(b => {
      const count = predictions.filter(p => p.error.improvement >= b.min && p.error.improvement < b.max).length;
      return { name: b.label, count };
    });
  }, [predictions]);

  // Top over-predicted (model expected much more improvement than actual).
  const overPredicted = useMemo(() => {
    return [...predictions]
      .sort((a, b) => b.error.improvement - a.error.improvement)
      .slice(0, 5)
      .map(p => ({
        label: `#${p.index} (${p.instance.slice(0, 12)})`,
        error: p.error.improvement,
        predicted: p.predicted.expected_improvement,
        actual: p.actual.improvement_amount,
      }));
  }, [predictions]);

  // Top under-predicted (model expected less improvement than actual).
  const underPredicted = useMemo(() => {
    return [...predictions]
      .sort((a, b) => a.error.improvement - b.error.improvement)
      .slice(0, 5)
      .map(p => ({
        label: `#${p.index} (${p.instance.slice(0, 12)})`,
        error: p.error.improvement,
        predicted: p.predicted.expected_improvement,
        actual: p.actual.improvement_amount,
      }));
  }, [predictions]);

  return (
    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
      {/* Predicted vs Actual Scatter */}
      <Card title="Predicted vs Actual Improvement">
        <p className="text-[10px] text-gray-500 mb-2">
          Points on the diagonal indicate perfect predictions.
        </p>
        <div className="h-48">
          <ResponsiveContainer width="100%" height="100%">
            <ScatterChart margin={{ top: 5, right: 10, bottom: 20, left: 20 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
              <XAxis
                dataKey="actual" type="number" name="Actual"
                tick={{ fontSize: 9, fill: '#9ca3af' }}
                label={{ value: 'Actual', position: 'bottom', fontSize: 9, fill: '#6b7280' }}
              />
              <YAxis
                dataKey="predicted" type="number" name="Predicted"
                tick={{ fontSize: 9, fill: '#9ca3af' }}
                label={{ value: 'Predicted', angle: -90, position: 'left', fontSize: 9, fill: '#6b7280' }}
              />
              <Tooltip
                contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '6px' }}
                labelStyle={{ fontSize: 9 }}
                itemStyle={{ fontSize: 10 }}
              />
              <Scatter data={scatterData} fill="#34d399" opacity={0.5} />
            </ScatterChart>
          </ResponsiveContainer>
        </div>
      </Card>

      {/* Probability Calibration */}
      <Card title="Prediction Calibration">
        <p className="text-[10px] text-gray-500 mb-2">
          If calibrated, P(improved)=60% should mean ~60% actually improve.
        </p>
        <div className="h-48">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={bucketData} margin={{ top: 5, right: 10, bottom: 20, left: 10 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
              <XAxis dataKey="name" tick={{ fontSize: 9, fill: '#9ca3af' }} />
              <YAxis tick={{ fontSize: 9, fill: '#9ca3af' }} domain={[0, 100]} />
              <Tooltip
                contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '6px' }}
                itemStyle={{ fontSize: 10 }}
                formatter={(value, name) => {
                  if (name === 'rate') return [`${value}%`, 'Actual Success Rate'];
                  return [value, String(name)];
                }}
              />
              <Bar dataKey="rate" fill="#34d399" radius={[2, 2, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </Card>

      {/* Error Histogram */}
      <Card title="Prediction Error Distribution">
        <p className="text-[10px] text-gray-500 mb-2">
          Errors clustered near zero indicate good predictions.
        </p>
        <div className="h-48">
          <ResponsiveContainer width="100%" height="100%">
            <BarChart data={errorHistogram} margin={{ top: 5, right: 10, bottom: 20, left: 10 }}>
              <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
              <XAxis dataKey="name" tick={{ fontSize: 8, fill: '#9ca3af' }} angle={-20} textAnchor="end" height={40} />
              <YAxis tick={{ fontSize: 9, fill: '#9ca3af' }} />
              <Tooltip
                contentStyle={{ backgroundColor: '#1f2937', border: '1px solid #374151', borderRadius: '6px' }}
                itemStyle={{ fontSize: 10 }}
              />
              <Bar dataKey="count" fill="#fbbf24" radius={[2, 2, 0, 0]} />
            </BarChart>
          </ResponsiveContainer>
        </div>
      </Card>

      {/* Top Mis-predictions */}
      <Card title="Biggest Mis-predictions">
        <div className="grid grid-cols-2 gap-3">
          <div>
            <h5 className="text-[10px] text-red-400 uppercase font-semibold mb-2">Over-predicted</h5>
            <div className="space-y-1">
              {overPredicted.map((p, i) => (
                <div key={i} className="text-[10px] flex justify-between">
                  <span className="text-gray-400 truncate w-28">{p.label}</span>
                  <span className="text-red-400">+{p.error.toFixed(0)}</span>
                </div>
              ))}
            </div>
          </div>
          <div>
            <h5 className="text-[10px] text-blue-400 uppercase font-semibold mb-2">Under-predicted</h5>
            <div className="space-y-1">
              {underPredicted.map((p, i) => (
                <div key={i} className="text-[10px] flex justify-between">
                  <span className="text-gray-400 truncate w-28">{p.label}</span>
                  <span className="text-blue-400">{p.error.toFixed(0)}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </Card>
    </div>
  );
}
