'use client';
import { ComposedChart, Line, Scatter, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid, Legend, ReferenceLine } from 'recharts';
import Card from '@/components/Card';
import { BranchEvent } from '@/lib/types';

interface TemperatureCurveProps {
  initialTemperature: number;
  effectiveCoolingRate: number;
  iterationsPerWorker: number;
  coolingMode: string;
  minTemperature: number;
  branches?: BranchEvent[];
  // Future overlays: plateaus, local-best improvements, acceptance-rate samples
}

function generateTheoreticalCurve(
  initialTemp: number,
  rate: number,
  iterations: number,
  minTemp: number
): { candidate: number; temperature: number }[] {
  // Sample ~200 points for smooth curve without overwhelming the DOM.
  const points: { candidate: number; temperature: number }[] = [];
  const step = Math.max(1, Math.floor(iterations / 200));

  for (let i = 0; i <= iterations; i += step) {
    let temp = initialTemp * Math.pow(1 - rate, i);
    if (temp < minTemp) temp = minTemp;
    points.push({ candidate: i, temperature: temp });
  }

  // Ensure final point is included.
  if (points[points.length - 1].candidate !== iterations) {
    let temp = initialTemp * Math.pow(1 - rate, iterations);
    if (temp < minTemp) temp = minTemp;
    points.push({ candidate: iterations, temperature: temp });
  }

  return points;
}

