'use client';

import { useMemo } from 'react';
import Card from '@/components/Card';
import Link from 'next/link';
import { RunSummary, DiscoveryRecord, WorkerLifecycle, TreeNode, DiversityRecord, PlateauEvent } from '@/lib/types';

interface Props {
  runId: string;
  summary: RunSummary;
  discoveries: DiscoveryRecord[];
  workers: WorkerLifecycle[];
  tree: TreeNode[];
  diversity: DiversityRecord[];
  plateaus: PlateauEvent[];
}

interface NarrativeLine {
  text: string;
  link: string;
  section: 'exploration' | 'diversity' | 'plateaus' | 'convergence' | 'workers' | 'result';
  icon: string;
}

interface Recommendation {
  text: string;
  reason: string;
  link: string;
  confidence: 'high' | 'moderate' | 'low';
}

function generateNarrative(props: Props): NarrativeLine[] {
  const { runId, summary: s, discoveries, workers, tree, diversity, plateaus } = props;
  const lines: NarrativeLine[] = [];
  const base = `/runs/${runId}`;

  // --- Exploration phase ---
  const globalBests = discoveries.filter(d => d.eventType === 'global_best');
  const totalTime = s.totalDurationMs;

  if (s.acceptWorseRate > 1) {
    lines.push({ text: `The optimiser began with strong exploration (accept-worse rate: ${s.acceptWorseRate.toFixed(2)}%).`, link: `${base}/search`, section: 'exploration', icon: '🧭' });
  } else if (s.acceptWorseRate > 0.1) {
    lines.push({ text: `The optimiser used moderate exploration (accept-worse rate: ${s.acceptWorseRate.toFixed(2)}%).`, link: `${base}/search`, section: 'exploration', icon: '🧭' });
  } else {
    lines.push({ text: `The optimiser was highly exploitative (accept-worse rate: ${s.acceptWorseRate.toFixed(3)}%).`, link: `${base}/search`, section: 'exploration', icon: '🎯' });
  }

  // --- Diversity ---
  const parentMap = new Map<number, number>();
  for (const t of tree) parentMap.set(t.pathID, t.parentID);
  const weeks = [...new Set(tree.map(t => t.week))].sort((a, b) => a - b);
  const entropies: number[] = [];
  for (const week of weeks) {
    const retained = tree.filter(t => t.week === week && t.retained);
    if (retained.length <= 1) { entropies.push(0); continue; }
    const fam = new Map<number, number>();
    for (const t of retained) { let c = t.pathID; let i = 0; while (parentMap.has(c) && parentMap.get(c)! >= 0 && i < 100) { c = parentMap.get(c)!; i++; } fam.set(c, (fam.get(c)||0)+1); }
    let e = 0; for (const cnt of fam.values()) { const p = cnt/retained.length; if (p > 0) e -= p * Math.log2(p); }
    entropies.push(e);
  }

  if (entropies.length > 2) {
    const highEntropyWeeks = entropies.filter(e => e > 0.7).length;
    if (highEntropyWeeks > 0) {
      const lastHigh = entropies.lastIndexOf(entropies.filter(e => e > 0.7).pop() || 0);
      lines.push({ text: `Entropy remained above 0.7 until Week ${lastHigh + 1}.`, link: `${base}/families`, section: 'diversity', icon: '🌿' });
    }
    const collapseWeek = entropies.findIndex(e => e < 0.3);
    if (collapseWeek >= 0) {
      lines.push({ text: `Diversity collapsed in Week ${collapseWeek + 1} (entropy dropped to ${entropies[collapseWeek].toFixed(2)}).`, link: `${base}/families`, section: 'diversity', icon: '⚠️' });
    }
  }

  // Near-duplicate check.
  const nearDups = diversity.filter(d => d.nearDuplicate).length;
  if (diversity.length > 0 && nearDups / diversity.length > 0.3) {
    lines.push({ text: `${((nearDups/diversity.length)*100).toFixed(0)}% of beam paths were near-duplicates.`, link: `${base}/diversity`, section: 'diversity', icon: '🔁' });
  }

  // --- Plateaus ---
  if (plateaus.length > 0) {
    const longest = Math.max(...plateaus.map(p => p.candsSinceImprove));
    const weekCounts = new Map<number, number>();
    for (const p of plateaus) weekCounts.set(p.week, (weekCounts.get(p.week) || 0) + 1);
    let worstWeek = 0, worstCount = 0;
    for (const [w, c] of weekCounts) { if (c > worstCount) { worstWeek = w; worstCount = c; } }

    if (longest > 5000) {
      lines.push({ text: `A major plateau occurred (${longest.toLocaleString()} iterations without improvement).`, link: `${base}/plateaus`, section: 'plateaus', icon: '🏔️' });
    }
    if (worstCount > plateaus.length * 0.3) {
      lines.push({ text: `Week ${worstWeek} contained ${((worstCount/plateaus.length)*100).toFixed(0)}% of all plateaus.`, link: `${base}/plateaus`, section: 'plateaus', icon: '🏔️' });
    }
  }

  // --- Convergence ---
  if (globalBests.length > 0) {
    const first = globalBests[0];
    const last = globalBests[globalBests.length - 1];
    const earlyBests = globalBests.filter(d => d.elapsedMs < totalTime * 0.3).length;
    const lateBests = globalBests.filter(d => d.elapsedMs > totalTime * 0.7).length;

    lines.push({ text: `The first global best was found at ${(first.elapsedMs/1000).toFixed(1)}s (penalty ${first.newBest}).`, link: `${base}/map`, section: 'convergence', icon: '🏆' });

    if (lateBests > 0) {
      lines.push({ text: `Late refinement produced ${lateBests} additional global best improvement${lateBests > 1 ? 's' : ''}.`, link: `${base}/search`, section: 'convergence', icon: '🔧' });
    }

    if (earlyBests > globalBests.length * 0.7) {
      lines.push({ text: `Most improvements (${earlyBests}/${globalBests.length}) occurred in the first 30% of runtime.`, link: `${base}/timeline`, section: 'convergence', icon: '⚡' });
    }

    // Winning lineage detection.
    const winningNodes = tree.filter(t => t.winning);
    if (winningNodes.length > 0) {
      const winWeek = Math.min(...winningNodes.map(t => t.week));
      lines.push({ text: `The winning lineage emerged in Week ${winWeek}.`, link: `${base}/genealogy`, section: 'convergence', icon: '👑' });
    }
  }

  // --- Workers ---
  const useful = workers.filter(w => w.bestPenalty < w.startPenalty).length;
  const globalFinders = workers.filter(w => w.producedGlobalBest).length;
  lines.push({ text: `${useful} of ${workers.length} workers (${((useful/workers.length)*100).toFixed(0)}%) contributed improvements.`, link: `${base}/workers`, section: 'workers', icon: '👷' });
  if (globalFinders > 0) {
    lines.push({ text: `${globalFinders} worker${globalFinders > 1 ? 's' : ''} found global bests.`, link: `${base}/archetypes`, section: 'workers', icon: '⭐' });
  }

  // --- Final result ---
  // Compute beam health approximation.
  const maxEntropy = Math.max(...entropies, 0);
  const beamHealth = Math.min(100, Math.round(maxEntropy * 30 + (1 - nearDups / Math.max(diversity.length, 1)) * 50 + (globalFinders / Math.max(workers.length, 1)) * 200));

  lines.push({ text: `The final solution achieved penalty ${s.totalPenalty.toLocaleString()} with Beam Health ${beamHealth}.`, link: `${base}/report`, section: 'result', icon: '📊' });

  return lines;
}

