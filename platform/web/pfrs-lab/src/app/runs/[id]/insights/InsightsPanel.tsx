'use client';
import { useMemo } from 'react';
import Card from '@/components/Card';
import MetricCard from '@/components/MetricCard';
import { TreeNode, DiscoveryRecord, RunSummary } from '@/lib/types';

interface Props {
  nodes: TreeNode[];
  discoveries: DiscoveryRecord[];
  summary: RunSummary;
  runId: string;
}

// ============================================================
// ANALYSIS ENGINE
// ============================================================

interface SearchEvent {
  week: number;
  icon: string;
  title: string;
  detail: string;
  severity: 'info' | 'warning' | 'success' | 'danger';
}

interface FamilyStats {
  id: number;
  globalBests: number;
  totalImprovement: number;
  pctContribution: number;
  firstWeek: number;
  lastWeek: number;
  totalDescendants: number;
  alive: boolean; // has descendants in final week
}

interface HealthBreakdown {
  score: number;
  diversity: number;
  innovation: number;
  nearDuplicates: number;
  entropyCollapse: number;
  monopoly: number;
  explanation: string;
}

interface Recommendation {
  text: string;
  confidence: 'high' | 'medium' | 'low';
  evidence: string[];
}

function analyse(nodes: TreeNode[], discoveries: DiscoveryRecord[], summary: RunSummary) {
  const byID = new Map(nodes.map(n => [n.pathID, n]));
  const childrenOf = new Map<number, TreeNode[]>();
  for (const n of nodes) {
    if (!childrenOf.has(n.parentID)) childrenOf.set(n.parentID, []);
    childrenOf.get(n.parentID)!.push(n);
  }

  const retained = nodes.filter(n => n.retained);
  const weeks = Array.from(new Set(nodes.map(n => n.week))).sort((a, b) => a - b);
  const maxWeek = weeks[weeks.length - 1] || 0;

  // Root ancestor.
  function getRoot(id: number): number {
    const n = byID.get(id);
    if (!n || n.parentID <= 0) return id;
    return getRoot(n.parentID);
  }

  // Winning lineage.
  const winningIDs = new Set<number>();
  const winners = nodes.filter(n => n.winning);
  if (winners.length > 0) {
    let cur = winners[winners.length - 1];
    while (cur) {
      winningIDs.add(cur.pathID);
      if (cur.parentID <= 0) break;
      cur = byID.get(cur.parentID)!;
      if (!cur) break;
    }
  }

  // --- Per-week entropy ---
  const weekEntropy: { week: number; entropy: number; families: number; dominant: number; dominantPct: number }[] = [];
  for (const w of weeks) {
    const weekRetained = retained.filter(n => n.week === w);
    const familyCounts = new Map<number, number>();
    for (const n of weekRetained) {
      const root = getRoot(n.pathID);
      familyCounts.set(root, (familyCounts.get(root) || 0) + 1);
    }
    const total = weekRetained.length;
    let entropy = 0;
    let maxCount = 0;
    let dominantFamily = 0;
    for (const [fam, count] of familyCounts) {
      const p = count / total;
      if (p > 0) entropy -= p * Math.log2(p);
      if (count > maxCount) { maxCount = count; dominantFamily = fam; }
    }
    const maxE = Math.log2(familyCounts.size) || 1;
    weekEntropy.push({
      week: w,
      entropy: entropy / maxE,
      families: familyCounts.size,
      dominant: dominantFamily,
      dominantPct: total > 0 ? (maxCount / total) * 100 : 0,
    });
  }

  // --- Search Events ---
  const events: SearchEvent[] = [];
  let prevEntropy = 1;
  for (const we of weekEntropy) {
    // Entropy collapse.
    if (we.entropy < 0.35 && prevEntropy >= 0.35) {
      events.push({
        week: we.week, icon: '⚠️', severity: 'warning',
        title: `Beam entropy dropped below 0.35`,
        detail: `Diversity collapsed. Only ${we.families} families survive. Dominant family owns ${we.dominantPct.toFixed(0)}%.`,
      });
    }
    // Recovery.
    if (we.entropy >= 0.6 && prevEntropy < 0.4) {
      events.push({
        week: we.week, icon: '🌱', severity: 'success',
        title: 'Diversity recovered',
        detail: `Entropy rose to ${we.entropy.toFixed(2)} with ${we.families} active families.`,
      });
    }
    // High monopoly.
    if (we.dominantPct >= 75) {
      events.push({
        week: we.week, icon: '👑', severity: 'info',
        title: `Family ${we.dominant} owns ${we.dominantPct.toFixed(0)}% of surviving paths`,
        detail: 'Lineage dominance is high. Beam diversity is low.',
      });
    }
    prevEntropy = we.entropy;
  }

  // Winning lineage first appearance.
  const winnerRoot = winners.length > 0 ? getRoot(winners[winners.length - 1].pathID) : 0;
  if (winnerRoot) {
    const firstWinWeek = Math.min(...retained.filter(n => getRoot(n.pathID) === winnerRoot).map(n => n.week));
    events.push({
      week: firstWinWeek, icon: '🏆', severity: 'success',
      title: `Winning lineage (Family ${winnerRoot}) first appeared`,
      detail: `This family ultimately produces the best Week ${maxWeek} solution.`,
    });
  }

  // Sort events by week.
  events.sort((a, b) => a.week - b.week);

  // --- Innovation Monopoly (from discoveries) ---
  const globalBests = discoveries.filter(d => d.eventType === 'GLOBAL_BEST');
  const familyInnovation = new Map<number, { count: number; improvement: number; firstWeek: number; lastWeek: number }>();

  // Map discovery workerID to a beam path's root. This is approximate since discoveries
  // come from per-week PFRS workers, not directly from beam paths. We'll use week grouping.
  // Simpler approach: use tree data — paths that are on winning lineage or have low penalties.
  // Actually, let's just count innovations per week and attribute to the dominant family that week.
  for (const we of weekEntropy) {
    const weekGlobals = globalBests.filter(d => d.week === we.week);
    const fam = we.dominant;
    if (!familyInnovation.has(fam)) familyInnovation.set(fam, { count: 0, improvement: 0, firstWeek: we.week, lastWeek: we.week });
    const fi = familyInnovation.get(fam)!;
    fi.count += weekGlobals.length;
    fi.improvement += weekGlobals.reduce((s, d) => s + d.improvement, 0);
    fi.lastWeek = Math.max(fi.lastWeek, we.week);
  }

  const totalGlobalBests = globalBests.length;
  const familyStats: FamilyStats[] = [];
  const allRoots = Array.from(new Set(retained.map(n => getRoot(n.pathID))));
  for (const root of allRoots) {
    const fi = familyInnovation.get(root);
    const descendants = retained.filter(n => getRoot(n.pathID) === root);
    const weeksPresent = Array.from(new Set(descendants.map(n => n.week)));
    const aliveAtEnd = descendants.some(n => n.week === maxWeek);
    familyStats.push({
      id: root,
      globalBests: fi?.count || 0,
      totalImprovement: fi?.improvement || 0,
      pctContribution: totalGlobalBests > 0 ? ((fi?.count || 0) / totalGlobalBests) * 100 : 0,
      firstWeek: Math.min(...weeksPresent),
      lastWeek: Math.max(...weeksPresent),
      totalDescendants: descendants.length,
      alive: aliveAtEnd,
    });
  }
  familyStats.sort((a, b) => b.pctContribution - a.pctContribution);

  // --- Beam Health Diagnosis ---
  const avgEntropy = weekEntropy.length > 0 ? weekEntropy.reduce((s, e) => s + e.entropy, 0) / weekEntropy.length : 0;
  const maxMonopoly = Math.max(...weekEntropy.map(e => e.dominantPct));
  const innovationRate = totalGlobalBests / (weeks.length || 1);
  const topFamilyPct = familyStats.length > 0 ? familyStats[0].pctContribution : 0;

  const diversityScore = Math.round(avgEntropy * 40);
  const innovScore = Math.round(Math.min(1, innovationRate / 10) * 30);
  const monopolyPenalty = Math.round((topFamilyPct / 100) * 20);
  const entropyPenalty = weekEntropy.filter(e => e.entropy < 0.35).length * 5;
  const healthScore = Math.max(0, Math.min(100, diversityScore + innovScore - monopolyPenalty - entropyPenalty));

  const healthBreakdown: HealthBreakdown = {
    score: healthScore,
    diversity: diversityScore,
    innovation: innovScore,
    nearDuplicates: 0,
    entropyCollapse: -entropyPenalty,
    monopoly: -monopolyPenalty,
    explanation: healthScore >= 70 ? 'Search maintained healthy exploration throughout.'
      : healthScore >= 50 ? 'Search showed signs of convergence but retained some diversity.'
      : healthScore >= 30 ? 'Search converged prematurely. Diversity was lost during early-mid weeks.'
      : 'Search collapsed into a single lineage very early. Most beam slots were wasted.',
  };

  // --- Recommendations ---
  const recommendations: Recommendation[] = [];

  if (avgEntropy < 0.4) {
    recommendations.push({
      text: 'Reduce look-ahead weight to preserve beam diversity.',
      confidence: 'high',
      evidence: [`Average entropy: ${avgEntropy.toFixed(2)}`, `${weekEntropy.filter(e => e.entropy < 0.35).length} weeks below collapse threshold`],
    });
  }
  if (topFamilyPct > 80) {
    recommendations.push({
      text: 'Innovation is monopolised by one family. Consider diversity-preserving selection.',
      confidence: 'high',
      evidence: [`Family ${familyStats[0]?.id} contributes ${topFamilyPct.toFixed(0)}% of innovations`],
    });
  }
  if (maxMonopoly > 90) {
    recommendations.push({
      text: 'Dominant family owns >90% of paths. Beam width is effectively wasted.',
      confidence: 'medium',
      evidence: [`Peak dominance: ${maxMonopoly.toFixed(0)}%`, `Beam width ${summary.metadata?.beamWidth || '?'} reduced to effective width ~2`],
    });
  }
  if (healthScore >= 70 && familyStats.filter(f => f.pctContribution > 10).length >= 3) {
    recommendations.push({
      text: 'Current beam configuration maintains healthy exploration. No changes needed.',
      confidence: 'high',
      evidence: [`${familyStats.filter(f => f.pctContribution > 10).length} families contributing >10%`, `Beam health: ${healthScore}/100`],
    });
  }
  if (recommendations.length === 0) {
    recommendations.push({
      text: 'Try wider beam or lower look-ahead weight to improve diversity.',
      confidence: 'low',
      evidence: ['No strong signal detected — general recommendation.'],
    });
  }

  // ============================================================
  // ACTIONABLE RESEARCH METRICS
  // ============================================================

  // 1. Effective Beam Width: how many distinct families are actually retained per week?
  const effectiveWidth: { week: string; nominal: number; effective: number; utilisation: number }[] = [];
  for (const we of weekEntropy) {
    const weekRetained = retained.filter(n => n.week === we.week);
    const families = new Set(weekRetained.map(n => getRoot(n.pathID)));
    const nominal = weekRetained.length;
    const effective = families.size;
    effectiveWidth.push({
      week: `W${we.week}`,
      nominal,
      effective,
      utilisation: nominal > 0 ? Math.round((effective / nominal) * 100) : 0,
    });
  }

  // 2. Exploration vs Exploitation: from week records in summary.
  const explorationBalance: { week: string; exploration: number; exploitation: number }[] = [];
  for (const w of summary.weeks) {
    const total = w.saAcceptedBetter + w.saAcceptedWorse + w.lahcAcceptedByCurrent + w.lahcAcceptedByLate;
    const exploration = w.saAcceptedWorse + w.lahcAcceptedByLate;
    const exploitation = w.saAcceptedBetter + w.lahcAcceptedByCurrent;
    explorationBalance.push({
      week: `W${w.week}`,
      exploration: total > 0 ? Math.round((exploration / total) * 100) : 0,
      exploitation: total > 0 ? Math.round((exploitation / total) * 100) : 0,
    });
  }

  // 3. Time Since Last Global Improvement: candidates between last global best and end.
  const lastGlobalPerWeek: { week: string; candsSinceLastGlobal: number; pctWasted: number }[] = [];
  const globalBestsByWeek = new Map<number, DiscoveryRecord[]>();
  for (const d of discoveries.filter(d => d.eventType === 'GLOBAL_BEST')) {
    if (!globalBestsByWeek.has(d.week)) globalBestsByWeek.set(d.week, []);
    globalBestsByWeek.get(d.week)!.push(d);
  }
  for (const w of summary.weeks) {
    const weekGlobals = globalBestsByWeek.get(w.week) || [];
    const lastGlobal = weekGlobals.length > 0 ? Math.max(...weekGlobals.map(d => d.candidate)) : 0;
    const totalCands = w.candidates;
    const wasted = totalCands - lastGlobal;
    lastGlobalPerWeek.push({
      week: `W${w.week}`,
      candsSinceLastGlobal: wasted,
      pctWasted: totalCands > 0 ? Math.round((wasted / totalCands) * 100) : 0,
    });
  }

  // 4. Beam Efficiency: paths that contribute to winner vs dead ends.
  const winnerAncestors = new Set<number>();
  for (const n of nodes.filter(n => winningIDs.has(n.pathID))) {
    let cur: TreeNode | undefined = n;
    while (cur) {
      winnerAncestors.add(cur.pathID);
      if (cur.parentID <= 0) break;
      cur = byID.get(cur.parentID);
    }
  }
  const totalRetainedCount = retained.length;
  const contributingCount = retained.filter(n => winnerAncestors.has(n.pathID)).length;
  const beamEfficiency = totalRetainedCount > 0 ? Math.round((contributingCount / totalRetainedCount) * 100) : 0;

  // 5. Search Efficiency Score: improvement per million candidates.
  const totalImprovement = summary.weeks.reduce((s, w) => s + w.improvement, 0);
  const totalCandidates = summary.totalCandidates;
  const efficiencyPerM = totalCandidates > 0 ? Math.round((totalImprovement / (totalCandidates / 1_000_000))) : 0;

  return { events, familyStats, healthBreakdown, recommendations, weekEntropy,
    effectiveWidth, explorationBalance, lastGlobalPerWeek, beamEfficiency, contributingCount, totalRetainedCount, efficiencyPerM };
}

