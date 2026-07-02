'use client';

import { useMemo } from 'react';
import Card from '@/components/Card';
import { WorkerLifecycle, DiscoveryRecord } from '@/lib/types';

type Archetype = 'Explorer' | 'Exploiter' | 'Innovator' | 'Grinder' | 'Branch Builder' | 'Dormant';

interface ClassifiedWorker {
  workerID: number;
  week: number;
  archetype: Archetype;
  runtimeMs: number;
  discoveries: number;
  globalBests: number;
  branches: number;
  improvement: number;
  acceptWorseRate: number;
  depth: number;
}

interface ArchetypeStats {
  archetype: Archetype;
  count: number;
  pct: number;
  avgDiscoveries: number;
  avgImprovement: number;
  avgLifetimeMs: number;
  globalBests: number;
  globalBestPct: number;
  color: string;
  icon: string;
  description: string;
}

interface Props {
  workers: WorkerLifecycle[];
  discoveries: DiscoveryRecord[];
}

const ARCHETYPE_CONFIG: Record<Archetype, { color: string; icon: string; description: string }> = {
  Explorer: { color: '#f59e0b', icon: '🧭', description: 'High accept-worse rate, broad search, few improvements' },
  Exploiter: { color: '#10b981', icon: '🎯', description: 'Low accept-worse, focused, consistent improvements' },
  Innovator: { color: '#fbbf24', icon: '💡', description: 'Produced global bests, key breakthroughs' },
  Grinder: { color: '#6b7280', icon: '⚙️', description: 'Long runtime, many candidates, moderate improvement' },
  'Branch Builder': { color: '#8b5cf6', icon: '🌿', description: 'Created branches, spawned new lineages' },
  Dormant: { color: '#374151', icon: '💤', description: 'No improvement, contributed nothing' },
};

function classifyWorker(w: WorkerLifecycle, discCount: number, globalCount: number): Archetype {
  const improved = w.bestPenalty < w.startPenalty;
  const runtimeMs = w.finishTimeMs - w.startTimeMs;

  // Innovator: produced a global best.
  if (w.producedGlobalBest || globalCount > 0) return 'Innovator';

  // Dormant: never improved.
  if (!improved && discCount === 0) return 'Dormant';

  // Branch Builder: created branches.
  if (w.branchCount >= 2) return 'Branch Builder';

  // Explorer vs Exploiter: based on accept-worse behaviour.
  // High hard reject rate + few discoveries = explorer (trying many moves).
  // We approximate from available data.
  const improvementRate = runtimeMs > 0 ? (w.startPenalty - w.bestPenalty) / (runtimeMs / 1000) : 0;

  if (discCount >= 3 && improvementRate > 0) return 'Exploiter';

  if (w.depth >= 2 && discCount <= 1) return 'Explorer';

  // Grinder: long runtime, some improvement.
  if (runtimeMs > 500 && improved) return 'Grinder';

  // Default to explorer if active but not clearly anything else.
  if (improved) return 'Exploiter';
  return 'Explorer';
}

