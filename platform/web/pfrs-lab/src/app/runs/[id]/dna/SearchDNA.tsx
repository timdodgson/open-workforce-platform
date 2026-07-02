'use client';

import { useMemo } from 'react';
import Card from '@/components/Card';
import { RunSummary, DiscoveryRecord, WorkerLifecycle, TreeNode, DiversityRecord } from '@/lib/types';

interface DNAMetric {
  id: string;
  name: string;
  score: number; // 0-100
  description: string;
}

interface Props {
  summary: RunSummary;
  discoveries: DiscoveryRecord[];
  workers: WorkerLifecycle[];
  tree: TreeNode[];
  diversity: DiversityRecord[];
}

function computeMetrics(props: Props): DNAMetric[] {
  const { summary, discoveries, workers, tree, diversity } = props;
  const weeks = summary.weeks;
  const metrics: DNAMetric[] = [];

  // 1. Exploration: how much of the search space was explored.
  // Based on: acceptance of worse solutions + branch diversity.
  const explorationScore = Math.min(100, Math.round(
    (summary.acceptWorseRate * 10) + // higher accept-worse = more exploration
    (summary.totalBranches / Math.max(summary.totalWorkers, 1)) * 20 // branches per worker
  ));
  metrics.push({
    id: 'exploration', name: 'Exploration',
    score: explorationScore,
    description: 'Willingness to explore worse solutions and create branches',
  });

  // 2. Exploitation: how efficiently improvements were found.
  // Based on: improvement per million candidates.
  const totalImprovement = weeks.reduce((s, w) => s + w.improvement, 0);
  const effPerM = summary.totalCandidates > 0
    ? totalImprovement / (summary.totalCandidates / 1_000_000) : 0;
  const exploitationScore = Math.min(100, Math.round(effPerM * 2));
  metrics.push({
    id: 'exploitation', name: 'Exploitation',
    score: exploitationScore,
    description: 'Efficiency at converting exploration into improvements',
  });

  // 3. Innovation: how many genuinely new solutions were discovered.
  // Based on: global best discoveries relative to total workers.
  const globalDiscoveries = discoveries.filter(d => d.eventType === 'global_best').length;
  const innovationScore = Math.min(100, Math.round(
    (globalDiscoveries / Math.max(summary.totalWorkers, 1)) * 500
  ));
  metrics.push({
    id: 'innovation', name: 'Innovation',
    score: innovationScore,
    description: 'Rate of finding genuinely new best solutions',
  });

  // 4. Lineage Diversity: how many distinct families survive.
  // Based on: unique parent IDs in retained paths.
  const retainedPaths = tree.filter(t => t.retained);
  const uniqueParents = new Set(retainedPaths.map(t => t.parentID)).size;
  const maxPossible = Math.max(retainedPaths.length, 1);
  const lineageScore = Math.min(100, Math.round((uniqueParents / maxPossible) * 100));
  metrics.push({
    id: 'lineage', name: 'Lineage Diversity',
    score: lineageScore,
    description: 'Number of distinct family lines surviving beam selection',
  });

  // 5. Worker Utilisation: percentage of workers that contributed.
  const producedBest = workers.filter(w => w.producedGlobalBest).length;
  const improved = workers.filter(w => w.bestPenalty < w.startPenalty).length;
  const utilisationScore = Math.min(100, Math.round(
    (improved / Math.max(workers.length, 1)) * 100
  ));
  metrics.push({
    id: 'utilisation', name: 'Worker Utilisation',
    score: utilisationScore,
    description: 'Percentage of workers that improved on their starting solution',
  });

  // 6. Convergence Speed: how quickly the run found its best result.
  // Based on: what percentage of total time was spent after finding the best.
  const bestDiscovery = discoveries.filter(d => d.eventType === 'global_best').slice(-1)[0];
  const totalTimeMs = summary.totalDurationMs;
  const timeToBest = bestDiscovery?.elapsedMs ?? totalTimeMs;
  const convergenceScore = totalTimeMs > 0
    ? Math.min(100, Math.round((1 - timeToBest / totalTimeMs) * 100))
    : 50;
  metrics.push({
    id: 'convergence', name: 'Convergence Speed',
    score: Math.max(0, convergenceScore),
    description: 'How early in the run the best solution was found',
  });

  // 7. Beam Stability: consistency of beam selection across weeks.
  // Based on: variance in cumulative penalties of retained paths.
  const retainedByWeek = new Map<number, number[]>();
  for (const t of retainedPaths) {
    const existing = retainedByWeek.get(t.week) || [];
    existing.push(t.cumulativePenalty);
    retainedByWeek.set(t.week, existing);
  }
  let stabilitySum = 0;
  let stabilityCount = 0;
  for (const penalties of retainedByWeek.values()) {
    if (penalties.length < 2) continue;
    const mean = penalties.reduce((a, b) => a + b, 0) / penalties.length;
    const variance = penalties.reduce((s, p) => s + (p - mean) ** 2, 0) / penalties.length;
    const cv = mean > 0 ? Math.sqrt(variance) / mean : 0;
    stabilitySum += 1 - Math.min(cv, 1);
    stabilityCount++;
  }
  const beamStabilityScore = stabilityCount > 0
    ? Math.round((stabilitySum / stabilityCount) * 100) : 50;
  metrics.push({
    id: 'beam_stability', name: 'Beam Stability',
    score: beamStabilityScore,
    description: 'Consistency of retained path quality across weeks',
  });

  // 8. Search Stability: consistency of improvement rate across the run.
  // Based on: whether improvements happen evenly or in bursts.
  if (discoveries.length > 1) {
    const halfIdx = Math.floor(discoveries.length / 2);
    const firstHalf = discoveries.filter(d => d.eventType === 'global_best').slice(0, halfIdx).length;
    const secondHalf = discoveries.filter(d => d.eventType === 'global_best').slice(halfIdx).length;
    const total = firstHalf + secondHalf;
    const balance = total > 0 ? Math.min(firstHalf, secondHalf) / Math.max(firstHalf, secondHalf, 1) : 0.5;
    metrics.push({
      id: 'search_stability', name: 'Search Stability',
      score: Math.round(balance * 100),
      description: 'Balance of improvement discovery across the run duration',
    });
  } else {
    metrics.push({
      id: 'search_stability', name: 'Search Stability',
      score: 50,
      description: 'Balance of improvement discovery across the run duration',
    });
  }

  return metrics;
}