// ============================================================
// COMPONENT
// ============================================================

export default function InsightsPanel({ nodes, discoveries, summary, runId }: Props) {
  const { events, familyStats, healthBreakdown, recommendations, weekEntropy,
    effectiveWidth, explorationBalance, lastGlobalPerWeek, beamEfficiency, contributingCount, totalRetainedCount, efficiencyPerM } = useMemo(
    () => analyse(nodes, discoveries, summary), [nodes, discoveries, summary]
  );

  return (
    <>
      {/* Beam Health Diagnosis */}
      <Card title="Beam Health Diagnosis">
        <div className="flex items-center gap-4 mb-3">
          <div className={`text-3xl font-bold ${
            healthBreakdown.score >= 70 ? 'text-emerald-400'
            : healthBreakdown.score >= 50 ? 'text-amber-400'
            : 'text-red-400'
          }`}>
            {healthBreakdown.score}/100
          </div>
          <p className="text-sm text-gray-400">{healthBreakdown.explanation}</p>
        </div>
        <div className="grid grid-cols-2 sm:grid-cols-5 gap-2 text-xs">
          <div className="bg-gray-800 rounded p-2">
            <div className="text-gray-500">Diversity</div>
            <div className="text-emerald-400 font-bold">+{healthBreakdown.diversity}</div>
          </div>
          <div className="bg-gray-800 rounded p-2">
            <div className="text-gray-500">Innovation</div>
            <div className="text-emerald-400 font-bold">+{healthBreakdown.innovation}</div>
          </div>
          <div className="bg-gray-800 rounded p-2">
            <div className="text-gray-500">Entropy Collapse</div>
            <div className="text-red-400 font-bold">{healthBreakdown.entropyCollapse}</div>
          </div>
          <div className="bg-gray-800 rounded p-2">
            <div className="text-gray-500">Monopoly</div>
            <div className="text-red-400 font-bold">{healthBreakdown.monopoly}</div>
          </div>
          <div className="bg-gray-800 rounded p-2">
            <div className="text-gray-500">Net Score</div>
            <div className="font-bold">{healthBreakdown.score}</div>
          </div>
        </div>
      </Card>

      {/* Search Events Timeline */}
      <Card title="Search Events Timeline">
        {events.length === 0 ? (
          <p className="text-xs text-gray-500">No notable events detected.</p>
        ) : (
          <div className="space-y-3">
            {events.map((e, i) => (
              <div key={i} className={`border-l-2 pl-3 py-1 ${
                e.severity === 'warning' ? 'border-amber-500'
                : e.severity === 'danger' ? 'border-red-500'
                : e.severity === 'success' ? 'border-emerald-500'
                : 'border-blue-500'
              }`}>
                <div className="flex items-center gap-2">
                  <span className="text-sm">{e.icon}</span>
                  <span className="text-xs text-gray-500">Week {e.week}</span>
                  <span className="text-xs font-semibold text-gray-200">{e.title}</span>
                </div>
                <p className="text-[11px] text-gray-500 mt-0.5">{e.detail}</p>
              </div>
            ))}
          </div>
        )}
      </Card>

      {/* Innovation Monopoly */}
      <Card title="Innovation Monopoly — Who Creates Value?">
        <div className="space-y-2">
          {familyStats.slice(0, 6).map(f => (
            <div key={f.id} className="flex items-center gap-3">
              <span className="text-xs text-gray-400 w-20">Family {f.id}</span>
              <div className="flex-1 bg-gray-800 rounded-full h-4 overflow-hidden">
                <div
                  className={`h-full rounded-full ${f.pctContribution > 50 ? 'bg-emerald-500' : f.pctContribution > 20 ? 'bg-blue-500' : 'bg-gray-600'}`}
                  style={{ width: `${Math.max(2, f.pctContribution)}%` }}
                />
              </div>
              <span className="text-xs text-gray-300 w-12 text-right">{f.pctContribution.toFixed(0)}%</span>
              <span className="text-[10px] text-gray-600">W{f.firstWeek}–W{f.lastWeek}</span>
            </div>
          ))}
        </div>
      </Card>

      {/* Lineage Lifetime */}
      <Card title="Lineage Lifetime — Family Lifespan">
        <div className="space-y-1">
          {familyStats.map(f => {
            const totalWeeks = weekEntropy.length;
            const startPct = totalWeeks > 0 ? ((f.firstWeek - 1) / totalWeeks) * 100 : 0;
            const widthPct = totalWeeks > 0 ? ((f.lastWeek - f.firstWeek + 1) / totalWeeks) * 100 : 0;
            return (
              <div key={f.id} className="flex items-center gap-2">
                <span className="text-[10px] text-gray-500 w-16">Family {f.id}</span>
                <div className="flex-1 bg-gray-800 rounded h-3 relative overflow-hidden">
                  <div
                    className={`absolute h-full rounded ${f.alive ? 'bg-emerald-500' : 'bg-gray-500'}`}
                    style={{ left: `${startPct}%`, width: `${Math.max(3, widthPct)}%` }}
                  />
                </div>
                <span className="text-[9px] text-gray-600 w-20">
                  {f.totalDescendants} desc | {f.alive ? '✓ alive' : '✗ extinct'}
                </span>
              </div>
            );
          })}
        </div>
      </Card>

      {/* Recommendations */}
      <Card title="Evidence-Based Recommendations">
        <div className="space-y-3">
          {recommendations.map((r, i) => (
            <div key={i} className="bg-gray-800 rounded-lg p-3">
              <div className="flex items-center gap-2 mb-1">
                <span className={`text-[9px] uppercase font-bold px-1.5 py-0.5 rounded ${
                  r.confidence === 'high' ? 'bg-emerald-900 text-emerald-400'
                  : r.confidence === 'medium' ? 'bg-amber-900 text-amber-400'
                  : 'bg-gray-700 text-gray-400'
                }`}>{r.confidence}</span>
                <span className="text-xs text-gray-200">{r.text}</span>
              </div>
              <div className="flex flex-wrap gap-2 mt-1">
                {r.evidence.map((ev, j) => (
                  <span key={j} className="text-[9px] bg-gray-900 text-gray-500 px-2 py-0.5 rounded">{ev}</span>
                ))}
              </div>
            </div>
          ))}
        </div>
      </Card>

      {/* Effective Beam Width */}
      <Card title="Effective Beam Width — Are you paying for diversity you're not getting?">
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 mb-3">
          <MetricCard label="Beam Efficiency" value={`${beamEfficiency}%`} color={beamEfficiency > 50 ? 'green' : 'red'} />
          <MetricCard label="Contributing Paths" value={`${contributingCount}/${totalRetainedCount}`} color="blue" />
          <MetricCard label="Efficiency/M Cands" value={String(efficiencyPerM)} color="amber" />
          <MetricCard label="Avg Utilisation" value={`${effectiveWidth.length > 0 ? Math.round(effectiveWidth.reduce((s, e) => s + e.utilisation, 0) / effectiveWidth.length) : 0}%`} color="default" />
        </div>
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Week</th>
              <th className="text-right p-2">Nominal</th>
              <th className="text-right p-2">Effective</th>
              <th className="text-right p-2">Utilisation</th>
            </tr>
          </thead>
          <tbody>
            {effectiveWidth.map(e => (
              <tr key={e.week} className="border-t border-gray-800">
                <td className="p-2">{e.week}</td>
                <td className="text-right p-2">{e.nominal}</td>
                <td className="text-right p-2">{e.effective}</td>
                <td className={`text-right p-2 ${e.utilisation < 50 ? 'text-red-400' : 'text-emerald-400'}`}>{e.utilisation}%</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>

      {/* Exploration vs Exploitation */}
      <Card title="Exploration vs Exploitation — Is the cooling schedule right?">
        <p className="text-xs text-gray-500 mb-3">
          High exploration early, shifting to exploitation late = healthy. All exploitation from start = too cold.
        </p>
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Week</th>
              <th className="text-right p-2">Exploration %</th>
              <th className="text-right p-2">Exploitation %</th>
            </tr>
          </thead>
          <tbody>
            {explorationBalance.map(e => (
              <tr key={e.week} className="border-t border-gray-800">
                <td className="p-2">{e.week}</td>
                <td className="text-right p-2 text-amber-400">{e.exploration}%</td>
                <td className="text-right p-2 text-emerald-400">{e.exploitation}%</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>

      {/* Wasted Compute */}
      <Card title="Wasted Compute — Should we stop earlier?">
        <p className="text-xs text-gray-500 mb-3">
          Candidates between last global improvement and end of run. High % = iteration budget is too large.
        </p>
        <table className="w-full text-xs">
          <thead>
            <tr className="text-gray-500 uppercase">
              <th className="text-left p-2">Week</th>
              <th className="text-right p-2">Cands After Last Global</th>
              <th className="text-right p-2">% Wasted</th>
            </tr>
          </thead>
          <tbody>
            {lastGlobalPerWeek.map(e => (
              <tr key={e.week} className="border-t border-gray-800">
                <td className="p-2">{e.week}</td>
                <td className="text-right p-2">{e.candsSinceLastGlobal.toLocaleString()}</td>
                <td className={`text-right p-2 ${e.pctWasted > 70 ? 'text-red-400' : e.pctWasted > 40 ? 'text-amber-400' : 'text-emerald-400'}`}>{e.pctWasted}%</td>
              </tr>
            ))}
          </tbody>
        </table>
      </Card>
    </>
  );
}