function formatCandidate(n: number): string {
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`;
  if (n >= 1_000) return `${(n / 1_000).toFixed(0)}K`;
  return String(n);
}

export default function TemperatureCurve({
  initialTemperature,
  effectiveCoolingRate,
  iterationsPerWorker,
  coolingMode,
  minTemperature,
  branches = [],
}: TemperatureCurveProps) {
  const data = generateTheoreticalCurve(
    initialTemperature,
    effectiveCoolingRate,
    iterationsPerWorker,
    minTemperature
  );

  // Calculate temperature at each branch event using the same cooling formula.
  const branchPoints = branches.map(b => ({
    candidate: b.candidate,
    temperature: Math.max(minTemperature, initialTemperature * Math.pow(1 - effectiveCoolingRate, b.candidate)),
    workerID: b.workerID,
    week: b.week,
    oldPenalty: b.oldPenalty,
    newPenalty: b.newPenalty,
    improvement: b.improvement,
  }));

  // Merge branch points into data for the composed chart.
  const mergedData = data.map(d => ({ ...d, branch: undefined as number | undefined }));
  for (const bp of branchPoints) {
    // Find nearest data point or insert.
    let closest = 0;
    let minDist = Infinity;
    for (let i = 0; i < mergedData.length; i++) {
      const dist = Math.abs(mergedData[i].candidate - bp.candidate);
      if (dist < minDist) { minDist = dist; closest = i; }
    }
    // If close enough, tag it; otherwise we'll use scatter overlay.
    if (minDist < iterationsPerWorker / 100) {
      mergedData[closest] = { ...mergedData[closest], branch: bp.temperature };
    }
  }

  const finalTemp = data[data.length - 1].temperature;

  return (
    <Card title="Temperature Curve">
      <div className="flex items-center gap-3 mb-3">
        <span className="text-[10px] uppercase tracking-wider text-gray-500 bg-gray-800 px-2 py-0.5 rounded">
          Source: ✓ Theoretical
        </span>
        <span className="text-[10px] text-gray-500">
          {coolingMode} cooling · rate {effectiveCoolingRate.toFixed(10)}
        </span>
      </div>

      <ResponsiveContainer width="100%" height={280}>
        <ComposedChart data={mergedData} margin={{ top: 10, right: 30, left: 10, bottom: 10 }}>
          <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
          <XAxis
            dataKey="candidate"
            stroke="#9ca3af"
            fontSize={10}
            tickFormatter={formatCandidate}
            label={{ value: 'Candidate', position: 'insideBottom', offset: -5, fill: '#9ca3af', fontSize: 10 }}
          />
          <YAxis
            stroke="#9ca3af"
            fontSize={10}
            scale="log"
            domain={[minTemperature, initialTemperature]}
            tickFormatter={(v: number) => v >= 1 ? v.toFixed(0) : v.toFixed(4)}
            label={{ value: 'Temperature (log)', angle: -90, position: 'insideLeft', fill: '#9ca3af', fontSize: 10 }}
          />
          <Tooltip
            contentStyle={{ background: '#1f2937', border: '1px solid #374151', fontSize: 11 }}
            formatter={(value: number, name: string) => {
              if (name === 'branch') return [value.toFixed(4), '🌿 Branch'];
              return [value.toFixed(6), 'Temperature'];
            }}
            labelFormatter={(label: number) => `Candidate ${formatCandidate(label)}`}
          />
          <Legend wrapperStyle={{ fontSize: 11 }} />
          <ReferenceLine y={initialTemperature} stroke="#60a5fa" strokeDasharray="4 2" label={{ value: `Initial: ${initialTemperature}`, fill: '#60a5fa', fontSize: 9, position: 'right' }} />
          <ReferenceLine y={minTemperature} stroke="#f87171" strokeDasharray="4 2" label={{ value: `Min: ${minTemperature}`, fill: '#f87171', fontSize: 9, position: 'right' }} />
          <Line
            type="monotone"
            dataKey="temperature"
            stroke="#fbbf24"
            strokeWidth={2}
            dot={false}
            name={`${coolingMode} decay`}
          />
          {branchPoints.length > 0 && (
            <Scatter
              data={branchPoints}
              dataKey="temperature"
              fill="#34d399"
              shape="diamond"
              name="Branch events"
            />
          )}
        </ComposedChart>
      </ResponsiveContainer>

      {/* Branch summary stats */}
      {branchPoints.length > 0 && (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mt-3 text-xs">
          <div className="bg-gray-800 rounded px-3 py-2">
            <div className="text-[9px] text-gray-500 uppercase">Total Branches</div>
            <div className="text-emerald-400 font-semibold">{branchPoints.length}</div>
          </div>
          <div className="bg-gray-800 rounded px-3 py-2">
            <div className="text-[9px] text-gray-500 uppercase">Avg Temp at Branch</div>
            <div className="text-emerald-400 font-semibold">{(branchPoints.reduce((s, b) => s + b.temperature, 0) / branchPoints.length).toFixed(4)}</div>
          </div>
          <div className="bg-gray-800 rounded px-3 py-2">
            <div className="text-[9px] text-gray-500 uppercase">Earliest Branch</div>
            <div className="text-emerald-400 font-semibold">{formatCandidate(Math.min(...branchPoints.map(b => b.candidate)))}</div>
          </div>
          <div className="bg-gray-800 rounded px-3 py-2">
            <div className="text-[9px] text-gray-500 uppercase">Latest Branch</div>
            <div className="text-emerald-400 font-semibold">{formatCandidate(Math.max(...branchPoints.map(b => b.candidate)))}</div>
          </div>
        </div>
      )}

      <div className="grid grid-cols-3 gap-3 mt-3 text-xs">
        <div className="bg-gray-800 rounded px-3 py-2">
          <div className="text-[9px] text-gray-500 uppercase">Initial</div>
          <div className="text-amber-400 font-semibold">{initialTemperature}</div>
        </div>
        <div className="bg-gray-800 rounded px-3 py-2">
          <div className="text-[9px] text-gray-500 uppercase">Final (theoretical)</div>
          <div className="text-amber-400 font-semibold">{finalTemp.toFixed(6)}</div>
        </div>
        <div className="bg-gray-800 rounded px-3 py-2">
          <div className="text-[9px] text-gray-500 uppercase">Iterations</div>
          <div className="text-amber-400 font-semibold">{formatCandidate(iterationsPerWorker)}</div>
        </div>
      </div>
    </Card>
  );
}
