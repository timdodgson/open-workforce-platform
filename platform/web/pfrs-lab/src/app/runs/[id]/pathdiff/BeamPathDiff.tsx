'use client';

import { useState, useMemo } from 'react';
import Card from '@/components/Card';
import { TreeNode } from '@/lib/types';

type ViewMode = 'side-by-side' | 'overlay';

interface PathChain {
  nodes: TreeNode[];
  totalPenalty: number;
  perWeek: Map<number, number>;
  family: number;
  isWinning: boolean;
}

interface Props {
  tree: TreeNode[];
}

function traceRoot(id: number, parentMap: Map<number, number>): number {
  let cur = id;
  let iter = 0;
  while (parentMap.has(cur) && parentMap.get(cur)! >= 0 && iter < 100) {
    cur = parentMap.get(cur)!; iter++;
  }
  return cur;
}

function buildPathChain(pathID: number, tree: TreeNode[], parentMap: Map<number, number>): PathChain {
  // Find the selected node.
  const node = tree.find(t => t.pathID === pathID);
  if (!node) return { nodes: [], totalPenalty: 0, perWeek: new Map(), family: 0, isWinning: false };

  // Trace ancestors back to root.
  const chain: TreeNode[] = [node];
  let cur = node.parentID;
  let iter = 0;
  while (cur >= 0 && iter < 100) {
    const parent = tree.find(t => t.pathID === cur);
    if (!parent) break;
    chain.unshift(parent);
    cur = parent.parentID;
    iter++;
  }

  const perWeek = new Map<number, number>();
  for (const n of chain) perWeek.set(n.week, n.weekPenalty);

  return {
    nodes: chain,
    totalPenalty: node.cumulativePenalty,
    perWeek,
    family: traceRoot(pathID, parentMap),
    isWinning: node.winning,
  };
}

function generateObservations(a: PathChain, b: PathChain): string[] {
  const obs: string[] = [];
  if (a.totalPenalty < b.totalPenalty) {
    obs.push(`Path A has ${b.totalPenalty - a.totalPenalty} lower cumulative penalty.`);
  } else if (b.totalPenalty < a.totalPenalty) {
    obs.push(`Path B has ${a.totalPenalty - b.totalPenalty} lower cumulative penalty.`);
  } else {
    obs.push('Both paths have identical cumulative penalty.');
  }

  // Find weeks where one is better.
  const weeks = new Set([...a.perWeek.keys(), ...b.perWeek.keys()]);
  let aBetterWeeks: number[] = [], bBetterWeeks: number[] = [];
  for (const w of weeks) {
    const pA = a.perWeek.get(w) || 0;
    const pB = b.perWeek.get(w) || 0;
    if (pA < pB) aBetterWeeks.push(w);
    if (pB < pA) bBetterWeeks.push(w);
  }

  if (aBetterWeeks.length > 0 && bBetterWeeks.length > 0) {
    const aDiff = bBetterWeeks.reduce((s, w) => s + ((a.perWeek.get(w) || 0) - (b.perWeek.get(w) || 0)), 0);
    const bDiff = aBetterWeeks.reduce((s, w) => s + ((b.perWeek.get(w) || 0) - (a.perWeek.get(w) || 0)), 0);
    obs.push(`Path A wins weeks ${aBetterWeeks.join(', ')} (saving ${bDiff}).`);
    obs.push(`Path B wins weeks ${bBetterWeeks.join(', ')} (saving ${Math.abs(aDiff)}).`);
  }

  // Trade-off detection.
  if (aBetterWeeks.length > 0 && bBetterWeeks.length > 0) {
    const bestAWeek = aBetterWeeks.reduce((best, w) => {
      const diff = (b.perWeek.get(w) || 0) - (a.perWeek.get(w) || 0);
      return diff > ((b.perWeek.get(best) || 0) - (a.perWeek.get(best) || 0)) ? w : best;
    }, aBetterWeeks[0]);
    const bestBWeek = bBetterWeeks[0];
    const aSaving = (b.perWeek.get(bestAWeek) || 0) - (a.perWeek.get(bestAWeek) || 0);
    const bCost = (a.perWeek.get(bestBWeek) || 0) - (b.perWeek.get(bestBWeek) || 0);
    obs.push(`Path A reduced Week ${bestAWeek} penalty by ${aSaving} at the expense of Week ${bestBWeek} (+${bCost}).`);
  }

  if (a.family !== b.family) {
    obs.push(`Paths come from different families (${a.family} vs ${b.family}).`);
  } else {
    obs.push('Both paths share the same root ancestor.');
  }

  return obs;
}

