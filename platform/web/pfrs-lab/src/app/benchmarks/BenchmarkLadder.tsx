'use client';

import { useMemo } from 'react';
import Card from '@/components/Card';
import { BenchmarkRun } from './page';

// --- Known-optimal / best-known reference values ---
const KNOWN_OPTIMAL: Record<string, { value: number; source: string }> = {
  // CVRP (CVRPLIB best-known solutions)
  'A-n10-k2': { value: 204, source: 'CVRPLIB optimal' },
  'A-n32-k5': { value: 784, source: 'CVRPLIB optimal' },
  'A-n33-k5': { value: 661, source: 'CVRPLIB optimal' },
  'A-n45-k6': { value: 944, source: 'CVRPLIB optimal' },
  'A-n60-k9': { value: 1354, source: 'CVRPLIB optimal' },
  'A-n80-k10': { value: 1763, source: 'CVRPLIB optimal' },
  // NRP (ILP baseline)
  'n012w8': { value: 3020, source: 'ILP (HiGHS, 5hr)' },
  'n005w4': { value: 385, source: 'ILP baseline' },
  // JSS (Taillard/OR-Library optimal solutions)
  'ft06': { value: 55, source: 'Optimal (Fisher & Thompson)' },
  'ft10': { value: 930, source: 'Optimal (Fisher & Thompson)' },
  'la01': { value: 666, source: 'Optimal (Lawrence)' },
  // VRPTW (Solomon best-known solutions — distance only, ignoring vehicle count)
  'C101': { value: 828, source: 'Solomon BKS' },
  'R101': { value: 1645, source: 'Solomon BKS' },
  'RC101': { value: 1696, source: 'Solomon BKS' },
};

interface InstanceRow {
  instance: string;
  problemType: string;
  customers: number;
  sa: number | null;
  lahc: number | null;
  tabu: number | null;
  portfolio: number | null;
  reference: number | null;
  referenceSource: string;
  bestHeuristic: number | null;
  gap: number | null;
  winner: string;
}

const MODES = ['sa', 'lahc', 'tabu', 'portfolio'] as const;

function parseCustomerCount(instance: string): number {
  const match = instance.match(/n(\d+)/i);
  if (match) return parseInt(match[1]);
  return 0;
}

