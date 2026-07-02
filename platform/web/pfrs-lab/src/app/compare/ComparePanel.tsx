'use client';
import { useState, useMemo } from 'react';
import { BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid, Legend, LineChart, Line, RadarChart, PolarGrid, PolarAngleAxis, PolarRadiusAxis, Radar } from 'recharts';
import Card from '@/components/Card';
import { RunSummary, TreeNode, DiversityRecord } from '@/lib/types';

const COLORS = ['#34d399', '#60a5fa', '#fbbf24', '#f87171', '#a78bfa', '#fb923c'];

interface RunData {
  id: string;
  summary: RunSummary;
  nodes: TreeNode[];
  diversity: DiversityRecord[];
}

interface Props {
  runs: RunData[];
}

interface RunMetrics {
  id: string;
  color: string;
  mode: string;
  beamWidth: number;
  iterations: number;
  totalPenalty: number;
  totalWorkers: number;
  totalCandidates: number;
  totalDurationMs: number;
  maxWeekPenalty: number;
  avgEntropy: number;
  healthScore: number;
  maxMonopoly: number;
  weekPenalties: number[];
  entropyByWeek: number[];
  innovationDiversity: number; // families with >10% contribution
  runtimeEfficiency: number; // improvement per second
}

function computeMetrics(runs: RunData[]): RunMetrics[] {
  return runs.map((run, idx) => {
    const s = run.summary;
    const nodes = run.nodes;
    const retained = nodes.filter(n => n.retained);
    const weeks = Array.from(new Set(nodes.map(n => n.week))).sort((a, b) => a - b);

    const byID = new Map(nodes.map(n => [n.pathID, n]));
    function getRoot(id: number): number {
      const n = byID.get(id);
      if (!n || n.parentID <= 0) return id;
      return getRoot(n.parentID);
    }

    const entropyByWeek: number[] = [];
    for (const w of weeks) {
      const weekRetained = retained.filter(n => n.week === w);
      const familyCounts = new Map<number, number>();
      for (const n of weekRetained) {
        const root = getRoot(n.pathID);
        familyCounts.set(root, (familyCounts.get(root) || 0) + 1);
      }
      const total = weekRetained.length;
      let entropy = 0;
      for (const count of familyCounts.values()) {
        const p = count / total;
        if (p > 0) entropy -= p * Math.log2(p);
      }
      const maxE = Math.log2(familyCounts.size) || 1;
      entropyByWeek.push(parseFloat((entropy / maxE).toFixed(3)));
    }

    const avgEntropy = entropyByWeek.length > 0 ? entropyByWeek.reduce((s, e) => s + e, 0) / entropyByWeek.length : 0;

    const allRoots = Array.from(new Set(retained.map(n => getRoot(n.pathID))));
    let maxFamilyPct = 0;
    const familySizes = new Map<number, number>();
    for (const root of allRoots) {
      const count = retained.filter(n => getRoot(n.pathID) === root).length;
      familySizes.set(root, count);
      const pct = (count / (retained.length || 1)) * 100;
      if (pct > maxFamilyPct) maxFamilyPct = pct;
    }

    const innovationDiversity = Array.from(familySizes.values()).filter(c => (c / (retained.length || 1)) > 0.1).length;
    const diversityScore = Math.round(avgEntropy * 30);
    const monopolyPenalty = Math.round((maxFamilyPct / 100) * 20);
    const healthScore = Math.max(0, Math.min(100, diversityScore + 30 - monopolyPenalty));
    const runtimeEfficiency = s.totalDurationMs > 0 ? Math.round((s.totalPenalty > 0 ? s.weeks.reduce((sum, w) => sum + w.improvement, 0) : 0) / (s.totalDurationMs / 1000)) : 0;

    return {
      id: run.id,
      color: COLORS[idx % COLORS.length],
      mode: s.metadata?.mode?.toUpperCase() || '—',
      beamWidth: s.metadata?.beamWidth || 0,
      iterations: s.metadata?.iterationsPerWorker || 0,
      totalPenalty: s.totalPenalty,
      totalWorkers: s.totalWorkers,
      totalCandidates: s.totalCandidates,
      totalDurationMs: s.totalDurationMs,
      maxWeekPenalty: s.maxWeekPenalty,
      avgEntropy: parseFloat(avgEntropy.toFixed(3)),
      healthScore,
      maxMonopoly: Math.round(maxFamilyPct),
      weekPenalties: s.weeks.map(w => w.finalPenalty),
      entropyByWeek,
      innovationDiversity,
      runtimeEfficiency,
    };
  });
}

