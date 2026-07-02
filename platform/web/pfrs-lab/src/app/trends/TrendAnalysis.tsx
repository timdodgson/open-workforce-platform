'use client';

import { useState, useMemo } from 'react';
import Card from '@/components/Card';
import { TrendPoint } from './page';

type Metric = 'penalty' | 'entropy' | 'workers' | 'durationMs' | 'candidates' | 'nearDuplicates';

interface Milestone {
  index: number;
  label: string;
}

interface Props {
  points: TrendPoint[];
}

const METRIC_CONFIG: Record<Metric, { label: string; color: string; format: (v: number) => string }> = {
  penalty: { label: 'Penalty', color: '#10b981', format: v => v.toLocaleString() },
  entropy: { label: 'Entropy', color: '#3b82f6', format: v => v.toFixed(2) },
  workers: { label: 'Workers', color: '#f59e0b', format: v => v.toLocaleString() },
  durationMs: { label: 'Runtime (s)', color: '#8b5cf6', format: v => (v / 1000).toFixed(1) },
  candidates: { label: 'Candidates (M)', color: '#06b6d4', format: v => (v / 1_000_000).toFixed(1) },
  nearDuplicates: { label: 'Near-Duplicates', color: '#ef4444', format: v => v.toLocaleString() },
};

function linearRegression(values: number[]): { slope: number; intercept: number; r2: number } {
  const n = values.length;
  if (n < 2) return { slope: 0, intercept: values[0] || 0, r2: 0 };
  let sumX = 0, sumY = 0, sumXY = 0, sumXX = 0, sumYY = 0;
  for (let i = 0; i < n; i++) {
    sumX += i; sumY += values[i];
    sumXY += i * values[i]; sumXX += i * i; sumYY += values[i] * values[i];
  }
  const slope = (n * sumXY - sumX * sumY) / (n * sumXX - sumX * sumX);
  const intercept = (sumY - slope * sumX) / n;
  const ssRes = values.reduce((s, v, i) => s + (v - (intercept + slope * i)) ** 2, 0);
  const ssTot = values.reduce((s, v) => s + (v - sumY / n) ** 2, 0);
  const r2 = ssTot > 0 ? 1 - ssRes / ssTot : 0;
  return { slope, intercept, r2 };
}

function movingAverage(values: number[], window: number): number[] {
  return values.map((_, i) => {
    const start = Math.max(0, i - window + 1);
    const slice = values.slice(start, i + 1);
    return slice.reduce((s, v) => s + v, 0) / slice.length;
  });
}

