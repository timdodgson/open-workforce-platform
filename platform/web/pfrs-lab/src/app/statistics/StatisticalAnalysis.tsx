'use client';

import { useState, useMemo } from 'react';
import Card from '@/components/Card';
import { RunEntry } from './page';

type GroupBy = 'config' | 'mode' | 'beamWidth' | 'instance' | 'coolingMode' | 'iterations';

interface GroupStats {
  key: string;
  runs: RunEntry[];
  penalties: number[];
  mean: number;
  median: number;
  best: number;
  worst: number;
  stdDev: number;
  variance: number;
  ci95Lower: number;
  ci95Upper: number;
  n: number;
}

interface TestResult {
  groupA: string;
  groupB: string;
  tStat: number;
  pValue: number;
  significant: boolean;
  effectSize: number; // Cohen's d
  meanDiff: number;
  meanDiffPct: number;
}

function computeStats(penalties: number[]): Omit<GroupStats, 'key' | 'runs' | 'penalties'> {
  const n = penalties.length;
  if (n === 0) return { mean: 0, median: 0, best: 0, worst: 0, stdDev: 0, variance: 0, ci95Lower: 0, ci95Upper: 0, n: 0 };

  const sorted = [...penalties].sort((a, b) => a - b);
  const mean = penalties.reduce((s, p) => s + p, 0) / n;
  const median = n % 2 === 0 ? (sorted[n/2 - 1] + sorted[n/2]) / 2 : sorted[Math.floor(n/2)];
  const best = sorted[0];
  const worst = sorted[n - 1];
  const variance = n > 1 ? penalties.reduce((s, p) => s + (p - mean) ** 2, 0) / (n - 1) : 0;
  const stdDev = Math.sqrt(variance);
  const se = stdDev / Math.sqrt(n);
  const t975 = n > 1 ? 1.96 : 0; // approximate z for 95% CI
  const ci95Lower = mean - t975 * se;
  const ci95Upper = mean + t975 * se;

  return { mean, median, best, worst, stdDev, variance, ci95Lower, ci95Upper, n };
}

function welchTTest(a: number[], b: number[]): { tStat: number; pValue: number } {
  const nA = a.length, nB = b.length;
  if (nA < 2 || nB < 2) return { tStat: 0, pValue: 1 };

  const meanA = a.reduce((s, v) => s + v, 0) / nA;
  const meanB = b.reduce((s, v) => s + v, 0) / nB;
  const varA = a.reduce((s, v) => s + (v - meanA) ** 2, 0) / (nA - 1);
  const varB = b.reduce((s, v) => s + (v - meanB) ** 2, 0) / (nB - 1);

  const se = Math.sqrt(varA / nA + varB / nB);
  if (se === 0) return { tStat: 0, pValue: 1 };

  const tStat = (meanA - meanB) / se;
  // Approximate p-value using normal distribution for large samples.
  const absT = Math.abs(tStat);
  const pValue = 2 * (1 - normalCDF(absT));

  return { tStat, pValue };
}

function normalCDF(x: number): number {
  // Approximation of the standard normal CDF.
  const a1 = 0.254829592, a2 = -0.284496736, a3 = 1.421413741;
  const a4 = -1.453152027, a5 = 1.061405429, p = 0.3275911;
  const sign = x < 0 ? -1 : 1;
  x = Math.abs(x) / Math.sqrt(2);
  const t = 1.0 / (1.0 + p * x);
  const y = 1.0 - (((((a5 * t + a4) * t) + a3) * t + a2) * t + a1) * t * Math.exp(-x * x);
  return 0.5 * (1.0 + sign * y);
}

function cohensD(a: number[], b: number[]): number {
  const meanA = a.reduce((s, v) => s + v, 0) / a.length;
  const meanB = b.reduce((s, v) => s + v, 0) / b.length;
  const varA = a.reduce((s, v) => s + (v - meanA) ** 2, 0) / (a.length - 1);
  const varB = b.reduce((s, v) => s + (v - meanB) ** 2, 0) / (b.length - 1);
  const pooledSD = Math.sqrt(((a.length - 1) * varA + (b.length - 1) * varB) / (a.length + b.length - 2));
  return pooledSD > 0 ? (meanA - meanB) / pooledSD : 0;
}

