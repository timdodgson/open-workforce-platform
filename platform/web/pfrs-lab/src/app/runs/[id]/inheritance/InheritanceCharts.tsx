'use client';
import { useMemo, useState } from 'react';
import {
  BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer,
  CartesianGrid, ScatterChart, Scatter, Legend, LineChart, Line,
} from 'recharts';
import Card from '@/components/Card';
import MetricCard from '@/components/MetricCard';
import { TreeNode } from '@/lib/types';

// ============================================================
// PHASE 2A: Foundation Analytics
// Each metric answers a specific research question.
// ============================================================

interface ParentStats {
  pathID: number;
  week: number;
  penalty: number;
  onWinningLineage: boolean;
  childCount: number;
  bestChildPenalty: number;
  avgChildPenalty: number;
  totalDescendants: number;
  descendantsReachingFinal: number;
  descendantsOnWinning: number;
  inheritanceScore: number;
  // Innovation.
  improvementGenerated: number;
  efficiency: number; // improvement per descendant
}

interface EntropyWeek {
  week: string;
  entropy: number;
  normalised: number; // 0-1
  familyCount: number;
}

interface KnowledgeTransfer {
  week: string;
  inherited: number; // % of penalty improvement that came from parent
  newDiscovery: number; // % from this week's optimisation
}

interface AncestorLifetime {
  pathID: number;
  birthWeek: number;
  deathWeek: number; // last week with living descendants
  generationsSurvived: number;
  onWinningLineage: boolean;
}