function generateRecommendations(props: Props): Recommendation[] {
  const { runId, summary: s, discoveries, workers, tree, diversity, plateaus } = props;
  const recs: Recommendation[] = [];
  const base = `/runs/${runId}`;

  const globalBests = discoveries.filter(d => d.eventType === 'global_best');
  const totalTime = s.totalDurationMs;
  const useful = workers.filter(w => w.bestPenalty < w.startPenalty).length;
  const utilRate = workers.length > 0 ? useful / workers.length : 0;
  const nearDups = diversity.filter(d => d.nearDuplicate).length;
  const dupRate = diversity.length > 0 ? nearDups / diversity.length : 0;
  const lastGlobal = globalBests.length > 0 ? globalBests[globalBests.length - 1] : null;

  if (dupRate > 0.4) {
    recs.push({ text: 'Consider increasing beam width.', reason: `${(dupRate*100).toFixed(0)}% near-duplicate paths suggest insufficient diversity.`, link: `${base}/diversity`, confidence: 'high' });
  }
  if (utilRate < 0.4) {
    recs.push({ text: 'Consider reducing total workers.', reason: `${((1-utilRate)*100).toFixed(0)}% compute wasted on unproductive workers.`, link: `${base}/efficiency`, confidence: 'moderate' });
  }
  if (lastGlobal && lastGlobal.elapsedMs > totalTime * 0.9) {
    recs.push({ text: 'Consider increasing time budget.', reason: `Best found at ${(lastGlobal.elapsedMs/totalTime*100).toFixed(0)}% through runtime.`, link: `${base}/timeline`, confidence: 'high' });
  }
  if (plateaus.length > 0) {
    const longPlateaus = plateaus.filter(p => p.candsSinceImprove > 5000);
    if (longPlateaus.length > plateaus.length * 0.3) {
      recs.push({ text: 'Consider lowering cooling rate.', reason: `${longPlateaus.length} long plateaus suggest premature convergence.`, link: `${base}/plateaus`, confidence: 'moderate' });
    }
  }
  if (s.metadata?.mode === 'sa' && s.acceptWorseRate < 0.05) {
    recs.push({ text: 'Consider raising initial temperature.', reason: `Accept-worse rate (${s.acceptWorseRate.toFixed(3)}%) is very low.`, link: `${base}/temperature`, confidence: 'low' });
  }
  if (recs.length === 0) {
    recs.push({ text: 'Maintain current configuration.', reason: 'No clear improvement opportunities detected.', link: `${base}/report`, confidence: 'moderate' });
  }

  return recs;
}