function getGroupKey(run: RunEntry, groupBy: GroupBy): string {
  const m = run.metadata;
  if (!m) return 'unknown';
  const meta = m as unknown as Record<string, unknown>;
  const domain = getRunDomain(run);
  const isNRP = domain === 'nrp';

  switch (groupBy) {
    case 'config': {
      if (!isNRP) {
        // For CVRP/JSS: mode is the primary differentiator.
        const mode = String(meta.mode || 'sa');
        return mode;
      }
      // NRP: composite key: mode + portfolio + strategy + final window
      const parts = [m.mode || 'sa'];
      if (m.portfolio) parts[0] = m.portfolio;
      if (m.beamStrategy && m.beamStrategy !== 'none') parts.push(m.beamStrategy);
      if (m.finalWindowWeeks && m.finalWindowWeeks > 1) parts.push(`fw${m.finalWindowWeeks}`);
      return parts.join('+');
    }
    case 'mode': return String(meta.mode || m.mode || 'unknown');
    case 'beamWidth': return isNRP ? `beam=${m.beamWidth || 1}` : 'n/a';
    case 'instance': return String(meta.instance || m.instance || 'unknown');
    case 'coolingMode': return !isNRP ? String(meta.mode || 'sa') : (m.coolingMode || 'unknown');
    case 'iterations': {
      const iters = Number(meta.iterations || m.iterationsPerWorker || 0);
      return `${(iters / 1000).toFixed(0)}K`;
    }
    default: return 'unknown';
  }
}

type DomainFilter = 'all' | 'nrp' | 'cvrp' | 'jss';

function getRunDomain(run: RunEntry): string {
  const meta = run.metadata as unknown as Record<string, unknown>;
  if (!meta) return 'nrp';
  const pt = String(meta.problemType || meta.mode || '').toLowerCase();
  if (pt === 'cvrp') return 'cvrp';
  if (pt === 'jss' || pt === 'jobshop') return 'jss';
  if (pt === 'vrptw') return 'vrptw';
  if (pt === 'ilp') return 'nrp'; // ILP runs are NRP-domain benchmarks
  return 'nrp';
}