export default function BenchmarkLadder({ runs }: { runs: BenchmarkRun[] }) {
  const rows = useMemo(() => {
    const instanceMap = new Map<string, BenchmarkRun[]>();
    for (const run of runs) {
      const key = `${run.problemType}:${run.instance}`;
      const existing = instanceMap.get(key) || [];
      existing.push(run);
      instanceMap.set(key, existing);
    }

    const result: InstanceRow[] = [];
    for (const [key, instanceRuns] of instanceMap) {
      const [problemType, instance] = key.split(':');

      const bestByMode: Record<string, number> = {};
      for (const run of instanceRuns) {
        const mode = normaliseMode(run.mode);
        if (!bestByMode[mode] || run.penalty < bestByMode[mode]) {
          bestByMode[mode] = run.penalty;
        }
      }

      const sa = bestByMode['sa'] || null;
      const lahc = bestByMode['lahc'] || null;
      const tabu = bestByMode['tabu'] || null;
      const portfolio = bestByMode['portfolio'] || null;

      // Reference: use known-optimal, fall back to ILP run if available.
      const known = KNOWN_OPTIMAL[instance];
      let reference = known?.value || null;
      let referenceSource = known?.source || '';
      if (!reference && bestByMode['ilp']) {
        reference = bestByMode['ilp'];
        referenceSource = 'ILP solve';
      }

      const heuristicValues = [sa, lahc, tabu, portfolio].filter((v): v is number => v !== null);
      const bestHeuristic = heuristicValues.length > 0 ? Math.min(...heuristicValues) : null;
      const gap = (reference && bestHeuristic) ? ((bestHeuristic - reference) / reference * 100) : null;

      let winner = '—';
      if (bestHeuristic !== null) {
        if (portfolio === bestHeuristic) winner = 'Portfolio';
        else if (lahc === bestHeuristic) winner = 'LAHC';
        else if (sa === bestHeuristic) winner = 'SA';
        else if (tabu === bestHeuristic) winner = 'Tabu';
      }

      result.push({
        instance, problemType,
        customers: parseCustomerCount(instance),
        sa, lahc, tabu, portfolio,
        reference, referenceSource,
        bestHeuristic, gap, winner,
      });
    }

    result.sort((a, b) => {
      if (a.problemType !== b.problemType) return a.problemType.localeCompare(b.problemType);
      return a.customers - b.customers;
    });

    return result;
  }, [runs]);

  // Leaderboard: wins per algorithm.
  const leaderboard = useMemo(() => {
    const wins: Record<string, number> = { SA: 0, LAHC: 0, Tabu: 0, Portfolio: 0 };
    for (const row of rows) {
      if (row.winner !== '—') {
        wins[row.winner] = (wins[row.winner] || 0) + 1;
      }
    }
    return Object.entries(wins)
      .map(([algo, count]) => ({ algo, count }))
      .sort((a, b) => b.count - a.count);
  }, [rows]);

  // Average gap.
  const avgGap = useMemo(() => {
    const gaps = rows.filter(r => r.gap !== null).map(r => r.gap!);
    if (gaps.length === 0) return null;
    return gaps.reduce((s, g) => s + g, 0) / gaps.length;
  }, [rows]);

  const nrpRows = rows.filter(r => r.problemType === 'nrp');
  const cvrpRows = rows.filter(r => r.problemType === 'cvrp');
  const jssRows = rows.filter(r => r.problemType === 'jss');
  const vrptwRows = rows.filter(r => r.problemType === 'vrptw');

  return (
    <div className="space-y-6">
      {/* Overall Leaderboard */}
      <Card title="Algorithm Leaderboard">
        <p className="text-xs text-gray-500 mb-4">
          Which algorithm wins (lowest objective) across all instances and problem domains.
        </p>
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-3">
          {leaderboard.map(({ algo, count }, i) => (
            <div key={algo} className="bg-gray-800 rounded p-3 text-center">
              <div className="text-[10px] text-gray-500 uppercase">{i === 0 ? '🏆 ' : ''}{algo}</div>
              <div className={`text-2xl font-bold ${i === 0 ? 'text-emerald-400' : 'text-gray-300'}`}>
                {count}
              </div>
              <div className="text-[9px] text-gray-600">{count === 1 ? 'win' : 'wins'}</div>
            </div>
          ))}
        </div>
        {avgGap !== null && (
          <div className="mt-4 bg-gray-800 rounded p-3">
            <div className="flex justify-between items-center">
              <span className="text-xs text-gray-500">Average gap to reference (best-known/optimal)</span>
              <span className={`text-lg font-bold ${avgGap < 15 ? 'text-emerald-400' : avgGap < 30 ? 'text-amber-400' : 'text-red-400'}`}>
                +{avgGap.toFixed(1)}%
              </span>
            </div>
          </div>
        )}
      </Card>

      {/* CVRP Ladder */}
      {cvrpRows.length > 0 && (
        <Card title="Vehicle Routing (CVRPLIB)">
          <p className="text-xs text-gray-500 mb-3">Objective: minimise total travel distance. Reference: published CVRPLIB optimal solutions.</p>
          <LadderTable rows={cvrpRows} />
        </Card>
      )}

      {/* NRP Ladder */}
      {nrpRows.length > 0 && (
        <Card title="Nurse Rostering (INRC-II)">
          <p className="text-xs text-gray-500 mb-3">Objective: minimise soft constraint penalty. Reference: ILP solve (HiGHS, time-limited).</p>
          <LadderTable rows={nrpRows} />
        </Card>
      )}

      {/* JSS Ladder */}
      {jssRows.length > 0 && (
        <Card title="Job Shop Scheduling (Taillard)">
          <p className="text-xs text-gray-500 mb-3">Objective: minimise makespan (total completion time). Reference: published optimal solutions.</p>
          <LadderTable rows={jssRows} />
        </Card>
      )}

      {/* VRPTW Ladder */}
      {vrptwRows.length > 0 && (
        <Card title="Vehicle Routing with Time Windows (Solomon)">
          <p className="text-xs text-gray-500 mb-3">Objective: minimise total travel distance with time window constraints. Reference: Solomon best-known solutions.</p>
          <LadderTable rows={vrptwRows} />
        </Card>
      )}

      {/* Summary stats */}
      <Card title="Platform Summary">
        <div className="grid grid-cols-2 sm:grid-cols-5 gap-3 text-xs">
          <Stat label="Problem Domains" value={new Set(rows.map(r => r.problemType)).size} colour="blue" />
          <Stat label="Instances" value={rows.length} colour="blue" />
          <Stat label="Total Runs" value={runs.length} colour="emerald" />
          <Stat label="With Reference" value={rows.filter(r => r.reference !== null).length} colour="amber" />
          <Stat label="At Optimum" value={rows.filter(r => r.gap !== null && r.gap === 0).length} colour="emerald" />
        </div>
      </Card>
    </div>
  );
}

