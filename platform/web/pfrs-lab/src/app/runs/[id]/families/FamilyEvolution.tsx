'use client';

import { useMemo } from 'react';
import Card from '@/components/Card';
import { TreeNode } from '@/lib/types';

interface FamilyWeekData {
  week: number;
  families: Map<number, number>; // rootAncestor -> count of retained paths
  totalRetained: number;
  entropy: number;
  dominantFamily: number;
  dominantSize: number;
  newBranches: number;
  deaths: number;
}

interface Props {
  tree: TreeNode[];
}

const FAMILY_COLORS = [
  '#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6',
  '#06b6d4', '#f97316', '#ec4899', '#84cc16', '#a855f7',
  '#14b8a6', '#e11d48', '#7c3aed', '#0891b2', '#dc2626',
];

function traceRoot(pathID: number, parentMap: Map<number, number>): number {
  let current = pathID;
  let iterations = 0;
  while (parentMap.has(current) && parentMap.get(current)! >= 0 && iterations < 100) {
    current = parentMap.get(current)!;
    iterations++;
  }
  return current;
}

function computeEntropy(families: Map<number, number>, total: number): number {
  if (total <= 1) return 0;
  let entropy = 0;
  for (const count of families.values()) {
    const p = count / total;
    if (p > 0) entropy -= p * Math.log2(p);
  }
  return entropy;
}

function buildFamilyData(tree: TreeNode[]): FamilyWeekData[] {
  // Build parent map for ancestor tracing.
  const parentMap = new Map<number, number>();
  for (const node of tree) {
    parentMap.set(node.pathID, node.parentID);
  }

  const weeks = [...new Set(tree.map(t => t.week))].sort((a, b) => a - b);
  const prevFamilies = new Map<number, number>();
  const result: FamilyWeekData[] = [];

  for (const week of weeks) {
    const weekNodes = tree.filter(t => t.week === week && t.retained);
    const families = new Map<number, number>();

    for (const node of weekNodes) {
      const root = traceRoot(node.pathID, parentMap);
      families.set(root, (families.get(root) || 0) + 1);
    }

    const totalRetained = weekNodes.length;
    const entropy = computeEntropy(families, totalRetained);

    // Find dominant.
    let dominantFamily = 0;
    let dominantSize = 0;
    for (const [fam, count] of families) {
      if (count > dominantSize) {
        dominantFamily = fam;
        dominantSize = count;
      }
    }

    // Count new families (not in previous week).
    let newBranches = 0;
    for (const fam of families.keys()) {
      if (!prevFamilies.has(fam)) newBranches++;
    }

    // Count deaths (in previous but not current).
    let deaths = 0;
    for (const fam of prevFamilies.keys()) {
      if (!families.has(fam)) deaths++;
    }

    result.push({ week, families, totalRetained, entropy, dominantFamily, dominantSize, newBranches, deaths });

    prevFamilies.clear();
    for (const [k, v] of families) prevFamilies.set(k, v);
  }

  return result;
}