export default function StatisticalAnalysis({ runs }: { runs: RunEntry[] }) {
  const [groupBy, setGroupBy] = useState<GroupBy>('config');
  const [domainFilter, setDomainFilter] = useState<DomainFilter>('all');

  // Detect available domains.
  const availableDomains = useMemo(() => {
    const domains = new Set(runs.map(r => getRunDomain(r)));
    return Array.from(domains).sort();
  }, [runs]);

  // Filter runs by domain. When 'all' is selected but multiple domains exist,
  // auto-select the domain with the most runs to avoid cross-domain comparison.
  const filteredRuns = useMemo(() => {
    if (domainFilter !== 'all') {
      return runs.filter(r => getRunDomain(r) === domainFilter);
    }
    // If only one domain exists, show all.
    if (availableDomains.length === 1) return runs;
    // Multiple domains: auto-select the one with most runs.
    const counts: Record<string, number> = {};
    for (const r of runs) {
      const d = getRunDomain(r);
      counts[d] = (counts[d] || 0) + 1;
    }
    const dominant = Object.entries(counts).sort((a, b) => b[1] - a[1])[0][0];
    return runs.filter(r => getRunDomain(r) === dominant);
  }, [runs, domainFilter, availableDomains]);

  // Determine active problem type from filtered runs.
  const problemType = useMemo((): string => {
    if (filteredRuns.length === 0) return 'nrp';
    const domains = filteredRuns.map(r => getRunDomain(r));
    const cvrp = domains.filter(d => d === 'cvrp').length;
    const jss = domains.filter(d => d === 'jss').length;
    const vrptw = domains.filter(d => d === 'vrptw').length;
    if (cvrp > domains.length / 2) return 'cvrp';
    if (jss > domains.length / 2) return 'jss';
    if (vrptw > domains.length / 2) return 'vrptw';
    return 'nrp';
  }, [filteredRuns]);

  // Objective label depends on problem type.
  const objectiveLabel = problemType === 'cvrp' ? 'Distance' : problemType === 'vrptw' ? 'Distance' : problemType === 'jss' ? 'Makespan' : 'Penalty';

  // Group runs.
  const groups = useMemo(() => {
    const map = new Map<string, RunEntry[]>();
    for (const run of filteredRuns) {
      const key = getGroupKey(run, groupBy);
      const existing = map.get(key) || [];
      existing.push(run);
      map.set(key, existing);
    }

    const stats: GroupStats[] = [];
    for (const [key, groupRuns] of map) {
      const penalties = groupRuns.map(r => r.summary.totalPenalty).filter(p => p > 0);
      const s = computeStats(penalties);
      stats.push({ key, runs: groupRuns, penalties, ...s });
    }
    return stats.sort((a, b) => a.mean - b.mean);
  }, [filteredRuns, groupBy]);

  // Significance tests (pairwise between groups with n >= 2).
  const tests = useMemo(() => {
    const results: TestResult[] = [];
    const testable = groups.filter(g => g.n >= 2);
    for (let i = 0; i < testable.length; i++) {
      for (let j = i + 1; j < testable.length; j++) {
        const a = testable[i], b = testable[j];
        const { tStat, pValue } = welchTTest(a.penalties, b.penalties);
        const d = cohensD(a.penalties, b.penalties);
        const meanDiff = a.mean - b.mean;
        const meanDiffPct = b.mean > 0 ? (meanDiff / b.mean) * 100 : 0;
        results.push({
          groupA: a.key, groupB: b.key,
          tStat, pValue, significant: pValue < 0.05,
          effectSize: d, meanDiff, meanDiffPct,
        });
      }
    }
    return results;
  }, [groups]);

  // Observations.
  const observations = useMemo(() => {
    const obs: string[] = [];
    const objName = objectiveLabel.toLowerCase();
    const best = groups[0];
    if (best && groups.length > 1) {
      obs.push(`Best average ${objName}: ${best.key} (${best.mean.toFixed(0)}, n=${best.n}).`);
    }
    for (const t of tests) {
      if (t.significant && Math.abs(t.meanDiffPct) > 0.5) {
        const better = t.meanDiff < 0 ? t.groupA : t.groupB;
        const worse = t.meanDiff < 0 ? t.groupB : t.groupA;
        const pct = Math.abs(t.meanDiffPct).toFixed(1);
        const conf = t.pValue < 0.01 ? 'high' : 'moderate';
        obs.push(`${better} improves average ${objName} by ${pct}% over ${worse} with ${conf} statistical confidence (p=${t.pValue.toFixed(4)}).`);
      }
    }
    if (tests.length > 0 && !tests.some(t => t.significant)) {
      obs.push('No statistically significant differences detected between groups (p > 0.05). More runs may be needed.');
    }
    return obs;
  }, [groups, tests, objectiveLabel]);

  // Chart helpers.
  const globalMax = Math.max(...groups.flatMap(g => g.penalties), 1);
  const globalMin = Math.min(...groups.flatMap(g => g.penalties), 0);

  return (
    <div className="space-y-4">
      {/* Conclusions first */}
      {observations.length > 0 && (
        <Card title="Conclusions">
          <div className="space-y-2">
            {observations.map((obs, i) => (
              <p key={i} className="text-sm text-gray-300 border-l-2 border-blue-600 pl-3">{obs}</p>
            ))}
          </div>
        </Card>
      )}

      {/* Group selector */}
      <Card title="Configuration">
        {/* Domain filter */}
        {availableDomains.length > 1 && (
          <div className="flex items-center gap-2 mb-3">
            <span className="text-[10px] text-gray-500">Domain:</span>
            {(['all', ...availableDomains] as DomainFilter[]).map(d => (
              <button key={d} onClick={() => setDomainFilter(d)}
                className={`px-3 py-1 rounded text-xs ${domainFilter === d ? 'bg-emerald-600 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700'}`}>
                {d === 'all' ? 'Auto' : d.toUpperCase()}
              </button>
            ))}
          </div>
        )}
        <div className="flex items-center gap-2 mb-3">
          <span className="text-[10px] text-gray-500">Group by:</span>
          {(['config', 'mode', 'beamWidth', 'instance', 'coolingMode', 'iterations'] as GroupBy[]).map(g => (
            <button key={g} onClick={() => setGroupBy(g)}
              className={`px-3 py-1 rounded text-xs ${groupBy === g ? 'bg-blue-600 text-white' : 'bg-gray-800 text-gray-400 hover:bg-gray-700'}`}>
              {g}
            </button>
          ))}
        </div>
        <p className="text-[9px] text-gray-600">
          {filteredRuns.length} runs ({problemType.toUpperCase()}), {groups.length} groups
          {availableDomains.length > 1 && domainFilter === 'all' && (
            <span className="text-amber-500 ml-2">Auto-filtered to dominant domain to avoid cross-domain comparison.</span>
          )}
        </p>
      </Card>

      {/* Box plots — only show groups with data */}
      {groups.filter(g => g.n > 0).length > 0 && (
      <Card title="Distribution (Box Plots)">
        <p className="text-xs text-gray-500 mb-3">
          {objectiveLabel} spread per configuration. Blue box = 95% CI. White line = median. Green dot = mean. Left is better.
        </p>
        <div className="space-y-3">
          {groups.filter(g => g.n > 0).map(g => {
            const range = globalMax - globalMin || 1;
            const boxLeft = ((g.ci95Lower - globalMin) / range) * 100;
            const boxRight = ((g.ci95Upper - globalMin) / range) * 100;
            const medianPos = ((g.median - globalMin) / range) * 100;
            const bestPos = ((g.best - globalMin) / range) * 100;
            const worstPos = ((g.worst - globalMin) / range) * 100;
            const meanPos = ((g.mean - globalMin) / range) * 100;

            return (
              <div key={g.key} className="flex items-center gap-2">
                <span className="w-24 text-xs text-gray-300 truncate" title={g.key}>{g.key}</span>
                <div className="flex-1 h-8 bg-gray-800 rounded relative">
                  {/* Whiskers */}
                  <div className="absolute top-1/2 h-px bg-gray-600" style={{ left: `${bestPos}%`, width: `${worstPos - bestPos}%` }} />
                  {/* Box (CI95) */}
                  <div className="absolute top-1 bottom-1 bg-blue-700 rounded opacity-60" style={{ left: `${boxLeft}%`, width: `${boxRight - boxLeft}%` }} />
                  {/* Median line */}
                  <div className="absolute top-0 bottom-0 w-0.5 bg-white" style={{ left: `${medianPos}%` }} />
                  {/* Mean dot */}
                  <div className="absolute top-1/2 -translate-y-1/2 w-2 h-2 bg-emerald-400 rounded-full" style={{ left: `${meanPos}%` }} />
                  {/* Data points */}
                  {g.penalties.map((p, i) => (
                    <div key={i} className="absolute top-1/2 -translate-y-1/2 w-1.5 h-1.5 bg-gray-400 rounded-full opacity-50"
                      style={{ left: `${((p - globalMin) / range) * 100}%` }} />
                  ))}
                </div>
                <span className="w-16 text-right text-[10px] text-gray-500">n={g.n}</span>
              </div>
            );
          })}
        </div>
        <div className="flex gap-4 mt-2 text-[9px] text-gray-500">
          <span className="flex items-center gap-1"><span className="w-2 h-2 bg-emerald-400 rounded-full" />Mean</span>
          <span className="flex items-center gap-1"><span className="w-2 h-0.5 bg-white" />Median</span>
          <span className="flex items-center gap-1"><span className="w-4 h-2 bg-blue-700 rounded opacity-60" />95% CI</span>
        </div>
      </Card>
      )}

      {/* Summary table */}
      <Card title="Group Statistics">
        <div className="overflow-x-auto">
          <table className="w-full text-[10px]">
            <thead>
              <tr className="text-gray-500 uppercase">
                <th className="text-left p-1.5">Group</th>
                <th className="text-right p-1.5">N</th>
                <th className="text-right p-1.5">Mean</th>
                <th className="text-right p-1.5">Median</th>
                <th className="text-right p-1.5">Best</th>
                <th className="text-right p-1.5">Worst</th>
                <th className="text-right p-1.5">Std Dev</th>
                <th className="text-right p-1.5">95% CI</th>
              </tr>
            </thead>
            <tbody>
              {groups.filter(g => g.n > 0).map((g, i) => (
                <tr key={g.key} className={`border-t border-gray-800 ${i === 0 ? 'bg-emerald-900/10' : ''}`}>
                  <td className="p-1.5 font-medium text-blue-400">{g.key} {i === 0 && '🥇'}</td>
                  <td className="text-right p-1.5">{g.n}</td>
                  <td className="text-right p-1.5 font-mono">{g.mean.toFixed(0)}</td>
                  <td className="text-right p-1.5 font-mono">{g.median.toFixed(0)}</td>
                  <td className="text-right p-1.5 text-emerald-400">{g.best.toLocaleString()}</td>
                  <td className="text-right p-1.5 text-red-400">{g.worst.toLocaleString()}</td>
                  <td className="text-right p-1.5">{g.stdDev.toFixed(1)}</td>
                  <td className="text-right p-1.5 text-gray-400">[{g.ci95Lower.toFixed(0)}, {g.ci95Upper.toFixed(0)}]</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </Card>

      {/* Histogram overlay — only show when there are multiple groups with data */}
      {groups.filter(g => g.n > 0).length >= 2 && (
      <Card title={`${objectiveLabel} Distribution`}>
        <p className="text-xs text-gray-500 mb-3">
          Histogram of all run {objectiveLabel.toLowerCase()} values grouped by configuration. Clusters further left are better. Overlapping clusters suggest the configurations produce similar results — check the significance table below for statistical confirmation.
        </p>
        <svg viewBox="0 0 700 180" className="w-full h-44 bg-gray-900 rounded border border-gray-800">
          {groups.map((g, gi) => {
            if (g.penalties.length === 0) return null;
            const numBins = 15;
            const binSize = (globalMax - globalMin) / numBins || 1;
            const bins = Array(numBins).fill(0);
            for (const p of g.penalties) {
              const bin = Math.min(Math.floor((p - globalMin) / binSize), numBins - 1);
              bins[bin]++;
            }
            const maxBin = Math.max(...bins, 1);
            const colors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4'];
            const color = colors[gi % colors.length];

            return bins.map((count, bi) => {
              if (count === 0) return null;
              const x = 40 + (bi / numBins) * 620;
              const w = 620 / numBins - 2;
              const h = (count / maxBin) * 140;
              return (
                <rect key={`${gi}-${bi}`} x={x + gi * 3} y={160 - h} width={Math.max(w / groups.length, 2)} height={h}
                  fill={color} opacity={0.6} rx={1}
                >
                  <title>{`${g.key}: ${count} runs in bin`}</title>
                </rect>
              );
            });
          })}
          <text x="350" y="178" textAnchor="middle" className="fill-gray-600 text-[8px]">{objectiveLabel}</text>
        </svg>
        <div className="flex gap-3 mt-1">
          {groups.map((g, i) => {
            const colors = ['#3b82f6', '#10b981', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4'];
            return (
              <span key={g.key} className="flex items-center gap-1 text-[9px] text-gray-400">
                <span className="w-2 h-2 rounded-sm" style={{ background: colors[i % colors.length] }} />{g.key}
              </span>
            );
          })}
        </div>
      </Card>
      )}

      {/* Significance tests */}
      {tests.length > 0 && tests.some(t => t.significant) && (
        <Card title="Statistical Significance (Welch's t-test)">
          <p className="text-xs text-gray-500 mb-3">
            Pairwise comparison between configurations. A ✓ in the Sig? column means the difference is statistically reliable (p &lt; 0.05). Cohen&apos;s d measures effect size: |d| &gt; 0.8 is a large practical difference. Negative mean diff means Group A is better.
          </p>
          <table className="w-full text-[10px]">
            <thead>
              <tr className="text-gray-500 uppercase">
                <th className="text-left p-1.5">Group A</th>
                <th className="text-left p-1.5">Group B</th>
                <th className="text-right p-1.5">Mean Diff</th>
                <th className="text-right p-1.5">Diff %</th>
                <th className="text-right p-1.5">t-stat</th>
                <th className="text-right p-1.5">p-value</th>
                <th className="text-right p-1.5">Cohen's d</th>
                <th className="text-center p-1.5">Sig?</th>
              </tr>
            </thead>
            <tbody>
              {tests.map((t, i) => (
                <tr key={i} className={`border-t border-gray-800 ${t.significant ? 'bg-emerald-900/10' : ''}`}>
                  <td className="p-1.5 text-blue-400">{t.groupA}</td>
                  <td className="p-1.5 text-rose-400">{t.groupB}</td>
                  <td className="text-right p-1.5 font-mono">{t.meanDiff.toFixed(0)}</td>
                  <td className="text-right p-1.5">{t.meanDiffPct.toFixed(1)}%</td>
                  <td className="text-right p-1.5 font-mono">{t.tStat.toFixed(3)}</td>
                  <td className={`text-right p-1.5 font-mono ${t.pValue < 0.05 ? 'text-emerald-400' : 'text-gray-500'}`}>
                    {t.pValue < 0.001 ? '<0.001' : t.pValue.toFixed(4)}
                  </td>
                  <td className="text-right p-1.5">{t.effectSize.toFixed(2)}</td>
                  <td className="text-center p-1.5">
                    {t.significant ? <span className="text-emerald-400">✓</span> : <span className="text-gray-600">✗</span>}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          <p className="text-[9px] text-gray-600 mt-2">
            Significance: p &lt; 0.05. Effect size: |d| &lt; 0.2 negligible, 0.2-0.5 small, 0.5-0.8 medium, &gt; 0.8 large.
          </p>
        </Card>
      )}

    </div>
  );
}