export default function WorkerArchetypes({ workers, discoveries }: Props) {
  const classified = useMemo(() => {
    const discMap = new Map<number, { local: number; global: number }>();
    for (const d of discoveries) {
      const e = discMap.get(d.workerID) || { local: 0, global: 0 };
      if (d.eventType === 'local_best') e.local++;
      if (d.eventType === 'global_best') e.global++;
      discMap.set(d.workerID, e);
    }

    return workers.map(w => {
      const disc = discMap.get(w.workerID) || { local: 0, global: 0 };
      const archetype = classifyWorker(w, disc.local + disc.global, disc.global);
      return {
        workerID: w.workerID,
        week: w.week,
        archetype,
        runtimeMs: w.finishTimeMs - w.startTimeMs,
        discoveries: disc.local + disc.global,
        globalBests: disc.global,
        branches: w.branchCount,
        improvement: Math.max(0, w.startPenalty - w.bestPenalty),
        acceptWorseRate: 0, // Not directly available per-worker.
        depth: w.depth,
      } as ClassifiedWorker;
    });
  }, [workers, discoveries]);

  // Compute per-archetype stats.
  const archetypeStats = useMemo(() => {
    const total = classified.length;
    const totalGlobalBests = classified.reduce((s, w) => s + w.globalBests, 0);
    const groups = new Map<Archetype, ClassifiedWorker[]>();
    for (const w of classified) {
      const existing = groups.get(w.archetype) || [];
      existing.push(w);
      groups.set(w.archetype, existing);
    }

    const stats: ArchetypeStats[] = [];
    for (const [archetype, members] of groups) {
      const cfg = ARCHETYPE_CONFIG[archetype];
      const count = members.length;
      const gb = members.reduce((s, w) => s + w.globalBests, 0);
      stats.push({
        archetype, count, pct: (count / total) * 100,
        avgDiscoveries: members.reduce((s, w) => s + w.discoveries, 0) / count,
        avgImprovement: members.reduce((s, w) => s + w.improvement, 0) / count,
        avgLifetimeMs: members.reduce((s, w) => s + w.runtimeMs, 0) / count,
        globalBests: gb,
        globalBestPct: totalGlobalBests > 0 ? (gb / totalGlobalBests) * 100 : 0,
        ...cfg,
      });
    }
    return stats.sort((a, b) => b.globalBestPct - a.globalBestPct);
  }, [classified]);

  // Per-week distribution.
  const weeks = useMemo(() => {
    const set = new Set(classified.map(w => w.week));
    return Array.from(set).sort((a, b) => a - b);
  }, [classified]);

  // Observations.
  const observations = useMemo(() => {
    const obs: string[] = [];
    const top = archetypeStats[0];
    if (top && top.globalBestPct > 50) {
      obs.push(`${top.archetype}s produced ${top.globalBestPct.toFixed(0)}% of all global bests.`);
    }
    const dormant = archetypeStats.find(a => a.archetype === 'Dormant');
    if (dormant && dormant.pct > 30) {
      obs.push(`${dormant.pct.toFixed(0)}% of workers were Dormant — significant compute waste.`);
    }
    const innovators = archetypeStats.find(a => a.archetype === 'Innovator');
    if (innovators && innovators.pct < 10) {
      obs.push(`Only ${innovators.pct.toFixed(1)}% of workers were Innovators — breakthroughs are rare.`);
    }
    const branchBuilders = archetypeStats.find(a => a.archetype === 'Branch Builder');
    if (branchBuilders && branchBuilders.count > 0) {
      obs.push(`${branchBuilders.count} Branch Builders spawned new lineages, maintaining diversity.`);
    }
    return obs;
  }, [archetypeStats]);

  function formatMs(ms: number): string {
    if (ms < 1000) return `${Math.round(ms)}ms`;
    return `${(ms / 1000).toFixed(1)}s`;
  }

  return (
    <div className="space-y-4">
      {/* Distribution pie (as stacked bar) */}
      <Card title="Worker Archetype Distribution">
        <div className="h-10 rounded-lg overflow-hidden flex mb-3">
          {archetypeStats.map(a => (
            <div key={a.archetype} style={{ width: `${a.pct}%`, background: a.color }}
              className="relative group" title={`${a.archetype}: ${a.count} (${a.pct.toFixed(1)}%)`}>
              {a.pct > 10 && (
                <span className="absolute inset-0 flex items-center justify-center text-[9px] font-bold text-white">
                  {a.icon}
                </span>
              )}
            </div>
          ))}
        </div>
        <div className="flex flex-wrap gap-3">
          {archetypeStats.map(a => (
            <span key={a.archetype} className="flex items-center gap-1 text-[10px] text-gray-400">
              <span className="w-3 h-3 rounded-sm" style={{ background: a.color }} />
              {a.icon} {a.archetype} ({a.count})
            </span>
          ))}
        </div>
      </Card>

      {/* Stats table */}
      <Card title="Archetype Statistics">
        <table className="w-full text-[10px]">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-1.5">Archetype</th>
              <th className="text-right p-1.5">Count</th>
              <th className="text-right p-1.5">%</th>
              <th className="text-right p-1.5">Avg Disc.</th>
              <th className="text-right p-1.5">Avg Improve</th>
              <th className="text-right p-1.5">Avg Lifetime</th>
              <th className="text-right p-1.5">Global Bests</th>
              <th className="text-right p-1.5">GB %</th>
            </tr>
          </thead>
          <tbody>
            {archetypeStats.map(a => (
              <tr key={a.archetype} className="border-t border-gray-800">
                <td className="p-1.5"><span className="mr-1">{a.icon}</span><span style={{ color: a.color }}>{a.archetype}</span></td>
                <td className="text-right p-1.5">{a.count}</td>
                <td className="text-right p-1.5">{a.pct.toFixed(1)}%</td>
                <td className="text-right p-1.5">{a.avgDiscoveries.toFixed(1)}</td>
                <td className="text-right p-1.5">{a.avgImprovement.toFixed(0)}</td>
                <td className="text-right p-1.5">{formatMs(a.avgLifetimeMs)}</td>
                <td className="text-right p-1.5 text-yellow-400">{a.globalBests}</td>
                <td className="text-right p-1.5 font-bold" style={{ color: a.color }}>{a.globalBestPct.toFixed(0)}%</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>

      {/* Per-week archetype timeline */}
      <Card title="Archetype Mix by Week">
        <div className="flex gap-1 h-24">
          {weeks.map(w => {
            const weekWorkers = classified.filter(c => c.week === w);
            const total = weekWorkers.length || 1;
            return (
              <div key={w} className="flex-1 flex flex-col-reverse rounded overflow-hidden">
                {archetypeStats.map(a => {
                  const count = weekWorkers.filter(c => c.archetype === a.archetype).length;
                  const height = (count / total) * 100;
                  return (
                    <div key={a.archetype} style={{ height: `${height}%`, background: a.color }}
                      title={`W${w} ${a.archetype}: ${count}`} />
                  );
                })}
              </div>
            );
          })}
        </div>
        <div className="flex justify-between text-[9px] text-gray-600 mt-1">
          {weeks.map(w => <span key={w}>W{w}</span>)}
        </div>
      </Card>

      {/* Success rates by archetype */}
      <Card title="Global Best Contribution">
        <div className="space-y-2">
          {archetypeStats.filter(a => a.globalBests > 0).map(a => (
            <div key={a.archetype} className="flex items-center gap-2">
              <span className="w-28 text-xs" style={{ color: a.color }}>{a.icon} {a.archetype}</span>
              <div className="flex-1 h-5 bg-gray-800 rounded overflow-hidden">
                <div className="h-full rounded" style={{ width: `${a.globalBestPct}%`, background: a.color }} />
              </div>
              <span className="w-16 text-right text-[10px] text-gray-400">{a.globalBestPct.toFixed(0)}% ({a.globalBests})</span>
            </div>
          ))}
        </div>
      </Card>

      {/* Observations */}
      <Card title="Observations">
        <div className="space-y-2">
          {observations.map((obs, i) => (
            <p key={i} className="text-sm text-gray-300">{obs}</p>
          ))}
        </div>
        <div className="mt-3 border-t border-gray-700 pt-3">
          <p className="text-[9px] text-gray-600">Archetype descriptions:</p>
          <div className="grid grid-cols-2 gap-1 mt-1">
            {archetypeStats.map(a => (
              <p key={a.archetype} className="text-[9px] text-gray-500">
                <span style={{ color: a.color }}>{a.icon} {a.archetype}:</span> {a.description}
              </p>
            ))}
          </div>
        </div>
      </Card>
    </div>
  );
}