function computeAll(nodes: TreeNode[]) {
  // --- Core lookups ---
  const byID = new Map<number, TreeNode>();
  const childrenOf = new Map<number, TreeNode[]>();
  for (const n of nodes) {
    byID.set(n.pathID, n);
    if (!childrenOf.has(n.parentID)) childrenOf.set(n.parentID, []);
    childrenOf.get(n.parentID)!.push(n);
  }

  const weeks = Array.from(new Set(nodes.map(n => n.week))).sort((a, b) => a - b);
  const maxWeek = weeks.length > 0 ? weeks[weeks.length - 1] : 0;

  // --- Winning lineage ---
  const winningLineageIDs = new Set<number>();
  const winners = nodes.filter(n => n.winning);
  if (winners.length > 0) {
    let current = winners[winners.length - 1];
    while (current) {
      winningLineageIDs.add(current.pathID);
      if (current.parentID <= 0) break;
      const parent = byID.get(current.parentID);
      if (!parent) break;
      current = parent;
    }
  }

  const winningLineage = nodes
    .filter(n => winningLineageIDs.has(n.pathID))
    .sort((a, b) => a.week - b.week);

  // --- Helpers ---
  function countDescendants(id: number): number {
    const children = childrenOf.get(id) || [];
    let count = children.length;
    for (const c of children) count += countDescendants(c.pathID);
    return count;
  }

  function countDescendantsAtWeek(id: number, targetWeek: number): number {
    const children = childrenOf.get(id) || [];
    let count = 0;
    for (const c of children) {
      if (c.week === targetWeek && c.retained) count++;
      count += countDescendantsAtWeek(c.pathID, targetWeek);
    }
    return count;
  }

  function countDescendantsOnWinning(id: number): number {
    const children = childrenOf.get(id) || [];
    let count = 0;
    for (const c of children) {
      if (winningLineageIDs.has(c.pathID)) count++;
      count += countDescendantsOnWinning(c.pathID);
    }
    return count;
  }

  function getDeathWeek(id: number): number {
    const children = childrenOf.get(id) || [];
    if (children.length === 0) return byID.get(id)?.week || 0;
    let maxDeath = 0;
    for (const c of children) {
      const d = getDeathWeek(c.pathID);
      if (d > maxDeath) maxDeath = d;
    }
    return maxDeath;
  }

  // Find root ancestor of any node.
  function getRootAncestor(id: number): number {
    const node = byID.get(id);
    if (!node || node.parentID <= 0) return id;
    const parent = byID.get(node.parentID);
    if (!parent) return id;
    return getRootAncestor(parent.pathID);
  }

  // ============================================================
  // 1. LINEAGE ENTROPY
  // Question: Is the beam still exploring, or has it collapsed?
  // ============================================================
  const entropyData: EntropyWeek[] = [];
  for (const w of weeks) {
    const weekRetained = nodes.filter(n => n.week === w && n.retained);
    if (weekRetained.length === 0) continue;

    // For each retained path, find its root ancestor (week 1 family).
    const familyCounts = new Map<number, number>();
    for (const n of weekRetained) {
      const root = getRootAncestor(n.pathID);
      familyCounts.set(root, (familyCounts.get(root) || 0) + 1);
    }

    const total = weekRetained.length;
    let entropy = 0;
    for (const count of familyCounts.values()) {
      const p = count / total;
      if (p > 0) entropy -= p * Math.log2(p);
    }

    const maxEntropy = Math.log2(familyCounts.size) || 1;
    entropyData.push({
      week: `W${w}`,
      entropy: parseFloat(entropy.toFixed(3)),
      normalised: parseFloat((entropy / maxEntropy).toFixed(3)),
      familyCount: familyCounts.size,
    });
  }

  // ============================================================
  // 2. PARENT STATS + INNOVATION INDEX
  // Question: Which families actually create improvements?
  // ============================================================
  const parentStats: ParentStats[] = [];
  for (const node of nodes) {
    if (!node.retained) continue;
    const children = (childrenOf.get(node.pathID) || []).filter(c => c.retained);
    if (children.length === 0 && node.week === maxWeek) continue;

    const childPenalties = children.map(c => c.cumulativePenalty);
    const avgChild = childPenalties.length > 0
      ? childPenalties.reduce((s, p) => s + p, 0) / childPenalties.length : 0;
    const bestChild = childPenalties.length > 0 ? Math.min(...childPenalties) : 0;

    const totalDesc = countDescendants(node.pathID);
    const descFinal = countDescendantsAtWeek(node.pathID, maxWeek);
    const descWinning = countDescendantsOnWinning(node.pathID);

    // Innovation: how much total improvement did this family generate?
    const improvementGenerated = children.length > 0
      ? children.reduce((s, c) => s + Math.max(0, node.cumulativePenalty - c.cumulativePenalty), 0)
      : 0;
    const efficiency = totalDesc > 0 ? improvementGenerated / totalDesc : 0;

    const score = (descFinal * 10) + (improvementGenerated * 2) + (descWinning * 50);

    parentStats.push({
      pathID: node.pathID,
      week: node.week,
      penalty: node.cumulativePenalty,
      onWinningLineage: winningLineageIDs.has(node.pathID),
      childCount: children.length,
      bestChildPenalty: bestChild,
      avgChildPenalty: Math.round(avgChild),
      totalDescendants: totalDesc,
      descendantsReachingFinal: descFinal,
      descendantsOnWinning: descWinning,
      inheritanceScore: score,
      improvementGenerated,
      efficiency: parseFloat(efficiency.toFixed(2)),
    });
  }

  // ============================================================
  // 3. ANCESTOR LIFETIME
  // Question: Is beam pruning too aggressive?
  // ============================================================
  const lifetimes: AncestorLifetime[] = [];
  for (const node of nodes) {
    if (!node.retained || node.week === maxWeek) continue;
    const deathWeek = getDeathWeek(node.pathID);
    lifetimes.push({
      pathID: node.pathID,
      birthWeek: node.week,
      deathWeek,
      generationsSurvived: deathWeek - node.week,
      onWinningLineage: winningLineageIDs.has(node.pathID),
    });
  }

  // ============================================================
  // 4. PARENT EFFICIENCY
  // Question: Do excellent parents or average parents produce better offspring?
  // ============================================================
  // (Uses parentStats.efficiency — improvement per descendant)

  // ============================================================
  // 5. WEEKLY KNOWLEDGE TRANSFER
  // Question: Are later weeks refining inherited quality or creating new knowledge?
  // ============================================================
  const knowledgeData: KnowledgeTransfer[] = [];
  for (const w of weeks) {
    if (w === weeks[0]) continue; // skip first week (no parent)
    const weekRetained = nodes.filter(n => n.week === w && n.retained);
    if (weekRetained.length === 0) continue;

    let totalInherited = 0;
    let totalNew = 0;

    for (const n of weekRetained) {
      const parent = byID.get(n.parentID);
      if (!parent) continue;
      // "Inherited" = parent's cumulative improvement from root.
      // "New" = this week's optimisation contribution.
      const parentContribution = parent.cumulativePenalty; // penalty parent passed down
      const thisWeekContribution = n.weekPenalty; // what this week added
      const totalPath = n.cumulativePenalty;
      if (totalPath > 0) {
        totalInherited += parentContribution / totalPath;
        totalNew += thisWeekContribution / totalPath;
      }
    }

    const count = weekRetained.length;
    const inheritedPct = count > 0 ? (totalInherited / count) * 100 : 0;
    const newPct = count > 0 ? (totalNew / count) * 100 : 0;

    knowledgeData.push({
      week: `W${w}`,
      inherited: parseFloat(inheritedPct.toFixed(1)),
      newDiscovery: parseFloat(newPct.toFixed(1)),
    });
  }

  // ============================================================
  // PHASE 2B: BEAM HEALTH SCORE + DIVERSITY COLLAPSE DETECTION
  // Composite 0-100 score per week.
  // Inputs: entropy, beam spread, near-duplicate %, family count, innovation rate.
  // ============================================================
  interface BeamHealth {
    week: string;
    score: number; // 0-100
    entropy: number;
    beamSpread: number; // normalised 0-1
    familyDiversity: number; // families / beam_width
    innovationRate: number; // new discoveries this week
    status: 'excellent' | 'healthy' | 'converging' | 'premature' | 'collapsed';
  }

  const beamHealthData: BeamHealth[] = [];
  for (let i = 0; i < weeks.length; i++) {
    const w = weeks[i];
    const weekRetained = nodes.filter(n => n.week === w && n.retained);
    const weekAll = nodes.filter(n => n.week === w);
    if (weekRetained.length === 0) continue;

    // Entropy (already computed).
    const entropyEntry = entropyData.find(e => e.week === `W${w}`);
    const normEntropy = entropyEntry?.normalised || 0;

    // Beam spread: range of cumulative penalties among retained.
    const penalties = weekRetained.map(n => n.cumulativePenalty);
    const minP = Math.min(...penalties);
    const maxP = Math.max(...penalties);
    const spread = maxP > 0 ? (maxP - minP) / maxP : 0; // normalised

    // Family diversity: unique families / total retained.
    const families = new Set(weekRetained.map(n => getRootAncestor(n.pathID)));
    const familyDiv = weekRetained.length > 0 ? families.size / weekRetained.length : 0;

    // Innovation rate: how many retained paths improved over their parent.
    let improved = 0;
    for (const n of weekRetained) {
      const parent = byID.get(n.parentID);
      if (parent && n.cumulativePenalty < parent.cumulativePenalty) improved++;
    }
    const innovRate = weekRetained.length > 0 ? improved / weekRetained.length : 0;

    // Composite score (weighted).
    const score = Math.round(
      (normEntropy * 30) +        // 30% weight on entropy
      (spread * 100 * 15 / 100) + // 15% on beam spread (normalised)
      (familyDiv * 25) +          // 25% on family diversity
      (innovRate * 100 * 30 / 100) // 30% on innovation rate
    );
    const clampedScore = Math.min(100, Math.max(0, score));

    let status: BeamHealth['status'] = 'excellent';
    if (clampedScore >= 90) status = 'excellent';
    else if (clampedScore >= 70) status = 'healthy';
    else if (clampedScore >= 50) status = 'converging';
    else if (clampedScore >= 30) status = 'premature';
    else status = 'collapsed';

    beamHealthData.push({
      week: `W${w}`,
      score: clampedScore,
      entropy: normEntropy,
      beamSpread: parseFloat(spread.toFixed(3)),
      familyDiversity: parseFloat(familyDiv.toFixed(2)),
      innovationRate: parseFloat(innovRate.toFixed(2)),
      status,
    });
  }

  return { parentStats, entropyData, lifetimes, knowledgeData, winningLineage, winningLineageIDs, maxWeek, beamHealthData };
}

