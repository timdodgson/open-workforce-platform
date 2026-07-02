'use client';

import { useState, useMemo } from 'react';
import Card from '@/components/Card';
import { TreeNode, WorkerLifecycle } from '@/lib/types';

type ColorMode = 'week' | 'penalty' | 'family' | 'winning';

interface GenealogyNode {
  pathID: number;
  parentID: number;
  week: number;
  penalty: number;
  cumulativePenalty: number;
  retained: boolean;
  winning: boolean;
  children: number[];
  family: number; // root ancestor
  depth: number;
  globalBests: number;
  localBests: number;
  workersStarted: number;
}

interface Props {
  tree: TreeNode[];
  workers: WorkerLifecycle[];
}

const WEEK_COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4', '#f97316', '#ec4899'];
const FAMILY_COLORS = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4', '#f97316', '#ec4899', '#84cc16', '#a855f7'];

function traceRoot(id: number, parentMap: Map<number, number>): number {
  let cur = id;
  let iter = 0;
  while (parentMap.has(cur) && parentMap.get(cur)! >= 0 && iter < 100) {
    cur = parentMap.get(cur)!;
    iter++;
  }
  return cur;
}

function traceDepth(id: number, parentMap: Map<number, number>): number {
  let cur = id;
  let depth = 0;
  while (parentMap.has(cur) && parentMap.get(cur)! >= 0 && depth < 100) {
    cur = parentMap.get(cur)!;
    depth++;
  }
  return depth;
}

function buildGenealogy(tree: TreeNode[], workers: WorkerLifecycle[]): GenealogyNode[] {
  const parentMap = new Map<number, number>();
  for (const t of tree) parentMap.set(t.pathID, t.parentID);

  // Worker stats per path.
  const globalBestsPerPath = new Map<number, number>();
  for (const w of workers) {
    if (w.producedGlobalBest) {
      // Approximate: link worker to path via week.
      // Not exact but gives a signal.
    }
  }

  const childrenMap = new Map<number, number[]>();
  for (const t of tree) {
    if (t.parentID >= 0) {
      const existing = childrenMap.get(t.parentID) || [];
      existing.push(t.pathID);
      childrenMap.set(t.parentID, existing);
    }
  }

  return tree.map(t => ({
    pathID: t.pathID,
    parentID: t.parentID,
    week: t.week,
    penalty: t.weekPenalty,
    cumulativePenalty: t.cumulativePenalty,
    retained: t.retained,
    winning: t.winning,
    children: childrenMap.get(t.pathID) || [],
    family: traceRoot(t.pathID, parentMap),
    depth: traceDepth(t.pathID, parentMap),
    globalBests: globalBestsPerPath.get(t.pathID) || 0,
    localBests: 0,
    workersStarted: t.workersStarted,
  }));
}