function rank(values: number[], direction: 'min' | 'max'): ('best' | 'second' | 'worst' | '')[] {
  const sorted = [...values].sort((a, b) => direction === 'min' ? a - b : b - a);
  return values.map(v => {
    if (v === sorted[0]) return 'best';
    if (v === sorted[sorted.length - 1] && sorted.length > 1) return 'worst';
    if (sorted.length > 2 && v === sorted[1]) return 'second';
    return '';
  });
}

function rankColor(r: string): string {
  if (r === 'best') return 'text-emerald-400 font-bold';
  if (r === 'worst') return 'text-red-400';
  if (r === 'second') return 'text-amber-400';
  return '';
}

export default function ComparePanel({ runs }: Props) {
  const [selected, setSelected] = useState<string[]>(runs.slice(0, Math.min(4, runs.length)).map(r => r.id));
  const [baseline, setBaseline] = useState<string>(selected[0] || '');

  const selectedRuns = runs.filter(r => selected.includes(r.id));
  const metrics = useMemo(() => computeMetrics(selectedRuns), [selectedRuns]);
  const baselineMetrics = metrics.find(m => m.id === baseline);

  // Radar data (normalised 0-100).
  const radarData = useMemo(() => {
    if (metrics.length === 0) return [];
    const maxPenalty = Math.max(...metrics.map(m => m.totalPenalty)) || 1;
    const maxRuntime = Math.max(...metrics.map(m => m.totalDurationMs)) || 1;

    const axes = ['Penalty', 'Beam Health', 'Entropy', 'Innovation', 'Efficiency', 'Stability'];
    return axes.map(axis => {
      const row: Record<string, number | string> = { axis };
      for (const m of metrics) {
        let val = 0;
        switch (axis) {
          case 'Penalty': val = Math.round((1 - m.totalPenalty / maxPenalty) * 100); break;
          case 'Beam Health': val = m.healthScore; break;
          case 'Entropy': val = Math.round(m.avgEntropy * 100); break;
          case 'Innovation': val = Math.round((m.innovationDiversity / Math.max(...metrics.map(x => x.innovationDiversity), 1)) * 100); break;
          case 'Efficiency': val = Math.round((1 - m.totalDurationMs / maxRuntime) * 100); break;
          case 'Stability': val = Math.round((1 - m.maxMonopoly / 100) * 100); break;
        }
        row[m.id] = Math.max(0, Math.min(100, val));
      }
      return row;
    });
  }, [metrics]);

  // Ranking.
  const overallScores = metrics.map(m => {
    const penaltyNorm = metrics.length > 0 ? (1 - m.totalPenalty / Math.max(...metrics.map(x => x.totalPenalty), 1)) * 30 : 0;
    const healthNorm = m.healthScore * 0.25;
    const entropyNorm = m.avgEntropy * 20;
    const diversityNorm = (m.innovationDiversity / Math.max(...metrics.map(x => x.innovationDiversity), 1)) * 15;
    const monopolyNorm = (1 - m.maxMonopoly / 100) * 10;
    return Math.round(penaltyNorm + healthNorm + entropyNorm + diversityNorm + monopolyNorm);
  });

  // Research notes.
  const notes = useMemo(() => {
    const observations: string[] = [];
    if (metrics.length < 2) return observations;
    const sorted = [...metrics].sort((a, b) => a.totalPenalty - b.totalPenalty);
    const best = sorted[0];
    const worst = sorted[sorted.length - 1];
    const penaltyDiff = worst.totalPenalty - best.totalPenalty;
    const pctDiff = ((penaltyDiff / worst.totalPenalty) * 100).toFixed(1);
    observations.push(`${best.id} achieves ${penaltyDiff} lower penalty than ${worst.id} (${pctDiff}% improvement).`);

    const bestEntropy = metrics.reduce((a, b) => a.avgEntropy > b.avgEntropy ? a : b);
    const worstEntropy = metrics.reduce((a, b) => a.avgEntropy < b.avgEntropy ? a : b);
    if (bestEntropy.id !== worstEntropy.id) {
      observations.push(`${bestEntropy.id} maintains ${((bestEntropy.avgEntropy - worstEntropy.avgEntropy) * 100).toFixed(0)}% higher average entropy than ${worstEntropy.id}.`);
    }

    const bestHealth = metrics.reduce((a, b) => a.healthScore > b.healthScore ? a : b);
    if (bestHealth.healthScore > 50) {
      observations.push(`${bestHealth.id} achieves healthy beam exploration (score ${bestHealth.healthScore}/100).`);
    } else {
      observations.push(`All runs show beam convergence issues (best health: ${bestHealth.healthScore}/100).`);
    }

    const runtimeRange = Math.max(...metrics.map(m => m.totalDurationMs)) - Math.min(...metrics.map(m => m.totalDurationMs));
    if (runtimeRange < Math.min(...metrics.map(m => m.totalDurationMs)) * 0.1) {
      observations.push(`Runtime is comparable across all configurations (within 10%).`);
    }

    return observations;
  }, [metrics]);

  // Recommendation.
  const recommendation = useMemo(() => {
    if (metrics.length < 2) return null;
    const bestIdx = overallScores.indexOf(Math.max(...overallScores));
    const best = metrics[bestIdx];
    const advantages: string[] = [];
    const tradeoffs: string[] = [];

    const penaltyRanks = rank(metrics.map(m => m.totalPenalty), 'min');
    const healthRanks = rank(metrics.map(m => m.healthScore), 'max');
    const entropyRanks = rank(metrics.map(m => m.avgEntropy), 'max');

    if (penaltyRanks[bestIdx] === 'best') advantages.push('Lowest penalty');
    else if (penaltyRanks[bestIdx] === 'worst') tradeoffs.push(`Higher penalty than best by ${(best.totalPenalty - Math.min(...metrics.map(m => m.totalPenalty))).toLocaleString()}`);
    if (healthRanks[bestIdx] === 'best') advantages.push(`Best Beam Health (${best.healthScore}/100)`);
    if (entropyRanks[bestIdx] === 'best') advantages.push(`Highest diversity (entropy ${best.avgEntropy})`);
    if (best.maxMonopoly < 70) advantages.push(`Low monopoly (${best.maxMonopoly}%)`);
    else tradeoffs.push(`High monopoly (${best.maxMonopoly}%)`);

    return { id: best.id, score: overallScores[bestIdx], advantages, tradeoffs };
  }, [metrics, overallScores]);

  // Chart data.
  const weekChartData = metrics[0]?.weekPenalties.map((_, i) => {
    const row: Record<string, number | string> = { week: `W${i + 1}` };
    for (const m of metrics) row[m.id] = m.weekPenalties[i] || 0;
    return row;
  }) || [];

  const entropyChartData = metrics[0]?.entropyByWeek.map((_, i) => {
    const row: Record<string, number | string> = { week: `W${i + 1}` };
    for (const m of metrics) row[m.id] = m.entropyByWeek[i] || 0;
    return row;
  }) || [];

  return (
    <>
      {/* Run selector */}
      <Card title="Select Runs (2-4)">
        <div className="flex flex-wrap gap-2 mb-2">
          {runs.map(run => {
            const isSelected = selected.includes(run.id);
            return (
              <button key={run.id}
                onClick={() => {
                  if (isSelected) { if (selected.length > 2) setSelected(selected.filter(id => id !== run.id)); }
                  else { if (selected.length < 4) setSelected([...selected, run.id]); }
                }}
                className={`text-xs px-3 py-1.5 rounded border transition-colors ${
                  isSelected ? 'bg-blue-600 border-blue-500 text-white' : 'bg-gray-800 border-gray-700 text-gray-400 hover:border-gray-500'
                }`}>
                {run.id}
              </button>
            );
          })}
        </div>
        <div className="flex items-center gap-2 text-[10px] text-gray-500">
          <span>Baseline:</span>
          <select value={baseline} onChange={e => setBaseline(e.target.value)}
            className="bg-gray-800 text-gray-300 rounded px-2 py-0.5 border border-gray-700 text-[10px]">
            {selected.map(id => <option key={id} value={id}>{id}</option>)}
          </select>
        </div>
      </Card>

      {/* Research Recommendation */}
      {recommendation && (
        <Card title="Research Recommendation">
          <div className="bg-gray-800/50 rounded-lg p-4">
            <div className="text-sm text-gray-200 mb-2">
              <span className="text-emerald-400 font-bold">{recommendation.id}</span> provides the best overall search quality (score: {recommendation.score}).
            </div>
            {recommendation.advantages.length > 0 && (
              <div className="mb-2">
                {recommendation.advantages.map((a, i) => (
                  <div key={i} className="text-xs text-emerald-400">✓ {a}</div>
                ))}
              </div>
            )}
            {recommendation.tradeoffs.length > 0 && (
              <div>
                {recommendation.tradeoffs.map((t, i) => (
                  <div key={i} className="text-xs text-amber-400">• {t}</div>
                ))}
              </div>
            )}
          </div>
        </Card>
      )}

      {/* Summary table with ranking + deltas */}
      <Card title="Metrics Comparison">
        <div className="overflow-x-auto">
          <table className="w-full text-xs">
            <thead>
              <tr className="text-gray-500 uppercase">
                <th className="text-left p-2 sticky left-0 bg-gray-850">Metric</th>
                {metrics.map(m => (
                  <th key={m.id} className="text-right p-2 min-w-[120px]" style={{ color: m.color }}>
                    {m.id}{m.id === baseline ? ' ★' : ''}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              <MetricRow label="Algorithm" values={metrics.map(m => m.mode)} />
              <MetricRow label="Beam Width" values={metrics.map(m => String(m.beamWidth))} />
              <MetricRow label="Iterations" values={metrics.map(m => `${(m.iterations / 1000).toFixed(0)}K`)} />
              <NumMetricRow label="Total Penalty" values={metrics.map(m => m.totalPenalty)} direction="min" baseline={baselineMetrics?.totalPenalty} />
              <NumMetricRow label="Worst Week" values={metrics.map(m => m.maxWeekPenalty)} direction="min" baseline={baselineMetrics?.maxWeekPenalty} />
              <NumMetricRow label="Beam Health" values={metrics.map(m => m.healthScore)} direction="max" baseline={baselineMetrics?.healthScore} />
              <NumMetricRow label="Avg Entropy" values={metrics.map(m => m.avgEntropy)} direction="max" baseline={baselineMetrics?.avgEntropy} />
              <NumMetricRow label="Max Monopoly %" values={metrics.map(m => m.maxMonopoly)} direction="min" baseline={baselineMetrics?.maxMonopoly} />
              <NumMetricRow label="Innovation Families" values={metrics.map(m => m.innovationDiversity)} direction="max" baseline={baselineMetrics?.innovationDiversity} />
              <NumMetricRow label="Runtime (s)" values={metrics.map(m => parseFloat((m.totalDurationMs / 1000).toFixed(1)))} direction="min" baseline={baselineMetrics ? parseFloat((baselineMetrics.totalDurationMs / 1000).toFixed(1)) : undefined} />
              <NumMetricRow label="Overall Score" values={overallScores} direction="max" />
            </tbody>
          </table>
        </div>
      </Card>

      {/* Radar chart */}
      <Card title="Strategy Profile">
        <ResponsiveContainer width="100%" height={300}>
          <RadarChart data={radarData}>
            <PolarGrid stroke="#374151" />
            <PolarAngleAxis dataKey="axis" tick={{ fill: '#9ca3af', fontSize: 10 }} />
            <PolarRadiusAxis angle={30} domain={[0, 100]} tick={{ fill: '#6b7280', fontSize: 9 }} />
            {metrics.map(m => (
              <Radar key={m.id} name={m.id} dataKey={m.id} stroke={m.color} fill={m.color} fillOpacity={0.15} strokeWidth={2} />
            ))}
            <Legend wrapperStyle={{ fontSize: 10 }} />
            <Tooltip contentStyle={{ background: '#1f2937', border: '1px solid #374151' }} />
          </RadarChart>
        </ResponsiveContainer>
      </Card>

      {/* Penalty by week */}
      <Card title="Penalty by Week">
        <ResponsiveContainer width="100%" height={200}>
          <BarChart data={weekChartData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="week" stroke="#9ca3af" fontSize={11} />
            <YAxis stroke="#9ca3af" fontSize={11} />
            <Tooltip contentStyle={{ background: '#1f2937', border: '1px solid #374151' }} />
            <Legend wrapperStyle={{ fontSize: 10 }} />
            {metrics.map(m => <Bar key={m.id} dataKey={m.id} fill={m.color} name={m.id} opacity={0.8} />)}
          </BarChart>
        </ResponsiveContainer>
      </Card>

      {/* Entropy overlay */}
      <Card title="Lineage Entropy Over Time">
        <ResponsiveContainer width="100%" height={180}>
          <LineChart data={entropyChartData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="week" stroke="#9ca3af" fontSize={11} />
            <YAxis stroke="#9ca3af" fontSize={11} domain={[0, 1]} />
            <Tooltip contentStyle={{ background: '#1f2937', border: '1px solid #374151' }} />
            <Legend wrapperStyle={{ fontSize: 10 }} />
            {metrics.map(m => <Line key={m.id} type="monotone" dataKey={m.id} stroke={m.color} strokeWidth={2} name={m.id} dot={{ r: 3 }} />)}
          </LineChart>
        </ResponsiveContainer>
      </Card>

      {/* Ranking */}
      <Card title="Overall Ranking">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Rank</th>
              <th className="text-left p-2">Run</th>
              <th className="text-right p-2">Score</th>
              <th className="text-left p-2">Strength</th>
              <th className="text-left p-2">Weakness</th>
            </tr>
          </thead>
          <tbody>
            {metrics
              .map((m, i) => ({ ...m, score: overallScores[i] }))
              .sort((a, b) => b.score - a.score)
              .map((m, i) => {
                const strength = m.healthScore >= 50 ? 'Healthy exploration'
                  : m.totalPenalty === Math.min(...metrics.map(x => x.totalPenalty)) ? 'Best penalty'
                  : m.avgEntropy > 0.5 ? 'Good diversity' : 'Competitive';
                const weakness = m.maxMonopoly > 80 ? 'High monopoly'
                  : m.healthScore < 30 ? 'Beam collapsed'
                  : m.totalPenalty === Math.max(...metrics.map(x => x.totalPenalty)) ? 'Highest penalty'
                  : '—';
                return (
                  <tr key={m.id} className="border-t border-gray-800">
                    <td className="p-2 font-bold">{i === 0 ? '🥇' : i === 1 ? '🥈' : i === 2 ? '🥉' : `${i + 1}`}</td>
                    <td className="p-2" style={{ color: m.color }}>{m.id}</td>
                    <td className="text-right p-2 font-bold">{m.score}</td>
                    <td className="p-2 text-emerald-400">{strength}</td>
                    <td className="p-2 text-red-400">{weakness}</td>
                  </tr>
                );
              })}
          </tbody>
        </table>
      </Card>

      {/* Research Notes */}
      {notes.length > 0 && (
        <Card title="Research Notes">
          <div className="space-y-2">
            {notes.map((note, i) => (
              <div key={i} className="text-xs text-gray-400 pl-3 border-l-2 border-blue-500">
                {note}
              </div>
            ))}
          </div>
        </Card>
      )}
    </>
  );
}

// --- Table Row Components ---

function MetricRow({ label, values }: { label: string; values: string[] }) {
  return (
    <tr className="border-t border-gray-800">
      <td className="p-2 text-gray-400 sticky left-0 bg-gray-850">{label}</td>
      {values.map((v, i) => <td key={i} className="text-right p-2">{v}</td>)}
    </tr>
  );
}

function NumMetricRow({ label, values, direction, baseline }: { label: string; values: number[]; direction: 'min' | 'max'; baseline?: number }) {
  const ranks = rank(values, direction);
  return (
    <tr className="border-t border-gray-800">
      <td className="p-2 text-gray-400 sticky left-0 bg-gray-850">{label}</td>
      {values.map((v, i) => {
        const delta = baseline !== undefined ? v - baseline : null;
        const pct = baseline !== undefined && baseline !== 0 ? ((v - baseline) / baseline) * 100 : null;
        const isGood = delta !== null && ((direction === 'min' && delta < 0) || (direction === 'max' && delta > 0));
        return (
          <td key={i} className={`text-right p-2 ${rankColor(ranks[i])}`}>
            <div>{typeof v === 'number' && Math.abs(v) >= 100 ? v.toLocaleString() : v}</div>
            {delta !== null && delta !== 0 && (
              <div className={`text-[9px] ${isGood ? 'text-emerald-600' : 'text-red-600'}`}>
                {delta > 0 ? '+' : ''}{typeof delta === 'number' && Math.abs(delta) >= 100 ? delta.toLocaleString() : delta.toFixed(1)}
                {pct !== null && ` (${pct > 0 ? '+' : ''}${pct.toFixed(1)}%)`}
              </div>
            )}
          </td>
        );
      })}
    </tr>
  );
}
