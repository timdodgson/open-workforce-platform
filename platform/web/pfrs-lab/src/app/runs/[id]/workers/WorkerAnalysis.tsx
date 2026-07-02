'use client';

import { useState, useMemo } from 'react';
import Card from '@/components/Card';
import { WorkerLifecycle, DiscoveryRecord } from '@/lib/types';

type SortKey = 'discoveries' | 'branches' | 'runtime' | 'improvement';
type SortDir = 'asc' | 'desc';

interface WorkerStats {
  workerID: number;
  week: number;
  depth: number;
  runtimeMs: number;
  discoveries: number;
  globalBests: number;
  localBests: number;
  branches: number;
  bestImprovement: number;
  startPenalty: number;
  bestPenalty: number;
  finalPenalty: number;
  producedGlobalBest: boolean;
  neverImproved: boolean;
}

interface Props {
  workers: WorkerLifecycle[];
  discoveries: DiscoveryRecord[];
}

function buildWorkerStats(workers: WorkerLifecycle[], discoveries: DiscoveryRecord[]): WorkerStats[] {
  // Count discoveries per worker.
  const discByWorker = new Map<number, { local: number; global: number; bestImprovement: number }>();
  for (const d of discoveries) {
    const existing = discByWorker.get(d.workerID) || { local: 0, global: 0, bestImprovement: 0 };
    if (d.eventType === 'local_best') existing.local++;
    if (d.eventType === 'global_best') existing.global++;
    existing.bestImprovement = Math.max(existing.bestImprovement, d.improvement);
    discByWorker.set(d.workerID, existing);
  }

  return workers.map(w => {
    const disc = discByWorker.get(w.workerID) || { local: 0, global: 0, bestImprovement: 0 };
    const runtimeMs = w.finishTimeMs - w.startTimeMs;
    const neverImproved = w.bestPenalty >= w.startPenalty;
    return {
      workerID: w.workerID,
      week: w.week,
      depth: w.depth,
      runtimeMs,
      discoveries: disc.local + disc.global,
      globalBests: disc.global,
      localBests: disc.local,
      branches: w.branchCount,
      bestImprovement: disc.bestImprovement || (w.startPenalty - w.bestPenalty),
      startPenalty: w.startPenalty,
      bestPenalty: w.bestPenalty,
      finalPenalty: w.finalPenalty,
      producedGlobalBest: w.producedGlobalBest,
      neverImproved,
    };
  });
}