// ============================================================
// COMPONENT
// ============================================================

export default function InheritanceCharts({ nodes }: { nodes: TreeNode[] }) {
  const { parentStats, entropyData, lifetimes, knowledgeData, winningLineage, winningLineageIDs, maxWeek, beamHealthData } = useMemo(
    () => computeAll(nodes), [nodes]
  );

  // Summary.
  const totalRetained = parentStats.length;
  const onLineage = parentStats.filter(p => p.onWinningLineage).length;
  const avgEntropy = entropyData.length > 0
    ? (entropyData.reduce((s, e) => s + e.normalised, 0) / entropyData.length).toFixed(2) : '—';
  const finalEntropy = entropyData.length > 0 ? entropyData[entropyData.length - 1].normalised.toFixed(2) : '—';
  const avgLifetime = lifetimes.length > 0
    ? (lifetimes.reduce((s, l) => s + l.generationsSurvived, 0) / lifetimes.length).toFixed(1) : '—';

  // Top innovators (by efficiency).
  const innovatorData = [...parentStats]
    .filter(p => p.totalDescendants > 0)
    .sort((a, b) => b.efficiency - a.efficiency)
    .slice(0, 20)
    .map(p => ({
      parent: `P${p.pathID}`,
      efficiency: p.efficiency,
      descendants: p.totalDescendants,
      onLineage: p.onWinningLineage,
    }));

  // Lifetime histogram.
  const lifetimeHist = new Map<number, number>();
  for (const l of lifetimes) {
    lifetimeHist.set(l.generationsSurvived, (lifetimeHist.get(l.generationsSurvived) || 0) + 1);
  }
  const lifetimeData = Array.from(lifetimeHist.entries())
    .sort((a, b) => a[0] - b[0])
    .map(([gen, count]) => ({ generations: `${gen}`, count }));

  // Parent efficiency scatter.
  const efficiencyData = parentStats
    .filter(p => p.totalDescendants > 0)
    .map(p => ({
      penalty: p.penalty,
      efficiency: p.efficiency,
      pathID: p.pathID,
      onLineage: p.onWinningLineage,
    }));

  return (
    <>
      {/* Summary */}
      <div className="grid grid-cols-2 sm:grid-cols-5 gap-3 mb-4">
        <MetricCard label="Retained Paths" value={String(totalRetained)} color="blue" />
        <MetricCard label="On Winning Lineage" value={String(onLineage)} color="green" />
        <MetricCard label="Avg Entropy" value={avgEntropy} color="amber" />
        <MetricCard label="Final Entropy" value={finalEntropy} color={parseFloat(finalEntropy) < 0.3 ? 'red' : 'green'} />
        <MetricCard label="Avg Lifetime" value={`${avgLifetime} gen`} color="default" />
      </div>

      {/* Winning Lineage */}
      <Card title="Winning Lineage">
        <div className="flex gap-2 overflow-x-auto pb-2">
          {winningLineage.map(n => (
            <div key={n.pathID} className="flex flex-col items-center min-w-[70px]">
              <div className="text-[9px] text-gray-500">W{n.week}</div>
              <div className={`w-8 h-8 rounded-full flex items-center justify-center text-[10px] font-bold ${
                n.winning ? 'bg-emerald-500 text-white' : 'bg-blue-500/80 text-white'
              }`}>
                {n.pathID}
              </div>
              <div className="text-[9px] text-gray-400 mt-1">{n.cumulativePenalty}</div>
            </div>
          ))}
        </div>
      </Card>

      {/* PHASE 2B: Beam Health Score */}
      <Card title="Beam Health Score — Search Quality Over Time">
        <p className="text-xs text-gray-500 mb-3">
          Composite 0-100 score. 90+ Excellent, 70+ Healthy, 50+ Converging, 30+ Premature convergence, &lt;30 Collapsed.
        </p>
        <div className="flex gap-1 mb-3">
          {beamHealthData.map(h => {
            const color = h.status === 'excellent' ? 'bg-emerald-500'
              : h.status === 'healthy' ? 'bg-blue-500'
              : h.status === 'converging' ? 'bg-amber-500'
              : h.status === 'premature' ? 'bg-orange-500'
              : 'bg-red-500';
            return (
              <div key={h.week} className="flex-1 text-center">
                <div className="text-[9px] text-gray-500 mb-1">{h.week}</div>
                <div className={`${color} rounded py-1 text-[10px] font-bold text-white`}>{h.score}</div>
                <div className="text-[8px] text-gray-600 mt-1">{h.status}</div>
              </div>
            );
          })}
        </div>
        <ResponsiveContainer width="100%" height={160}>
          <LineChart data={beamHealthData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="week" stroke="#9ca3af" fontSize={11} />
            <YAxis stroke="#9ca3af" fontSize={11} domain={[0, 100]} />
            <Tooltip contentStyle={{ background: '#1f2937', border: '1px solid #374151' }} />
            <Line type="monotone" dataKey="score" stroke="#34d399" strokeWidth={2} dot={{ r: 4, fill: '#34d399' }} />
          </LineChart>
        </ResponsiveContainer>
      </Card>

      {/* 1. Lineage Entropy */}
      <Card title="Lineage Entropy — Ancestral Diversity Over Time">
        <p className="text-xs text-gray-500 mb-3">
          High entropy = diverse ancestry. Low entropy = beam has collapsed into one dominant bloodline.
        </p>
        <ResponsiveContainer width="100%" height={200}>
          <LineChart data={entropyData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="week" stroke="#9ca3af" fontSize={11} />
            <YAxis stroke="#9ca3af" fontSize={11} domain={[0, 1]} />
            <Tooltip contentStyle={{ background: '#1f2937', border: '1px solid #374151' }} />
            <Legend wrapperStyle={{ fontSize: 11 }} />
            <Line type="monotone" dataKey="normalised" stroke="#fbbf24" strokeWidth={2} name="Normalised Entropy" dot={{ r: 4 }} />
          </LineChart>
        </ResponsiveContainer>
        <div className="flex gap-4 mt-2 text-[10px] text-gray-500">
          {entropyData.map(e => (
            <span key={e.week}>{e.week}: {e.familyCount} families</span>
          ))}
        </div>
      </Card>

      {/* 5. Weekly Knowledge Transfer */}
      <Card title="Weekly Knowledge Transfer — Inherited vs New Discovery">
        <p className="text-xs text-gray-500 mb-3">
          Are later weeks refining inherited quality, or still creating genuinely new knowledge?
        </p>
        <ResponsiveContainer width="100%" height={200}>
          <BarChart data={knowledgeData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="week" stroke="#9ca3af" fontSize={11} />
            <YAxis stroke="#9ca3af" fontSize={11} unit="%" />
            <Tooltip contentStyle={{ background: '#1f2937', border: '1px solid #374151' }} />
            <Legend wrapperStyle={{ fontSize: 11 }} />
            <Bar dataKey="inherited" stackId="a" fill="#60a5fa" name="Inherited %" />
            <Bar dataKey="newDiscovery" stackId="a" fill="#34d399" name="New Discovery %" />
          </BarChart>
        </ResponsiveContainer>
      </Card>

      {/* 2. Innovation Index / Parent Efficiency */}
      <Card title="Innovation Index — Improvement per Descendant">
        <p className="text-xs text-gray-500 mb-3">
          Distinguishes large families from useful families. High efficiency = creates improvements, not just copies.
        </p>
        <ResponsiveContainer width="100%" height={220}>
          <ScatterChart>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis type="number" dataKey="penalty" name="Parent Penalty" stroke="#9ca3af" fontSize={11} />
            <YAxis type="number" dataKey="efficiency" name="Efficiency" stroke="#9ca3af" fontSize={11} />
            <Tooltip contentStyle={{ background: '#1f2937', border: '1px solid #374151' }} />
            <Legend wrapperStyle={{ fontSize: 11 }} />
            <Scatter data={efficiencyData.filter(d => d.onLineage)} fill="#34d399" name="Winning Lineage" opacity={1} />
            <Scatter data={efficiencyData.filter(d => !d.onLineage)} fill="#60a5fa" name="Other" opacity={0.5} />
          </ScatterChart>
        </ResponsiveContainer>
      </Card>

      {/* 3. Ancestor Lifetime */}
      <Card title="Ancestor Lifetime — How Long Do Paths Survive?">
        <p className="text-xs text-gray-500 mb-3">
          Distribution of how many generations each ancestor's lineage persists. Short lifetimes may indicate aggressive pruning.
        </p>
        <ResponsiveContainer width="100%" height={180}>
          <BarChart data={lifetimeData}>
            <CartesianGrid strokeDasharray="3 3" stroke="#374151" />
            <XAxis dataKey="generations" stroke="#9ca3af" fontSize={11} label={{ value: 'Generations', position: 'bottom', fontSize: 10, fill: '#6b7280' }} />
            <YAxis stroke="#9ca3af" fontSize={11} />
            <Tooltip contentStyle={{ background: '#1f2937', border: '1px solid #374151' }} />
            <Bar dataKey="count" fill="#a78bfa" name="Paths" radius={[3, 3, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </Card>

      {/* Lineage Dominance Table */}
      <Card title="Top Innovators — Ranked by Inheritance Score">
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Path</th>
              <th className="p-2">Week</th>
              <th className="text-right p-2">Penalty</th>
              <th className="text-right p-2">Children</th>
              <th className="text-right p-2">Descendants</th>
              <th className="text-right p-2">Improvement</th>
              <th className="text-right p-2">Efficiency</th>
              <th className="text-right p-2">Score</th>
            </tr>
          </thead>
          <tbody>
            {[...parentStats]
              .sort((a, b) => b.inheritanceScore - a.inheritanceScore)
              .slice(0, 15)
              .map(p => (
                <tr key={p.pathID} className={`border-t border-gray-800 ${p.onWinningLineage ? 'bg-emerald-900/20' : ''}`}>
                  <td className="p-2 font-mono">{p.pathID}</td>
                  <td className="p-2 text-center">{p.week}</td>
                  <td className="text-right p-2">{p.penalty.toLocaleString()}</td>
                  <td className="text-right p-2">{p.childCount}</td>
                  <td className="text-right p-2">{p.totalDescendants}</td>
                  <td className="text-right p-2">{p.improvementGenerated}</td>
                  <td className="text-right p-2">{p.efficiency.toFixed(1)}</td>
                  <td className="text-right p-2 font-bold">{p.inheritanceScore}</td>
                </tr>
              ))}
          </tbody>
        </table>
      </Card>

      {/* PHASE 2C: Evolution Flow — Visual representation of path ancestry */}
      <Card title="Evolution Flow — Path Ancestry by Week">
        <p className="text-xs text-gray-500 mb-3">
          Each row is a week. Each node is a retained path. Lines connect parents to children.
          Width indicates descendant count. Green = winning lineage.
        </p>
        <EvolutionFlow nodes={nodes} winningLineageIDs={winningLineageIDs} />
      </Card>

      {/* PHASE 2C: Innovation Timeline — Global bests coloured by lineage */}
      <Card title="Innovation Timeline — Global Bests by Lineage">
        <p className="text-xs text-gray-500 mb-3">
          Each improvement event coloured by which ancestral family produced it. Reveals whether innovation comes from one family or many.
        </p>
        <InnovationTimeline nodes={nodes} winningLineageIDs={winningLineageIDs} />
      </Card>
    </>
  );
}


// ============================================================
// PHASE 2C SUB-COMPONENTS
// ============================================================

const LINEAGE_COLORS = ['#34d399', '#60a5fa', '#fbbf24', '#f87171', '#a78bfa', '#fb923c', '#2dd4bf', '#e879f9', '#84cc16', '#f472b6'];

function EvolutionFlow({ nodes, winningLineageIDs }: { nodes: TreeNode[]; winningLineageIDs: Set<number> }) {
  const [selected, setSelected] = useState<number | null>(null);

  const weeks = Array.from(new Set(nodes.map(n => n.week))).sort((a, b) => a - b);
  const retained = nodes.filter(n => n.retained);
  const byID = new Map(nodes.map(n => [n.pathID, n]));

  // Count descendants for width.
  const childrenOf = new Map<number, TreeNode[]>();
  for (const n of nodes) {
    if (!childrenOf.has(n.parentID)) childrenOf.set(n.parentID, []);
    childrenOf.get(n.parentID)!.push(n);
  }
  function countDesc(id: number): number {
    const children = childrenOf.get(id) || [];
    let c = children.length;
    for (const ch of children) c += countDesc(ch.pathID);
    return c;
  }

  // Assign colour by root ancestor.
  function getRoot(id: number): number {
    const n = byID.get(id);
    if (!n || n.parentID <= 0) return id;
    return getRoot(n.parentID);
  }
  const rootIDs = Array.from(new Set(retained.map(n => getRoot(n.pathID))));
  const colorMap = new Map<number, string>();
  rootIDs.forEach((r, i) => colorMap.set(r, LINEAGE_COLORS[i % LINEAGE_COLORS.length]));

  // Compute highlighted set: ancestors + descendants of selected node.
  const highlighted = useMemo(() => {
    if (selected === null) return null;
    const set = new Set<number>([selected]);
    // Walk up to root (ancestors).
    let cur = byID.get(selected);
    while (cur && cur.parentID > 0) {
      set.add(cur.parentID);
      cur = byID.get(cur.parentID);
    }
    // Walk down to leaves (descendants).
    function addDescendants(id: number) {
      const children = (childrenOf.get(id) || []).filter(c => c.retained);
      for (const c of children) {
        set.add(c.pathID);
        addDescendants(c.pathID);
      }
    }
    addDescendants(selected);
    return set;
  }, [selected, byID, childrenOf]);

  return (
    <div className="overflow-x-auto">
      {selected !== null && (
        <button onClick={() => setSelected(null)} className="text-[10px] text-gray-500 hover:text-white mb-2 underline">
          Clear selection
        </button>
      )}
      <div className="flex gap-6 min-w-[600px]">
        {weeks.map(w => {
          const weekNodes = retained.filter(n => n.week === w)
            .sort((a, b) => a.cumulativePenalty - b.cumulativePenalty);
          return (
            <div key={w} className="flex flex-col items-center gap-1 min-w-[60px]">
              <div className="text-[9px] text-gray-500 font-bold mb-1">W{w}</div>
              {weekNodes.map(n => {
                const desc = countDesc(n.pathID);
                const size = Math.max(20, Math.min(40, 16 + desc * 3));
                const root = getRoot(n.pathID);
                const color = winningLineageIDs.has(n.pathID) ? '#34d399' : (colorMap.get(root) || '#6b7280');
                const isHighlighted = highlighted === null || highlighted.has(n.pathID);
                const isSelected = selected === n.pathID;
                return (
                  <button key={n.pathID}
                    onClick={() => setSelected(selected === n.pathID ? null : n.pathID)}
                    className={`rounded-full flex items-center justify-center text-[8px] font-bold text-white border transition-all cursor-pointer ${
                      isSelected ? 'ring-2 ring-white border-white' : 'border-gray-700 hover:border-gray-400'
                    }`}
                    style={{
                      width: size, height: size,
                      backgroundColor: color,
                      opacity: isHighlighted ? 1 : 0.15,
                    }}
                    title={`P${n.pathID} | Cum: ${n.cumulativePenalty} | Desc: ${desc} | Parent: ${n.parentID}`}
                  >
                    {n.pathID}
                  </button>
                );
              })}
            </div>
          );
        })}
      </div>
    </div>
  );
}

function InnovationTimeline({ nodes, winningLineageIDs }: { nodes: TreeNode[]; winningLineageIDs: Set<number> }) {
  // Show retained paths that performed better than average for their week, coloured by root family.
  const byID = new Map(nodes.map(n => [n.pathID, n]));
  const retained = nodes.filter(n => n.retained && n.parentID > 0);

  function getRoot(id: number): number {
    const n = byID.get(id);
    if (!n || n.parentID <= 0) return id;
    return getRoot(n.parentID);
  }

  const rootIDs = Array.from(new Set(retained.map(n => getRoot(n.pathID))));
  const colorMap = new Map<number, string>();
  rootIDs.forEach((r, i) => colorMap.set(r, LINEAGE_COLORS[i % LINEAGE_COLORS.length]));

  // For each week, compute average week penalty. Paths below average are "innovators".
  const weeks = Array.from(new Set(retained.map(n => n.week))).sort((a, b) => a - b);
  const weekAvg = new Map<number, number>();
  for (const w of weeks) {
    const weekNodes = retained.filter(n => n.week === w);
    const avg = weekNodes.reduce((s, n) => s + n.weekPenalty, 0) / weekNodes.length;
    weekAvg.set(w, avg);
  }

  // Innovation = how much better than average this path is.
  const innovations = retained
    .map(n => {
      const avg = weekAvg.get(n.week) || 0;
      const improvement = Math.round(avg - n.weekPenalty);
      if (improvement <= 0) return null;
      const root = getRoot(n.pathID);
      return {
        week: n.week,
        pathID: n.pathID,
        improvement,
        penalty: n.weekPenalty,
        color: colorMap.get(root) || '#6b7280',
        onWinning: winningLineageIDs.has(n.pathID),
        familyID: root,
      };
    })
    .filter(Boolean) as { week: number; pathID: number; improvement: number; penalty: number; color: string; onWinning: boolean; familyID: number }[];

  if (innovations.length === 0) {
    return <p className="text-xs text-gray-500">No improvement events found.</p>;
  }

  // Group by family for legend.
  const familiesUsed = Array.from(new Set(innovations.map(i => i.familyID)));

  return (
    <div>
      <div className="flex gap-3 mb-2 flex-wrap">
        {familiesUsed.map(f => (
          <span key={f} className="text-[9px] flex items-center gap-1">
            <span className="w-3 h-3 rounded-full inline-block" style={{ backgroundColor: colorMap.get(f) }}></span>
            Family {f}
          </span>
        ))}
      </div>
      <div className="flex items-end gap-[2px] h-[160px] overflow-x-auto">
        {innovations.sort((a, b) => a.week - b.week || a.pathID - b.pathID).map((inv, i) => {
          const maxImp = Math.max(...innovations.map(x => x.improvement));
          const height = Math.max(4, (inv.improvement / maxImp) * 140);
          return (
            <div key={i} className="flex flex-col items-center justify-end" style={{ height: '100%' }}>
              <div
                className={`rounded-t-sm ${inv.onWinning ? 'ring-1 ring-white' : ''}`}
                style={{ width: 6, height, backgroundColor: inv.color, opacity: inv.onWinning ? 1 : 0.7 }}
                title={`W${inv.week} P${inv.pathID} | Improvement: ${inv.improvement} | Family: ${inv.familyID}`}
              />
              {inv.week !== innovations[i - 1]?.week && (
                <div className="text-[7px] text-gray-600 mt-1">W{inv.week}</div>
              )}
            </div>
          );
        })}
      </div>
    </div>
  );
}
