'use client';

import { useMemo } from 'react';
import Card from '@/components/Card';
import { BenchmarkRun } from './page';

interface InstanceRow {
  instance: string;
  problemType: string;
  customers: number; // for sorting (parsed from instance name)
  constructive: number | null;
  sa: number | null;
  lahc: number | null;
  tabu: number | null;
  portfolio: number | null;
  ilp: number | null;
  bestHeuristic: number | null;
  gap: number | null; // % gap from best heuristic to ILP
  winner: string;
}

const MODES = ['sa', 'lahc', 'tabu', 'portfolio', 'ilp'] as const;

function parseCustomerCount(instance: string): number {
  // Try to extract number from instance name patterns like "A-n32-k5", "n012w8"
  const cvrpMatch = instance.match(/n(\d+)/i);
  if (cvrpMatch) return parseInt(cvrpMatch[1]);
  return 0;
}

export default function BenchmarkLadder({ runs }: { runs: BenchmarkRun[] }) {
  const rows = useMemo(() => {
    // Group by instance.
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

      // Find best per mode (lowest objective).
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
      const ilp = bestByMode['ilp'] || null;

      // Best heuristic (non-ILP).
      const heuristicValues = [sa, lahc, tabu, portfolio].filter((v): v is number => v !== null);
      const bestHeuristic = heuristicValues.length > 0 ? Math.min(...heuristicValues) : null;

      // Gap to ILP.
      const gap = (ilp && bestHeuristic) ? ((bestHeuristic - ilp) / ilp * 100) : null;

      // Winner.
      let winner = '—';
      if (bestHeuristic !== null) {
        if (sa === bestHeuristic) winner = 'SA';
        else if (lahc === bestHeuristic) winner = 'LAHC';
        else if (portfolio === bestHeuristic) winner = 'Portfolio';
        else if (tabu === bestHeuristic) winner = 'Tabu';
      }

      result.push({
        instance,
        problemType,
        customers: parseCustomerCount(instance),
        constructive: null, // TODO: store constructive baseline in run metadata
        sa, lahc, tabu, portfolio, ilp,
        bestHeuristic, gap, winner,
      });
    }

    // Sort by problem type, then by customer count.
    result.sort((a, b) => {
      if (a.problemType !== b.problemType) return a.problemType.localeCompare(b.problemType);
      return a.customers - b.customers;
    });

    return result;
  }, [runs]);

  // Separate by problem type.
  const nrpRows = rows.filter(r => r.problemType === 'nrp');
  const cvrpRows = rows.filter(r => r.problemType === 'cvrp');

  return (
    <div className="space-y-6">
      <Card title="Benchmark Ladder">
        <p className="text-xs text-gray-500 mb-4">
          Best objective value per algorithm per instance. Lower is better.
          Gap% shows how close the best heuristic gets to the ILP optimal/bounded solution.
          The winner column shows which heuristic achieved the lowest objective.
        </p>
      </Card>

      {nrpRows.length > 0 && (
        <Card title="Nurse Rostering (INRC-II)">
          <p className="text-xs text-gray-500 mb-3">Objective: minimise soft constraint penalty. Lower is better.</p>
          <LadderTable rows={nrpRows} objectiveLabel="Penalty" />
        </Card>
      )}

      {cvrpRows.length > 0 && (
        <Card title="Vehicle Routing (CVRPLIB)">
          <p className="text-xs text-gray-500 mb-3">Objective: minimise total travel distance. Lower is better.</p>
          <LadderTable rows={cvrpRows} objectiveLabel="Distance" />
        </Card>
      )}

      {rows.length > 0 && (
        <Card title="Summary">
          <div className="grid grid-cols-2 sm:grid-cols-4 gap-3 text-xs">
            <div className="bg-gray-800 rounded p-3">
              <div className="text-[9px] text-gray-500 uppercase">Instances</div>
              <div className="text-blue-400 font-bold text-lg">{rows.length}</div>
            </div>
            <div className="bg-gray-800 rounded p-3">
              <div className="text-[9px] text-gray-500 uppercase">Total Runs</div>
              <div className="text-emerald-400 font-bold text-lg">{runs.length}</div>
            </div>
            <div className="bg-gray-800 rounded p-3">
              <div className="text-[9px] text-gray-500 uppercase">ILP Benchmarks</div>
              <div className="text-amber-400 font-bold text-lg">{rows.filter(r => r.ilp !== null).length}</div>
            </div>
            <div className="bg-gray-800 rounded p-3">
              <div className="text-[9px] text-gray-500 uppercase">Avg Gap to ILP</div>
              <div className="text-rose-400 font-bold text-lg">
                {(() => {
                  const gaps = rows.filter(r => r.gap !== null).map(r => r.gap!);
                  if (gaps.length === 0) return '—';
                  return `${(gaps.reduce((s, g) => s + g, 0) / gaps.length).toFixed(1)}%`;
                })()}
              </div>
            </div>
          </div>
        </Card>
      )}
    </div>
  );
}

function LadderTable({ rows, objectiveLabel }: { rows: InstanceRow[]; objectiveLabel: string }) {
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
            <th className="text-right p-2 text-blue-400">ILP</th>
            <th className="text-right p-2">Gap%</th>
            <th className="text-center p-2">Winner</th>
          </tr>
        </thead>
        <tbody>
          {rows.map(row => {
            const best = row.bestHeuristic;
            return (
              <tr key={row.instance} className="border-t border-gray-800 hover:bg-gray-800/50">
                <td className="p-2 font-mono text-gray-300">{row.instance}</td>
                <CellValue value={row.sa} best={best} />
                <CellValue value={row.lahc} best={best} />
                <CellValue value={row.tabu} best={best} />
                <CellValue value={row.portfolio} best={best} />
                <td className="text-right p-2 text-blue-400 font-semibold">
                  {row.ilp?.toLocaleString() || '—'}
                </td>
                <td className="text-right p-2">
                  {row.gap !== null ? (
                    <span className={row.gap < 15 ? 'text-emerald-400' : row.gap < 30 ? 'text-amber-400' : 'text-red-400'}>
                      +{row.gap.toFixed(1)}%
                    </span>
                  ) : '—'}
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

function CellValue({ value, best }: { value: number | null; best: number | null }) {
  if (value === null) {
    return <td className="text-right p-2 text-gray-700">—</td>;
  }
  const isBest = best !== null && value === best;
  return (
    <td className={`text-right p-2 ${isBest ? 'text-emerald-400 font-semibold' : 'text-gray-300'}`}>
      {value.toLocaleString()}
    </td>
  );
}

function normaliseMode(mode: string): string {
  const lower = mode.toLowerCase();
  if (lower.includes('portfolio') || lower.includes('sa,lahc') || lower.includes('sa,lahc,tabu')) return 'portfolio';
  if (lower === 'sa' || lower === 'simulated-annealing') return 'sa';
  if (lower === 'lahc' || lower === 'late-acceptance') return 'lahc';
  if (lower === 'tabu' || lower === 'tabu-search') return 'tabu';
  if (lower === 'ilp') return 'ilp';
  // NRP modes from tune-pfrs.
  if (lower.includes('+')) return 'portfolio'; // composite configs like "sa,lahc,tabu+lookahead+fw2"
  return lower;
}
