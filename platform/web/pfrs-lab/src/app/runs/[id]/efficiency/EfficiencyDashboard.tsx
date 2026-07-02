'use client';

import { useMemo } from 'react';
import Card from '@/components/Card';
import { WorkerLifecycle, DiscoveryRecord, RunSummary } from '@/lib/types';

interface Props {
  workers: WorkerLifecycle[];
  discoveries: DiscoveryRecord[];
  summary: RunSummary;
}

interface WorkerEfficiency {
  workerID: number;
  week: number;
  runtimeMs: number;
  discoveries: number;
  globalBests: number;
  improvement: number;
  branches: number;
  candidates: number;
  improvementPerSec: number;
  discoveryPerSec: number;
  useful: boolean;
}

export default function EfficiencyDashboard({ workers, discoveries, summary }: Props) {
  const workerStats = useMemo(() => {
    const discMap = new Map<number, { local: number; global: number; totalImp: number }>();
    for (const d of discoveries) {
      const e = discMap.get(d.workerID) || { local: 0, global: 0, totalImp: 0 };
      if (d.eventType === 'local_best') e.local++;
      if (d.eventType === 'global_best') e.global++;
      e.totalImp += d.improvement;
      discMap.set(d.workerID, e);
    }

    return workers.map(w => {
      const d = discMap.get(w.workerID) || { local: 0, global: 0, totalImp: 0 };
      const runtimeMs = w.finishTimeMs - w.startTimeMs;
      const runtimeSec = runtimeMs / 1000;
      const totalDisc = d.local + d.global;
      return {
        workerID: w.workerID,
        week: w.week,
        runtimeMs,
        discoveries: totalDisc,
        globalBests: d.global,
        improvement: d.totalImp || (w.startPenalty - w.bestPenalty),
        branches: w.branchCount,
        candidates: w.finishCandidate,
        improvementPerSec: runtimeSec > 0 ? d.totalImp / runtimeSec : 0,
        discoveryPerSec: runtimeSec > 0 ? totalDisc / runtimeSec : 0,
        useful: w.bestPenalty < w.startPenalty,
      } as WorkerEfficiency;
    });
  }, [workers, discoveries]);

  // Aggregate metrics.
  const total = workerStats.length;
  const useful = workerStats.filter(w => w.useful).length;
  const neverImproved = total - useful;
  const producedGlobal = workerStats.filter(w => w.globalBests > 0).length;
  const producedLocal = workerStats.filter(w => w.discoveries > 0).length;
  const totalRuntimeMs = workerStats.reduce((s, w) => s + w.runtimeMs, 0);
  const totalDiscoveries = workerStats.reduce((s, w) => s + w.discoveries, 0);
  const totalGlobalBests = workerStats.reduce((s, w) => s + w.globalBests, 0);
  const totalImprovement = workerStats.reduce((s, w) => s + w.improvement, 0);
  const totalBranches = workerStats.reduce((s, w) => s + w.branches, 0);

  const improvPerWorker = total > 0 ? (totalImprovement / total).toFixed(1) : '0';
  const improvPerSec = totalRuntimeMs > 0 ? (totalImprovement / (totalRuntimeMs / 1000)).toFixed(2) : '0';
  const branchesPerWorker = total > 0 ? (totalBranches / total).toFixed(1) : '0';
  const globalPerCpuSec = totalRuntimeMs > 0 ? (totalGlobalBests / (totalRuntimeMs / 1000)).toFixed(4) : '0';
  const discoveryEff = total > 0 ? ((producedLocal / total) * 100).toFixed(1) : '0';
  const beamEff = total > 0 ? ((producedGlobal / total) * 100).toFixed(1) : '0';

  // CPU utilisation estimate (wall time vs sum of worker time).
  const wallTimeMs = summary.totalDurationMs;
  const cpuUtil = wallTimeMs > 0 ? Math.min(100, (totalRuntimeMs / wallTimeMs / (summary.metadata?.cpus || 1)) * 100) : 0;

  // Identify notable workers.
  const sorted = [...workerStats].sort((a, b) => b.improvement - a.improvement);
  const mostProductive = sorted[0];
  const leastProductive = [...workerStats].sort((a, b) => a.improvement - b.improvement)[0];

  // Most expensive discovery: highest runtime for a single global best.
  const globalWorkers = workerStats.filter(w => w.globalBests > 0);
  const mostExpensive = [...globalWorkers].sort((a, b) => b.runtimeMs - a.runtimeMs)[0];
  const bestROI = [...globalWorkers].sort((a, b) => (b.improvement / Math.max(b.runtimeMs, 1)) - (a.improvement / Math.max(a.runtimeMs, 1)))[0];

  // Histogram: improvement distribution.
  const improvements = workerStats.map(w => w.improvement).filter(i => i > 0);
  const histBins = useMemo(() => {
    if (improvements.length === 0) return [];
    const max = Math.max(...improvements);
    const bins = Array(20).fill(0);
    const binSize = max / 20;
    for (const imp of improvements) {
      const bin = Math.min(Math.floor(imp / binSize), 19);
      bins[bin]++;
    }
    return bins;
  }, [improvements]);

  function formatMs(ms: number): string {
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(1)}s`;
  }

  return (
    <div className="space-y-4">
      {/* Key metrics */}
      <div className="grid grid-cols-4 gap-3">
        <Card title="Workers Created"><p className="text-2xl font-bold text-blue-400">{total}</p></Card>
        <Card title="Useful Workers"><p className="text-2xl font-bold text-emerald-400">{useful} ({((useful/total)*100).toFixed(0)}%)</p></Card>
        <Card title="Never Improved"><p className="text-2xl font-bold text-red-400">{neverImproved}</p></Card>
        <Card title="CPU Utilisation"><p className="text-2xl font-bold text-amber-400">{cpuUtil.toFixed(1)}%</p></Card>
      </div>

      <div className="grid grid-cols-4 gap-3">
        <Card title="Improv/Worker"><p className="text-xl font-bold text-gray-300">{improvPerWorker}</p></Card>
        <Card title="Improv/Second"><p className="text-xl font-bold text-gray-300">{improvPerSec}</p></Card>
        <Card title="Branches/Worker"><p className="text-xl font-bold text-gray-300">{branchesPerWorker}</p></Card>
        <Card title="Globals/CPU·s"><p className="text-xl font-bold text-gray-300">{globalPerCpuSec}</p></Card>
      </div>

      <div className="grid grid-cols-3 gap-3">
        <Card title="Discovery Efficiency"><p className="text-xl font-bold text-purple-400">{discoveryEff}%</p><p className="text-[9px] text-gray-500">workers finding any improvement</p></Card>
        <Card title="Beam Efficiency"><p className="text-xl font-bold text-yellow-400">{beamEff}%</p><p className="text-[9px] text-gray-500">workers finding global best</p></Card>
        <Card title="Global Best Finders"><p className="text-xl font-bold text-yellow-400">{producedGlobal}</p></Card>
      </div>

      {/* Worker utilisation stacked bar */}
      <Card title="Worker Contribution">
        <div className="h-8 rounded-lg overflow-hidden flex mb-2">
          <div className="bg-yellow-500" style={{ width: `${(producedGlobal / total) * 100}%` }} title={`Global best: ${producedGlobal}`} />
          <div className="bg-emerald-600" style={{ width: `${((producedLocal - producedGlobal) / total) * 100}%` }} title={`Local only: ${producedLocal - producedGlobal}`} />
          <div className="bg-blue-700" style={{ width: `${((useful - producedLocal) / total) * 100}%` }} title={`Improved: ${useful - producedLocal}`} />
          <div className="bg-gray-700" style={{ width: `${(neverImproved / total) * 100}%` }} title={`Wasted: ${neverImproved}`} />
        </div>
        <div className="flex gap-3 text-[9px] text-gray-400">
          <span className="flex items-center gap-1"><span className="w-2 h-2 bg-yellow-500 rounded-sm" />Global ({producedGlobal})</span>
          <span className="flex items-center gap-1"><span className="w-2 h-2 bg-emerald-600 rounded-sm" />Local ({producedLocal - producedGlobal})</span>
          <span className="flex items-center gap-1"><span className="w-2 h-2 bg-blue-700 rounded-sm" />Improved ({useful - producedLocal})</span>
          <span className="flex items-center gap-1"><span className="w-2 h-2 bg-gray-700 rounded-sm" />Wasted ({neverImproved})</span>
        </div>
      </Card>

      {/* Improvement distribution histogram */}
      <Card title="Improvement Distribution per Worker">
        <div className="flex items-end gap-px h-24">
          {histBins.map((count, i) => {
            const maxBin = Math.max(...histBins, 1);
            return (
              <div key={i} className="flex-1">
                <div className="bg-emerald-500 rounded-t" style={{ height: `${(count / maxBin) * 100}%`, minHeight: count > 0 ? '2px' : '0' }} title={`${count} workers`} />
              </div>
            );
          })}
        </div>
        <div className="flex justify-between text-[9px] text-gray-600 mt-1">
          <span>0</span>
          <span>Max improvement</span>
        </div>
      </Card>

      {/* Scatter: runtime vs improvement */}
      <Card title="Runtime vs Improvement (per worker)">
        <svg viewBox="0 0 600 200" className="w-full h-48 bg-gray-900 rounded border border-gray-800">
          {workerStats.filter(w => w.improvement > 0).slice(0, 300).map((w, i) => {
            const maxRuntime = Math.max(...workerStats.map(ws => ws.runtimeMs), 1);
            const maxImp = Math.max(...workerStats.map(ws => ws.improvement), 1);
            const x = 40 + (w.runtimeMs / maxRuntime) * 520;
            const y = 180 - (w.improvement / maxImp) * 160;
            return (
              <circle key={i} cx={x} cy={y} r={w.globalBests > 0 ? 5 : 3}
                fill={w.globalBests > 0 ? '#fbbf24' : '#3b82f6'}
                opacity={0.7}
              >
                <title>{`W${w.week}#${w.workerID}: ${formatMs(w.runtimeMs)}, imp=${w.improvement}`}</title>
              </circle>
            );
          })}
          <text x="300" y="198" textAnchor="middle" className="fill-gray-600 text-[8px]">Runtime</text>
          <text x="10" y="100" textAnchor="middle" transform="rotate(-90, 10, 100)" className="fill-gray-600 text-[8px]">Improvement</text>
        </svg>
      </Card>

      {/* Notable workers */}
      <Card title="🏆 Notable Workers">
        <div className="grid grid-cols-2 gap-4">
          {mostProductive && (
            <div className="bg-gray-800/50 rounded p-3">
              <p className="text-[10px] text-gray-500 uppercase mb-1">Most Productive</p>
              <p className="text-sm font-mono text-emerald-400">Worker {mostProductive.workerID} (W{mostProductive.week})</p>
              <p className="text-[10px] text-gray-400">Improvement: {mostProductive.improvement} | {mostProductive.discoveries} discoveries | {formatMs(mostProductive.runtimeMs)}</p>
            </div>
          )}
          {leastProductive && (
            <div className="bg-gray-800/50 rounded p-3">
              <p className="text-[10px] text-gray-500 uppercase mb-1">Least Productive</p>
              <p className="text-sm font-mono text-red-400">Worker {leastProductive.workerID} (W{leastProductive.week})</p>
              <p className="text-[10px] text-gray-400">Improvement: {leastProductive.improvement} | {formatMs(leastProductive.runtimeMs)}</p>
            </div>
          )}
          {mostExpensive && (
            <div className="bg-gray-800/50 rounded p-3">
              <p className="text-[10px] text-gray-500 uppercase mb-1">Most Expensive Discovery</p>
              <p className="text-sm font-mono text-amber-400">Worker {mostExpensive.workerID} (W{mostExpensive.week})</p>
              <p className="text-[10px] text-gray-400">Runtime: {formatMs(mostExpensive.runtimeMs)} for {mostExpensive.globalBests} global best(s)</p>
            </div>
          )}
          {bestROI && (
            <div className="bg-gray-800/50 rounded p-3">
              <p className="text-[10px] text-gray-500 uppercase mb-1">Best Return on Compute</p>
              <p className="text-sm font-mono text-blue-400">Worker {bestROI.workerID} (W{bestROI.week})</p>
              <p className="text-[10px] text-gray-400">Improvement: {bestROI.improvement} in {formatMs(bestROI.runtimeMs)} ({bestROI.improvementPerSec.toFixed(1)}/s)</p>
            </div>
          )}
        </div>
      </Card>

      {/* Top 10 leaderboard */}
      <Card title="Top 10 by Improvement/Second">
        <div className="space-y-1">
          {sorted.filter(w => w.improvement > 0).slice(0, 10).map((w, i) => {
            const maxImpPerSec = sorted[0]?.improvementPerSec || 1;
            return (
              <div key={w.workerID} className="flex items-center gap-2">
                <span className="w-5 text-[10px] text-gray-500 text-right">
                  {i === 0 ? '🥇' : i === 1 ? '🥈' : i === 2 ? '🥉' : `${i+1}.`}
                </span>
                <span className="w-20 text-[10px] font-mono text-blue-400">W{w.week}#{w.workerID}</span>
                <div className="flex-1 h-3 bg-gray-800 rounded overflow-hidden">
                  <div className="h-full bg-emerald-600 rounded"
                    style={{ width: `${(w.improvementPerSec / maxImpPerSec) * 100}%` }} />
                </div>
                <span className="w-20 text-right text-[9px] text-gray-400">{w.improvementPerSec.toFixed(1)}/s</span>
              </div>
            );
          })}
        </div>
      </Card>
    </div>
  );
}