function generateNarrative(metrics: DNAMetric[], summary: RunSummary): string {
  const sorted = [...metrics].sort((a, b) => b.score - a.score);
  const top = sorted[0];
  const bottom = sorted[sorted.length - 1];

  const parts: string[] = [];

  // Opening.
  const overall = Math.round(metrics.reduce((s, m) => s + m.score, 0) / metrics.length);
  if (overall >= 70) parts.push('This run performed well overall.');
  else if (overall >= 50) parts.push('This run showed mixed performance.');
  else parts.push('This run struggled in several areas.');

  // Strength.
  if (top.score >= 70) {
    parts.push(`Its strongest trait was ${top.name.toLowerCase()} (${top.score}/100).`);
  }

  // Weakness.
  if (bottom.score < 40) {
    parts.push(`The weakest area was ${bottom.name.toLowerCase()} (${bottom.score}/100).`);
  }

  // Specific observations.
  const exploration = metrics.find(m => m.id === 'exploration');
  const convergence = metrics.find(m => m.id === 'convergence');
  const lineage = metrics.find(m => m.id === 'lineage');

  if (exploration && exploration.score >= 70 && convergence && convergence.score >= 60) {
    parts.push('It explored aggressively early and converged steadily.');
  } else if (exploration && exploration.score < 30) {
    parts.push('The search was conservative, focusing on local improvements.');
  }

  if (lineage && lineage.score >= 70) {
    parts.push('Good diversity was maintained throughout the beam.');
  } else if (lineage && lineage.score < 30) {
    parts.push('Lineage diversity collapsed — a single family dominated.');
  }

  if (convergence && convergence.score < 20) {
    parts.push('The best solution was found very late — more compute time may help.');
  }

  return parts.join(' ');
}

function metricColor(score: number): string {
  if (score >= 70) return 'text-emerald-400';
  if (score >= 50) return 'text-blue-400';
  if (score >= 30) return 'text-amber-400';
  return 'text-red-400';
}

function barBg(score: number): string {
  if (score >= 70) return 'bg-emerald-500';
  if (score >= 50) return 'bg-blue-500';
  if (score >= 30) return 'bg-amber-500';
  return 'bg-red-500';
}