export default function WorkerAnalysis({ workers, discoveries }: Props) {
  const [sortKey, setSortKey] = useState<SortKey>('discoveries');
  const [sortDir, setSortDir] = useState<SortDir>('desc');

  const stats = useMemo(() => buildWorkerStats(workers, discoveries), [workers, discoveries]);

  const totalWorkers = stats.length;
  const neverImproved = stats.filter(s => s.neverImproved).length;
  const foundLocalBest = stats.filter(s => s.localBests > 0).length;
  const foundGlobalBest = stats.filter(s => s.producedGlobalBest).length;
  const avgRuntime = stats.length > 0 ? stats.reduce((s, w) => s + w.runtimeMs, 0) / stats.length : 0;
  const totalDiscoveries = stats.reduce((s, w) => s + w.discoveries, 0);
  const successRate = totalWorkers > 0 ? ((totalWorkers - neverImproved) / totalWorkers * 100) : 0;

  // Sort workers.
  const sorted = useMemo(() => {
    return [...stats].sort((a, b) => {
      const dir = sortDir === 'desc' ? -1 : 1;
      switch (sortKey) {
        case 'discoveries': return (a.discoveries - b.discoveries) * dir;
        case 'branches': return (a.branches - b.branches) * dir;
        case 'runtime': return (a.runtimeMs - b.runtimeMs) * dir;
        case 'improvement': return (a.bestImprovement - b.bestImprovement) * dir;
        default: return 0;
      }
    });
  }, [stats, sortKey, sortDir]);

  // Top 10 explorers (by discoveries).
  const top10 = useMemo(() => {
    return [...stats].sort((a, b) => b.discoveries - a.discoveries).slice(0, 10);
  }, [stats]);

  // Discovery timeline bins (by elapsed time).
  const discoveryBins = useMemo(() => {
    if (discoveries.length === 0) return [];
    const maxTime = Math.max(...discoveries.map(d => d.elapsedMs));
    const numBins = 30;
    const binSize = maxTime / numBins;
    const bins = Array(numBins).fill(0);
    for (const d of discoveries) {
      const bin = Math.min(Math.floor(d.elapsedMs / binSize), numBins - 1);
      bins[bin]++;
    }
    return bins;
  }, [discoveries]);

  // Improvement histogram.
  const improvementBins = useMemo(() => {
    const improvements = stats.filter(s => s.bestImprovement > 0).map(s => s.bestImprovement);
    if (improvements.length === 0) return [];
    const max = Math.max(...improvements);
    const numBins = 20;
    const binSize = max / numBins;
    const bins = Array(numBins).fill(0);
    for (const imp of improvements) {
      const bin = Math.min(Math.floor(imp / binSize), numBins - 1);
      bins[bin]++;
    }
    return bins;
  }, [stats]);

  function toggleSort(key: SortKey) {
    if (sortKey === key) setSortDir(sortDir === 'desc' ? 'asc' : 'desc');
    else { setSortKey(key); setSortDir('desc'); }
  }

  function formatMs(ms: number): string {
    if (ms < 1000) return `${ms}ms`;
    return `${(ms / 1000).toFixed(1)}s`;
  }

  return (
    <div className="space-y-4">
      {/* Summary cards */}
      <div className="grid grid-cols-4 gap-3">
        <Card title="Total Workers">
          <p className="text-2xl font-bold text-blue-400">{totalWorkers}</p>
        </Card>
        <Card title="Success Rate">
          <p className="text-2xl font-bold text-emerald-400">{successRate.toFixed(1)}%</p>
          <p className="text-[10px] text-gray-500">{totalWorkers - neverImproved} improved</p>
        </Card>
        <Card title="Global Best Finders">
          <p className="text-2xl font-bold text-yellow-400">{foundGlobalBest}</p>
          <p className="text-[10px] text-gray-500">{foundLocalBest} found local best</p>
        </Card>
        <Card title="Total Discoveries">
          <p className="text-2xl font-bold text-purple-400">{totalDiscoveries}</p>
          <p className="text-[10px] text-gray-500">avg {formatMs(avgRuntime)} runtime</p>
        </Card>
      </div>

      {/* Worker utilisation: who did nothing vs who contributed */}
      <Card title="Worker Utilisation">
        <div className="flex gap-1 h-6 rounded overflow-hidden mb-2">
          <div className="bg-emerald-600" style={{ width: `${(foundGlobalBest / totalWorkers) * 100}%` }}
            title={`${foundGlobalBest} found global best`} />
          <div className="bg-blue-600" style={{ width: `${((foundLocalBest - foundGlobalBest) / totalWorkers) * 100}%` }}
            title={`${foundLocalBest - foundGlobalBest} found local best only`} />
          <div className="bg-gray-700" style={{ width: `${(neverImproved / totalWorkers) * 100}%` }}
            title={`${neverImproved} never improved`} />
        </div>
        <div className="flex gap-4 text-[10px] text-gray-400">
          <span className="flex items-center gap-1"><span className="w-2 h-2 bg-emerald-600 rounded-sm" />Global best ({foundGlobalBest})</span>
          <span className="flex items-center gap-1"><span className="w-2 h-2 bg-blue-600 rounded-sm" />Local best ({foundLocalBest - foundGlobalBest})</span>
          <span className="flex items-center gap-1"><span className="w-2 h-2 bg-gray-700 rounded-sm" />No improvement ({neverImproved})</span>
        </div>
      </Card>

      {/* Discovery timeline */}
      <Card title="Discovery Timeline">
        <div className="flex items-end gap-px h-20">
          {discoveryBins.map((count, i) => {
            const maxBin = Math.max(...discoveryBins, 1);
            return (
              <div key={i} className="flex-1 flex flex-col justify-end">
                <div
                  className="bg-purple-500 rounded-t min-w-[2px]"
                  style={{ height: `${(count / maxBin) * 100}%`, minHeight: count > 0 ? '2px' : '0' }}
                  title={`${count} discoveries`}
                />
              </div>
            );
          })}
        </div>
        <div className="flex justify-between text-[9px] text-gray-600 mt-1">
          <span>Start</span>
          <span>End</span>
        </div>
      </Card>

      {/* Improvement histogram */}
      <Card title="Improvement Distribution">
        <div className="flex items-end gap-px h-20">
          {improvementBins.map((count, i) => {
            const maxBin = Math.max(...improvementBins, 1);
            return (
              <div key={i} className="flex-1 flex flex-col justify-end">
                <div
                  className="bg-emerald-500 rounded-t min-w-[2px]"
                  style={{ height: `${(count / maxBin) * 100}%`, minHeight: count > 0 ? '2px' : '0' }}
                  title={`${count} workers`}
                />
              </div>
            );
          })}
        </div>
        <div className="flex justify-between text-[9px] text-gray-600 mt-1">
          <span>Small</span>
          <span>Large improvement</span>
        </div>
      </Card>

      {/* Top 10 Explorers */}
      <Card title="🏆 Top 10 Explorers">
        <div className="space-y-2">
          {top10.map((w, i) => {
            const maxDisc = top10[0]?.discoveries || 1;
            return (
              <div key={w.workerID} className="flex items-center gap-2">
                <span className="w-5 text-xs text-gray-500 text-right">
                  {i === 0 ? '🥇' : i === 1 ? '🥈' : i === 2 ? '🥉' : `${i + 1}.`}
                </span>
                <span className="w-16 text-xs font-mono text-blue-400">W{w.week}#{w.workerID}</span>
                <div className="flex-1 h-4 bg-gray-800 rounded overflow-hidden relative">
                  <div
                    className={`h-full rounded ${w.producedGlobalBest ? 'bg-yellow-500' : 'bg-blue-600'}`}
                    style={{ width: `${(w.discoveries / maxDisc) * 100}%` }}
                  />
                  <span className="absolute right-1 top-0 h-full flex items-center text-[9px] text-gray-300">
                    {w.discoveries} disc / {w.branches} br / imp {w.bestImprovement}
                  </span>
                </div>
                {w.producedGlobalBest && <span className="text-[9px] text-yellow-400">★ global</span>}
              </div>
            );
          })}
        </div>
      </Card>

      {/* Sortable full table */}
      <Card title="All Workers">
        <div className="flex gap-2 mb-2">
          {(['discoveries', 'branches', 'runtime', 'improvement'] as SortKey[]).map(key => (
            <button
              key={key}
              onClick={() => toggleSort(key)}
              className={`px-2 py-1 rounded text-[10px] ${
                sortKey === key ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700'
              }`}
            >
              {key} {sortKey === key ? (sortDir === 'desc' ? '↓' : '↑') : ''}
            </button>
          ))}
        </div>
        <div className="max-h-96 overflow-y-auto">
          <table className="w-full text-[10px]">
            <thead className="sticky top-0 bg-gray-900">
              <tr className="text-gray-500 uppercase">
                <th className="text-left p-1">ID</th>
                <th className="text-right p-1">Week</th>
                <th className="text-right p-1">Depth</th>
                <th className="text-right p-1">Runtime</th>
                <th className="text-right p-1">Disc.</th>
                <th className="text-right p-1">Global</th>
                <th className="text-right p-1">Branches</th>
                <th className="text-right p-1">Best Imp.</th>
                <th className="text-right p-1">Start→Best</th>
              </tr>
            </thead>
            <tbody>
              {sorted.slice(0, 100).map(w => (
                <tr key={`${w.week}-${w.workerID}`} className={`border-t border-gray-800 ${
                  w.producedGlobalBest ? 'bg-yellow-900/10' : w.neverImproved ? 'opacity-50' : ''
                }`}>
                  <td className="p-1 font-mono text-blue-400">{w.workerID}</td>
                  <td className="text-right p-1">{w.week}</td>
                  <td className="text-right p-1">{w.depth}</td>
                  <td className="text-right p-1">{formatMs(w.runtimeMs)}</td>
                  <td className="text-right p-1 font-medium">{w.discoveries}</td>
                  <td className="text-right p-1">{w.globalBests > 0 ? `★${w.globalBests}` : '—'}</td>
                  <td className="text-right p-1">{w.branches}</td>
                  <td className="text-right p-1 text-emerald-400">{w.bestImprovement > 0 ? `−${w.bestImprovement}` : '—'}</td>
                  <td className="text-right p-1 text-gray-400">{w.startPenalty}→{w.bestPenalty}</td>
                </tr>
              ))}
            </tbody>
          </table>
          {sorted.length > 100 && (
            <p className="text-[9px] text-gray-600 p-2 text-center">Showing first 100 of {sorted.length}</p>
          )}
        </div>
      </Card>
    </div>
  );
}