export default function FamilyEvolution({ tree }: Props) {
  const familyData = useMemo(() => buildFamilyData(tree), [tree]);

  // Collect all unique families across all weeks for consistent colours.
  const allFamilies = useMemo(() => {
    const set = new Set<number>();
    for (const wd of familyData) {
      for (const fam of wd.families.keys()) set.add(fam);
    }
    return Array.from(set).sort((a, b) => a - b);
  }, [familyData]);

  const familyColorMap = useMemo(() => {
    const map = new Map<number, string>();
    allFamilies.forEach((fam, i) => {
      map.set(fam, FAMILY_COLORS[i % FAMILY_COLORS.length]);
    });
    return map;
  }, [allFamilies]);

  const maxEntropy = Math.max(...familyData.map(d => d.entropy), 1);
  const maxRetained = Math.max(...familyData.map(d => d.totalRetained), 1);

  // Monopoly detection: weeks where one family has > 60%.
  const monopolyWeeks = familyData.filter(d => d.totalRetained > 0 && d.dominantSize / d.totalRetained > 0.6);

  return (
    <div className="space-y-4">
      {/* Summary */}
      <div className="grid grid-cols-4 gap-3">
        <Card title="Families Seen">
          <p className="text-2xl font-bold text-purple-400">{allFamilies.length}</p>
        </Card>
        <Card title="Surviving (Final)">
          <p className="text-2xl font-bold text-emerald-400">
            {familyData.length > 0 ? familyData[familyData.length - 1].families.size : 0}
          </p>
        </Card>
        <Card title="Monopoly Weeks">
          <p className="text-2xl font-bold text-red-400">{monopolyWeeks.length}</p>
          <p className="text-[9px] text-gray-500">&gt;60% single family</p>
        </Card>
        <Card title="Max Entropy">
          <p className="text-2xl font-bold text-blue-400">{maxEntropy.toFixed(2)}</p>
        </Card>
      </div>

      {/* Stacked area chart */}
      <Card title="Family Size Over Time (Stacked)">
        <div className="h-48 flex items-end gap-1">
          {familyData.map(wd => (
            <div key={wd.week} className="flex-1 flex flex-col-reverse h-full">
              {allFamilies.map(fam => {
                const count = wd.families.get(fam) || 0;
                const height = wd.totalRetained > 0 ? (count / maxRetained) * 100 : 0;
                return (
                  <div
                    key={fam}
                    style={{ height: `${height}%`, background: familyColorMap.get(fam) }}
                    className="w-full transition-all duration-300"
                    title={`Family ${fam}: ${count} paths (W${wd.week})`}
                  />
                );
              })}
            </div>
          ))}
        </div>
        <div className="flex justify-between text-[9px] text-gray-600 mt-1">
          {familyData.map(wd => (
            <span key={wd.week}>W{wd.week}</span>
          ))}
        </div>
      </Card>

      {/* Entropy overlay */}
      <Card title="Lineage Entropy">
        <div className="h-24 flex items-end gap-1">
          {familyData.map(wd => {
            const height = maxEntropy > 0 ? (wd.entropy / maxEntropy) * 100 : 0;
            const isLow = wd.entropy < maxEntropy * 0.3;
            return (
              <div key={wd.week} className="flex-1 flex flex-col items-center justify-end h-full">
                <span className="text-[8px] text-gray-500 mb-1">{wd.entropy.toFixed(2)}</span>
                <div
                  className={`w-full rounded-t ${isLow ? 'bg-red-500' : 'bg-cyan-500'}`}
                  style={{ height: `${Math.max(height, 3)}%` }}
                  title={`Week ${wd.week}: entropy ${wd.entropy.toFixed(3)}`}
                />
              </div>
            );
          })}
        </div>
        <div className="flex justify-between text-[9px] text-gray-600 mt-1">
          <span>Week 1</span>
          <span>Week {familyData.length}</span>
        </div>
        <p className="text-[9px] text-gray-600 mt-2">
          Higher entropy = more diverse beam. Red = potential monopoly risk.
        </p>
      </Card>

      {/* Monopoly timeline */}
      <Card title="Dominance Timeline">
        <div className="space-y-1">
          {familyData.map(wd => {
            const dominantPct = wd.totalRetained > 0 ? (wd.dominantSize / wd.totalRetained) * 100 : 0;
            const isMonopoly = dominantPct > 60;
            return (
              <div key={wd.week} className="flex items-center gap-2">
                <span className="w-8 text-[10px] text-gray-500">W{wd.week}</span>
                <div className="flex-1 h-4 bg-gray-800 rounded overflow-hidden relative">
                  <div
                    className={`h-full rounded ${isMonopoly ? 'bg-red-500' : 'bg-blue-600'}`}
                    style={{ width: `${dominantPct}%` }}
                  />
                  <span className="absolute right-1 top-0 h-full flex items-center text-[9px] text-gray-300">
                    {dominantPct.toFixed(0)}%
                  </span>
                </div>
                <span className="w-16 text-[9px] text-gray-500">
                  Fam {wd.dominantFamily}
                </span>
                {isMonopoly && <span className="text-[9px] text-red-400">⚠</span>}
              </div>
            );
          })}
        </div>
      </Card>

      {/* Tree growth: new branches vs deaths */}
      <Card title="Tree Growth (Births vs Deaths)">
        <div className="h-32 flex items-center gap-1">
          {familyData.map(wd => {
            const maxChange = Math.max(...familyData.map(d => Math.max(d.newBranches, d.deaths)), 1);
            const birthHeight = (wd.newBranches / maxChange) * 45;
            const deathHeight = (wd.deaths / maxChange) * 45;
            return (
              <div key={wd.week} className="flex-1 flex flex-col items-center h-full justify-center gap-0.5">
                {/* Births (up) */}
                <div className="flex flex-col justify-end h-[45%]">
                  <div
                    className="w-full bg-emerald-500 rounded-t"
                    style={{ height: `${Math.max(birthHeight, wd.newBranches > 0 ? 4 : 0)}%` }}
                    title={`+${wd.newBranches} new families`}
                  />
                </div>
                <div className="w-full h-px bg-gray-700" />
                {/* Deaths (down) */}
                <div className="flex flex-col justify-start h-[45%]">
                  <div
                    className="w-full bg-red-500 rounded-b"
                    style={{ height: `${Math.max(deathHeight, wd.deaths > 0 ? 4 : 0)}%` }}
                    title={`−${wd.deaths} extinct families`}
                  />
                </div>
              </div>
            );
          })}
        </div>
        <div className="flex justify-between text-[9px] text-gray-600 mt-1">
          {familyData.map(wd => (
            <span key={wd.week}>W{wd.week}</span>
          ))}
        </div>
        <div className="flex gap-4 mt-2 text-[9px] text-gray-500">
          <span className="flex items-center gap-1"><span className="w-2 h-2 bg-emerald-500 rounded-sm" />New families</span>
          <span className="flex items-center gap-1"><span className="w-2 h-2 bg-red-500 rounded-sm" />Extinct families</span>
        </div>
      </Card>

      {/* Family legend */}
      <Card title="Family Key">
        <div className="flex flex-wrap gap-2">
          {allFamilies.slice(0, 20).map(fam => (
            <span key={fam} className="flex items-center gap-1 text-[10px] text-gray-400">
              <span className="w-3 h-3 rounded-sm" style={{ background: familyColorMap.get(fam) }} />
              {fam}
            </span>
          ))}
          {allFamilies.length > 20 && (
            <span className="text-[10px] text-gray-600">+{allFamilies.length - 20} more</span>
          )}
        </div>
      </Card>
    </div>
  );
}