export default function SearchDNA({ summary, discoveries, workers, tree, diversity }: Props) {
  const metrics = useMemo(
    () => computeMetrics({ summary, discoveries, workers, tree, diversity }),
    [summary, discoveries, workers, tree, diversity]
  );

  const narrative = useMemo(() => generateNarrative(metrics, summary), [metrics, summary]);
  const overall = Math.round(metrics.reduce((s, m) => s + m.score, 0) / metrics.length);

  return (
    <div className="space-y-4">
      {/* Overall score + narrative */}
      <Card title="Search DNA Profile">
        <div className="flex items-start gap-6">
          <div className="text-center shrink-0">
            <p className={`text-4xl font-bold ${metricColor(overall)}`}>{overall}</p>
            <p className="text-[10px] text-gray-500 mt-1">Overall Score</p>
          </div>
          <p className="text-sm text-gray-300 leading-relaxed">{narrative}</p>
        </div>
      </Card>

      {/* Radar chart (SVG) */}
      <Card title="DNA Radar">
        <div className="flex justify-center">
          <svg viewBox="0 0 300 300" className="w-72 h-72">
            {/* Background rings */}
            {[25, 50, 75, 100].map(ring => (
              <polygon
                key={ring}
                points={metrics.map((_, i) => {
                  const angle = (Math.PI * 2 * i) / metrics.length - Math.PI / 2;
                  const r = (ring / 100) * 120;
                  return `${150 + r * Math.cos(angle)},${150 + r * Math.sin(angle)}`;
                }).join(' ')}
                fill="none"
                stroke="#374151"
                strokeWidth="0.5"
              />
            ))}
            {/* Axis lines */}
            {metrics.map((_, i) => {
              const angle = (Math.PI * 2 * i) / metrics.length - Math.PI / 2;
              return (
                <line
                  key={i}
                  x1="150" y1="150"
                  x2={150 + 120 * Math.cos(angle)}
                  y2={150 + 120 * Math.sin(angle)}
                  stroke="#374151" strokeWidth="0.5"
                />
              );
            })}
            {/* Data polygon */}
            <polygon
              points={metrics.map((m, i) => {
                const angle = (Math.PI * 2 * i) / metrics.length - Math.PI / 2;
                const r = (m.score / 100) * 120;
                return `${150 + r * Math.cos(angle)},${150 + r * Math.sin(angle)}`;
              }).join(' ')}
              fill="rgba(59, 130, 246, 0.2)"
              stroke="#3b82f6"
              strokeWidth="2"
            />
            {/* Data points + labels */}
            {metrics.map((m, i) => {
              const angle = (Math.PI * 2 * i) / metrics.length - Math.PI / 2;
              const r = (m.score / 100) * 120;
              const labelR = 135;
              return (
                <g key={m.id}>
                  <circle
                    cx={150 + r * Math.cos(angle)}
                    cy={150 + r * Math.sin(angle)}
                    r="4" fill="#3b82f6"
                  />
                  <text
                    x={150 + labelR * Math.cos(angle)}
                    y={150 + labelR * Math.sin(angle)}
                    textAnchor="middle" dominantBaseline="middle"
                    className="fill-gray-400 text-[9px]"
                  >
                    {m.name.split(' ')[0]}
                  </text>
                </g>
              );
            })}
          </svg>
        </div>
      </Card>

      {/* Progress bars */}
      <Card title="Metric Breakdown">
        <div className="space-y-3">
          {metrics.map(m => (
            <div key={m.id}>
              <div className="flex justify-between items-center mb-1">
                <span className="text-xs text-gray-300">{m.name}</span>
                <span className={`text-xs font-bold ${metricColor(m.score)}`}>{m.score}/100</span>
              </div>
              <div className="h-2.5 bg-gray-800 rounded-full overflow-hidden">
                <div
                  className={`h-full ${barBg(m.score)} rounded-full transition-all duration-500`}
                  style={{ width: `${m.score}%` }}
                />
              </div>
              <p className="text-[9px] text-gray-600 mt-0.5">{m.description}</p>
            </div>
          ))}
        </div>
      </Card>

      {/* Summary cards grid */}
      <Card title="Key Numbers">
        <div className="grid grid-cols-4 gap-3 text-center">
          <div>
            <p className="text-lg font-bold text-emerald-400">{summary.totalPenalty.toLocaleString()}</p>
            <p className="text-[9px] text-gray-500">Total Penalty</p>
          </div>
          <div>
            <p className="text-lg font-bold text-blue-400">{summary.totalWorkers}</p>
            <p className="text-[9px] text-gray-500">Workers</p>
          </div>
          <div>
            <p className="text-lg font-bold text-purple-400">
              {discoveries.filter(d => d.eventType === 'global_best').length}
            </p>
            <p className="text-[9px] text-gray-500">Global Bests</p>
          </div>
          <div>
            <p className="text-lg font-bold text-amber-400">
              {(summary.totalDurationMs / 1000).toFixed(1)}s
            </p>
            <p className="text-[9px] text-gray-500">Runtime</p>
          </div>
        </div>
      </Card>
    </div>
  );
}