export default function BeamPathDiff({ tree }: Props) {
  const [viewMode, setViewMode] = useState<ViewMode>('side-by-side');

  // Get terminal paths (latest week for each path ID).
  const terminalPaths = useMemo(() => {
    const maxWeek = Math.max(...tree.map(t => t.week));
    return tree.filter(t => t.week === maxWeek).sort((a, b) => a.cumulativePenalty - b.cumulativePenalty);
  }, [tree]);

  const [pathAId, setPathAId] = useState(terminalPaths[0]?.pathID ?? 0);
  const [pathBId, setPathBId] = useState(terminalPaths[1]?.pathID ?? terminalPaths[0]?.pathID ?? 0);

  const parentMap = useMemo(() => {
    const m = new Map<number, number>();
    for (const t of tree) m.set(t.pathID, t.parentID);
    return m;
  }, [tree]);

  const chainA = useMemo(() => buildPathChain(pathAId, tree, parentMap), [pathAId, tree, parentMap]);
  const chainB = useMemo(() => buildPathChain(pathBId, tree, parentMap), [pathBId, tree, parentMap]);

  const weeks = useMemo(() => {
    const set = new Set([...chainA.perWeek.keys(), ...chainB.perWeek.keys()]);
    return Array.from(set).sort((a, b) => a - b);
  }, [chainA, chainB]);

  const observations = useMemo(() => generateObservations(chainA, chainB), [chainA, chainB]);
  const maxPenalty = Math.max(...weeks.map(w => Math.max(chainA.perWeek.get(w) || 0, chainB.perWeek.get(w) || 0)), 1);

  return (
    <div className="space-y-4">
      {/* Selectors */}
      <Card title="Beam Path Diff">
        <div className="flex gap-4 items-center mb-3">
          <div className="flex-1">
            <label className="text-[10px] text-blue-400 uppercase block mb-0.5">Path A</label>
            <select value={pathAId} onChange={e => setPathAId(parseInt(e.target.value))}
              className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs">
              {terminalPaths.map(p => (
                <option key={p.pathID} value={p.pathID}>
                  #{p.pathID} pen={p.cumulativePenalty} {p.winning ? '🏆' : ''}
                </option>
              ))}
            </select>
          </div>
          <span className="text-gray-600">vs</span>
          <div className="flex-1">
            <label className="text-[10px] text-rose-400 uppercase block mb-0.5">Path B</label>
            <select value={pathBId} onChange={e => setPathBId(parseInt(e.target.value))}
              className="w-full bg-gray-800 border border-gray-700 rounded px-2 py-1 text-xs">
              {terminalPaths.map(p => (
                <option key={p.pathID} value={p.pathID}>
                  #{p.pathID} pen={p.cumulativePenalty} {p.winning ? '🏆' : ''}
                </option>
              ))}
            </select>
          </div>
        </div>
        <div className="flex gap-2">
          {(['side-by-side', 'overlay'] as ViewMode[]).map(m => (
            <button key={m} onClick={() => setViewMode(m)}
              className={`px-3 py-1 rounded text-[10px] ${viewMode === m ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400'}`}>{m}</button>
          ))}
        </div>
      </Card>

      {/* Summary comparison */}
      <Card title="Summary">
        <div className="grid grid-cols-3 gap-4 text-center text-xs">
          <div>
            <p className="text-[10px] text-blue-400 uppercase">Path A (#{pathAId})</p>
            <p className="text-xl font-bold text-blue-400">{chainA.totalPenalty}</p>
            <p className="text-[9px] text-gray-500">Family: {chainA.family} {chainA.isWinning && '🏆'}</p>
          </div>
          <div>
            <p className="text-[10px] text-gray-500 uppercase">Delta</p>
            <p className={`text-xl font-bold ${chainA.totalPenalty <= chainB.totalPenalty ? 'text-emerald-400' : 'text-red-400'}`}>
              {chainA.totalPenalty - chainB.totalPenalty > 0 ? '+' : ''}{chainA.totalPenalty - chainB.totalPenalty}
            </p>
          </div>
          <div>
            <p className="text-[10px] text-rose-400 uppercase">Path B (#{pathBId})</p>
            <p className="text-xl font-bold text-rose-400">{chainB.totalPenalty}</p>
            <p className="text-[9px] text-gray-500">Family: {chainB.family} {chainB.isWinning && '🏆'}</p>
          </div>
        </div>
      </Card>

      {/* Per-week comparison */}
      <Card title="Per-Week Penalty">
        {viewMode === 'side-by-side' ? (
          <div className="space-y-1">
            {weeks.map(w => {
              const pA = chainA.perWeek.get(w) || 0;
              const pB = chainB.perWeek.get(w) || 0;
              const diff = pA - pB;
              return (
                <div key={w} className="flex items-center gap-2">
                  <span className="w-8 text-[10px] text-gray-500">W{w}</span>
                  <div className="flex-1 flex gap-1 items-center">
                    <div className="flex-1 h-4 bg-gray-800 rounded overflow-hidden">
                      <div className="h-full bg-blue-600 rounded" style={{ width: `${(pA / maxPenalty) * 100}%` }} />
                    </div>
                    <span className="w-10 text-right text-[9px] text-blue-400">{pA}</span>
                  </div>
                  <div className="flex-1 flex gap-1 items-center">
                    <span className="w-10 text-[9px] text-rose-400">{pB}</span>
                    <div className="flex-1 h-4 bg-gray-800 rounded overflow-hidden">
                      <div className="h-full bg-rose-600 rounded" style={{ width: `${(pB / maxPenalty) * 100}%` }} />
                    </div>
                  </div>
                  <span className={`w-12 text-right text-[9px] font-mono ${diff < 0 ? 'text-emerald-400' : diff > 0 ? 'text-red-400' : 'text-gray-600'}`}>
                    {diff > 0 ? `+${diff}` : diff === 0 ? '=' : diff}
                  </span>
                </div>
              );
            })}
          </div>
        ) : (
          <div className="flex items-end gap-1 h-32">
            {weeks.map(w => {
              const pA = chainA.perWeek.get(w) || 0;
              const pB = chainB.perWeek.get(w) || 0;
              return (
                <div key={w} className="flex-1 flex gap-px items-end h-full">
                  <div className="flex-1 bg-blue-600 rounded-t" style={{ height: `${(pA / maxPenalty) * 100}%` }} title={`A: ${pA}`} />
                  <div className="flex-1 bg-rose-600 rounded-t" style={{ height: `${(pB / maxPenalty) * 100}%` }} title={`B: ${pB}`} />
                </div>
              );
            })}
          </div>
        )}
      </Card>

      {/* Delta chart */}
      <Card title="Penalty Delta (A - B)">
        <div className="flex items-center gap-1 h-20">
          {weeks.map(w => {
            const diff = (chainA.perWeek.get(w) || 0) - (chainB.perWeek.get(w) || 0);
            const maxDiff = Math.max(...weeks.map(wk => Math.abs((chainA.perWeek.get(wk) || 0) - (chainB.perWeek.get(wk) || 0))), 1);
            const height = Math.abs(diff) / maxDiff * 35;
            const isUp = diff > 0;
            return (
              <div key={w} className="flex-1 flex flex-col items-center justify-center h-full">
                {isUp && <div className="bg-red-500 rounded-t w-full" style={{ height: `${height}px` }} />}
                <div className="w-full h-px bg-gray-700" />
                {!isUp && diff !== 0 && <div className="bg-emerald-500 rounded-b w-full" style={{ height: `${height}px` }} />}
                <span className="text-[8px] text-gray-600 mt-0.5">W{w}</span>
              </div>
            );
          })}
        </div>
        <p className="text-[9px] text-gray-600 text-center mt-1">Red = A worse, Green = A better</p>
      </Card>

      {/* Ancestry */}
      <Card title="Ancestry">
        <div className="grid grid-cols-2 gap-4">
          <div>
            <p className="text-[10px] text-blue-400 uppercase mb-1">Path A lineage</p>
            <div className="flex gap-1 flex-wrap">
              {chainA.nodes.map(n => (
                <span key={n.pathID} className={`px-1.5 py-0.5 rounded text-[9px] ${n.winning ? 'bg-yellow-900 text-yellow-300' : 'bg-gray-800 text-gray-400'}`}>
                  {n.pathID}
                </span>
              ))}
            </div>
          </div>
          <div>
            <p className="text-[10px] text-rose-400 uppercase mb-1">Path B lineage</p>
            <div className="flex gap-1 flex-wrap">
              {chainB.nodes.map(n => (
                <span key={n.pathID} className={`px-1.5 py-0.5 rounded text-[9px] ${n.winning ? 'bg-yellow-900 text-yellow-300' : 'bg-gray-800 text-gray-400'}`}>
                  {n.pathID}
                </span>
              ))}
            </div>
          </div>
        </div>
      </Card>

      {/* Observations */}
      <Card title="Observations">
        <div className="space-y-1">
          {observations.map((obs, i) => (
            <p key={i} className="text-sm text-gray-300">{obs}</p>
          ))}
        </div>
      </Card>
    </div>
  );
}
