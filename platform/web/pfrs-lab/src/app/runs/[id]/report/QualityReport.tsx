'use client';

import { useMemo } from 'react';
import Card from '@/components/Card';
import { RunSummary, DiscoveryRecord, WorkerLifecycle, TreeNode, DiversityRecord, PlateauEvent } from '@/lib/types';

interface Props {
  summary: RunSummary;
  discoveries: DiscoveryRecord[];
  workers: WorkerLifecycle[];
  tree: TreeNode[];
  diversity: DiversityRecord[];
  plateaus: PlateauEvent[];
  runId: string;
}

interface Recommendation {
  action: string;
  reason: string;
  confidence: 'high' | 'moderate' | 'low';
  metric: string;
}

function confidenceColor(c: string): string {
  switch (c) { case 'high': return 'text-emerald-400'; case 'moderate': return 'text-amber-400'; default: return 'text-gray-500'; }
}

function sectionGrade(score: number): { grade: string; color: string } {
  if (score >= 80) return { grade: 'A', color: 'text-emerald-400' };
  if (score >= 60) return { grade: 'B', color: 'text-blue-400' };
  if (score >= 40) return { grade: 'C', color: 'text-amber-400' };
  if (score >= 20) return { grade: 'D', color: 'text-orange-400' };
  return { grade: 'F', color: 'text-red-400' };
}

function generateRecommendations(
  summary: RunSummary, discoveries: DiscoveryRecord[], workers: WorkerLifecycle[],
  tree: TreeNode[], diversity: DiversityRecord[], plateaus: PlateauEvent[]
): Recommendation[] {
  const recs: Recommendation[] = [];
  const s = summary;

  // Beam width analysis.
  const retained = tree.filter(t => t.retained);
  const parentMap = new Map<number, number>();
  for (const t of tree) parentMap.set(t.pathID, t.parentID);
  const maxWeek = Math.max(...tree.map(t => t.week), 0);
  const finalRetained = retained.filter(t => t.week === maxWeek);
  const families = new Set<number>();
  for (const t of finalRetained) {
    let cur = t.pathID; let iter = 0;
    while (parentMap.has(cur) && parentMap.get(cur)! >= 0 && iter < 100) { cur = parentMap.get(cur)!; iter++; }
    families.add(cur);
  }
  if (families.size <= 1 && finalRetained.length > 1) {
    recs.push({ action: 'Increase Beam Width', reason: `Only ${families.size} family survived to final week — diversity collapsed`, confidence: 'high', metric: `${families.size} surviving families` });
  }

  // Worker utilisation.
  const useful = workers.filter(w => w.bestPenalty < w.startPenalty).length;
  const utilRate = workers.length > 0 ? useful / workers.length : 0;
  if (utilRate < 0.4) {
    recs.push({ action: 'Reduce Total Workers', reason: `${((1-utilRate)*100).toFixed(0)}% of workers never improved — compute is wasted`, confidence: 'moderate', metric: `${useful}/${workers.length} useful` });
  }

  // Plateau analysis.
  const longPlateaus = plateaus.filter(p => p.candsSinceImprove > 5000);
  if (longPlateaus.length > plateaus.length * 0.3) {
    recs.push({ action: 'Lower Cooling Rate', reason: `${((longPlateaus.length/plateaus.length)*100).toFixed(0)}% of plateaus are long (>5K iterations) — temperature may drop too fast`, confidence: 'moderate', metric: `${longPlateaus.length} long plateaus` });
  }

  // Convergence speed.
  const globalBests = discoveries.filter(d => d.eventType === 'global_best');
  if (globalBests.length > 0) {
    const lastGlobal = globalBests[globalBests.length - 1];
    const totalTime = s.totalDurationMs;
    if (lastGlobal.elapsedMs > totalTime * 0.9) {
      recs.push({ action: 'Increase Time Budget', reason: 'Best solution found in last 10% of runtime — more time may yield improvements', confidence: 'high', metric: `best at ${(lastGlobal.elapsedMs/1000).toFixed(1)}s of ${(totalTime/1000).toFixed(1)}s` });
    }
    if (lastGlobal.elapsedMs < totalTime * 0.3) {
      recs.push({ action: 'Maintain Current Configuration', reason: 'Best solution found early — search converged efficiently', confidence: 'moderate', metric: `best at ${(lastGlobal.elapsedMs/1000).toFixed(1)}s` });
    }
  }

  // Diversity check.
  const nearDups = diversity.filter(d => d.nearDuplicate).length;
  if (nearDups > diversity.length * 0.4) {
    recs.push({ action: 'Enable Diversity Slots', reason: `${((nearDups/diversity.length)*100).toFixed(0)}% of beam paths are near-duplicates`, confidence: 'high', metric: `${nearDups} near-duplicates` });
  }

  // Refinement suggestion.
  const worstWeek = s.weeks.reduce((w, c) => c.finalPenalty > w.finalPenalty ? c : w, s.weeks[0]);
  if (worstWeek && worstWeek.finalPenalty > s.totalPenalty * 0.2) {
    recs.push({ action: 'Enable Refinement', reason: `Week ${worstWeek.week} contributes ${((worstWeek.finalPenalty/s.totalPenalty)*100).toFixed(0)}% of total penalty — post-processing may help`, confidence: 'low', metric: `W${worstWeek.week}: ${worstWeek.finalPenalty}` });
  }

  if (recs.length === 0) {
    recs.push({ action: 'Maintain Current Configuration', reason: 'No obvious improvement opportunities detected from telemetry', confidence: 'moderate', metric: 'all metrics nominal' });
  }

  return recs;
}