export default function TrendAnalysis({ points }: Props) {
  const [metric, setMetric] = useState<Metric>('penalty');
  const [showRegression, setShowRegression] = useState(true);
  const [showMA, setShowMA] = useState(true);
  const [maWindow, setMAWindow] = useState(3);
  const [milestones, setMilestones] = useState<Milestone[]>([]);
  const [newMilestone, setNewMilestone] = useState('');
  const [milestoneIdx, setMilestoneIdx] = useState(0);

  const values = useMemo(() => points.map(p => p[metric] as number), [points, metric]);
  const regression = useMemo(() => linearRegression(values), [values]);
  const ma = useMemo(() => movingAverage(values, maWindow), [values, maWindow]);

  const cfg = METRIC_CONFIG[metric];
  const maxVal = Math.max(...values, 1);
  const minVal = Math.min(...values, 0);
  const range = maxVal - minVal || 1;

  // SVG dimensions.
  const W = 900, H = 250, PL = 60, PR = 20, PT = 20, PB = 40;
  const plotW = W - PL - PR;
  const plotH = H - PT - PB;

  function toX(i: number): number { return PL + (i / Math.max(points.length - 1, 1)) * plotW; }
  function toY(v: number): number { return PT + (1 - (v - minVal) / range) * plotH; }

  function addMilestone() {
    if (!newMilestone.trim()) return;
    setMilestones([...milestones, { index: milestoneIdx, label: newMilestone.trim() }]);
    setNewMilestone('');
  }

  // Trend observation.
  const trendObs = useMemo(() => {
    const obs: string[] = [];
    if (regression.slope < 0 && regression.r2 > 0.3) {
      obs.push(`${cfg.label} is trending downward (slope: ${regression.slope.toFixed(2)}, R²=${regression.r2.toFixed(2)}).`);
    } else if (regression.slope > 0 && regression.r2 > 0.3) {
      obs.push(`${cfg.label} is trending upward (slope: ${regression.slope.toFixed(2)}, R²=${regression.r2.toFixed(2)}).`);
    } else {
      obs.push(`No clear trend in ${cfg.label.toLowerCase()} (R²=${regression.r2.toFixed(2)}).`);
    }
    const best = Math.min(...values);
    const bestIdx = values.indexOf(best);
    obs.push(`Best ${cfg.label.toLowerCase()}: ${cfg.format(best)} (run ${points[bestIdx]?.id || bestIdx}).`);
    return obs;
  }, [regression, values, cfg, points]);

  return (
    <div className="space-y-4">
      {/* Controls */}
      <Card title="Trend Analysis">
        <div className="flex flex-wrap gap-3 mb-3">
          <div className="flex items-center gap-1">
            <span className="text-[10px] text-gray-500">Metric:</span>
            {(Object.keys(METRIC_CONFIG) as Metric[]).map(m => (
              <button key={m} onClick={() => setMetric(m)}
                className={`px-2 py-0.5 rounded text-[10px] ${metric === m ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400'}`}>
                {METRIC_CONFIG[m].label}
              </button>
            ))}
          </div>
          <div className="flex items-center gap-1">
            <label className="text-[10px] text-gray-500 flex items-center gap-1">
              <input type="checkbox" checked={showRegression} onChange={e => setShowRegression(e.target.checked)} />
              Regression
            </label>
            <label className="text-[10px] text-gray-500 flex items-center gap-1">
              <input type="checkbox" checked={showMA} onChange={e => setShowMA(e.target.checked)} />
              MA({maWindow})
            </label>
            <input type="range" min={2} max={5} value={maWindow} onChange={e => setMAWindow(parseInt(e.target.value))}
              className="w-16 h-1 accent-blue-500" />
          </div>
        </div>

        {/* Chart */}
        <svg viewBox={`0 0 ${W} ${H}`} className="w-full h-[250px] bg-gray-900 rounded border border-gray-800">
          {/* Grid */}
          {[0, 0.25, 0.5, 0.75, 1].map(frac => {
            const y = PT + (1 - frac) * plotH;
            const val = minVal + frac * range;
            return (
              <g key={frac}>
                <line x1={PL} y1={y} x2={W - PR} y2={y} stroke="#1f2937" strokeWidth="0.5" />
                <text x={PL - 5} y={y} textAnchor="end" dominantBaseline="middle" className="fill-gray-600 text-[8px]">
                  {cfg.format(val)}
                </text>
              </g>
            );
          })}

          {/* Milestones */}
          {milestones.map((m, i) => {
            const x = toX(m.index);
            return (
              <g key={i}>
                <line x1={x} y1={PT} x2={x} y2={H - PB} stroke="#a855f7" strokeWidth="1" strokeDasharray="4,4" opacity={0.6} />
                <text x={x + 3} y={PT + 10} className="fill-purple-400 text-[7px]">{m.label}</text>
              </g>
            );
          })}

          {/* Regression line */}
          {showRegression && values.length >= 2 && (
            <line
              x1={toX(0)} y1={toY(regression.intercept)}
              x2={toX(values.length - 1)} y2={toY(regression.intercept + regression.slope * (values.length - 1))}
              stroke="#ef4444" strokeWidth="1" strokeDasharray="6,3" opacity={0.7}
            />
          )}

          {/* Moving average */}
          {showMA && ma.length > 1 && (
            <polyline
              points={ma.map((v, i) => `${toX(i)},${toY(v)}`).join(' ')}
              fill="none" stroke="#f59e0b" strokeWidth="1.5" opacity={0.7}
            />
          )}

          {/* Data line */}
          {values.length > 1 && (
            <polyline
              points={values.map((v, i) => `${toX(i)},${toY(v)}`).join(' ')}
              fill="none" stroke={cfg.color} strokeWidth="2"
            />
          )}

          {/* Data points */}
          {values.map((v, i) => (
            <circle key={i} cx={toX(i)} cy={toY(v)} r={4}
              fill={cfg.color} opacity={0.9}
            >
              <title>{`${points[i].id}: ${cfg.format(v)}`}</title>
            </circle>
          ))}

          {/* X-axis labels */}
          {points.map((p, i) => (
            points.length <= 15 || i % Math.ceil(points.length / 10) === 0 ? (
              <text key={i} x={toX(i)} y={H - PB + 15} textAnchor="middle" className="fill-gray-600 text-[7px]">
                {p.id.length > 8 ? p.id.slice(0, 8) + '…' : p.id}
              </text>
            ) : null
          ))}
        </svg>

        <div className="flex gap-4 mt-2 text-[9px] text-gray-500">
          <span className="flex items-center gap-1"><span className="w-4 h-0.5" style={{ background: cfg.color }} />Data</span>
          {showMA && <span className="flex items-center gap-1"><span className="w-4 h-0.5 bg-amber-500" />Moving Avg</span>}
          {showRegression && <span className="flex items-center gap-1"><span className="w-4 h-0.5 bg-red-500 border-dashed" />Regression</span>}
        </div>
      </Card>

      {/* Milestones editor */}
      <Card title="Milestones">
        <div className="flex gap-2 mb-2">
          <select value={milestoneIdx} onChange={e => setMilestoneIdx(parseInt(e.target.value))}
            className="bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs">
            {points.map((p, i) => <option key={i} value={i}>{p.id}</option>)}
          </select>
          <input value={newMilestone} onChange={e => setNewMilestone(e.target.value)}
            placeholder="e.g. Refinement introduced"
            className="flex-1 bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs" />
          <button onClick={addMilestone}
            className="px-3 py-1 bg-purple-600 hover:bg-purple-500 text-white rounded text-xs">Add</button>
        </div>
        {milestones.length > 0 && (
          <div className="flex flex-wrap gap-1">
            {milestones.map((m, i) => (
              <span key={i} className="px-2 py-0.5 bg-purple-900/30 border border-purple-700 rounded text-[9px] text-purple-300">
                {m.label} (run {m.index})
                <button onClick={() => setMilestones(milestones.filter((_, j) => j !== i))} className="ml-1 text-gray-500">×</button>
              </span>
            ))}
          </div>
        )}
      </Card>

      {/* Statistics */}
      <Card title="Statistics">
        <div className="grid grid-cols-4 gap-3 text-xs text-center">
          <div><p className="text-lg font-bold" style={{ color: cfg.color }}>{cfg.format(Math.min(...values))}</p><p className="text-[9px] text-gray-500">Best</p></div>
          <div><p className="text-lg font-bold text-gray-300">{cfg.format(values.reduce((s, v) => s + v, 0) / values.length)}</p><p className="text-[9px] text-gray-500">Mean</p></div>
          <div><p className="text-lg font-bold text-gray-300">{cfg.format(Math.max(...values))}</p><p className="text-[9px] text-gray-500">Worst</p></div>
          <div><p className="text-lg font-bold text-gray-400">{regression.r2.toFixed(3)}</p><p className="text-[9px] text-gray-500">R²</p></div>
        </div>
      </Card>

      {/* Observations */}
      <Card title="Observations">
        <div className="space-y-1">
          {trendObs.map((obs, i) => (
            <p key={i} className="text-sm text-gray-300">{obs}</p>
          ))}
        </div>
      </Card>

      {/* Run list */}
      <Card title="Runs (chronological)">
        <div className="max-h-48 overflow-y-auto">
          <table className="w-full text-[10px]">
            <thead className="sticky top-0 bg-gray-900">
              <tr className="text-gray-500 uppercase">
                <th className="text-left p-1">#</th>
                <th className="text-left p-1">Run</th>
                <th className="text-right p-1">Penalty</th>
                <th className="text-right p-1">Mode</th>
                <th className="text-right p-1">Beam</th>
                <th className="text-right p-1">Workers</th>
              </tr>
            </thead>
            <tbody>
              {points.map((p, i) => (
                <tr key={p.id} className="border-t border-gray-800">
                  <td className="p-1 text-gray-600">{i + 1}</td>
                  <td className="p-1 text-blue-400 font-mono truncate max-w-[120px]">{p.id}</td>
                  <td className="text-right p-1">{p.penalty.toLocaleString()}</td>
                  <td className="text-right p-1 text-gray-400">{p.mode}</td>
                  <td className="text-right p-1">{p.beamWidth}</td>
                  <td className="text-right p-1">{p.workers}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>
    </div>
  );
}