const SECTION_LABELS: Record<string, { label: string; color: string }> = {
  exploration: { label: 'Exploration', color: 'border-amber-600' },
  diversity: { label: 'Diversity', color: 'border-blue-600' },
  plateaus: { label: 'Plateaus', color: 'border-orange-600' },
  convergence: { label: 'Convergence', color: 'border-emerald-600' },
  workers: { label: 'Workers', color: 'border-purple-600' },
  result: { label: 'Result', color: 'border-gray-500' },
};

export default function ExplainRun(props: Props) {
  const narrative = useMemo(() => generateNarrative(props), [props]);
  const recommendations = useMemo(() => generateRecommendations(props), [props]);

  // Group narrative by section.
  const sections = useMemo(() => {
    const grouped = new Map<string, NarrativeLine[]>();
    for (const line of narrative) {
      const existing = grouped.get(line.section) || [];
      existing.push(line);
      grouped.set(line.section, existing);
    }
    return grouped;
  }, [narrative]);

  return (
    <div className="space-y-4">
      <Card title="Explain This Run">
        <p className="text-xs text-gray-500 mb-4">
          Every statement below is derived from measured telemetry. Click links to see supporting data.
        </p>

        {/* Narrative sections */}
        {Array.from(sections.entries()).map(([section, lines]) => {
          const cfg = SECTION_LABELS[section] || { label: section, color: 'border-gray-600' };
          return (
            <div key={section} className={`border-l-2 ${cfg.color} pl-4 mb-4`}>
              <p className="text-[10px] text-gray-500 uppercase mb-1">{cfg.label}</p>
              <div className="space-y-1.5">
                {lines.map((line, i) => (
                  <div key={i} className="flex items-start gap-2">
                    <span className="shrink-0">{line.icon}</span>
                    <p className="text-sm text-gray-200 leading-relaxed">
                      {line.text}
                      <Link href={line.link} className="ml-2 text-blue-400 hover:text-blue-300 text-[10px]">
                        [view →]
                      </Link>
                    </p>
                  </div>
                ))}
              </div>
            </div>
          );
        })}
      </Card>

      {/* Recommendations */}
      <Card title="Recommendations">
        <div className="space-y-3">
          {recommendations.map((rec, i) => (
            <div key={i} className="flex items-start gap-3 border border-gray-700 rounded p-3">
              <span className={`text-[10px] px-2 py-0.5 rounded shrink-0 ${
                rec.confidence === 'high' ? 'bg-emerald-900/30 text-emerald-400' :
                rec.confidence === 'moderate' ? 'bg-amber-900/30 text-amber-400' :
                'bg-gray-800 text-gray-500'
              }`}>{rec.confidence}</span>
              <div>
                <p className="text-sm text-gray-200 font-medium">{rec.text}</p>
                <p className="text-[10px] text-gray-500 mt-0.5">
                  {rec.reason}
                  <Link href={rec.link} className="ml-2 text-blue-400 hover:text-blue-300">[evidence →]</Link>
                </p>
              </div>
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}