export default function BranchGenealogy({ tree, workers }: Props) {
  const [colorMode, setColorMode] = useState<ColorMode>('winning');
  const [selected, setSelected] = useState<GenealogyNode | null>(null);
  const [collapsed, setCollapsed] = useState<Set<number>>(new Set());
  const [zoom, setZoom] = useState(1);

  const nodes = useMemo(() => buildGenealogy(tree, workers), [tree, workers]);

  // Family stats.
  const families = useMemo(() => {
    const fam = new Map<number, GenealogyNode[]>();
    for (const n of nodes) {
      const existing = fam.get(n.family) || [];
      existing.push(n);
      fam.set(n.family, existing);
    }
    return fam;
  }, [nodes]);

  const totalFamilies = families.size;
  const largestFamily = useMemo(() => {
    let best = 0, bestId = 0;
    for (const [id, members] of families) {
      if (members.length > best) { best = members.length; bestId = id; }
    }
    return { id: bestId, size: best };
  }, [families]);
  const avgFamilySize = totalFamilies > 0 ? (nodes.length / totalFamilies).toFixed(1) : '0';

  // Extinction: families with no retained nodes in the final week.
  const maxWeek = Math.max(...nodes.map(n => n.week), 0);
  const extinctFamilies = useMemo(() => {
    let count = 0;
    for (const [, members] of families) {
      const hasRetainedFinal = members.some(m => m.week === maxWeek && m.retained);
      if (!hasRetainedFinal) count++;
    }
    return count;
  }, [families, maxWeek]);

  // Winning lineage path IDs.
  const winningPath = useMemo(() => {
    const set = new Set<number>();
    const winNodes = nodes.filter(n => n.winning);
    for (const n of winNodes) {
      let cur = n.pathID;
      set.add(cur);
      const parentMap = new Map<number, number>();
      for (const node of nodes) parentMap.set(node.pathID, node.parentID);
      let iter = 0;
      while (parentMap.has(cur) && parentMap.get(cur)! >= 0 && iter < 100) {
        cur = parentMap.get(cur)!;
        set.add(cur);
        iter++;
      }
    }
    return set;
  }, [nodes]);

  function getNodeColor(n: GenealogyNode): string {
    switch (colorMode) {
      case 'week': return WEEK_COLORS[(n.week - 1) % WEEK_COLORS.length];
      case 'penalty': {
        const maxP = Math.max(...nodes.map(nd => nd.penalty), 1);
        const intensity = n.penalty / maxP;
        return `hsl(${120 - intensity * 120}, 70%, 50%)`;
      }
      case 'family': return FAMILY_COLORS[Array.from(families.keys()).indexOf(n.family) % FAMILY_COLORS.length];
      case 'winning': {
        if (n.winning) return '#fbbf24';
        if (winningPath.has(n.pathID)) return '#f59e0b';
        if (!n.retained) return '#374151';
        return '#3b82f6';
      }
      default: return '#3b82f6';
    }
  }

  function toggleCollapse(pathID: number) {
    setCollapsed(prev => {
      const next = new Set(prev);
      if (next.has(pathID)) next.delete(pathID); else next.add(pathID);
      return next;
    });
  }

  // Layout: arrange by week (x) and position within week (y).
  const weekGroups = useMemo(() => {
    const groups = new Map<number, GenealogyNode[]>();
    for (const n of nodes) {
      const existing = groups.get(n.week) || [];
      existing.push(n);
      groups.set(n.week, existing);
    }
    return groups;
  }, [nodes]);

  const weeks = [...weekGroups.keys()].sort((a, b) => a - b);
  const maxPerWeek = Math.max(...Array.from(weekGroups.values()).map(g => g.length), 1);

  const NODE_W = 80;
  const NODE_H = Math.max(30, 500 / maxPerWeek);
  const SVG_W = weeks.length * NODE_W + 100;
  const SVG_H = maxPerWeek * NODE_H + 60;

  return (
    <div className="space-y-4">
      {/* Summary */}
      <div className="grid grid-cols-4 gap-3">
        <Card title="Total Families"><p className="text-2xl font-bold text-purple-400">{totalFamilies}</p></Card>
        <Card title="Largest Family"><p className="text-2xl font-bold text-blue-400">{largestFamily.size} nodes</p><p className="text-[9px] text-gray-500">ID: {largestFamily.id}</p></Card>
        <Card title="Avg Family Size"><p className="text-2xl font-bold text-gray-300">{avgFamilySize}</p></Card>
        <Card title="Extinct Families"><p className="text-2xl font-bold text-red-400">{extinctFamilies}</p><p className="text-[9px] text-gray-500">{totalFamilies > 0 ? ((extinctFamilies / totalFamilies) * 100).toFixed(0) : 0}% rate</p></Card>
      </div>

      {/* Controls */}
      <Card title="Interactive Tree">
        <div className="flex gap-2 mb-3">
          <div className="flex items-center gap-1">
            <span className="text-[10px] text-gray-500">Color:</span>
            {(['winning', 'week', 'penalty', 'family'] as ColorMode[]).map(m => (
              <button key={m} onClick={() => setColorMode(m)}
                className={`px-2 py-0.5 rounded text-[10px] ${colorMode === m ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400'}`}>{m}</button>
            ))}
          </div>
          <div className="flex items-center gap-1 ml-auto">
            <button onClick={() => setZoom(z => Math.min(z + 0.2, 3))} className="px-2 py-0.5 bg-gray-800 hover:bg-gray-700 rounded text-[10px]">+</button>
            <button onClick={() => setZoom(z => Math.max(z - 0.2, 0.5))} className="px-2 py-0.5 bg-gray-800 hover:bg-gray-700 rounded text-[10px]">−</button>
            <button onClick={() => setCollapsed(new Set())} className="px-2 py-0.5 bg-gray-800 hover:bg-gray-700 rounded text-[10px]">Expand All</button>
            <button onClick={() => setCollapsed(new Set(nodes.filter(n => n.children.length > 0).map(n => n.pathID)))} className="px-2 py-0.5 bg-gray-800 hover:bg-gray-700 rounded text-[10px]">Collapse All</button>
          </div>
        </div>

        {/* SVG Tree */}
        <div className="overflow-auto border border-gray-800 rounded" style={{ maxHeight: '500px' }}>
          <svg
            viewBox={`0 0 ${SVG_W} ${SVG_H}`}
            style={{ width: `${SVG_W * zoom}px`, height: `${SVG_H * zoom}px`, minWidth: '100%' }}
          >
            {/* Edges */}
            {nodes.map(n => {
              if (n.parentID < 0) return null;
              const parent = nodes.find(p => p.pathID === n.parentID);
              if (!parent) return null;
              if (collapsed.has(parent.pathID)) return null;

              const parentWeekIdx = weeks.indexOf(parent.week);
              const childWeekIdx = weeks.indexOf(n.week);
              const parentGroup = weekGroups.get(parent.week) || [];
              const childGroup = weekGroups.get(n.week) || [];
              const parentYIdx = parentGroup.indexOf(parent);
              const childYIdx = childGroup.indexOf(n);

              const x1 = parentWeekIdx * NODE_W + 60;
              const y1 = parentYIdx * NODE_H + 40;
              const x2 = childWeekIdx * NODE_W + 60;
              const y2 = childYIdx * NODE_H + 40;

              const isWinning = winningPath.has(n.pathID) && winningPath.has(parent.pathID);
              return (
                <line key={`e-${n.pathID}`}
                  x1={x1} y1={y1} x2={x2} y2={y2}
                  stroke={isWinning ? '#fbbf24' : '#374151'}
                  strokeWidth={isWinning ? 2 : 0.5}
                  opacity={isWinning ? 1 : 0.4}
                />
              );
            })}

            {/* Nodes */}
            {nodes.map(n => {
              const weekIdx = weeks.indexOf(n.week);
              const group = weekGroups.get(n.week) || [];
              const yIdx = group.indexOf(n);
              if (collapsed.has(n.parentID) && n.parentID >= 0) return null;

              const cx = weekIdx * NODE_W + 60;
              const cy = yIdx * NODE_H + 40;
              const isSelected = selected?.pathID === n.pathID;
              const isDead = !n.retained && n.week < maxWeek;

              return (
                <g key={n.pathID} className="cursor-pointer" onClick={() => setSelected(n)}>
                  <circle cx={cx} cy={cy}
                    r={isSelected ? 10 : n.winning ? 8 : 6}
                    fill={getNodeColor(n)}
                    opacity={isDead ? 0.3 : 1}
                    stroke={isSelected ? '#fff' : 'none'}
                    strokeWidth={isSelected ? 2 : 0}
                  />
                  {n.children.length > 0 && !collapsed.has(n.pathID) && (
                    <text x={cx + 8} y={cy - 6} className="fill-gray-600 text-[7px]"
                      onClick={(e) => { e.stopPropagation(); toggleCollapse(n.pathID); }}>
                      ▼
                    </text>
                  )}
                  {collapsed.has(n.pathID) && (
                    <text x={cx + 8} y={cy - 6} className="fill-gray-500 text-[7px]"
                      onClick={(e) => { e.stopPropagation(); toggleCollapse(n.pathID); }}>
                      ▶ {n.children.length}
                    </text>
                  )}
                  <title>Path {n.pathID} W{n.week} pen={n.penalty}</title>
                </g>
              );
            })}

            {/* Week labels */}
            {weeks.map((w, i) => (
              <text key={w} x={i * NODE_W + 60} y={20}
                textAnchor="middle" className="fill-gray-500 text-[8px]">W{w}</text>
            ))}
          </svg>
        </div>
      </Card>

      {/* Selected node detail */}
      {selected && (
        <Card title={`Path ${selected.pathID}`}>
          <div className="grid grid-cols-4 gap-3 text-xs">
            <div><p className="text-gray-500">Path ID</p><p className="font-mono text-blue-400">{selected.pathID}</p></div>
            <div><p className="text-gray-500">Parent</p><p className="font-mono">{selected.parentID >= 0 ? selected.parentID : 'Root'}</p></div>
            <div><p className="text-gray-500">Week</p><p>{selected.week}</p></div>
            <div><p className="text-gray-500">Depth</p><p>{selected.depth}</p></div>
            <div><p className="text-gray-500">Week Penalty</p><p className="text-emerald-400">{selected.penalty}</p></div>
            <div><p className="text-gray-500">Cumulative</p><p>{selected.cumulativePenalty}</p></div>
            <div><p className="text-gray-500">Family</p><p className="font-mono text-purple-400">{selected.family}</p></div>
            <div><p className="text-gray-500">Children</p><p>{selected.children.length}</p></div>
            <div><p className="text-gray-500">Retained</p><p className={selected.retained ? 'text-emerald-400' : 'text-red-400'}>{selected.retained ? 'Yes' : 'No'}</p></div>
            <div><p className="text-gray-500">Winning</p><p className={selected.winning ? 'text-yellow-400' : 'text-gray-500'}>{selected.winning ? '🏆 Yes' : 'No'}</p></div>
            <div><p className="text-gray-500">Workers</p><p>{selected.workersStarted}</p></div>
            <div><p className="text-gray-500">On Winning Line</p><p className={winningPath.has(selected.pathID) ? 'text-yellow-400' : 'text-gray-500'}>{winningPath.has(selected.pathID) ? 'Yes' : 'No'}</p></div>
          </div>
          {selected.children.length > 0 && (
            <div className="mt-3">
              <p className="text-[10px] text-gray-500 mb-1">Children: {selected.children.join(', ')}</p>
            </div>
          )}
          <button onClick={() => setSelected(null)}
            className="mt-3 px-3 py-1 bg-gray-800 hover:bg-gray-700 rounded text-[10px] text-gray-400">Close</button>
        </Card>
      )}

      {/* Legend */}
      <Card title="Legend">
        <div className="flex flex-wrap gap-4 text-[10px] text-gray-400">
          {colorMode === 'winning' && (
            <>
              <span className="flex items-center gap-1"><span className="w-3 h-3 rounded-full bg-yellow-400" />Winning node</span>
              <span className="flex items-center gap-1"><span className="w-3 h-3 rounded-full bg-amber-500" />Winning lineage</span>
              <span className="flex items-center gap-1"><span className="w-3 h-3 rounded-full bg-blue-500" />Retained</span>
              <span className="flex items-center gap-1"><span className="w-3 h-3 rounded-full bg-gray-700" />Dead branch</span>
            </>
          )}
          {colorMode === 'week' && weeks.map(w => (
            <span key={w} className="flex items-center gap-1"><span className="w-3 h-3 rounded-full" style={{ background: WEEK_COLORS[(w-1) % WEEK_COLORS.length] }} />Week {w}</span>
          ))}
          {colorMode === 'penalty' && <span>Green = low penalty, Red = high penalty</span>}
          {colorMode === 'family' && <span>Each colour = distinct root ancestor</span>}
        </div>
      </Card>
    </div>
  );
}
