'use client';

import { useState, useMemo } from 'react';
import Card from '@/components/Card';
import { TrendPoint } from './page';

type MetricKey = 'gap' | 'objective' | 'efficiency';

interface Props {
  points: TrendPoint[];
  bestKnown: Record<string, number>;
}

// Platform milestones (manually maintained).
const MILESTONES = [
  { label: 'Portfolio mode', domain: 'all' },
  { label: 'Search Intelligence v1', domain: 'all' },
  { label: 'Adaptive mode', domain: 'all' },
  { label: 'Learned allocator', domain: 'all' },
];

function linearRegression(values: number[]): { slope: number; r2: number } {
  const n = values.length;
  if (n < 2) return { slope: 0, r2: 0 };
  let sumX = 0, sumY = 0, sumXY = 0, sumXX = 0;
  for (let i = 0; i < n; i++) { sumX += i; sumY += values[i]; sumXY += i * values[i]; sumXX += i * i; }
  const slope = (n * sumXY - sumX * sumY) / (n * sumXX - sumX * sumX);
  const intercept = (sumY - slope * sumX) / n;
  const ssRes = values.reduce((s, v, i) => s + (v - (intercept + slope * i)) ** 2, 0);
  const ssTot = values.reduce((s, v) => s + (v - sumY / n) ** 2, 0);
  const r2 = ssTot > 0 ? 1 - ssRes / ssTot : 0;
  return { slope, r2 };
}