export default function QualityReport({ summary, discoveries, workers, tree, diversity, plateaus, runId }: Props) {
  const s = summary;
  const recs = useMemo(() => generateRecommendations(summary, discoveries, workers, tree, diversity, plateaus), [summary, discoveries, workers, tree, diversity, plateaus]);

  // Section scores.
  const globalBests = discoveries.filter(d => d.eventType === 'global_best').length;
  const useful = workers.filter(w => w.bestPenalty < w.startPenalty).length;
  const utilRate = workers.length > 0 ? (useful / workers.length) * 100 : 50;
  const retained = tree.filter(t => t.retained);
  const nearDups = diversity.filter(d => d.nearDuplicate).length;
  const dupRate = diversity.length > 0 ? (1 - nearDups / diversity.length) * 100 : 50;

  const searchScore = Math.min(100, Math.round(globalBests * 5 + s.acceptWorseRate * 10));
  const beamScore = Math.min(100, Math.round(dupRate));
  const workerScore = Math.min(100, Math.round(utilRate));
  const overallScore = Math.round((searchScore + beamScore + workerScore) / 3);

  return (
    <div className="space-y-4">
      {/* Header */}
      <Card title={`Quality Report: ${runId}`}>
        <div className="flex items-center gap-6">
          <div className="text-center">
            <p className={`text-4xl font-bold ${sectionGrade(overallScore).color}`}>{sectionGrade(overallScore).grade}</p>
            <p className="text-[10px] text-gray-500">Overall ({overallScore}/100)</p>
          </div>
          <div className="grid grid-cols-3 gap-4 flex-1">
            <div className="text-center">
              <p className={`text-xl font-bold ${sectionGrade(searchScore).color}`}>{sectionGrade(searchScore).grade}</p>
              <p className="text-[9px] text-gray-500">Search</p>
            </div>
            <div className="text-center">
              <p className={`text-xl font-bold ${sectionGrade(beamScore).color}`}>{sectionGrade(beamScore).grade}</p>
              <p className="text-[9px] text-gray-500">Beam</p>
            </div>
            <div className="text-center">
              <p className={`text-xl font-bold ${sectionGrade(workerScore).color}`}>{sectionGrade(workerScore).grade}</p>
              <p className="text-[9px] text-gray-500">Workers</p>
            </div>
          </div>
        </div>
      </Card>

      {/* Summary section */}
      <Card title="§1 Summary">
        <div className="grid grid-cols-4 gap-3 text-center text-xs">
          <div><p className="text-lg font-bold text-emerald-400">{s.totalPenalty.toLocaleString()}</p><p className="text-[9px] text-gray-500">Penalty</p></div>
          <div><p className="text-lg font-bold text-blue-400">{s.numWeeks}</p><p className="text-[9px] text-gray-500">Weeks</p></div>
          <div><p className="text-lg font-bold text-gray-300">{(s.totalDurationMs/1000).toFixed(1)}s</p><p className="text-[9px] text-gray-500">Runtime</p></div>
          <div><p className="text-lg font-bold text-purple-400">{s.metadata?.mode?.toUpperCase() || '—'}</p><p className="text-[9px] text-gray-500">Algorithm</p></div>
        </div>
      </Card>

      {/* Search Behaviour */}
      <Card title="§2 Search Behaviour">
        <div className="grid grid-cols-3 gap-3 text-xs">
          <div><p className="text-gray-500">Global Bests Found</p><p className="font-bold text-yellow-400">{globalBests}</p></div>
          <div><p className="text-gray-500">Total Candidates</p><p className="font-mono">{(s.totalCandidates/1_000_000).toFixed(1)}M</p></div>
          <div><p className="text-gray-500">Accept Worse Rate</p><p>{s.acceptWorseRate.toFixed(2)}%</p></div>
          <div><p className="text-gray-500">Hard Reject Rate</p><p>{s.hardRejectRate.toFixed(1)}%</p></div>
          <div><p className="text-gray-500">Improvement/M Cands</p><p>{s.totalCandidates > 0 ? (s.weeks.reduce((sum,w)=>sum+w.improvement,0)/(s.totalCandidates/1_000_000)).toFixed(1) : '—'}</p></div>
          <div><p className="text-gray-500">Worst Week</p><p className="text-red-400">W{s.maxWeekNum}: {s.maxWeekPenalty}</p></div>
        </div>
      </Card>

      {/* Beam Behaviour */}
      <Card title="§3 Beam Behaviour">
        <div className="grid grid-cols-3 gap-3 text-xs">
          <div><p className="text-gray-500">Beam Width</p><p className="font-bold">{s.metadata?.beamWidth || '—'}</p></div>
          <div><p className="text-gray-500">Total Paths</p><p>{tree.length}</p></div>
          <div><p className="text-gray-500">Retained</p><p>{retained.length}</p></div>
          <div><p className="text-gray-500">Near-Duplicates</p><p className={nearDups > diversity.length*0.3 ? 'text-red-400' : 'text-gray-300'}>{nearDups}</p></div>
          <div><p className="text-gray-500">Diversity Rate</p><p>{dupRate.toFixed(0)}%</p></div>
          <div><p className="text-gray-500">Winning Path</p><p className="text-yellow-400">{tree.find(t=>t.winning)?.pathID ?? '—'}</p></div>
        </div>
      </Card>

      {/* Worker Behaviour */}
      <Card title="§4 Worker Behaviour">
        <div className="grid grid-cols-3 gap-3 text-xs">
          <div><p className="text-gray-500">Total Workers</p><p>{workers.length}</p></div>
          <div><p className="text-gray-500">Useful</p><p className="text-emerald-400">{useful} ({utilRate.toFixed(0)}%)</p></div>
          <div><p className="text-gray-500">Global Best Finders</p><p className="text-yellow-400">{workers.filter(w=>w.producedGlobalBest).length}</p></div>
          <div><p className="text-gray-500">Avg Lifetime</p><p>{workers.length > 0 ? (workers.reduce((s,w)=>s+(w.finishTimeMs-w.startTimeMs),0)/workers.length/1000).toFixed(2) : '—'}s</p></div>
          <div><p className="text-gray-500">Total Branches</p><p>{s.totalBranches}</p></div>
          <div><p className="text-gray-500">Branches/Worker</p><p>{workers.length > 0 ? (s.totalBranches/workers.length).toFixed(1) : '—'}</p></div>
        </div>
      </Card>

      {/* Penalty Analysis */}
      <Card title="§5 Penalty Analysis">
        <div className="flex items-end gap-1 h-20 mb-2">
          {s.weeks.map(w => {
            const maxP = s.maxWeekPenalty || 1;
            const height = (w.finalPenalty / maxP) * 100;
            const isWorst = w.finalPenalty === s.maxWeekPenalty;
            return (
              <div key={w.week} className="flex-1 flex flex-col items-center justify-end h-full">
                <span className="text-[8px] text-gray-600">{w.finalPenalty}</span>
                <div className={`w-full rounded-t ${isWorst ? 'bg-red-500' : 'bg-blue-600'}`}
                  style={{ height: `${Math.max(height, 4)}%` }} />
                <span className="text-[8px] text-gray-600 mt-0.5">W{w.week}</span>
              </div>
            );
          })}
        </div>
        <p className="text-[10px] text-gray-500">
          Avg: {(s.totalPenalty / s.numWeeks).toFixed(0)} | Max: W{s.maxWeekNum} ({s.maxWeekPenalty}) |
          Spread: {s.maxWeekPenalty - Math.min(...s.weeks.map(w=>w.finalPenalty))}
        </p>
      </Card>

      {/* Timeline Highlights */}
      <Card title="§6 Timeline Highlights">
        <div className="space-y-1 text-xs">
          {globalBests > 0 && (
            <p>🏆 First global best at {(discoveries.find(d=>d.eventType==='global_best')?.elapsedMs||0)/1000}s</p>
          )}
          {globalBests > 0 && (
            <p>🏆 Last global best at {(discoveries.filter(d=>d.eventType==='global_best').slice(-1)[0]?.elapsedMs||0)/1000}s</p>
          )}
          <p>👷 {workers.length} workers across {s.numWeeks} weeks</p>
          <p>🏔️ {plateaus.length} plateau events recorded</p>
          {plateaus.length > 0 && (
            <p>⏱️ Longest plateau: {Math.max(...plateaus.map(p=>p.candsSinceImprove)).toLocaleString()} iterations</p>
          )}
        </div>
      </Card>

      {/* DNA Summary */}
      <Card title="§7 DNA Summary">
        <div className="grid grid-cols-3 gap-3">
          {[
            { label: 'Search', score: searchScore },
            { label: 'Beam', score: beamScore },
            { label: 'Workers', score: workerScore },
          ].map(m => {
            const g = sectionGrade(m.score);
            return (
              <div key={m.label}>
                <div className="flex justify-between text-[10px] mb-0.5">
                  <span className="text-gray-400">{m.label}</span>
                  <span className={g.color}>{m.score}/100</span>
                </div>
                <div className="h-2 bg-gray-800 rounded-full overflow-hidden">
                  <div className={`h-full rounded-full ${m.score >= 60 ? 'bg-emerald-500' : m.score >= 40 ? 'bg-amber-500' : 'bg-red-500'}`}
                    style={{ width: `${m.score}%` }} />
                </div>
              </div>
            );
          })}
        </div>
      </Card>

      {/* Recommendations */}
      <Card title="§8 Recommendations">
        <div className="space-y-3">
          {recs.map((rec, i) => (
            <div key={i} className="border border-gray-700 rounded p-3">
              <div className="flex items-start justify-between">
                <p className="text-sm font-medium text-gray-200">{rec.action}</p>
                <span className={`text-[10px] px-2 py-0.5 rounded ${confidenceColor(rec.confidence)} bg-gray-800`}>
                  {rec.confidence}
                </span>
              </div>
              <p className="text-xs text-gray-400 mt-1">{rec.reason}</p>
              <p className="text-[9px] text-gray-600 mt-0.5 font-mono">Metric: {rec.metric}</p>
            </div>
          ))}
        </div>
      </Card>
    </div>
  );
}