function LadderTable({ rows }: { rows: InstanceRow[] }) {
  return (
    <div className="overflow-x-auto">
      <table className="w-full text-xs">
        <thead>
          <tr className="text-gray-500 uppercase">
            <th className="text-left p-2">Instance</th>
            <th className="text-right p-2">SA</th>
            <th className="text-right p-2">LAHC</th>
            <th className="text-right p-2">Tabu</th>
            <th className="text-right p-2">Portfolio</th>
            <th className="text-right p-2 text-blue-400">Reference</th>
            <th className="text-right p-2">Gap%</th>
            <th className="text-center p-2">Winner</th>
          </tr>
        </thead>
        <tbody>
          {rows.map(row => {
            const best = row.bestHeuristic;
            return (
              <tr key={row.instance} className="border-t border-gray-800 hover:bg-gray-800/50">
                <td className="p-2">
                  <span className="font-mono text-gray-300">{row.instance}</span>
                  <span className="text-[9px] text-gray-600 ml-2">{row.customers > 0 ? `${row.customers} nodes` : ''}</span>
                </td>
                <CellValue value={row.sa} best={best} reference={row.reference} />
                <CellValue value={row.lahc} best={best} reference={row.reference} />
                <CellValue value={row.tabu} best={best} reference={row.reference} />
                <CellValue value={row.portfolio} best={best} reference={row.reference} />
                <td className="text-right p-2">
                  {row.reference ? (
                    <div>
                      <span className="text-blue-400 font-semibold">{row.reference.toLocaleString()}</span>
                      <div className="text-[8px] text-gray-600">{row.referenceSource}</div>
                    </div>
                  ) : <span className="text-gray-700">—</span>}
                </td>
                <td className="text-right p-2">
                  {row.gap !== null ? (
                    <span className={`font-semibold ${row.gap === 0 ? 'text-emerald-400' : row.gap < 15 ? 'text-emerald-400' : row.gap < 30 ? 'text-amber-400' : 'text-red-400'}`}>
                      {row.gap === 0 ? '✓ optimal' : `+${row.gap.toFixed(1)}%`}
                    </span>
                  ) : <span className="text-gray-700">—</span>}
                </td>
                <td className="text-center p-2">
                  <span className="bg-emerald-900/50 text-emerald-400 px-2 py-0.5 rounded text-[10px] font-semibold">
                    {row.winner}
                  </span>
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

function CellValue({ value, best, reference }: { value: number | null; best: number | null; reference: number | null }) {
  if (value === null) {
    return <td className="text-right p-2 text-gray-700">—</td>;
  }
  const isBest = best !== null && value === best;
  const isOptimal = reference !== null && value === reference;
  const gapToRef = reference ? ((value - reference) / reference * 100) : null;

  // Sparkline: simple bar showing quality relative to range (reference → 2×reference).
  const sparkWidth = reference && reference > 0
    ? Math.max(5, Math.min(100, 100 - ((value - reference) / reference * 100)))
    : 50;

  return (
    <td className={`text-right p-2 ${isOptimal ? 'text-emerald-400 font-bold' : isBest ? 'text-emerald-400 font-semibold' : 'text-gray-300'}`}>
      <div>{value.toLocaleString()}</div>
      {reference && reference > 0 && (
        <div className="mt-0.5 h-1 w-full bg-gray-800 rounded overflow-hidden">
          <div
            className={`h-full rounded ${isOptimal ? 'bg-emerald-400' : isBest ? 'bg-emerald-500' : gapToRef && gapToRef < 15 ? 'bg-blue-500' : 'bg-amber-500'}`}
            style={{ width: `${sparkWidth}%` }}
          />
        </div>
      )}
      {gapToRef !== null && gapToRef > 0 && (
        <div className={`text-[9px] ${gapToRef < 10 ? 'text-gray-500' : gapToRef < 25 ? 'text-amber-600' : 'text-red-600'}`}>
          +{gapToRef.toFixed(1)}%
        </div>
      )}
      {isOptimal && <div className="text-[9px] text-emerald-600">★ optimal</div>}
    </td>
  );
}

function Stat({ label, value, colour }: { label: string; value: number; colour: string }) {
  const colorClass = colour === 'blue' ? 'text-blue-400' : colour === 'emerald' ? 'text-emerald-400' : 'text-amber-400';
  return (
    <div className="bg-gray-800 rounded p-3">
      <div className="text-[9px] text-gray-500 uppercase">{label}</div>
      <div className={`text-lg font-bold ${colorClass}`}>{value}</div>
    </div>
  );
}

function normaliseMode(mode: string): string {
  const lower = mode.toLowerCase();
  if (lower.includes('portfolio') || lower.includes('sa,lahc') || lower.includes('sa,lahc,tabu')) return 'portfolio';
  if (lower === 'sa' || lower === 'simulated-annealing') return 'sa';
  if (lower === 'lahc' || lower === 'late-acceptance') return 'lahc';
  if (lower === 'tabu' || lower === 'tabu-search') return 'tabu';
  if (lower === 'ilp') return 'ilp';
  if (lower.includes('+')) return 'portfolio';
  return lower;
}