export default function TrendAnalysis({ points, bestKnown }: Props) {
  const domains = useMemo(() => [...new Set(points.map(p => p.domain))].sort(), [points]);
  const [selectedDomain, setSelectedDomain] = useState(domains[0] || 'all');
  const [metric, setMetric] = useState<MetricKey>('gap');

  // Filter points by domain.
  const filtered = useMemo(() => {
    if (selectedDomain === 'all') return points;
    return points.filter(p => p.domain === selectedDomain);
  }, [points, selectedDomain]);

  // Compute values based on metric.
  const chartData = useMemo(() => {
    return filtered.map((p, i) => {
      const bk = bestKnown[p.instance] || 0;
      let value: number;
      switch (metric) {
        case 'gap':
          value = bk > 0 ? ((p.objective - bk) / bk) * 100 : 0;
          break;
        case 'objective':
          // Normalise if mixed domain (divide by best-known to get ratio).
          value = selectedDomain === 'all' && bk > 0 ? p.objective / bk : p.objective;
          break;
        case 'efficiency':
          value = p.runtime > 0 ? p.objective / p.runtime : 0;
          break;
      }
      return { ...p, value, displayIndex: i };
    }).filter(d => metric !== 'gap' || d.value > 0 || bestKnown[d.instance]); // Skip gap for unknown BKS
  }, [filtered, metric, selectedDomain, bestKnown]);

  // Regression on values.
  const values = chartData.map(d => d.value);
  const reg = linearRegression(values);
  const improving = metric === 'efficiency' ? reg.slope > 0 : reg.slope < 0;

  // Conclusion.
  const conclusion = useMemo(() => {
    if (chartData.length < 3) return 'Insufficient data for trend analysis.';
    const label = metric === 'gap' ? 'gap to best-known' : metric === 'efficiency' ? 'compute efficiency' : 'objective';
    if (Math.abs(reg.r2) < 0.1) return `No clear trend in ${label} (R² = ${reg.r2.toFixed(2)}). Results are stable.`;
    if (improving) return `Platform is improving: ${label} trending ${metric === 'efficiency' ? 'up' : 'down'} (R² = ${reg.r2.toFixed(2)}).`;
    return `${label} trending ${metric === 'efficiency' ? 'down' : 'up'} — investigate potential regression (R² = ${reg.r2.toFixed(2)}).`;
  }, [chartData, reg, metric, improving]);

  // Chart dimensions.
  const W = 800, H = 220, PL = 50, PR = 20, PT = 15, PB = 30;
  const plotW = W - PL - PR, plotH = H - PT - PB;
  const maxVal = Math.max(...values, 0.01);
  const minVal = Math.min(...values, 0);
  const range = maxVal - minVal || 1;
  const toX = (i: number) => PL + (i / Math.max(chartData.length - 1, 1)) * plotW;
  const toY = (v: number) => PT + (1 - (v - minVal) / range) * plotH;

  // Y-axis label.
  const yLabel = metric === 'gap' ? 'Gap %' : metric === 'efficiency' ? 'Obj/ms' : selectedDomain === 'all' ? 'Normalised' : 'Objective';

  return (
    <div className="space-y-4">
      {/* Conclusion */}
      <Card title="Is the platform improving?">
        <p className={`text-sm ${improving ? 'text-emerald-400' : reg.r2 < 0.1 ? 'text-gray-400' : 'text-amber-400'}`}>
          {conclusion}
        </p>
        <p className="text-[10px] text-gray-500 mt-1">
          {chartData.length} runs · {selectedDomain === 'all' ? 'all domains' : selectedDomain.toUpperCase()} · {metric === 'gap' ? 'gap to best-known' : metric}
        </p>
      </Card>

      {/* Controls */}
      <Card title="Research History">
        <div className="flex flex-wrap gap-4 mb-4">
          {/* Domain filter */}
          <div>
            <label className="text-[9px] text-gray-500 uppercase block mb-1">Domain</label>
            <div className="flex gap-1">
              {domains.length > 1 && (
                <button onClick={() => setSelectedDomain('all')}
                  className={`px-2 py-1 rounded text-[10px] ${selectedDomain === 'all' ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400'}`}>
                  All (normalised)
                </button>
              )}
              {domains.map(d => (
                <button key={d} onClick={() => setSelectedDomain(d)}
                  className={`px-2 py-1 rounded text-[10px] ${selectedDomain === d ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400'}`}>
                  {d.toUpperCase()}
                </button>
              ))}
            </div>
          </div>

          {/* Metric */}
          <div>
            <label className="text-[9px] text-gray-500 uppercase block mb-1">Metric</label>
            <div className="flex gap-1">
              <button onClick={() => setMetric('gap')}
                className={`px-2 py-1 rounded text-[10px] ${metric === 'gap' ? 'bg-emerald-600 text-white' : 'bg-gray-800 text-gray-400'}`}>
                Gap to BKS
              </button>
              <button onClick={() => setMetric('objective')}
                className={`px-2 py-1 rounded text-[10px] ${metric === 'objective' ? 'bg-emerald-600 text-white' : 'bg-gray-800 text-gray-400'}`}>
                Objective
              </button>
              <button onClick={() => setMetric('efficiency')}
                className={`px-2 py-1 rounded text-[10px] ${metric === 'efficiency' ? 'bg-emerald-600 text-white' : 'bg-gray-800 text-gray-400'}`}>
                Efficiency
              </button>
            </div>
          </div>
        </div>

        {/* Chart */}
        {chartData.length > 0 ? (
          <div className="overflow-x-auto">
            <svg viewBox={`0 0 ${W} ${H}`} className="w-full" style={{ minWidth: '500px' }}>
              {/* Grid */}
              {[0, 0.25, 0.5, 0.75, 1].map(f => {
                const y = PT + f * plotH;
                const val = maxVal - f * range;
                return (
                  <g key={f}>
                    <line x1={PL} y1={y} x2={W - PR} y2={y} stroke="#1f2937" strokeWidth={0.5} />
                    <text x={PL - 5} y={y + 3} textAnchor="end" fill="#6b7280" fontSize={8}>{val.toFixed(metric === 'gap' ? 1 : 0)}</text>
                  </g>
                );
              })}

              {/* Y label */}
              <text x={12} y={H / 2} textAnchor="middle" fill="#6b7280" fontSize={9} transform={`rotate(-90, 12, ${H / 2})`}>{yLabel}</text>

              {/* Regression line */}
              {chartData.length >= 3 && (
                <line
                  x1={toX(0)} y1={toY(reg.slope * 0 + (values.reduce((s, v) => s + v, 0) / values.length) - reg.slope * (values.length - 1) / 2)}
                  x2={toX(chartData.length - 1)} y2={toY(reg.slope * (values.length - 1) + (values.reduce((s, v) => s + v, 0) / values.length) - reg.slope * (values.length - 1) / 2)}
                  stroke={improving ? '#10b981' : '#f59e0b'} strokeWidth={1} strokeDasharray="4,3" opacity={0.6}
                />
              )}

              {/* Data points */}
              {chartData.map((d, i) => (
                <circle key={d.id} cx={toX(i)} cy={toY(d.value)} r={3}
                  fill={d.domain === 'cvrp' ? '#10b981' : d.domain === 'jss' ? '#f59e0b' : d.domain === 'vrptw' ? '#f43f5e' : '#3b82f6'}
                  opacity={0.7}>
                  <title>{`${d.id}\n${d.domain.toUpperCase()} ${d.instance} ${d.mode}\n${metric === 'gap' ? d.value.toFixed(1) + '%' : d.value.toFixed(0)}`}</title>
                </circle>
              ))}
            </svg>
          </div>
        ) : (
          <p className="text-xs text-gray-500 text-center py-4">No data with known best values for gap calculation. Try "Objective" metric instead.</p>
        )}

        {/* Legend */}
        <div className="flex gap-3 mt-2 text-[9px] text-gray-500">
          <span className="flex items-center gap-1"><span className="w-2 h-2 rounded-full bg-blue-500" />NRP</span>
          <span className="flex items-center gap-1"><span className="w-2 h-2 rounded-full bg-emerald-500" />CVRP</span>
          <span className="flex items-center gap-1"><span className="w-2 h-2 rounded-full bg-amber-500" />JSS</span>
          <span className="flex items-center gap-1"><span className="w-2 h-2 rounded-full bg-rose-500" />VRPTW</span>
          <span className="flex items-center gap-1">
            <span className="w-4 h-0 border-t border-dashed" style={{ borderColor: improving ? '#10b981' : '#f59e0b' }} />
            Trend (R²={reg.r2.toFixed(2)})
          </span>
        </div>
      </Card>

      {/* Summary stats */}
      {chartData.length > 0 && (
        <Card title="Summary">
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
            <StatCard label="Runs" value={String(chartData.length)} />
            <StatCard label={metric === 'gap' ? 'Best Gap' : 'Best'} value={metric === 'gap' ? `${Math.min(...values).toFixed(1)}%` : Math.min(...values).toFixed(0)} />
            <StatCard label={metric === 'gap' ? 'Worst Gap' : 'Worst'} value={metric === 'gap' ? `${Math.max(...values).toFixed(1)}%` : Math.max(...values).toFixed(0)} />
            <StatCard label="Trend" value={improving ? '↓ Improving' : reg.r2 < 0.1 ? '→ Stable' : '↑ Check'} />
          </div>
        </Card>
      )}
    </div>
  );
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <div className="bg-gray-800 rounded p-3">
      <div className="text-[9px] text-gray-500 uppercase">{label}</div>
      <div className="text-sm font-bold text-gray-200">{value}</div>
    </div>
  );
}
